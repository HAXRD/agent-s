package api

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorHandlerMiddleware handles panics and errors
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "Internal server error",
					"data":    nil,
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// LoggerMiddleware logs HTTP requests
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		log.Printf("%s %s %d %v", method, path, status, latency)
	}
}
