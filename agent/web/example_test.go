package web

import (
	"fmt"
	"time"
)

// ExampleWebConfig demonstrates how to create and use web configuration
func ExampleWebConfig() {
	// Create web configuration
	config := WebConfig{
		Timeout:      30 * time.Second,
		MaxRetries:   3,
		UserAgent:    "AI-Agent/1.0",
		Proxy:        "",
		AllowBlocked: false,
		RateLimit:    100,
	}

	// Print configuration details
	fmt.Printf("User Agent: %s\n", config.UserAgent)
	fmt.Printf("Timeout: %v\n", config.Timeout)
	fmt.Printf("Max Retries: %d\n", config.MaxRetries)

	// Output:
	// User Agent: AI-Agent/1.0
	// Timeout: 30s
	// Max Retries: 3
}

// ExampleSearchOptions demonstrates how to create search options
func ExampleSearchOptions() {
	// Create search options
	options := SearchOptions{
		MaxResults:    10,
		Timeout:       30 * time.Second,
		Language:      "en",
		Region:        "US",
		UserAgent:     "Mozilla/5.0",
		AllowBlocked:  true,
	}

	// Print search options
	fmt.Printf("Max Results: %d\n", options.MaxResults)
	fmt.Printf("Language: %s\n", options.Language)
	fmt.Printf("Allow Blocked: %t\n", options.AllowBlocked)

	// Output:
	// Max Results: 10
	// Language: en
	// Allow Blocked: true
}