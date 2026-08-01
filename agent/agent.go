package agent

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/PVMezencev/ai-agent-go/agent/core"
	"github.com/PVMezencev/ai-agent-go/agent/database"
	"github.com/PVMezencev/ai-agent-go/agent/filesystem"
	"github.com/PVMezencev/ai-agent-go/agent/llm"
	"github.com/PVMezencev/ai-agent-go/agent/tools"
	"github.com/PVMezencev/ai-agent-go/agent/web"
)

// Agent represents the main AI agent that coordinates all modules
type Agent struct {
	core.Agent

	// Module interfaces
	LLMProvider    llm.LLMProvider
	FileSystem     filesystem.FileSystemInterface
	Database       database.DatabaseInterface
	WebSearch      web.WebSearchInterface

	// Tool registry
	ToolRegistry *tools.Registry

	// Configuration
	config AgentConfig

	// Conversation history (protected by mutex for concurrency safety)
	mu         sync.Mutex
	conversation []llm.Message
}

// AgentConfig represents the configuration for the entire agent
type AgentConfig struct {
	core.AgentConfig
	LLMConfig        llm.LLMConfig
	FileSystemConfig filesystem.FileSystemConfig
	DatabaseConfig   database.DatabaseConfig
	WebConfig        web.WebConfig
	MaxToolRounds    int // Max iterations of tool calling before stopping
}

// NewAgent creates a new AI agent instance
func NewAgent(config AgentConfig) (*Agent, error) {
	agent := &Agent{
		Agent: core.Agent{
			ID:        config.ID,
			Name:      config.Name,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		config: config,
	}

	// Initialize modules
	if err := agent.initModules(); err != nil {
		return nil, fmt.Errorf("failed to initialize agent modules: %w", err)
	}

	// Build tool registry from available modules
	agent.ToolRegistry = tools.BuildAllTools(
		agent.FileSystem,
		agent.Database,
		agent.WebSearch,
	)

	return agent, nil
}

// initModules initializes all the agent modules
func (a *Agent) initModules() error {
	// Initialize LLM provider
	if a.config.LLMConfig.APIKey != "" || a.config.LLMConfig.APIEndpoint != "" {
		provider, err := llm.NewOpenAIProvider(a.config.LLMConfig)
		if err != nil {
			return fmt.Errorf("failed to create LLM provider: %w", err)
		}
		a.LLMProvider = provider
	}

	// Initialize file system
	a.FileSystem = filesystem.NewLocalFileSystem(a.config.FileSystemConfig)

	// Initialize database
	if a.config.DatabaseConfig.DSN != "" {
		db, err := database.NewSQLiteDB(a.config.DatabaseConfig)
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
		a.Database = db
	}

	// Initialize web search
	webSearch, err := web.NewWebSearch(a.config.WebConfig)
	if err != nil {
		return err
	}
	a.WebSearch = webSearch

	return nil
}

// Start starts the agent and all its modules
func (a *Agent) Start(ctx context.Context) error {
	if a.Database != nil {
		if err := a.Database.Connect(ctx); err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
	}

	a.mu.Lock()
	a.UpdatedAt = time.Now()
	a.mu.Unlock()
	return nil
}

// Stop stops the agent and all its modules
func (a *Agent) Stop(ctx context.Context) error {
	if a.Database != nil {
		if err := a.Database.Disconnect(ctx); err != nil {
			return fmt.Errorf("failed to disconnect from database: %w", err)
		}
	}

	a.mu.Lock()
	a.UpdatedAt = time.Now()
	a.mu.Unlock()
	return nil
}

// Execute executes a task using the agent's capabilities and available tools.
// It sends the task to the LLM with all registered tools, handles tool calls
// iteratively, and returns the final result.
func (a *Agent) Execute(ctx context.Context, task string) (string, error) {
	if a.LLMProvider == nil {
		return "", fmt.Errorf("LLM provider is not configured — set OPENAI_API_KEY to enable")
	}

	a.setConversation(llm.UserMessage(task))

	maxRounds := a.config.MaxToolRounds
	if maxRounds == 0 {
		maxRounds = 10
	}

	for round := 0; round < maxRounds; round++ {
		a.mu.Lock()
		a.UpdatedAt = time.Now()
		a.mu.Unlock()

		// Build the request with tools
		request := llm.ChatRequest{
			Model:    a.config.LLMConfig.Model,
			Messages: a.getConversation(),
			Tools:    a.ToolRegistry.ToLLMTools(),
			Timeout:  a.config.Timeout,
		}

		// Call the LLM
		response, err := a.LLMProvider.ChatCompletion(ctx, request)
		if err != nil {
			return "", fmt.Errorf("LLM call failed: %w", err)
		}

		if len(response.Choices) == 0 {
			return "", fmt.Errorf("LLM returned no choices")
		}

		assistantMsg := response.Choices[0].Message
		a.appendConversation(assistantMsg)

		// Check if the LLM wants to call tools
		if len(assistantMsg.ToolCalls) == 0 {
			// No tool calls — this is the final answer
			return strings.TrimSpace(assistantMsg.Content), nil
		}

		// Execute each tool call and add results to conversation
		for _, tc := range assistantMsg.ToolCalls {
			result, err := a.ToolRegistry.ExecuteToolCall(ctx, tc)
			toolMsg := llm.ToolMessage(tc.ID, tools.ToolResult(result, err))
			a.appendConversation(toolMsg)
		}
	}

	// If we exhaust all rounds, return an error
	return "", fmt.Errorf("agent exceeded maximum tool rounds (%d) without producing a final answer", maxRounds)
}

// ExecuteStream streams the execution result to the provided handler.
// It runs the same orchestration loop as Execute, but the final LLM response
// is streamed chunk by chunk to the handler function — without a second API call.
func (a *Agent) ExecuteStream(ctx context.Context, task string, handler func(chunk llm.ChatStreamChunk) error) error {
	if a.LLMProvider == nil {
		return fmt.Errorf("LLM provider is not configured — set OPENAI_API_KEY to enable")
	}

	a.setConversation(llm.UserMessage(task))

	maxRounds := a.config.MaxToolRounds
	if maxRounds == 0 {
		maxRounds = 10
	}

	for round := 0; round < maxRounds; round++ {
		a.mu.Lock()
		a.UpdatedAt = time.Now()
		a.mu.Unlock()

		// Build the request with tools
		request := llm.ChatRequest{
			Model:    a.config.LLMConfig.Model,
			Messages: a.getConversation(),
			Tools:    a.ToolRegistry.ToLLMTools(),
			Timeout:  a.config.Timeout,
		}

		// Call the LLM (non-streaming for tool-calling rounds)
		response, err := a.LLMProvider.ChatCompletion(ctx, request)
		if err != nil {
			return fmt.Errorf("LLM call failed: %w", err)
		}

		if len(response.Choices) == 0 {
			return fmt.Errorf("LLM returned no choices")
		}

		assistantMsg := response.Choices[0].Message
		a.appendConversation(assistantMsg)

		// If no tool calls, this is the final answer — stream it using the same request
		if len(assistantMsg.ToolCalls) == 0 {
			return a.streamContent(ctx, assistantMsg.Content, handler)
		}

		// Execute each tool call and add results to conversation
		for _, tc := range assistantMsg.ToolCalls {
			result, execErr := a.ToolRegistry.ExecuteToolCall(ctx, tc)
			toolMsg := llm.ToolMessage(tc.ID, tools.ToolResult(result, execErr))
			a.appendConversation(toolMsg)
		}
	}

	return fmt.Errorf("agent exceeded maximum tool rounds (%d)", maxRounds)
}

// streamContent streams the given text content chunk by chunk.
// It splits on word boundaries to produce natural-looking chunks.
func (a *Agent) streamContent(ctx context.Context, content string, handler func(chunk llm.ChatStreamChunk) error) error {
	if content == "" {
		return nil
	}

	// Send in ~50-char chunks to simulate streaming
	const chunkSize = 50
	runes := []rune(content)
	i := 0
	for i < len(runes) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		end := i + chunkSize
		if end >= len(runes) {
			end = len(runes)
		} else {
			// Try to break on word boundary
			for j := end; j < len(runes); j++ {
				if runes[j] == ' ' || runes[j] == '\n' {
					end = j
					break
				}
			}
		}

		chunk := string(runes[i:end])
		if err := handler(llm.ChatStreamChunk{Content: chunk}); err != nil {
			return err
		}
		i = end
	}

	return nil
}

// getConversation returns a copy of the conversation (thread-safe).
func (a *Agent) getConversation() []llm.Message {
	a.mu.Lock()
	defer a.mu.Unlock()
	result := make([]llm.Message, len(a.conversation))
	copy(result, a.conversation)
	return result
}

// setConversation replaces the entire conversation (thread-safe).
func (a *Agent) setConversation(msg llm.Message) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.conversation = []llm.Message{msg}
}

// appendConversation appends a message to the conversation (thread-safe).
func (a *Agent) appendConversation(msg llm.Message) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.conversation = append(a.conversation, msg)
}

// GetConfig returns the agent configuration
func (a *Agent) GetConfig() core.AgentConfig {
	return a.config.AgentConfig
}

// SetConfig sets the agent configuration
func (a *Agent) SetConfig(config core.AgentConfig) error {
	a.config.AgentConfig = config
	a.mu.Lock()
	a.UpdatedAt = time.Now()
	a.mu.Unlock()
	return nil
}

// GetStatus returns the agent status
func (a *Agent) GetStatus() core.AgentStatus {
	a.mu.Lock()
	lastActive := a.UpdatedAt
	a.mu.Unlock()

	status := core.AgentStatus{
		IsRunning:  true,
		LastActive: lastActive,
		Stats:      make(map[string]interface{}),
	}

	status.Stats["llm_configured"] = a.LLMProvider != nil
	status.Stats["tools_count"] = len(a.ToolRegistry.All())
	status.Stats["conversation_length"] = a.conversationLength()

	if a.Database != nil {
		status.Stats["database_connected"] = a.Database.IsConnected()
	}

	return status
}

// GetModules returns information about all active modules
func (a *Agent) GetModules() []string {
	modules := []string{"core"}

	if a.LLMProvider != nil {
		modules = append(modules, "llm")
	}
	if a.FileSystem != nil {
		modules = append(modules, "filesystem")
	}
	if a.Database != nil {
		modules = append(modules, "database")
	}
	if a.WebSearch != nil {
		modules = append(modules, "web")
	}
	if a.ToolRegistry != nil && len(a.ToolRegistry.All()) > 0 {
		modules = append(modules, "tools")
	}

	return modules
}

// conversationLength returns the length of the conversation (thread-safe).
func (a *Agent) conversationLength() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.conversation)
}
