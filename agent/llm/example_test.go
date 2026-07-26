package llm

import (
	"fmt"
	"time"
)

// ExampleLLMConfig demonstrates how to create and use LLM configuration
func ExampleLLMConfig() {
	// Create LLM configuration
	config := LLMConfig{
		APIKey:      "your-openai-api-key",
		APIEndpoint: "https://api.openai.com/v1",
		Model:       "gpt-4",
		Timeout:     30 * time.Second,
		MaxRetries:  3,
	}

	// Print configuration details
	fmt.Printf("Model: %s\n", config.Model)
	fmt.Printf("Timeout: %v\n", config.Timeout)
	fmt.Printf("Max Retries: %d\n", config.MaxRetries)

	// Output:
	// Model: gpt-4
	// Timeout: 30s
	// Max Retries: 3
}

// ExampleChatRequest demonstrates how to create a chat request
func ExampleChatRequest() {
	// Create a chat request
	request := ChatRequest{
		Model:   "gpt-4",
		Stream:  false,
		Timeout: 30 * time.Second,
		Messages: []Message{
			{
				Role:    "user",
				Content: "Hello, how are you?",
			},
			{
				Role:    "assistant",
				Content: "I'm doing well, thank you for asking!",
			},
		},
	}

	// Print request details
	fmt.Printf("Model: %s\n", request.Model)
	fmt.Printf("Number of messages: %d\n", len(request.Messages))

	// Output:
	// Model: gpt-4
	// Number of messages: 2
}