package database

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteDB implements the DatabaseInterface for SQLite database
type SQLiteDB struct {
	config DatabaseConfig
	db     *sql.DB
}

// NewSQLiteDB creates a new SQLite database instance
func NewSQLiteDB(config DatabaseConfig) (*SQLiteDB, error) {
	return &SQLiteDB{
		config: config,
	}, nil
}

// Connect establishes a connection to the database
func (db *SQLiteDB) Connect(ctx context.Context) error {
	// Create connection with timeout
	ctx, cancel := context.WithTimeout(ctx, db.config.Timeout)
	defer cancel()

	var err error
	db.db, err = sql.Open("sqlite3", db.config.DSN)
	if err != nil {
		return err
	}

	// Set connection pool settings
	db.db.SetMaxOpenConns(db.config.MaxOpenConns)
	db.db.SetMaxIdleConns(db.config.MaxIdleConns)
	db.db.SetConnMaxLifetime(db.config.ConnMaxLifetime)

	// Test the connection
	return db.db.PingContext(ctx)
}

// Disconnect closes the database connection
func (db *SQLiteDB) Disconnect(ctx context.Context) error {
	if db.db != nil {
		return db.db.Close()
	}
	return nil
}

// IsConnected checks if the database is connected
func (db *SQLiteDB) IsConnected() bool {
	return db.db != nil
}

// Query executes a query that returns rows
func (db *SQLiteDB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	ctx, cancel := context.WithTimeout(ctx, db.config.Timeout)
	defer cancel()

	return db.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that returns a single row
func (db *SQLiteDB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	ctx, cancel := context.WithTimeout(ctx, db.config.Timeout)
	defer cancel()

	return db.db.QueryRowContext(ctx, query, args...)
}

// Exec executes a query without returning rows
func (db *SQLiteDB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	ctx, cancel := context.WithTimeout(ctx, db.config.Timeout)
	defer cancel()

	return db.db.ExecContext(ctx, query, args...)
}

// BeginTx starts a transaction
func (db *SQLiteDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	ctx, cancel := context.WithTimeout(ctx, db.config.Timeout)
	defer cancel()

	return db.db.BeginTx(ctx, opts)
}

// Transaction executes a function within a transaction
func (db *SQLiteDB) Transaction(ctx context.Context, fn func(*sql.Tx) error) error {
	ctx, cancel := context.WithTimeout(ctx, db.config.Timeout)
	defer cancel()

	tx, err := db.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	return fn(tx)
}

// GetConfig returns the database configuration
func (db *SQLiteDB) GetConfig() DatabaseConfig {
	return db.config
}

// SetConfig sets the database configuration
func (db *SQLiteDB) SetConfig(config DatabaseConfig) error {
	db.config = config
	return nil
}

// Ping checks the database connection
func (db *SQLiteDB) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, db.config.Timeout)
	defer cancel()

	return db.db.PingContext(ctx)
}