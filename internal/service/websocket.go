package service

import (
	"crypto-monitor/internal/models"
	"crypto-monitor/internal/repository"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Throttle interval: maximum one update per second
	throttleInterval = 1 * time.Second
)

// Client represents a WebSocket client connection
type Client struct {
	conn     *websocket.Conn
	send     chan []byte
	subs     map[string]bool // Map of "symbol:interval" -> subscribed
	mu       sync.RWMutex
	lastSent map[string]time.Time // Track last sent time per subscription for throttling
}

// WebSocketService manages WebSocket connections and message broadcasting
type WebSocketService struct {
	clients       map[*Client]bool
	broadcast     chan []byte
	register      chan *Client
	unregister    chan *Client
	binanceSvc    *BinanceService
	klineRepo     *repository.KlineRepository
	mu            sync.RWMutex
	subscriptions map[string]map[*Client]bool // Map of "symbol:interval" -> clients
	subsMu        sync.RWMutex
}

// NewWebSocketService creates a new WebSocket service instance
func NewWebSocketService(binanceSvc *BinanceService, klineRepo *repository.KlineRepository) *WebSocketService {
	return &WebSocketService{
		clients:       make(map[*Client]bool),
		broadcast:     make(chan []byte, 256),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		binanceSvc:    binanceSvc,
		klineRepo:     klineRepo,
		subscriptions: make(map[string]map[*Client]bool),
	}
}

// ClientMessage represents a message from client
type ClientMessage struct {
	Action   string `json:"action"`   // "subscribe" or "unsubscribe"
	Symbol   string `json:"symbol"`   // e.g., "BTCUSDT"
	Interval string `json:"interval"` // e.g., "1m", "5m", "1h"
}

// ServerMessage represents a message to client
type ServerMessage struct {
	Type     string      `json:"type"` // "subscribed", "unsubscribed", "kline_update", "error"
	Symbol   string      `json:"symbol,omitempty"`
	Interval string      `json:"interval,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	Message  string      `json:"message,omitempty"`
}

// HandleConnection handles a new WebSocket connection
func (ws *WebSocketService) HandleConnection(conn *websocket.Conn) {
	client := &Client{
		conn:     conn,
		send:     make(chan []byte, 256),
		subs:     make(map[string]bool),
		lastSent: make(map[string]time.Time),
	}

	ws.register <- client

	// Start goroutines for reading and writing
	go client.writePump(ws)
	client.readPump(ws)
}

// Run starts the WebSocket service
func (ws *WebSocketService) Run() {
	for {
		select {
		case client := <-ws.register:
			ws.mu.Lock()
			ws.clients[client] = true
			ws.mu.Unlock()
			log.Printf("WebSocket client connected. Total clients: %d", len(ws.clients))

		case client := <-ws.unregister:
			ws.mu.Lock()
			if _, ok := ws.clients[client]; ok {
				delete(ws.clients, client)
				close(client.send)
				// Remove client from all subscriptions
				ws.removeClientFromAllSubscriptions(client)
			}
			ws.mu.Unlock()
			log.Printf("WebSocket client disconnected. Total clients: %d", len(ws.clients))

		case message := <-ws.broadcast:
			ws.mu.RLock()
			for client := range ws.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(ws.clients, client)
					ws.removeClientFromAllSubscriptions(client)
				}
			}
			ws.mu.RUnlock()
		}
	}
}

// SubscribeToBinanceStream subscribes to Binance WebSocket stream for a symbol and interval
func (ws *WebSocketService) SubscribeToBinanceStream(symbol, interval string) error {
	// Start a goroutine to handle Binance stream
	go func() {
		err := ws.binanceSvc.SubscribeKlineStream(symbol, interval, func(kline models.Kline) {
			// Store to database
			if err := ws.klineRepo.SafeCreateOrUpdateKline(&kline); err != nil {
				log.Printf("Error storing kline to database: %v", err)
			}

			// Broadcast to subscribed clients with throttling
			ws.broadcastKlineUpdate(kline)
		})

		if err != nil {
			log.Printf("Error subscribing to Binance stream for %s %s: %v", symbol, interval, err)
		}
	}()

	return nil
}

// broadcastKlineUpdate broadcasts kline update to all subscribed clients with throttling
func (ws *WebSocketService) broadcastKlineUpdate(kline models.Kline) {
	key := fmt.Sprintf("%s:%s", kline.Symbol, kline.Interval)

	ws.subsMu.RLock()
	clients, exists := ws.subscriptions[key]
	ws.subsMu.RUnlock()

	if !exists || len(clients) == 0 {
		return
	}

	// Create message
	msg := ServerMessage{
		Type:     "kline_update",
		Symbol:   kline.Symbol,
		Interval: kline.Interval,
		Data: map[string]interface{}{
			"open_time":  kline.OpenTime,
			"close_time": kline.CloseTime,
			"open":       fmt.Sprintf("%.8f", kline.OpenPrice),
			"high":       fmt.Sprintf("%.8f", kline.HighPrice),
			"low":        fmt.Sprintf("%.8f", kline.LowPrice),
			"close":      fmt.Sprintf("%.8f", kline.ClosePrice),
			"volume":     fmt.Sprintf("%.8f", kline.Volume),
		},
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling kline update message: %v", err)
		return
	}

	// Send to each subscribed client with throttling check
	ws.subsMu.RLock()
	for client := range clients {
		client.mu.Lock()
		lastSent, exists := client.lastSent[key]
		shouldSend := !exists || time.Since(lastSent) >= throttleInterval
		client.mu.Unlock()

		if shouldSend {
			select {
			case client.send <- msgBytes:
				client.mu.Lock()
				client.lastSent[key] = time.Now()
				client.mu.Unlock()
			default:
				// Channel full, skip this client
			}
		}
	}
	ws.subsMu.RUnlock()
}

// readPump reads messages from the WebSocket connection
func (c *Client) readPump(ws *WebSocketService) {
	defer func() {
		ws.unregister <- c
		c.conn.Close()
	}()

	// Set read deadline and pong handler
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse client message
		var clientMsg ClientMessage
		if err := json.Unmarshal(message, &clientMsg); err != nil {
			log.Printf("Error parsing client message: %v", err)
			sendError(c, "Invalid message format")
			continue
		}

		// Handle subscribe/unsubscribe
		switch clientMsg.Action {
		case "subscribe":
			ws.handleSubscribe(c, clientMsg.Symbol, clientMsg.Interval)
		case "unsubscribe":
			ws.handleUnsubscribe(c, clientMsg.Symbol, clientMsg.Interval)
		default:
			sendError(c, fmt.Sprintf("Unknown action: %s", clientMsg.Action))
		}
	}
}

// writePump writes messages to the WebSocket connection
func (c *Client) writePump(ws *WebSocketService) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleSubscribe handles client subscription
func (ws *WebSocketService) handleSubscribe(client *Client, symbol, interval string) {
	if symbol == "" || interval == "" {
		sendError(client, "Symbol and interval are required")
		return
	}

	key := fmt.Sprintf("%s:%s", symbol, interval)

	// Add to client's subscriptions
	client.mu.Lock()
	client.subs[key] = true
	client.mu.Unlock()

	// Add client to subscription map
	ws.subsMu.Lock()
	if ws.subscriptions[key] == nil {
		ws.subscriptions[key] = make(map[*Client]bool)
		// First client for this subscription, start Binance stream
		go ws.SubscribeToBinanceStream(symbol, interval)
	}
	ws.subscriptions[key][client] = true
	ws.subsMu.Unlock()

	// Send confirmation
	msg := ServerMessage{
		Type:     "subscribed",
		Symbol:   symbol,
		Interval: interval,
	}
	sendMessage(client, msg)

	log.Printf("Client subscribed to %s %s", symbol, interval)
}

// handleUnsubscribe handles client unsubscription
func (ws *WebSocketService) handleUnsubscribe(client *Client, symbol, interval string) {
	if symbol == "" || interval == "" {
		sendError(client, "Symbol and interval are required")
		return
	}

	key := fmt.Sprintf("%s:%s", symbol, interval)

	// Remove from client's subscriptions
	client.mu.Lock()
	delete(client.subs, key)
	delete(client.lastSent, key)
	client.mu.Unlock()

	// Remove client from subscription map
	ws.subsMu.Lock()
	if clients, exists := ws.subscriptions[key]; exists {
		delete(clients, client)
		if len(clients) == 0 {
			// No more clients, could stop Binance stream here if needed
			delete(ws.subscriptions, key)
		}
	}
	ws.subsMu.Unlock()

	// Send confirmation
	msg := ServerMessage{
		Type:     "unsubscribed",
		Symbol:   symbol,
		Interval: interval,
	}
	sendMessage(client, msg)

	log.Printf("Client unsubscribed from %s %s", symbol, interval)
}

// removeClientFromAllSubscriptions removes client from all subscriptions
func (ws *WebSocketService) removeClientFromAllSubscriptions(client *Client) {
	ws.subsMu.Lock()
	for key, clients := range ws.subscriptions {
		delete(clients, client)
		if len(clients) == 0 {
			delete(ws.subscriptions, key)
		}
	}
	ws.subsMu.Unlock()
}

// sendMessage sends a message to a client
func sendMessage(client *Client, msg ServerMessage) {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	select {
	case client.send <- msgBytes:
	default:
		log.Printf("Client send channel full, dropping message")
	}
}

// sendError sends an error message to a client
func sendError(client *Client, errorMsg string) {
	msg := ServerMessage{
		Type:    "error",
		Message: errorMsg,
	}
	sendMessage(client, msg)
}
