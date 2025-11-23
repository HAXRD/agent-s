package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"crypto-monitor/internal/api"
	"crypto-monitor/internal/repository"
	"crypto-monitor/internal/service"
	"crypto-monitor/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func main() {
	// Initialize database connection
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	// Test database connection
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Database connection test successful")

	// Initialize services
	klineRepo := repository.NewKlineRepository(db)
	binanceSvc := service.NewBinanceService()
	wsSvc := service.NewWebSocketService(binanceSvc, klineRepo)

	// Start WebSocket service
	go wsSvc.Run()
	log.Println("WebSocket service started")

	// Initialize Gin router
	r := gin.Default()

	// Setup API routes
	api.SetupRoutes(r, klineRepo)

	// Setup WebSocket route
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// Allow all origins for development
			// In production, you should validate the origin
			return true
		},
	}

	r.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("WebSocket upgrade failed: %v", err)
			return
		}
		wsSvc.HandleConnection(conn)
	})

	// Get server port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start HTTP server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
