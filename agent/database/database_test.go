package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDatabaseConfig_Create(t *testing.T) {
	// Test creating database configuration
	config := DatabaseConfig{
		DSN:           "file:test.db?cache=shared",
		MaxOpenConns:  10,
		MaxIdleConns:  5,
		ConnMaxLifetime: 5 * time.Minute,
		Timeout:       10 * time.Second,
	}

	assert.Equal(t, "file:test.db?cache=shared", config.DSN)
	assert.Equal(t, 10, config.MaxOpenConns)
	assert.Equal(t, 5, config.MaxIdleConns)
	assert.Equal(t, 5*time.Minute, config.ConnMaxLifetime)
	assert.Equal(t, 10*time.Second, config.Timeout)
}

func TestDatabaseStats_Create(t *testing.T) {
	// Test creating database stats
	stats := DatabaseStats{
		OpenConnections: 2,
		IdleConnections: 1,
		WaitCount:       5,
		WaitDuration:    100 * time.Millisecond,
		MaxOpenConns:    10,
	}

	assert.Equal(t, 2, stats.OpenConnections)
	assert.Equal(t, 1, stats.IdleConnections)
	assert.Equal(t, int64(5), stats.WaitCount)
	assert.Equal(t, 100*time.Millisecond, stats.WaitDuration)
	assert.Equal(t, 10, stats.MaxOpenConns)
}