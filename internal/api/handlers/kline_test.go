package handlers

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

	"github.com/gin-gonic/gin"
)

// setupTestHandler creates a test handler with test database
func setupTestHandler(t *testing.T) (*KlineHandler, *gin.Engine) {
	// Setup database
	os.Setenv("DB_HOST", "127.0.0.1")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "postgres")
	os.Setenv("DB_NAME", "crypto_monitor")

	db, err := database.InitDB()
	if err != nil {
		t.Skipf("Skipping test: database connection failed: %v", err)
		return nil, nil
	}

	klineRepo := repository.NewKlineRepository(db)
	handler := NewKlineHandler(klineRepo)

	// Create test data
	now := time.Now().UnixMilli()
	for i := 0; i < 5; i++ {
		kline := &models.Kline{
			Symbol:     "BTCUSDT",
			Interval:   "1m",
			OpenTime:   now - int64(i*60000),
			CloseTime:  now - int64(i*60000) + 60000,
			OpenPrice:  50000.0 + float64(i),
			HighPrice:  51000.0 + float64(i),
			LowPrice:   49000.0 + float64(i),
			ClosePrice: 50500.0 + float64(i),
			Volume:     100.5 + float64(i),
		}
		klineRepo.CreateOrUpdateKline(kline)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/klines", handler.GetKlines)
	router.GET("/api/v1/symbols", handler.GetSymbols)

	return handler, router
}

// TestKlineHandler_GetKlines tests GET /api/v1/klines endpoint
func TestKlineHandler_GetKlines(t *testing.T) {
	_, router := setupTestHandler(t)
	if router == nil {
		return
	}

	// Test successful request
	req, _ := http.NewRequest("GET", "/api/v1/klines?symbol=BTCUSDT&interval=1m&limit=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	var response APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Code != 200 {
		t.Errorf("Expected response code 200, got %d", response.Code)
	}

	if response.Message != "success" {
		t.Errorf("Expected message 'success', got '%s'", response.Message)
	}

	// Verify data is an array
	data, ok := response.Data.([]interface{})
	if !ok {
		t.Error("Expected data to be an array")
	}

	if len(data) == 0 {
		t.Error("Expected at least one kline in response")
	}
}

// TestKlineHandler_GetKlines_MissingParams tests error handling for missing parameters
func TestKlineHandler_GetKlines_MissingParams(t *testing.T) {
	_, router := setupTestHandler(t)
	if router == nil {
		return
	}

	// Test missing symbol
	req, _ := http.NewRequest("GET", "/api/v1/klines?interval=1m", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", w.Code)
	}

	var response APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Code != 400 {
		t.Errorf("Expected response code 400, got %d", response.Code)
	}

	// Test missing interval
	req, _ = http.NewRequest("GET", "/api/v1/klines?symbol=BTCUSDT", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", w.Code)
	}
}

// TestKlineHandler_GetKlines_InvalidParams tests error handling for invalid parameters
func TestKlineHandler_GetKlines_InvalidParams(t *testing.T) {
	_, router := setupTestHandler(t)
	if router == nil {
		return
	}

	// Test invalid start_time
	req, _ := http.NewRequest("GET", "/api/v1/klines?symbol=BTCUSDT&interval=1m&start_time=invalid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", w.Code)
	}

	// Test invalid limit
	req, _ = http.NewRequest("GET", "/api/v1/klines?symbol=BTCUSDT&interval=1m&limit=-1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code 400, got %d", w.Code)
	}
}

// TestKlineHandler_GetSymbols tests GET /api/v1/symbols endpoint
func TestKlineHandler_GetSymbols(t *testing.T) {
	_, router := setupTestHandler(t)
	if router == nil {
		return
	}

	req, _ := http.NewRequest("GET", "/api/v1/symbols", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	var response APIResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Code != 200 {
		t.Errorf("Expected response code 200, got %d", response.Code)
	}

	if response.Message != "success" {
		t.Errorf("Expected message 'success', got '%s'", response.Message)
	}

	// Verify data is an array
	data, ok := response.Data.([]interface{})
	if !ok {
		t.Error("Expected data to be an array")
	}

	if len(data) == 0 {
		t.Error("Expected at least one symbol in response")
	}

	// Verify symbol structure
	if len(data) > 0 {
		symbol, ok := data[0].(map[string]interface{})
		if !ok {
			t.Error("Expected symbol to be an object")
		}

		if symbol["symbol"] == nil {
			t.Error("Expected symbol to have 'symbol' field")
		}
	}
}
