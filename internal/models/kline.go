package models

import (
	"time"
)

// Kline represents a candlestick/K-line data point
type Kline struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Symbol     string    `gorm:"type:varchar(20);not null;index:idx_symbol_interval" json:"symbol"`
	Interval   string    `gorm:"type:varchar(10);not null;index:idx_symbol_interval" json:"interval"`
	OpenTime   int64     `gorm:"not null;index:idx_symbol_interval_time" json:"open_time"`
	CloseTime  int64     `gorm:"not null" json:"close_time"`
	OpenPrice  float64   `gorm:"type:decimal(20,8);not null" json:"open"`
	HighPrice  float64   `gorm:"type:decimal(20,8);not null" json:"high"`
	LowPrice   float64   `gorm:"type:decimal(20,8);not null" json:"low"`
	ClosePrice float64   `gorm:"type:decimal(20,8);not null" json:"close"`
	Volume     float64   `gorm:"type:decimal(20,8);not null" json:"volume"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (Kline) TableName() string {
	return "klines"
}
