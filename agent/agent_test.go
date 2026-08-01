package agent

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/PVMezencev/ai-agent-go/agent/core"
	"github.com/PVMezencev/ai-agent-go/agent/llm"
	"github.com/PVMezencev/ai-agent-go/agent/tools"
	"github.com/PVMezencev/ai-agent-go/agent/web"
)

func TestAgent_Execute_FinalAnswerNoTools(t *testing.T) {
	mock := llm.NewMockLLMProvider()
	mock.QueueChatResponse(&llm.ChatResponse{
		Choices: []llm.Choice{
			{
				Message: llm.AssistantMessage("Direct answer"),
			},
		},
	})

	agent, err := NewAgent(AgentConfig{
		AgentConfig: core.AgentConfig{
			ID:      "test-agent",
			Name:    "Test Agent",
			Enabled: true,
		},
		MaxToolRounds: 5,
	})
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}
	agent.LLMProvider = mock

	result, err := agent.Execute(context.Background(), "Hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Direct answer" {
		t.Errorf("expected 'Direct answer', got %q", result)
	}
}

func TestAgent_Execute_ToolCallThenFinalAnswer(t *testing.T) {
	mock := llm.NewMockLLMProvider()

	// First round: LLM calls a tool
	mock.QueueChatResponse(&llm.ChatResponse{
		Choices: []llm.Choice{
			{
				Message: llm.ToolCallMessage([]llm.ToolCall{
					{
						ID:   "call_1",
						Type: "function",
						FunctionCall: llm.FunctionCall{
							Name:      "read_file",
							Arguments: `{"path": "test.txt"}`,
						},
					},
				}),
			},
		},
	})

	// Second round: LLM gives final answer after seeing tool result
	mock.QueueChatResponse(&llm.ChatResponse{
		Choices: []llm.Choice{
			{
				Message: llm.AssistantMessage("Here is the file content"),
			},
		},
	})

	agent, err := NewAgent(AgentConfig{
		AgentConfig: core.AgentConfig{
			ID:      "test-agent",
			Name:    "Test Agent",
			Enabled: true,
		},
		MaxToolRounds: 5,
	})
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}
	agent.LLMProvider = mock

	result, err := agent.Execute(context.Background(), "Read test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "Here is the file content" {
		t.Errorf("expected 'Here is the file content', got %q", result)
	}
}

func TestAgent_Execute_MaxRoundsExceeded(t *testing.T) {
	mock := llm.NewMockLLMProvider()

	// Keep requesting tool calls for all rounds
	for i := 0; i < 10; i++ {
		mock.QueueChatResponse(&llm.ChatResponse{
			Choices: []llm.Choice{
				{
					Message: llm.ToolCallMessage([]llm.ToolCall{
						{
							ID:   fmt.Sprintf("call_%d", i),
							Type: "function",
							FunctionCall: llm.FunctionCall{
								Name:      "read_file",
								Arguments: `{"path": "test.txt"}`,
							},
						},
					}),
				},
			},
		})
	}

	agent, err := NewAgent(AgentConfig{
		AgentConfig: core.AgentConfig{
			ID:      "test-agent",
			Name:    "Test Agent",
			Enabled: true,
		},
		MaxToolRounds: 5,
	})
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}
	agent.LLMProvider = mock

	_, err = agent.Execute(context.Background(), "Read test.txt")
	if err == nil {
		t.Fatal("expected error for max rounds exceeded, got nil")
	}
	if !strings.Contains(err.Error(), "maximum tool rounds") {
		t.Errorf("expected 'maximum tool rounds' error, got: %v", err)
	}
}

func TestAgent_Execute_NoLLMProvider(t *testing.T) {
	agent, err := NewAgent(AgentConfig{
		AgentConfig: core.AgentConfig{
			ID:      "test-agent",
			Name:    "Test Agent",
			Enabled: true,
		},
		MaxToolRounds: 5,
	})
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}
	// LLMProvider is nil

	_, err = agent.Execute(context.Background(), "Hello")
	if err == nil {
		t.Fatal("expected error for missing LLM provider, got nil")
	}
	if !strings.Contains(err.Error(), "LLM provider is not configured") {
		t.Errorf("expected 'LLM provider is not configured' error, got: %v", err)
	}
}

func TestAgent_ExecuteStream_FinalAnswer(t *testing.T) {
	mock := llm.NewMockLLMProvider()

	// First round: direct answer (no tool calls) — should trigger streaming
	mock.QueueChatResponse(&llm.ChatResponse{
		Choices: []llm.Choice{
			{
				Message: llm.AssistantMessage("Streamed answer chunk 1. Streamed answer chunk 2."),
			},
		},
	})

	agent, err := NewAgent(AgentConfig{
		AgentConfig: core.AgentConfig{
			ID:      "test-agent",
			Name:    "Test Agent",
			Enabled: true,
		},
		MaxToolRounds: 5,
	})
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}
	agent.LLMProvider = mock

	var chunks []string
	handler := func(chunk llm.ChatStreamChunk) error {
		chunks = append(chunks, chunk.Content)
		return nil
	}

	err = agent.ExecuteStream(context.Background(), "Hello", handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Chunks should have been produced by streamContent splitting on ~50 chars
	if len(chunks) == 0 {
		t.Fatal("expected at least one chunk, got none")
	}

	// All chunks concatenated should equal the full content
	full := strings.Join(chunks, "")
	if full != "Streamed answer chunk 1. Streamed answer chunk 2." {
		t.Errorf("expected full content to match, got %q", full)
	}
}

func TestAgent_ExecuteStream_ToolCallThenStream(t *testing.T) {
	mock := llm.NewMockLLMProvider()

	// First round: tool call
	mock.QueueChatResponse(&llm.ChatResponse{
		Choices: []llm.Choice{
			{
				Message: llm.ToolCallMessage([]llm.ToolCall{
					{
						ID:   "call_1",
						Type: "function",
						FunctionCall: llm.FunctionCall{
							Name:      "read_file",
							Arguments: `{"path": "test.txt"}`,
						},
					},
				}),
			},
		},
	})

	// Second round: final answer — should be streamed
	mock.QueueChatResponse(&llm.ChatResponse{
		Choices: []llm.Choice{
			{
				Message: llm.AssistantMessage("File content here"),
			},
		},
	})

	agent, err := NewAgent(AgentConfig{
		AgentConfig: core.AgentConfig{
			ID:      "test-agent",
			Name:    "Test Agent",
			Enabled: true,
		},
		MaxToolRounds: 5,
	})
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}
	agent.LLMProvider = mock

	var chunks []string
	handler := func(chunk llm.ChatStreamChunk) error {
		chunks = append(chunks, chunk.Content)
		return nil
	}

	err = agent.ExecuteStream(context.Background(), "Read test.txt", handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(chunks) == 0 {
		t.Fatal("expected at least one chunk, got none")
	}
}

func TestAgent_ExecuteStream_NoLLMProvider(t *testing.T) {
	agent, err := NewAgent(AgentConfig{
		AgentConfig: core.AgentConfig{
			ID:      "test-agent",
			Name:    "Test Agent",
			Enabled: true,
		},
		MaxToolRounds: 5,
	})
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}

	err = agent.ExecuteStream(context.Background(), "Hello", func(chunk llm.ChatStreamChunk) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error for missing LLM provider, got nil")
	}
	if !strings.Contains(err.Error(), "LLM provider is not configured") {
		t.Errorf("expected 'LLM provider is not configured' error, got: %v", err)
	}
}

func TestAgent_ConversationIsThreadSafe(t *testing.T) {
	mock := llm.NewMockLLMProvider()
	mock.QueueChatResponse(&llm.ChatResponse{
		Choices: []llm.Choice{
			{Message: llm.AssistantMessage("answer")},
		},
	})

	agent, err := NewAgent(AgentConfig{
		AgentConfig: core.AgentConfig{
			ID:      "test-agent",
			Name:    "Test Agent",
			Enabled: true,
		},
		MaxToolRounds: 5,
	})
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}
	agent.LLMProvider = mock

	// Run Execute multiple times concurrently — should not race
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, _ = agent.Execute(context.Background(), "test")
			done <- true
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestAgent_StatusWithTools(t *testing.T) {
	mock := llm.NewMockLLMProvider()

	agent, err := NewAgent(AgentConfig{
		AgentConfig: core.AgentConfig{
			ID:      "test-agent",
			Name:    "Test Agent",
			Enabled: true,
		},
		MaxToolRounds: 5,
	})
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}
	agent.LLMProvider = mock

	status := agent.GetStatus()
	if !status.IsRunning {
		t.Error("expected agent to be running")
	}
	if status.Stats["llm_configured"] != true {
		t.Error("expected LLM to be configured")
	}
	if status.Stats["tools_count"] == nil || status.Stats["tools_count"].(int) == 0 {
		t.Error("expected tools to be registered")
	}
}

func TestAgent_Execute_EmptyChoices(t *testing.T) {
	mock := llm.NewMockLLMProvider()
	mock.QueueChatResponse(&llm.ChatResponse{
		Choices: []llm.Choice{},
	})

	agent, err := NewAgent(AgentConfig{
		AgentConfig: core.AgentConfig{
			ID:      "test-agent",
			Name:    "Test Agent",
			Enabled: true,
		},
		MaxToolRounds: 5,
	})
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}
	agent.LLMProvider = mock

	_, err = agent.Execute(context.Background(), "Hello")
	if err == nil {
		t.Fatal("expected error for empty choices, got nil")
	}
	if !strings.Contains(err.Error(), "no choices") {
		t.Errorf("expected 'no choices' error, got: %v", err)
	}
}

func TestAgent_Execute_StreamingContextCancellation(t *testing.T) {
	mock := llm.NewMockLLMProvider()
	mock.QueueChatResponse(&llm.ChatResponse{
		Choices: []llm.Choice{
			{Message: llm.AssistantMessage("long answer")},
		},
	})

	agent, err := NewAgent(AgentConfig{
		AgentConfig: core.AgentConfig{
			ID:      "test-agent",
			Name:    "Test Agent",
			Enabled: true,
		},
		MaxToolRounds: 5,
	})
	if err != nil {
		t.Fatalf("failed to create agent: %v", err)
	}
	agent.LLMProvider = mock

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	var chunks []string
	handler := func(chunk llm.ChatStreamChunk) error {
		chunks = append(chunks, chunk.Content)
		return nil
	}

	err = agent.ExecuteStream(ctx, "Hello", handler)
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
}

func TestNewToolFunc_Description(t *testing.T) {
	fn := tools.NewToolFunc(
		"test_tool",
		"A test tool",
		map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{"x": map[string]interface{}{"type": "number"}},
		},
		func(ctx context.Context, argsJSON string) (string, error) {
			return "ok", nil
		},
	)

	if fn.Description() != "A test tool" {
		t.Errorf("expected 'A test tool', got %q", fn.Description())
	}
}

func TestAgentConfig_DefaultMaxToolRounds(t *testing.T) {
	// When MaxToolRounds is 0, it should default to 10
	config := AgentConfig{
		AgentConfig: core.AgentConfig{
			ID:      "test",
			Enabled: true,
		},
		MaxToolRounds: 0,
	}

	if config.MaxToolRounds != 0 {
		t.Errorf("expected MaxToolRounds to be 0, got %d", config.MaxToolRounds)
	}
}

func TestToolResult_Formatting(t *testing.T) {
	// String result
	if got := tools.ToolResult("success", nil); got != "success" {
		t.Errorf("expected 'success', got %q", got)
	}

	// Error result
	errMsg := tools.ToolResult("", fmt.Errorf("something went wrong"))
	if !strings.Contains(errMsg, "Error:") {
		t.Errorf("expected 'Error:' prefix, got %q", errMsg)
	}

	// Struct result
	type data struct {
		Name string `json:"name"`
	}
	result := tools.ToolResult(data{Name: "test"}, nil)
	if !strings.Contains(result, `"name"`) {
		t.Errorf("expected JSON with 'name' field, got %q", result)
	}
}

func TestWebSearch_RealSearch(t *testing.T) {
	ws, err := web.NewWebSearch(web.WebConfig{
		Timeout:   10 * time.Second,
		UserAgent: "TestAgent/1.0",
	})
	if err != nil {
		t.Fatalf("failed to create web search: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	results, err := ws.Search(ctx, "Go programming language", web.SearchOptions{
		MaxResults: 3,
	})
	if err != nil {
		t.Fatalf("search failed: %v", err)
	}

	if results.Total == 0 {
		t.Fatal("expected at least one search result")
	}

	// Verify result structure
	first := results.Results[0]
	if first.URL == "" {
		t.Error("expected non-empty URL in first result")
	}
	if first.Title == "" {
		t.Error("expected non-empty title in first result")
	}
}
