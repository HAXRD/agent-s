package service

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"crypto-monitor/internal/models"

	"github.com/gorilla/websocket"
)

// BinanceService handles Binance API interactions
type BinanceService struct {
	apiURL     string
	wsURL      string
	httpClient *http.Client
}

// BinanceKlineResponse represents a single kline from Binance REST API
type BinanceKlineResponse []interface{}

// NewBinanceService creates a new Binance service instance
// Defaults to testnet if .env file is not present
func NewBinanceService() *BinanceService {
	apiURL := os.Getenv("BINANCE_API_URL")
	wsURL := os.Getenv("BINANCE_WS_URL")

	// Check if .env file exists
	_, err := os.Stat(".env")
	useTestnet := os.IsNotExist(err)

	// If no .env file exists, default to testnet
	if useTestnet {
		if apiURL == "" {
			apiURL = "https://testnet.binance.vision"
		}
		if wsURL == "" {
			wsURL = "wss://stream.testnet.binance.vision/ws"
		}
		log.Printf("Using Binance Testnet (no .env file found)")
	} else {
		// Production defaults
		if apiURL == "" {
			apiURL = "https://api.binance.com"
		}
		if wsURL == "" {
			wsURL = "wss://stream.binance.com:9443/ws"
		}
		log.Printf("Using Binance Production")
	}

	return &BinanceService{
		apiURL: apiURL,
		wsURL:  wsURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetKlines fetches historical kline data from Binance REST API
func (s *BinanceService) GetKlines(symbol, interval string, startTime, endTime *int64, limit int) ([]models.Kline, error) {
	url := fmt.Sprintf("%s/api/v3/klines", s.apiURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Build query parameters
	q := req.URL.Query()
	q.Add("symbol", symbol)
	q.Add("interval", interval)
	if startTime != nil {
		q.Add("startTime", strconv.FormatInt(*startTime, 10))
	}
	if endTime != nil {
		q.Add("endTime", strconv.FormatInt(*endTime, 10))
	}
	if limit > 0 {
		q.Add("limit", strconv.Itoa(limit))
	} else {
		q.Add("limit", "1000")
	}
	req.URL.RawQuery = q.Encode()

	// Retry with exponential backoff
	maxRetries := 3
	retryDelay := time.Second

	for i := 0; i < maxRetries; i++ {
		resp, err := s.httpClient.Do(req)
		if err != nil {
			if i < maxRetries-1 {
				log.Printf("Binance API call failed (attempt %d/%d): %v", i+1, maxRetries, err)
				time.Sleep(retryDelay * time.Duration(1<<i)) // Exponential backoff
				continue
			}
			return nil, fmt.Errorf("failed to fetch klines after %d attempts: %w", maxRetries, err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			if i < maxRetries-1 {
				log.Printf("Binance API returned status %d (attempt %d/%d)", resp.StatusCode, i+1, maxRetries)
				time.Sleep(retryDelay * time.Duration(1<<i))
				continue
			}
			return nil, fmt.Errorf("binance API returned status %d", resp.StatusCode)
		}

		var binanceKlines []BinanceKlineResponse
		if err := json.NewDecoder(resp.Body).Decode(&binanceKlines); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		resp.Body.Close()

		// Convert Binance format to internal model
		klines := make([]models.Kline, 0, len(binanceKlines))
		for _, bk := range binanceKlines {
			kline, err := s.convertBinanceKlineToModel(bk, symbol, interval)
			if err != nil {
				log.Printf("Failed to convert kline: %v", err)
				continue
			}
			klines = append(klines, kline)
		}

		log.Printf("Successfully fetched %d klines for %s %s", len(klines), symbol, interval)
		return klines, nil
	}

	return nil, fmt.Errorf("failed to fetch klines after %d attempts", maxRetries)
}

// convertBinanceKlineToModel converts Binance kline format to internal model
func (s *BinanceService) convertBinanceKlineToModel(bk BinanceKlineResponse, symbol, interval string) (models.Kline, error) {
	if len(bk) < 12 {
		return models.Kline{}, fmt.Errorf("invalid kline data: expected at least 12 fields, got %d", len(bk))
	}

	// Helper function to safely convert interface{} to int64
	getInt64 := func(v interface{}) int64 {
		switch val := v.(type) {
		case string:
			result, _ := strconv.ParseInt(val, 10, 64)
			return result
		case float64:
			return int64(val)
		case int64:
			return val
		case int:
			return int64(val)
		default:
			return 0
		}
	}

	// Helper function to safely convert interface{} to float64
	getFloat64 := func(v interface{}) float64 {
		switch val := v.(type) {
		case string:
			result, _ := strconv.ParseFloat(val, 64)
			return result
		case float64:
			return val
		case float32:
			return float64(val)
		case int:
			return float64(val)
		case int64:
			return float64(val)
		default:
			return 0
		}
	}

	openTime := getInt64(bk[0])
	closeTime := getInt64(bk[6])
	openPrice := getFloat64(bk[1])
	highPrice := getFloat64(bk[2])
	lowPrice := getFloat64(bk[3])
	closePrice := getFloat64(bk[4])
	volume := getFloat64(bk[5])

	return models.Kline{
		Symbol:     symbol,
		Interval:   interval,
		OpenTime:   openTime,
		CloseTime:  closeTime,
		OpenPrice:  openPrice,
		HighPrice:  highPrice,
		LowPrice:   lowPrice,
		ClosePrice: closePrice,
		Volume:     volume,
	}, nil
}

// SubscribeKlineStream subscribes to Binance WebSocket kline stream
// Uses raw stream format: /ws/<streamName> which returns direct data payload
func (s *BinanceService) SubscribeKlineStream(symbol, interval string, callback func(models.Kline)) error {
	// Build stream name: <symbol>@kline_<interval>
	// Binance requires lowercase symbols
	symbolLower := strings.ToLower(symbol)
	streamName := fmt.Sprintf("%s@kline_%s", symbolLower, interval)

	// Use raw stream format: /ws/<streamName>
	// This returns direct data payload, not wrapped in {"stream":"...","data":{...}}
	wsURL := fmt.Sprintf("%s/%s", s.wsURL, streamName)

	log.Printf("Connecting to Binance WebSocket: %s", wsURL)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to Binance WebSocket: %w", err)
	}
	defer conn.Close()

	log.Printf("Connected to Binance WebSocket stream: %s", streamName)

	// Read messages
	// Raw stream format: message is directly the kline data payload
	for {
		var msg struct {
			EventType string `json:"e"`
			EventTime int64  `json:"E"`
			Symbol    string `json:"s"`
			Kline     struct {
				StartTime  int64  `json:"t"`
				EndTime    int64  `json:"T"`
				Symbol     string `json:"s"`
				Interval   string `json:"i"`
				OpenPrice  string `json:"o"`
				ClosePrice string `json:"c"`
				HighPrice  string `json:"h"`
				LowPrice   string `json:"l"`
				Volume     string `json:"v"`
				IsClosed   bool   `json:"x"`
			} `json:"k"`
		}

		if err := conn.ReadJSON(&msg); err != nil {
			log.Printf("Error reading WebSocket message: %v", err)
			return fmt.Errorf("failed to read message: %w", err)
		}

		// Process kline update (only process closed klines)
		if msg.EventType == "kline" && msg.Kline.IsClosed {
			kline := models.Kline{
				Symbol:     msg.Kline.Symbol,
				Interval:   msg.Kline.Interval,
				OpenTime:   msg.Kline.StartTime,
				CloseTime:  msg.Kline.EndTime,
				OpenPrice:  parseFloat(msg.Kline.OpenPrice),
				HighPrice:  parseFloat(msg.Kline.HighPrice),
				LowPrice:   parseFloat(msg.Kline.LowPrice),
				ClosePrice: parseFloat(msg.Kline.ClosePrice),
				Volume:     parseFloat(msg.Kline.Volume),
			}

			log.Printf("Received kline update: %s %s at %d", kline.Symbol, kline.Interval, kline.OpenTime)
			callback(kline)
		}
	}
}

// parseFloat safely parses a string to float64
func parseFloat(s string) float64 {
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Printf("Failed to parse float: %s, error: %v", s, err)
		return 0
	}
	return val
}
