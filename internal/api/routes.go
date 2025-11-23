package api

import (
	"crypto-monitor/internal/api/handlers"
	"crypto-monitor/internal/repository"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(r *gin.Engine, klineRepo *repository.KlineRepository) {
	// Apply middleware
	r.Use(LoggerMiddleware())
	r.Use(ErrorHandlerMiddleware())

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Initialize handlers
		klineHandler := handlers.NewKlineHandler(klineRepo)

		// Kline endpoints
		v1.GET("/klines", klineHandler.GetKlines)
		v1.GET("/symbols", klineHandler.GetSymbols)
	}
}
