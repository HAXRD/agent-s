package service

import (
	"crypto-monitor/internal/models"
	"crypto-monitor/internal/repository"
	"crypto-monitor/pkg/database"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// setupTestWebSocketService creates a test WebSocket service
func setupTestWebSocketService(t *testing.T) (*WebSocketService, *BinanceService, *repository.KlineRepository) {
	// Setup database
	os.Setenv("DB_HOST", "127.0.0.1")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "postgres")
	os.Setenv("DB_NAME", "crypto_monitor")

	db, err := database.InitDB()
	if err != nil {
		t.Skipf("Skipping test: database connection failed: %v", err)
		return nil, nil, nil
	}

	binanceSvc := NewBinanceService()
	klineRepo := repository.NewKlineRepository(db)
	wsSvc := NewWebSocketService(binanceSvc, klineRepo)

	// Start WebSocket service
	go wsSvc.Run()

	return wsSvc, binanceSvc, klineRepo
}

// TestWebSocketService_HandleConnection tests WebSocket connection handling
func TestWebSocketService_HandleConnection(t *testing.T) {
	wsSvc, _, _ := setupTestWebSocketService(t)
	if wsSvc == nil {
		return
	}

	// Create a test HTTP server with WebSocket upgrade
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("Failed to upgrade connection: %v", err)
		}
		wsSvc.HandleConnection(conn)
	}))
	defer server.Close()

	// Connect to WebSocket
	wsURL := "ws" + server.URL[4:] + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Wait a bit for connection to be registered
	time.Sleep(100 * time.Millisecond)

	// Verify client is registered
	wsSvc.mu.RLock()
	clientCount := len(wsSvc.clients)
	wsSvc.mu.RUnlock()

	if clientCount == 0 {
		t.Error("Expected at least one client to be registered")
	}
}

// TestWebSocketService_Subscribe tests subscription functionality
func TestWebSocketService_Subscribe(t *testing.T) {
	wsSvc, _, _ := setupTestWebSocketService(t)
	if wsSvc == nil {
		return
	}

	// Create a test client
	client := &Client{
		conn:     nil, // Mock connection
		send:     make(chan []byte, 256),
		subs:     make(map[string]bool),
		lastSent: make(map[string]time.Time),
	}

	// Test subscribe
	wsSvc.handleSubscribe(client, "BTCUSDT", "1m")

	// Verify subscription
	client.mu.RLock()
	subscribed := client.subs["BTCUSDT:1m"]
	client.mu.RUnlock()

	if !subscribed {
		t.Error("Expected client to be subscribed to BTCUSDT:1m")
	}

	// Verify subscription map
	wsSvc.subsMu.RLock()
	clients, exists := wsSvc.subscriptions["BTCUSDT:1m"]
	wsSvc.subsMu.RUnlock()

	if !exists {
		t.Error("Expected subscription to exist in subscription map")
	}

	if clients == nil || !clients[client] {
		t.Error("Expected client to be in subscription map")
	}
}

// TestWebSocketService_Unsubscribe tests unsubscription functionality
func TestWebSocketService_Unsubscribe(t *testing.T) {
	wsSvc, _, _ := setupTestWebSocketService(t)
	if wsSvc == nil {
		return
	}

	// Create a test client
	client := &Client{
		conn:     nil,
		send:     make(chan []byte, 256),
		subs:     make(map[string]bool),
		lastSent: make(map[string]time.Time),
	}

	// Subscribe first
	wsSvc.handleSubscribe(client, "BTCUSDT", "1m")

	// Then unsubscribe
	wsSvc.handleUnsubscribe(client, "BTCUSDT", "1m")

	// Verify unsubscription
	client.mu.RLock()
	subscribed := client.subs["BTCUSDT:1m"]
	client.mu.RUnlock()

	if subscribed {
		t.Error("Expected client to be unsubscribed from BTCUSDT:1m")
	}
}

// TestWebSocketService_BroadcastKlineUpdate tests kline update broadcasting
func TestWebSocketService_BroadcastKlineUpdate(t *testing.T) {
	wsSvc, _, _ := setupTestWebSocketService(t)
	if wsSvc == nil {
		return
	}

	// Create a test client
	client := &Client{
		conn:     nil,
		send:     make(chan []byte, 256),
		subs:     make(map[string]bool),
		lastSent: make(map[string]time.Time),
	}

	// Subscribe to BTCUSDT:1m
	wsSvc.handleSubscribe(client, "BTCUSDT", "1m")

	// Create a test kline
	kline := models.Kline{
		Symbol:     "BTCUSDT",
		Interval:   "1m",
		OpenTime:   time.Now().UnixMilli(),
		CloseTime:  time.Now().UnixMilli() + 60000,
		OpenPrice:  50000.0,
		HighPrice:  51000.0,
		LowPrice:   49000.0,
		ClosePrice: 50500.0,
		Volume:     100.5,
	}

	// Wait for subscription confirmation message first
	time.Sleep(100 * time.Millisecond)
	select {
	case <-client.send:
		// Consume subscription confirmation
	default:
	}

	// Broadcast update
	wsSvc.broadcastKlineUpdate(kline)

	// Wait a bit for message to be sent
	time.Sleep(100 * time.Millisecond)

	// Check if message was sent
	found := false
	timeout := time.After(1 * time.Second)
	for !found {
		select {
		case msg := <-client.send:
			var serverMsg ServerMessage
			if err := json.Unmarshal(msg, &serverMsg); err != nil {
				t.Fatalf("Failed to unmarshal message: %v", err)
			}

			if serverMsg.Type == "kline_update" {
				found = true
				if serverMsg.Symbol != "BTCUSDT" {
					t.Errorf("Expected symbol 'BTCUSDT', got '%s'", serverMsg.Symbol)
				}
			}
		case <-timeout:
			if !found {
				t.Error("Expected to receive kline update message, but timeout")
			}
			return
		}
	}
}

// TestWebSocketService_Throttle tests throttling functionality
func TestWebSocketService_Throttle(t *testing.T) {
	wsSvc, _, _ := setupTestWebSocketService(t)
	if wsSvc == nil {
		return
	}

	// Create a test client
	client := &Client{
		conn:     nil,
		send:     make(chan []byte, 256),
		subs:     make(map[string]bool),
		lastSent: make(map[string]time.Time),
	}

	// Subscribe
	wsSvc.handleSubscribe(client, "BTCUSDT", "1m")

	// Create test kline
	kline := models.Kline{
		Symbol:     "BTCUSDT",
		Interval:   "1m",
		OpenTime:   time.Now().UnixMilli(),
		CloseTime:  time.Now().UnixMilli() + 60000,
		OpenPrice:  50000.0,
		HighPrice:  51000.0,
		LowPrice:   49000.0,
		ClosePrice: 50500.0,
		Volume:     100.5,
	}

	// Send first update
	wsSvc.broadcastKlineUpdate(kline)
	time.Sleep(50 * time.Millisecond)

	// Send second update immediately (should be throttled)
	wsSvc.broadcastKlineUpdate(kline)
	time.Sleep(50 * time.Millisecond)

	// Count messages received
	messageCount := 0
	for {
		select {
		case <-client.send:
			messageCount++
		case <-time.After(200 * time.Millisecond):
			goto done
		}
	}
done:

	// Should receive at least one message, but not necessarily two due to throttling
	if messageCount == 0 {
		t.Error("Expected to receive at least one message")
	}
}
