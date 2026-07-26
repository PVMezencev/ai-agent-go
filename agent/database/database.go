package database

import (
	"context"
	"database/sql"
	"time"
)

// DatabaseInterface defines the interface for database operations
type DatabaseInterface interface {
	// Connection management
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	IsConnected() bool

	// Query operations
	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
	Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error)

	// Transaction operations
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Transaction(ctx context.Context, fn func(*sql.Tx) error) error

	// Configuration
	GetConfig() DatabaseConfig
	SetConfig(config DatabaseConfig) error

	// Health check
	Ping(ctx context.Context) error
}

// DatabaseConfig represents the configuration for database operations
type DatabaseConfig struct {
	DSN          string
	MaxOpenConns int
	MaxIdleConns int
	ConnMaxLifetime time.Duration
	Timeout      time.Duration
}

// DatabaseStats represents database statistics
type DatabaseStats struct {
	OpenConnections int
	IdleConnections int
	WaitCount       int64
	WaitDuration    time.Duration
	MaxOpenConns    int
}