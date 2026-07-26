package core

import (
	"fmt"
)

// ExampleAgent demonstrates how to create and use an Agent
func ExampleAgent() {
	// Create an agent with configuration
	agent := Agent{
		ID:        "example-agent",
		Name:      "Example AI Agent",
		CreatedAt: nil, // This would be set to time.Now() in real usage
		UpdatedAt: nil, // This would be set to time.Now() in real usage
	}

	// Print agent information
	fmt.Printf("Agent ID: %s\n", agent.ID)
	fmt.Printf("Agent Name: %s\n", agent.Name)

	// Output:
	// Agent ID: example-agent
	// Agent Name: Example AI Agent
}