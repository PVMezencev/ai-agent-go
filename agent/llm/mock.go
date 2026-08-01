package llm

import (
	"context"
	"sync"
)

// MockLLMProvider is a test double for LLMProvider that returns
// pre-programmed responses. It supports chaining tool calls → final answer.
type MockLLMProvider struct {
	mu sync.Mutex

	// Sequence of responses to return in order. Each call to ChatCompletion
	// or ChatCompletionStream pops the next response from the queue.
	// When the queue is empty, returns an empty response (no choices).
	queue []mockResponse

	// Config
	config LLMConfig
}

type mockResponse struct {
	chat    *ChatResponse
	streamFn func(handler func(chunk ChatStreamChunk) error) error
}

// NewMockLLMProvider creates a new mock LLM provider
func NewMockLLMProvider() *MockLLMProvider {
	return &MockLLMProvider{}
}

// SetConfig implements LLMProvider
func (m *MockLLMProvider) SetConfig(config LLMConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = config
	return nil
}

// GetModelInfo implements LLMProvider
func (m *MockLLMProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		Name:       "mock",
		Provider:   "mock",
		MaxTokens:  128000,
	}
}

// QueueChatResponse adds a non-streaming response to the queue
func (m *MockLLMProvider) QueueChatResponse(r *ChatResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.queue = append(m.queue, mockResponse{chat: r})
}

// QueueStreamResponse adds a streaming response to the queue
func (m *MockLLMProvider) QueueStreamResponse(fn func(handler func(chunk ChatStreamChunk) error) error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.queue = append(m.queue, mockResponse{streamFn: fn})
}

// ChatCompletion implements LLMProvider
func (m *MockLLMProvider) ChatCompletion(ctx context.Context, request ChatRequest) (*ChatResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.queue) == 0 {
		return &ChatResponse{Choices: []Choice{}}, nil
	}

	resp := m.queue[0]
	m.queue = m.queue[1:]

	if resp.chat != nil {
		return resp.chat, nil
	}
	return &ChatResponse{Choices: []Choice{}}, nil
}

// ChatCompletionStream implements LLMProvider
func (m *MockLLMProvider) ChatCompletionStream(ctx context.Context, request ChatRequest, handler func(chunk ChatStreamChunk) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.queue) == 0 {
		return nil
	}

	resp := m.queue[0]
	m.queue = m.queue[1:]

	if resp.streamFn != nil {
		return resp.streamFn(handler)
	}
	return nil
}

// Embeddings implements LLMProvider
func (m *MockLLMProvider) Embeddings(ctx context.Context, request EmbeddingRequest) (*EmbeddingResponse, error) {
	return &EmbeddingResponse{Model: "mock", Data: []EmbeddingData{}}, nil
}

// Reset clears the response queue
func (m *MockLLMProvider) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.queue = nil
}
