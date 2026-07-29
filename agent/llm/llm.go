package llm

import (
	"context"
	"encoding/json"
	"time"
)

// LLMProvider represents the interface for LLM providers
type LLMProvider interface {
	// ChatCompletion generates a chat completion response
	ChatCompletion(ctx context.Context, request ChatRequest) (*ChatResponse, error)

	// ChatCompletionStream streams chat completion responses
	ChatCompletionStream(ctx context.Context, request ChatRequest, handler func(chunk ChatStreamChunk) error) error

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
	Tools    []ToolDef
	Stream   bool
	Timeout  time.Duration
}

// ToolDef represents a tool definition for function calling
type ToolDef struct {
	Type     string
	Function FunctionDef
}

// FunctionDef represents a function definition for tool calling
type FunctionDef struct {
	Name        string
	Description string
	Parameters  map[string]interface{}
}

// Message represents a single message in a chat
type Message struct {
	Role         string
	Content      string
	ToolCalls    []ToolCall
	ToolCallID   string
}

// ToolCall represents a tool call from the LLM
type ToolCall struct {
	ID             string
	Type           string
	FunctionCall   FunctionCall
}

// FunctionCall represents a function call within a tool call
type FunctionCall struct {
Name      string
	Arguments string
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

// AssistantMessage creates an assistant message
func AssistantMessage(content string) Message {
	return Message{Role: "assistant", Content: content}
}

// UserMessage creates a user message
func UserMessage(content string) Message {
	return Message{Role: "user", Content: content}
}

// ToolMessage creates a tool result message
func ToolMessage(toolCallID, content string) Message {
	return Message{Role: "tool", Content: content, ToolCallID: toolCallID}
}

// ToolCallMessage creates an assistant message with tool calls
func ToolCallMessage(toolCalls []ToolCall) Message {
	return Message{Role: "assistant", ToolCalls: toolCalls}
}

// MarshalToolDef converts a ToolDef to JSON-serializable format for the API
func MarshalToolDef(td ToolDef) (json.RawMessage, error) {
	params := td.Function.Parameters
	if params == nil {
		params = map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		}
	}

	return json.Marshal(map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name":        td.Function.Name,
			"description": td.Function.Description,
			"parameters":  params,
		},
	})
}
