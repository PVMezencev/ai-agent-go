package web

import (
	"context"
	"time"
)

// WebSearchInterface defines the interface for web search operations
type WebSearchInterface interface {
	// Search operations
	Search(ctx context.Context, query string, options SearchOptions) (*SearchResults, error)
	SearchWithBrowser(ctx context.Context, query string, options SearchOptions) (*SearchResults, error)

	// Scraping operations
	Scrape(ctx context.Context, url string, options ScrapeOptions) (*ScrapeResult, error)

	// Configuration
	GetConfig() WebConfig
	SetConfig(config WebConfig) error

	// Proxy operations
	SetProxy(proxy string) error
}

// SearchOptions represents search options
type SearchOptions struct {
	MaxResults    int
	Timeout       time.Duration
	Language      string
	Region        string
	UserAgent     string
	Headers       map[string]string
	AllowBlocked  bool
}

// SearchResults represents search results
type SearchResults struct {
	Query     string
	Results   []SearchResult
	Total     int
	NextPage  string
	Processed time.Time
}

// SearchResult represents a single search result
type SearchResult struct {
	Title       string
	URL         string
	Description string
	Snippet     string
	Source      string
	Thumbnail   string
	LastUpdated time.Time
}

// ScrapeOptions represents scraping options
type ScrapeOptions struct {
	Timeout     time.Duration
	UserAgent   string
	Headers     map[string]string
	WaitFor     string
	JavaScript  bool
	Images      bool
	Scripts     bool
}

// ScrapeResult represents scraping results
type ScrapeResult struct {
	URL         string
	Title       string
	Content     string
	HTML        string
	Links       []string
	Images      []string
	Meta        map[string]string
	Processed   time.Time
	StatusCode  int
}

// WebConfig represents the configuration for web operations
type WebConfig struct {
	Timeout      time.Duration
	MaxRetries   int
	UserAgent    string
	Proxy        string
	Headers      map[string]string
	AllowBlocked bool
	RateLimit    int
}