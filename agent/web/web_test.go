package web

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWebConfig_Create(t *testing.T) {
	// Test creating web configuration
	config := WebConfig{
		Timeout:      30 * time.Second,
		MaxRetries:   3,
		UserAgent:    "AI-Agent/1.0",
		Proxy:        "http://proxy.example.com:8080",
		AllowBlocked: true,
		RateLimit:    100,
	}

	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, 3, config.MaxRetries)
	assert.Equal(t, "AI-Agent/1.0", config.UserAgent)
	assert.Equal(t, "http://proxy.example.com:8080", config.Proxy)
	assert.True(t, config.AllowBlocked)
	assert.Equal(t, 100, config.RateLimit)
}

func TestSearchOptions_Create(t *testing.T) {
	// Test creating search options
	options := SearchOptions{
		MaxResults:    10,
		Timeout:       30 * time.Second,
		Language:      "en",
		Region:        "US",
		UserAgent:     "Mozilla/5.0",
		AllowBlocked:  true,
	}

	assert.Equal(t, 10, options.MaxResults)
	assert.Equal(t, 30*time.Second, options.Timeout)
	assert.Equal(t, "en", options.Language)
	assert.Equal(t, "US", options.Region)
	assert.Equal(t, "Mozilla/5.0", options.UserAgent)
	assert.True(t, options.AllowBlocked)
}

func TestScrapeOptions_Create(t *testing.T) {
	// Test creating scrape options
	options := ScrapeOptions{
		Timeout:     30 * time.Second,
		UserAgent:   "Mozilla/5.0",
		WaitFor:     "load",
		JavaScript:  true,
		Images:      true,
		Scripts:     true,
	}

	assert.Equal(t, 30*time.Second, options.Timeout)
	assert.Equal(t, "Mozilla/5.0", options.UserAgent)
	assert.Equal(t, "load", options.WaitFor)
	assert.True(t, options.JavaScript)
	assert.True(t, options.Images)
	assert.True(t, options.Scripts)
}