package llm

import (
	"context"
	"time"
)

// LLMProvider represents the interface for LLM providers
type LLMProvider interface {
	// ChatCompletion generates a chat completion response
	ChatCompletion(ctx context.Context, request ChatRequest) (*ChatResponse, error)

	// Embeddings generates embeddings for text
	Embeddings(ctx context.Context, request EmbeddingRequest) (*EmbeddingResponse, error)

	// GetModelInfo returns information about the model
	GetModelInfo() ModelInfo

	// SetConfig configures the provider
	SetConfig(config LLMConfig) error
}

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Model    string
	Messages []Message
	Stream   bool
	Timeout  time.Duration
}

// Message represents a single message in a chat
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
	ID      string    `json:"id"`
	Model   string    `json:"model"`
	Created time.Time `json:"created"`
	Choices []Choice  `json:"choices"`
}

// Choice represents a single choice in the chat response
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// EmbeddingRequest represents an embeddings request
type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// EmbeddingResponse represents an embeddings response
type EmbeddingResponse struct {
	Model string        `json:"model"`
	Data  []EmbeddingData `json:"data"`
}

// EmbeddingData represents a single embedding
type EmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

// ModelInfo represents information about a model
type ModelInfo struct {
	Name        string
	Description string
	MaxTokens   int
	Provider    string
}

// LLMConfig represents the configuration for LLM providers
type LLMConfig struct {
	APIKey      string
	APIEndpoint string
	Model       string
	Timeout     time.Duration
	MaxRetries  int
}