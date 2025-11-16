package api

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(r *gin.Engine) {
	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Kline endpoints
		// TODO: Implement kline handlers
		// v1.GET("/klines", handlers.GetKlines)
		// v1.GET("/symbols", handlers.GetSymbols)
	}
}
