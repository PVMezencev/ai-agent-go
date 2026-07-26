package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/PVMezencev/ai-agent-go/agent"
	"github.com/PVMezencev/ai-agent-go/agent/core"
	"github.com/PVMezencev/ai-agent-go/agent/llm"
	"github.com/PVMezencev/ai-agent-go/agent/filesystem"
	"github.com/PVMezencev/ai-agent-go/agent/database"
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
	MaxLoopCount   int // Maximum number of loop iterations to prevent infinite loops
}

func main() {
	// Parse command line arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: ai-agent <task>")
		fmt.Println("Example: ai-agent \"Write a report about Go programming\"")
		fmt.Println("")
		fmt.Println("Features:")
		fmt.Println("- Automatic error handling and retry logic")
		fmt.Println("- User confirmation for potentially dangerous actions")
		fmt.Println("- LLM-based result correction when needed")
		fmt.Println("- Loop prevention to avoid infinite execution")
		return
	}

	task := strings.Join(os.Args[1:], " ")

	// Create default configuration
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

	// Execute the task with safety checks
	result, err := executeTaskWithSafety(agentInstance, task, config)
	if err != nil {
		log.Fatalf("Task execution failed: %v", err)
	}

	fmt.Println("Result:")
	fmt.Println(result)
}

// createDefaultConfig creates a default configuration for the CLI agent
func createDefaultConfig() CLIConfig {
	return CLIConfig{
		AgentConfig: agent.AgentConfig{
			AgentConfig: core.AgentConfig{
				ID:          "cli-agent-" + time.Now().Format("20060102150405"),
				Name:        "CLI AI Agent",
				Enabled:     true,
				MaxRetries:  3,
				Timeout:     30 * time.Second,
			},
		},
		LLMConfig: llm.LLMConfig{
			APIKey:      os.Getenv("OPENAI_API_KEY"),
			Model:       "gpt-4",
			Timeout:     30 * time.Second,
			MaxRetries:  3,
		},
		FileSystemConfig: filesystem.FileSystemConfig{
			BasePath: "/tmp",
			Timeout:  10 * time.Second,
		},
		DatabaseConfig: database.DatabaseConfig{
			DSN:           "file:cli-agent.db?cache=shared",
			MaxOpenConns:  10,
			MaxIdleConns:  5,
			ConnMaxLifetime: 5 * time.Minute,
			Timeout:       10 * time.Second,
		},
		WebConfig: web.WebConfig{
			Timeout:      30 * time.Second,
			MaxRetries:   3,
			UserAgent:    "AI-Agent-CLI/1.0",
			AllowBlocked: false,
		},
		MaxRetries:   3,
		Timeout:      30 * time.Second,
		MaxLoopCount: 10, // Prevent more than 10 execution loops
	}
}

// executeTaskWithSafety executes a task with safety checks and error handling
func executeTaskWithSafety(a *agent.Agent, task string, config CLIConfig) (string, error) {
	ctx := context.Background()

	// Loop counter to prevent infinite execution
	loopCount := 0

	// Safety check: Verify that the task is not potentially dangerous
	if isDangerousTask(task) {
		fmt.Printf("Warning: The task '%s' might be dangerous.\n", task)
		fmt.Print("Do you want to proceed? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			return "", fmt.Errorf("task execution cancelled by user")
		}
	}

	// Main execution loop with safety checks
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Prevent infinite loops
		loopCount++
		if loopCount > config.MaxLoopCount {
			return "", fmt.Errorf("maximum loop count (%d) exceeded - possible infinite loop detected", config.MaxLoopCount)
		}

		fmt.Printf("Executing task (attempt %d/%d, loop %d/%d): %s\n",
			attempt+1, config.MaxRetries+1, loopCount, config.MaxLoopCount, task)

		result, err := a.Execute(ctx, task)
		if err == nil {
			// Check if the result needs correction by LLM
			correctedResult, correctionErr := correctResultWithLLM(a, task, result, config)
			if correctionErr == nil {
				return correctedResult, nil
			} else {
				fmt.Printf("Warning: Failed to correct result with LLM: %v\n", correctionErr)
				return result, nil // Return original result if correction fails
			}
		}

		if attempt < config.MaxRetries {
			fmt.Printf("Attempt %d failed: %v. Retrying...\n", attempt+1, err)
			time.Sleep(time.Duration(attempt+1) * time.Second) // Exponential backoff
		}
	}

	return "", fmt.Errorf("task failed after %d attempts", config.MaxRetries+1)
}

// isDangerousTask checks if a task might be dangerous
func isDangerousTask(task string) bool {
	dangerousKeywords := []string{
		"delete", "remove", "format", "erase", "remove",
		"system", "shell", "exec", "run", "command",
		"download", "install", "uninstall", "modify", "change",
		"write", "save", "create", "update", "destroy",
		"grant", "revoke", "delete", "remove",
	}

	taskLower := strings.ToLower(task)
	for _, keyword := range dangerousKeywords {
		if strings.Contains(taskLower, keyword) {
			return true
		}
	}
	return false
}

// correctResultWithLLM uses LLM to correct or validate the result
func correctResultWithLLM(a *agent.Agent, task, result string, config CLIConfig) (string, error) {
	// Check if we have an LLM provider available
	if a.LLMProvider == nil {
		return result, nil // No LLM available, return original result
	}

	// Create correction prompt
	correctionPrompt := fmt.Sprintf(`
You are an AI assistant reviewing the output of another AI agent.
Please review the following task and result and provide corrections if needed:

Task: %s
Result: %s

If the result is correct and complete, return it as is.
If there are issues, suggest improvements or corrections.

Respond only with the corrected result.
`, task, result)

	// Send correction request to LLM
	ctx := context.Background()
	request := llm.ChatRequest{
		Model:    config.LLMConfig.Model,
		Messages: []llm.Message{
			{
				Role:    "user",
				Content: correctionPrompt,
			},
		},
		Timeout: config.Timeout,
	}

	response, err := a.LLMProvider.ChatCompletion(ctx, request)
	if err != nil {
		return result, fmt.Errorf("LLM correction failed: %v", err)
	}

	// Extract the corrected content
	if len(response.Choices) > 0 {
		return response.Choices[0].Message.Content, nil
	}

	return result, nil
}