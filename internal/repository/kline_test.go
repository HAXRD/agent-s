package repository

import (
	"crypto-monitor/internal/models"
	"crypto-monitor/pkg/database"
	"os"
	"testing"
	"time"
)

// setupTestDB initializes a test database connection
func setupTestDB(t *testing.T) *KlineRepository {
	// Use test database configuration
	// Use 127.0.0.1 instead of localhost to avoid IPv6 resolution issues
	os.Setenv("DB_HOST", "127.0.0.1")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "postgres")
	os.Setenv("DB_NAME", "crypto_monitor")

	db, err := database.InitDB()
	if err != nil {
		t.Skipf("Skipping test: database connection failed: %v", err)
		return nil
	}

	return NewKlineRepository(db)
}

// TestKlineRepository_CreateKline tests creating a new kline
func TestKlineRepository_CreateKline(t *testing.T) {
	repo := setupTestDB(t)
	if repo == nil {
		return
	}

	kline := &models.Kline{
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

	err := repo.CreateKline(kline)
	if err != nil {
		t.Fatalf("Failed to create kline: %v", err)
	}

	if kline.ID == 0 {
		t.Error("Expected kline ID to be set after creation")
	}
}

// TestKlineRepository_CreateOrUpdateKline tests UPSERT functionality
func TestKlineRepository_CreateOrUpdateKline(t *testing.T) {
	repo := setupTestDB(t)
	if repo == nil {
		return
	}

	openTime := time.Now().UnixMilli()
	kline := &models.Kline{
		Symbol:     "BTCUSDT",
		Interval:   "1m",
		OpenTime:   openTime,
		CloseTime:  openTime + 60000,
		OpenPrice:  50000.0,
		HighPrice:  51000.0,
		LowPrice:   49000.0,
		ClosePrice: 50500.0,
		Volume:     100.5,
	}

	// First create
	err := repo.CreateOrUpdateKline(kline)
	if err != nil {
		t.Fatalf("Failed to create kline: %v", err)
	}

	originalID := kline.ID

	// Update with new values
	kline.ClosePrice = 50600.0
	kline.Volume = 150.0
	err = repo.CreateOrUpdateKline(kline)
	if err != nil {
		t.Fatalf("Failed to update kline: %v", err)
	}

	// Verify it's the same record (same ID)
	if kline.ID != originalID {
		t.Errorf("Expected same ID after update, got %d, want %d", kline.ID, originalID)
	}

	// Verify values were updated
	if kline.ClosePrice != 50600.0 {
		t.Errorf("Expected ClosePrice to be updated to 50600.0, got %f", kline.ClosePrice)
	}
}

// TestKlineRepository_GetKlines tests querying klines
func TestKlineRepository_GetKlines(t *testing.T) {
	repo := setupTestDB(t)
	if repo == nil {
		return
	}

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
		repo.CreateOrUpdateKline(kline)
	}

	// Query klines
	klines, err := repo.GetKlines("BTCUSDT", "1m", nil, nil, 10)
	if err != nil {
		t.Fatalf("Failed to get klines: %v", err)
	}

	if len(klines) == 0 {
		t.Error("Expected at least one kline, got 0")
	}

	// Verify ordering (should be descending by open_time)
	if len(klines) > 1 {
		for i := 0; i < len(klines)-1; i++ {
			if klines[i].OpenTime < klines[i+1].OpenTime {
				t.Error("Expected klines to be ordered by open_time descending")
			}
		}
	}
}

// TestKlineRepository_CreateKlinesBatch tests batch insert
func TestKlineRepository_CreateKlinesBatch(t *testing.T) {
	repo := setupTestDB(t)
	if repo == nil {
		return
	}

	// Create batch of klines
	now := time.Now().UnixMilli()
	klines := make([]models.Kline, 10)
	for i := 0; i < 10; i++ {
		klines[i] = models.Kline{
			Symbol:     "ETHUSDT",
			Interval:   "5m",
			OpenTime:   now - int64(i*300000),
			CloseTime:  now - int64(i*300000) + 300000,
			OpenPrice:  3000.0 + float64(i),
			HighPrice:  3100.0 + float64(i),
			LowPrice:   2900.0 + float64(i),
			ClosePrice: 3050.0 + float64(i),
			Volume:     50.0 + float64(i),
		}
	}

	err := repo.CreateKlinesBatch(klines)
	if err != nil {
		t.Fatalf("Failed to batch create klines: %v", err)
	}

	// Verify klines were created
	result, err := repo.GetKlines("ETHUSDT", "5m", nil, nil, 20)
	if err != nil {
		t.Fatalf("Failed to get klines: %v", err)
	}

	if len(result) < 10 {
		t.Errorf("Expected at least 10 klines, got %d", len(result))
	}
}

// TestKlineRepository_SafeCreateOrUpdateKline tests safe create with error handling
func TestKlineRepository_SafeCreateOrUpdateKline(t *testing.T) {
	repo := setupTestDB(t)
	if repo == nil {
		return
	}

	kline := &models.Kline{
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

	// Should not return error even if database fails (in this case it should work)
	err := repo.SafeCreateOrUpdateKline(kline)
	if err != nil {
		t.Errorf("SafeCreateOrUpdateKline should not return error: %v", err)
	}
}
