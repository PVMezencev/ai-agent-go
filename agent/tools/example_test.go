package tools_test

import (
	"context"
	"fmt"

	"github.com/PVMezencev/ai-agent-go/agent/tools"
)

func ExampleRegistry() {
	reg := tools.NewRegistry()

	reg.Register(tools.NewToolFunc(
		"greet",
		"Greet a person by name",
		map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the person to greet",
				},
			},
			"required": []string{"name"},
		},
		func(ctx context.Context, argsJSON string) (string, error) {
			return "Hello!", nil
		},
	))

	fmt.Printf("Registered %d tool(s)\n", len(reg.All()))
	// Output: Registered 1 tool(s)
}
