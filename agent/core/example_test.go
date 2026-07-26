package core

import (
	"fmt"
	"time"
)

// ExampleAgent demonstrates how to create and use an Agent
func ExampleAgent() {
	// Create an agent with configuration
	agent := Agent{
		ID:        "example-agent",
		Name:      "Example AI Agent",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Print agent information
	fmt.Printf("Agent ID: %s\n", agent.ID)
	fmt.Printf("Agent Name: %s\n", agent.Name)
	fmt.Printf("Created at: %v\n", agent.CreatedAt)

	// Output:
	// Agent ID: example-agent
	// Agent Name: Example AI Agent
	// Created at: <current time>
}