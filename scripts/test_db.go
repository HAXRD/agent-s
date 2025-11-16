package main

import (
	"fmt"
	"log"
	"os"

	"crypto-monitor/pkg/database"
)

func main() {
	// Load environment variables from .env if exists
	// This is a simple test script, so we rely on environment variables

	fmt.Println("Testing database connection...")

	// Initialize database
	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.CloseDB()

	// Test connection
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("✓ Database connection successful!")

	// Check if klines table exists
	var tableExists bool
	err = db.Raw(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'klines'
		)
	`).Scan(&tableExists).Error

	if err != nil {
		log.Fatalf("Failed to check table existence: %v", err)
	}

	if tableExists {
		fmt.Println("✓ klines table exists")
	} else {
		fmt.Println("✗ klines table does not exist")
		os.Exit(1)
	}

	// Check indexes
	var indexCount int64
	db.Raw(`
		SELECT COUNT(*) 
		FROM pg_indexes 
		WHERE tablename = 'klines'
	`).Scan(&indexCount)

	fmt.Printf("✓ Found %d indexes on klines table\n", indexCount)

	fmt.Println("\nDatabase test completed successfully!")
}
