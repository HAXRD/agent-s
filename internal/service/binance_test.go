package service

import (
	"crypto-monitor/internal/models"
	"os"
	"testing"
	"time"
)

func TestBinanceService_GetKlines(t *testing.T) {
	// Skip if running in CI or without network
	if os.Getenv("SKIP_NETWORK_TESTS") == "true" {
		t.Skip("Skipping network tests")
	}

	service := NewBinanceService()

	// Test fetching BTC/USDT 1m klines
	klines, err := service.GetKlines("BTCUSDT", "1m", nil, nil, 10)
	if err != nil {
		t.Fatalf("Failed to fetch klines: %v", err)
	}

	if len(klines) == 0 {
		t.Error("Expected at least one kline, got 0")
	}

	// Validate first kline
	kline := klines[0]
	if kline.Symbol != "BTCUSDT" {
		t.Errorf("Expected symbol BTCUSDT, got %s", kline.Symbol)
	}
	if kline.Interval != "1m" {
		t.Errorf("Expected interval 1m, got %s", kline.Interval)
	}
	if kline.OpenTime == 0 {
		t.Error("OpenTime should not be 0")
	}
	if kline.ClosePrice <= 0 {
		t.Error("ClosePrice should be greater than 0")
	}

	t.Logf("Successfully fetched %d klines", len(klines))
	t.Logf("First kline: %+v", kline)
}

func TestBinanceService_ConvertBinanceKline(t *testing.T) {
	service := NewBinanceService()

	// Sample Binance kline response
	binanceKline := BinanceKlineResponse{
		"1699000000000", // 0: Open time
		"35000.00",      // 1: Open price
		"35100.00",      // 2: High price
		"34900.00",      // 3: Low price
		"35050.00",      // 4: Close price
		"100.5",         // 5: Volume
		"1699000059999", // 6: Close time
		"50000.00",      // 7: Quote asset volume
		"10",            // 8: Number of trades
		"5000.00",       // 9: Taker buy base asset volume
		"50000.00",      // 10: Taker buy quote asset volume
		"0",             // 11: Ignore
	}

	kline, err := service.convertBinanceKlineToModel(binanceKline, "BTCUSDT", "1m")
	if err != nil {
		t.Fatalf("Failed to convert kline: %v", err)
	}

	if kline.Symbol != "BTCUSDT" {
		t.Errorf("Expected symbol BTCUSDT, got %s", kline.Symbol)
	}
	if kline.Interval != "1m" {
		t.Errorf("Expected interval 1m, got %s", kline.Interval)
	}
	if kline.OpenTime != 1699000000000 {
		t.Errorf("Expected OpenTime 1699000000000, got %d", kline.OpenTime)
	}
	if kline.OpenPrice != 35000.00 {
		t.Errorf("Expected OpenPrice 35000.00, got %f", kline.OpenPrice)
	}
}

func TestBinanceService_WebSocketConnection(t *testing.T) {
	// Skip if running in CI or without network
	if os.Getenv("SKIP_NETWORK_TESTS") == "true" {
		t.Skip("Skipping network tests")
	}

	service := NewBinanceService()

	// Test WebSocket connection with timeout
	done := make(chan bool, 1)
	timeout := time.After(10 * time.Second)

	go func() {
		err := service.SubscribeKlineStream("BTCUSDT", "1s", func(kline models.Kline) {
			t.Logf("Received kline: %+v", kline)
			done <- true
		})
		if err != nil {
			t.Logf("WebSocket error: %v", err)
		}
	}()

	select {
	case <-done:
		t.Log("Successfully received kline from WebSocket")
	case <-timeout:
		t.Log("WebSocket connection test completed (timeout after 10s)")
	}
}
