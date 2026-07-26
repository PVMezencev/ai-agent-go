package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/PVMezencev/ai-agent-go/agent/core"
	"github.com/PVMezencev/ai-agent-go/agent/llm"
	"github.com/PVMezencev/ai-agent-go/agent/filesystem"
	"github.com/PVMezencev/ai-agent-go/agent/database"
	"github.com/PVMezencev/ai-agent-go/agent/web"
)

// Agent represents the main AI agent that coordinates all modules
type Agent struct {
	core.Agent
	core.AgentInterface

	// Module interfaces
	LLMProvider     llm.LLMProvider
	FileSystem      filesystem.FileSystemInterface
	Database        database.DatabaseInterface
	WebSearch       web.WebSearchInterface

	// Configuration
	config AgentConfig
}

// AgentConfig represents the configuration for the entire agent
type AgentConfig struct {
	core.AgentConfig
	LLMConfig     llm.LLMConfig
	FileSystemConfig filesystem.FileSystemConfig
	DatabaseConfig database.DatabaseConfig
	WebConfig     web.WebConfig
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

	return agent, nil
}

// initModules initializes all the agent modules
func (a *Agent) initModules() error {
	// Initialize LLM provider
	if a.config.LLMConfig.APIKey != "" {
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
	a.WebSearch = web.NewWebSearch(a.config.WebConfig)

	return nil
}

// Start starts the agent and all its modules
func (a *Agent) Start(ctx context.Context) error {
	// Start all modules
	if a.Database != nil {
		if err := a.Database.Connect(ctx); err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
	}

	a.UpdatedAt = time.Now()
	return nil
}

// Stop stops the agent and all its modules
func (a *Agent) Stop(ctx context.Context) error {
	// Stop all modules
	if a.Database != nil {
		if err := a.Database.Disconnect(ctx); err != nil {
			return fmt.Errorf("failed to disconnect from database: %w", err)
		}
	}

	a.UpdatedAt = time.Now()
	return nil
}

// Execute executes a task using the agent's capabilities
func (a *Agent) Execute(ctx context.Context, task string) (string, error) {
	// In a real implementation, this would:
	// 1. Parse the task
	// 2. Use appropriate modules to execute it
	// 3. Return results

	// For this example, we'll return a simulated response
	return fmt.Sprintf("Agent executed task: %s", task), nil
}

// GetConfig returns the agent configuration
func (a *Agent) GetConfig() core.AgentConfig {
	return a.config.AgentConfig
}

// SetConfig sets the agent configuration
func (a *Agent) SetConfig(config core.AgentConfig) error {
	a.config.AgentConfig = config
	a.UpdatedAt = time.Now()
	return nil
}

// GetStatus returns the agent status
func (a *Agent) GetStatus() core.AgentStatus {
	status := core.AgentStatus{
		IsRunning:  true,
		LastActive: a.UpdatedAt,
		Stats:      make(map[string]interface{}),
	}

	// Add module-specific stats
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

	return modules
}