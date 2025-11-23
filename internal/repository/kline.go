package repository

import (
	"crypto-monitor/internal/models"
	"errors"
	"fmt"
	"log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	errDBConnectionUnavailable = "database connection is not available"
)

// KlineRepository handles database operations for Kline models
type KlineRepository struct {
	db *gorm.DB
}

// NewKlineRepository creates a new KlineRepository instance
func NewKlineRepository(db *gorm.DB) *KlineRepository {
	return &KlineRepository{db: db}
}

// CreateKline creates a new kline record in the database
// Returns error if the kline already exists (based on symbol, interval, open_time)
func (r *KlineRepository) CreateKline(kline *models.Kline) error {
	if r.db == nil {
		return fmt.Errorf(errDBConnectionUnavailable)
	}

	if err := r.db.Create(kline).Error; err != nil {
		// Check if it's a unique constraint violation
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("kline already exists: %w", err)
		}
		return fmt.Errorf("failed to create kline: %w", err)
	}

	return nil
}

// CreateOrUpdateKline creates a new kline or updates existing one using UPSERT logic
// Uses ON CONFLICT DO UPDATE for PostgreSQL
func (r *KlineRepository) CreateOrUpdateKline(kline *models.Kline) error {
	if r.db == nil {
		return fmt.Errorf(errDBConnectionUnavailable)
	}

	// Use GORM's Clauses with OnConflict for UPSERT
	// This will insert or update based on unique constraint (symbol, interval, open_time)
	result := r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "symbol"}, {Name: "interval"}, {Name: "open_time"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"close_time":  kline.CloseTime,
			"open_price":  kline.OpenPrice,
			"high_price":  kline.HighPrice,
			"low_price":   kline.LowPrice,
			"close_price": kline.ClosePrice,
			"volume":      kline.Volume,
			"updated_at":  clause.Expr{SQL: "CURRENT_TIMESTAMP"},
		}),
	}).Create(kline)

	if result.Error != nil {
		return fmt.Errorf("failed to create or update kline: %w", result.Error)
	}

	return nil
}

// GetKlines queries historical kline data with optional filters
// Parameters:
//   - symbol: trading pair symbol (e.g., "BTCUSDT")
//   - interval: time interval (e.g., "1m", "5m", "1h")
//   - startTime: optional start time in milliseconds (nil to ignore)
//   - endTime: optional end time in milliseconds (nil to ignore)
//   - limit: maximum number of records to return (0 for no limit)
func (r *KlineRepository) GetKlines(symbol, interval string, startTime, endTime *int64, limit int) ([]models.Kline, error) {
	if r.db == nil {
		return nil, fmt.Errorf(errDBConnectionUnavailable)
	}

	var klines []models.Kline
	query := r.db.Model(&models.Kline{})

	// Apply filters
	if symbol != "" {
		query = query.Where("symbol = ?", symbol)
	}
	if interval != "" {
		query = query.Where("interval = ?", interval)
	}
	if startTime != nil {
		query = query.Where("open_time >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("open_time <= ?", *endTime)
	}

	// Order by open_time descending (most recent first)
	query = query.Order("open_time DESC")

	// Apply limit
	if limit > 0 {
		query = query.Limit(limit)
	}

	// Execute query
	if err := query.Find(&klines).Error; err != nil {
		return nil, fmt.Errorf("failed to query klines: %w", err)
	}

	return klines, nil
}

// CreateKlinesBatch performs batch insert with UPSERT logic for multiple klines
// This is optimized for inserting large numbers of klines efficiently
func (r *KlineRepository) CreateKlinesBatch(klines []models.Kline) error {
	if r.db == nil {
		return fmt.Errorf(errDBConnectionUnavailable)
	}

	if len(klines) == 0 {
		return nil
	}

	// Use batch insert with ON CONFLICT for PostgreSQL
	// Process in chunks to avoid memory issues with very large batches
	const batchSize = 1000
	for i := 0; i < len(klines); i += batchSize {
		end := i + batchSize
		if end > len(klines) {
			end = len(klines)
		}

		batch := klines[i:end]
		result := r.db.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "symbol"}, {Name: "interval"}, {Name: "open_time"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"close_time":  clause.Expr{SQL: "excluded.close_time"},
				"open_price":  clause.Expr{SQL: "excluded.open_price"},
				"high_price":  clause.Expr{SQL: "excluded.high_price"},
				"low_price":   clause.Expr{SQL: "excluded.low_price"},
				"close_price": clause.Expr{SQL: "excluded.close_price"},
				"volume":      clause.Expr{SQL: "excluded.volume"},
				"updated_at":  clause.Expr{SQL: "CURRENT_TIMESTAMP"},
			}),
		}).Create(&batch)

		if result.Error != nil {
			return fmt.Errorf("failed to batch insert klines (batch %d-%d): %w", i, end-1, result.Error)
		}
	}

	return nil
}

// IsConnected checks if the database connection is available
func (r *KlineRepository) IsConnected() bool {
	if r.db == nil {
		return false
	}

	sqlDB, err := r.db.DB()
	if err != nil {
		return false
	}

	if err := sqlDB.Ping(); err != nil {
		return false
	}

	return true
}

// SafeCreateOrUpdateKline safely creates or updates a kline with error handling
// If database connection fails, it logs the error but doesn't return it
// This allows the application to continue running even if database is unavailable
func (r *KlineRepository) SafeCreateOrUpdateKline(kline *models.Kline) error {
	if !r.IsConnected() {
		log.Printf("Warning: Database connection not available, skipping kline storage: %s %s %d",
			kline.Symbol, kline.Interval, kline.OpenTime)
		return nil // Return nil to allow application to continue
	}

	if err := r.CreateOrUpdateKline(kline); err != nil {
		log.Printf("Error storing kline (continuing execution): %v", err)
		return nil // Return nil to allow application to continue
	}

	return nil
}

// SafeCreateKlinesBatch safely performs batch insert with error handling
// If database connection fails, it logs the error but doesn't return it
func (r *KlineRepository) SafeCreateKlinesBatch(klines []models.Kline) error {
	if !r.IsConnected() {
		log.Printf("Warning: Database connection not available, skipping batch kline storage (%d klines)",
			len(klines))
		return nil // Return nil to allow application to continue
	}

	if err := r.CreateKlinesBatch(klines); err != nil {
		log.Printf("Error batch storing klines (continuing execution): %v", err)
		return nil // Return nil to allow application to continue
	}

	return nil
}
