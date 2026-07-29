package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/PVMezencev/ai-agent-go/agent"
	"github.com/PVMezencev/ai-agent-go/agent/core"
	"github.com/PVMezencev/ai-agent-go/agent/database"
	"github.com/PVMezencev/ai-agent-go/agent/filesystem"
	"github.com/PVMezencev/ai-agent-go/agent/llm"
	"github.com/PVMezencev/ai-agent-go/agent/tools"
	"github.com/PVMezencev/ai-agent-go/agent/web"
)

// CLIConfig contains configuration for the CLI agent
type CLIConfig struct {
	AgentConfig    agent.AgentConfig
	LLMConfig      llm.LLMConfig
	FileSystemConfig filesystem.FileSystemConfig
	DatabaseConfig database.DatabaseConfig
	WebConfig      web.WebConfig
	MaxRetries     int
	Timeout        time.Duration
	MaxLoopCount   int
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	task := strings.Join(os.Args[1:], " ")
	config := createDefaultConfig()

	// Initialize the agent
	agentInstance, err := agent.NewAgent(config.AgentConfig)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	// Start the agent
	ctx := context.Background()
	if err := agentInstance.Start(ctx); err != nil {
		log.Fatalf("Failed to start agent: %v", err)
	}
	defer func() {
		if err := agentInstance.Stop(ctx); err != nil {
			log.Printf("Warning: Failed to stop agent: %v", err)
		}
	}()

	// Execute with streaming and safety
	if err := executeWithStreaming(agentInstance, task, config); err != nil {
		log.Fatalf("Task execution failed: %v", err)
	}
}

func printUsage() {
	fmt.Println("Usage: ai-agent <task>")
	fmt.Println()
	fmt.Println("AI Agent CLI — execute tasks with LLM-powered reasoning and tools.")
	fmt.Println()
	fmt.Println("Available tools:")
	fmt.Println("  - read_file     Read file contents")
	fmt.Println("  - write_file    Write content to a file")
	fmt.Println("  - list_files    List directory contents")
	fmt.Println("  - delete_file   Delete a file (requires confirmation)")
	fmt.Println("  - create_directory  Create a new directory")
	fmt.Println("  - query_sql     Execute SQL SELECT queries")
	fmt.Println("  - exec_sql      Execute SQL statements (requires confirmation)")
	fmt.Println("  - web_search    Search the web")
	fmt.Println("  - scrape_url    Scrape a web page")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  OPENAI_API_KEY  Your OpenAI API key (required)")
	fmt.Println("  OPENAI_MODEL    Model to use (default: gpt-4o)")
	fmt.Println("  OPENAI_ENDPOINT Custom API endpoint (optional)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  ai-agent \"Summarize the contents of README.md\"")
	fmt.Println("  ai-agent \"Search for the latest Go best practices\"")
	fmt.Println("  ai-agent \"Create a SQL query to find active users\"")
}

func createDefaultConfig() CLIConfig {
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-4o"
	}

	return CLIConfig{
		AgentConfig: agent.AgentConfig{
			AgentConfig: core.AgentConfig{
				ID:          "cli-agent-" + time.Now().Format("20060102150405"),
				Name:        "CLI AI Agent",
				Enabled:     true,
				MaxRetries:  3,
				Timeout:     120 * time.Second,
			},
			MaxToolRounds: 10,
		},
		LLMConfig: llm.LLMConfig{
			APIKey:      os.Getenv("OPENAI_API_KEY"),
			APIEndpoint: os.Getenv("OPENAI_ENDPOINT"),
			Model:       model,
			Timeout:     120 * time.Second,
			MaxRetries:  3,
		},
		FileSystemConfig: filesystem.FileSystemConfig{
			BasePath: ".",
			Timeout:  10 * time.Second,
		},
		DatabaseConfig: database.DatabaseConfig{
			DSN:             "file:cli-agent.db?cache=shared",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
			Timeout:         10 * time.Second,
		},
		WebConfig: web.WebConfig{
			Timeout:      30 * time.Second,
			MaxRetries:   3,
			UserAgent:    "AI-Agent-CLI/1.0",
			AllowBlocked: false,
		},
		MaxRetries:   3,
		Timeout:      120 * time.Second,
		MaxLoopCount: 10,
	}
}

func executeWithStreaming(a *agent.Agent, task string, config CLIConfig) error {
	// Build a safety-aware tool registry that wraps destructive tools
	safeRegistry := wrapToolsWithSafety(a.ToolRegistry)
	a.ToolRegistry = safeRegistry

	loopCount := 0
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		loopCount++
		if loopCount > config.MaxLoopCount {
			return fmt.Errorf("maximum loop count (%d) exceeded — possible infinite loop", config.MaxLoopCount)
		}

		fmt.Print("\rThinking...")

		// Stream the result
		err := a.ExecuteStream(context.Background(), task, func(chunk llm.ChatStreamChunk) error {
			// Print each chunk immediately
			if chunk.Content != "" {
				fmt.Print(chunk.Content)
			}
			return nil
		})

		if err != nil {
			lastErr = err
			fmt.Printf("\n\rAttempt %d failed: %v. Retrying...\n", attempt+1, err)
			time.Sleep(time.Duration(attempt+1) * time.Second)
			continue
		} else {
			fmt.Println() // Newline after streaming
			return nil
		}
	}

	fmt.Println()
	return fmt.Errorf("task failed after %d attempts: %v", config.MaxRetries+1, lastErr)
}

// wrapToolsWithSafety wraps the tool registry to add confirmation for destructive operations
func wrapToolsWithSafety(original *tools.Registry) *tools.Registry {
	safe := tools.NewRegistry()

	destructiveTools := map[string]bool{
		"delete_file": true,
		"exec_sql":    true,
		"write_file":  true,
	}

	for _, t := range original.All() {
		if destructiveTools[t.Name()] {
			// Wrap with confirmation
			originalTool := t
			wrapped := tools.NewToolFunc(
				originalTool.Name(),
				originalTool.Description()+ " Requires user confirmation before execution.",
				originalTool.Parameters(),
				func(ctx context.Context, argsJSON string) (string, error) {
					// Prompt for confirmation
					fmt.Printf("\n[SAFETY] Tool '%s' is about to be executed. This is a potentially destructive operation.\n", originalTool.Name())
					fmt.Print("Confirm? (y/N): ")

					scanner := bufio.NewScanner(os.Stdin)
					if scanner.Scan() {
						response := strings.ToLower(strings.TrimSpace(scanner.Text()))
						if response != "y" && response != "yes" {
							return "", fmt.Errorf("operation cancelled by user")
						}
					}
					fmt.Println()

					return originalTool.Execute(ctx, argsJSON)
				},
			)
			safe.Register(wrapped)
		} else {
			safe.Register(t)
		}
	}

	return safe
}
