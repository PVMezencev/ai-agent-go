package agent

import (
	"fmt"
	"time"

	"github.com/PVMezencev/ai-agent-go/agent/core"
	"github.com/PVMezencev/ai-agent-go/agent/database"
	"github.com/PVMezencev/ai-agent-go/agent/filesystem"
	"github.com/PVMezencev/ai-agent-go/agent/web"
)

// ExampleAgent demonstrates how to create and configure a complete AI agent
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
		FileSystemConfig: filesystem.FileSystemConfig{
			BasePath: "/tmp",
			Timeout:  10 * time.Second,
		},
		DatabaseConfig: database.DatabaseConfig{
			DSN:             "file:example.db?cache=shared",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
			Timeout:         10 * time.Second,
		},
		WebConfig: web.WebConfig{
			Timeout:      30 * time.Second,
			MaxRetries:   3,
			UserAgent:    "AI-Agent/1.0",
			AllowBlocked: false,
		},
		MaxToolRounds: 10,
	}

	// Create agent (without LLM API key — agent will be created but Execute requires LLM)
	a, err := NewAgent(config)
	if err != nil {
		fmt.Printf("Error creating agent: %v\n", err)
		return
	}

	modules := a.GetModules()
	fmt.Printf("Agent modules: %v\n", modules)

	status := a.GetStatus()
	fmt.Printf("Tools available: %v\n", status.Stats["tools_count"])

	// Output:
	// Agent modules: [core filesystem database web tools]
	// Tools available: 9
}
