package llm

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLLMConfig_Create(t *testing.T) {
	// Test creating LLM configuration
	config := LLMConfig{
		APIKey:      "test-key",
		APIEndpoint: "https://api.openai.com",
		Model:       "gpt-4",
		Timeout:     30 * time.Second,
		MaxRetries:  3,
	}

	assert.Equal(t, "test-key", config.APIKey)
	assert.Equal(t, "https://api.openai.com", config.APIEndpoint)
	assert.Equal(t, "gpt-4", config.Model)
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, 3, config.MaxRetries)
}

func TestChatRequest_Create(t *testing.T) {
	// Test creating chat request
	request := ChatRequest{
		Model:   "gpt-4",
		Stream:  true,
		Timeout: 30 * time.Second,
		Messages: []Message{
			{
				Role:    "user",
				Content: "Hello, world!",
			},
		},
	}

	assert.Equal(t, "gpt-4", request.Model)
	assert.True(t, request.Stream)
	assert.Equal(t, 30*time.Second, request.Timeout)
	assert.Len(t, request.Messages, 1)
	assert.Equal(t, "user", request.Messages[0].Role)
	assert.Equal(t, "Hello, world!", request.Messages[0].Content)
}

func TestMessage_Create(t *testing.T) {
	// Test creating message
	message := Message{
		Role:    "assistant",
		Content: "Hello, user!",
	}

	assert.Equal(t, "assistant", message.Role)
	assert.Equal(t, "Hello, user!", message.Content)
}

func TestModelInfo_Create(t *testing.T) {
	// Test creating model info
	info := ModelInfo{
		Name:        "gpt-4",
		Description: "OpenAI GPT-4 model",
		MaxTokens:   4096,
		Provider:    "OpenAI",
	}

	assert.Equal(t, "gpt-4", info.Name)
	assert.Equal(t, "OpenAI GPT-4 model", info.Description)
	assert.Equal(t, 4096, info.MaxTokens)
	assert.Equal(t, "OpenAI", info.Provider)
}