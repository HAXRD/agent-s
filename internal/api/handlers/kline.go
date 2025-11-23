package handlers

import (
	"crypto-monitor/internal/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// APIResponse represents a unified API response format
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// KlineHandler handles K-line related API requests
type KlineHandler struct {
	klineRepo *repository.KlineRepository
}

// NewKlineHandler creates a new KlineHandler instance
func NewKlineHandler(klineRepo *repository.KlineRepository) *KlineHandler {
	return &KlineHandler{
		klineRepo: klineRepo,
	}
}

// GetKlines handles GET /api/v1/klines request
// Query parameters:
//   - symbol (required): trading pair symbol, e.g., "BTCUSDT"
//   - interval (required): time interval, e.g., "1m", "5m", "1h"
//   - start_time (optional): start timestamp in milliseconds
//   - end_time (optional): end timestamp in milliseconds
//   - limit (optional): maximum number of records, default 1000
func (h *KlineHandler) GetKlines(c *gin.Context) {
	// Validate required parameters
	symbol := c.Query("symbol")
	if symbol == "" {
		respondError(c, http.StatusBadRequest, "symbol parameter is required")
		return
	}

	interval := c.Query("interval")
	if interval == "" {
		respondError(c, http.StatusBadRequest, "interval parameter is required")
		return
	}

	// Parse optional parameters
	var startTime *int64
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		val, err := strconv.ParseInt(startTimeStr, 10, 64)
		if err != nil {
			respondError(c, http.StatusBadRequest, "invalid start_time parameter")
			return
		}
		startTime = &val
	}

	var endTime *int64
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		val, err := strconv.ParseInt(endTimeStr, 10, 64)
		if err != nil {
			respondError(c, http.StatusBadRequest, "invalid end_time parameter")
			return
		}
		endTime = &val
	}

	limit := 1000 // default limit
	if limitStr := c.Query("limit"); limitStr != "" {
		val, err := strconv.Atoi(limitStr)
		if err != nil || val <= 0 {
			respondError(c, http.StatusBadRequest, "invalid limit parameter")
			return
		}
		limit = val
	}

	// Query klines from repository
	klines, err := h.klineRepo.GetKlines(symbol, interval, startTime, endTime, limit)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "failed to query klines: "+err.Error())
		return
	}

	// Convert to response format
	responseData := make([]map[string]interface{}, 0, len(klines))
	for _, kline := range klines {
		responseData = append(responseData, map[string]interface{}{
			"open_time":  kline.OpenTime,
			"close_time": kline.CloseTime,
			"open":       formatPrice(kline.OpenPrice),
			"high":       formatPrice(kline.HighPrice),
			"low":        formatPrice(kline.LowPrice),
			"close":      formatPrice(kline.ClosePrice),
			"volume":     formatPrice(kline.Volume),
		})
	}

	respondSuccess(c, responseData)
}

// GetSymbols handles GET /api/v1/symbols request
// Returns list of supported trading pairs
func (h *KlineHandler) GetSymbols(c *gin.Context) {
	// Supported symbols
	symbols := []map[string]string{
		{
			"symbol":      "BTCUSDT",
			"base_asset":  "BTC",
			"quote_asset": "USDT",
		},
		{
			"symbol":      "ETHUSDT",
			"base_asset":  "ETH",
			"quote_asset": "USDT",
		},
		{
			"symbol":      "BNBUSDT",
			"base_asset":  "BNB",
			"quote_asset": "USDT",
		},
	}

	respondSuccess(c, symbols)
}

// respondSuccess sends a successful API response
func respondSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Code:    200,
		Message: "success",
		Data:    data,
	})
}

// respondError sends an error API response
func respondError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, APIResponse{
		Code:    statusCode,
		Message: message,
		Data:    nil,
	})
}

// formatPrice formats a float64 price to string with 8 decimal places
func formatPrice(price float64) string {
	return strconv.FormatFloat(price, 'f', 8, 64)
}
