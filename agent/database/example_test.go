package database

import (
	"fmt"
	"time"
)

// ExampleDatabaseConfig demonstrates how to create and use database configuration
func ExampleDatabaseConfig() {
	// Create database configuration
	config := DatabaseConfig{
		DSN:           "file:example.db?cache=shared",
		MaxOpenConns:  10,
		MaxIdleConns:  5,
		ConnMaxLifetime: 5 * time.Minute,
		Timeout:       10 * time.Second,
	}

	// Print configuration details
	fmt.Printf("DSN: %s\n", config.DSN)
	fmt.Printf("Max Open Connections: %d\n", config.MaxOpenConns)
	fmt.Printf("Timeout: %v\n", config.Timeout)

	// Output:
	// DSN: file:example.db?cache=shared
	// Max Open Connections: 10
	// Timeout: 10s
}

// ExampleDatabaseStats demonstrates how to work with database statistics
func ExampleDatabaseStats() {
	// Create database stats
	stats := DatabaseStats{
		OpenConnections: 2,
		IdleConnections: 1,
		WaitCount:       5,
		WaitDuration:    100 * time.Millisecond,
		MaxOpenConns:    10,
	}

	// Print statistics
	fmt.Printf("Open Connections: %d\n", stats.OpenConnections)
	fmt.Printf("Idle Connections: %d\n", stats.IdleConnections)
	fmt.Printf("Wait Duration: %v\n", stats.WaitDuration)

	// Output:
	// Open Connections: 2
	// Idle Connections: 1
	// Wait Duration: 100ms
}