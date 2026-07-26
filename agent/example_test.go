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

// ExampleAgent demonstrates how to create and use a complete AI agent
func ExampleAgent() {
	// Create configuration for the agent
	config := AgentConfig{
		AgentConfig: core.AgentConfig{
			ID:          "example-agent",
			Name:        "Example AI Agent",
			Enabled:     true,
			MaxRetries:  3,
			Timeout:     30 * time.Second,
		},
		LLMConfig: llm.LLMConfig{
			APIKey:      "your-openai-api-key",
			Model:       "gpt-4",
			Timeout:     30 * time.Second,
			MaxRetries:  3,
		},
		FileSystemConfig: filesystem.FileSystemConfig{
			BasePath: "/tmp",
			Timeout:  10 * time.Second,
		},
		DatabaseConfig: database.DatabaseConfig{
			DSN:           "file:example.db?cache=shared",
			MaxOpenConns:  10,
			MaxIdleConns:  5,
			ConnMaxLifetime: 5 * time.Minute,
			Timeout:       10 * time.Second,
		},
		WebConfig: web.WebConfig{
			Timeout:      30 * time.Second,
			MaxRetries:   3,
			UserAgent:    "AI-Agent/1.0",
			AllowBlocked: false,
		},
	}

	// Create agent
	agent, err := NewAgent(config)
	if err != nil {
		fmt.Printf("Error creating agent: %v\n", err)
		return
	}

	// Start agent
	ctx := context.Background()
	err = agent.Start(ctx)
	if err != nil {
		fmt.Printf("Error starting agent: %v\n", err)
		return
	}
	defer func() {
		agent.Stop(ctx)
	}()

	// Execute a task
	result, err := agent.Execute(ctx, "Hello, world!")
	if err != nil {
		fmt.Printf("Error executing task: %v\n", err)
		return
	}

	fmt.Printf("Task result: %s\n", result)

	// Output:
	// Task result: Agent executed task: Hello, world!
}