package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/PVMezencev/ai-agent-go/agent/llm"
)

// Tool represents a callable tool that the agent can use
type Tool interface {
	// Name returns the unique tool name
	Name() string

	// Description returns a human-readable description for the LLM
	Description() string

	// Parameters returns the JSON schema for the tool's input parameters
	Parameters() map[string]interface{}

	// Execute runs the tool with the given arguments (JSON string) and returns the result
	Execute(ctx context.Context, argsJSON string) (string, error)

	// ToLLMTool converts this tool to an LLM ToolDef for function calling
	ToLLMTool() llm.ToolDef
}

// ToolFunc is a convenience type for implementing a simple tool
type ToolFunc struct {
	name        string
	description string
	parameters  map[string]interface{}
	fn          func(ctx context.Context, argsJSON string) (string, error)
}

// NewToolFunc creates a new tool from a function
func NewToolFunc(name, description string, parameters map[string]interface{}, fn func(ctx context.Context, argsJSON string) (string, error)) *ToolFunc {
	return &ToolFunc{
		name:        name,
		description: description,
		parameters:  parameters,
		fn:          fn,
	}
}

// Name returns the tool name
func (f *ToolFunc) Name() string { return f.name }

// Description returns the tool description
func (f *ToolFunc) Description() string { return f.description }

// Parameters returns the tool parameters schema
func (f *ToolFunc) Parameters() map[string]interface{} { return f.parameters }

// Execute runs the tool
func (f *ToolFunc) Execute(ctx context.Context, argsJSON string) (string, error) {
	return f.fn(ctx, argsJSON)
}

// ToLLMTool converts to an LLM ToolDef
func (f *ToolFunc) ToLLMTool() llm.ToolDef {
	return llm.ToolDef{
		Type: "function",
		Function: llm.FunctionDef{
			Name:        f.name,
			Description: f.description,
			Parameters:  f.parameters,
		},
	}
}

// Registry holds a set of tools and provides lookup by name
type Registry struct {
	tools      map[string]Tool
	toolsSlice []Tool // preserve insertion order
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register adds a tool to the registry
func (r *Registry) Register(t Tool) {
	if _, exists := r.tools[t.Name()]; exists {
		// Silently overwrite — last registration wins
	}
	r.tools[t.Name()] = t
	r.toolsSlice = append(r.toolsSlice, t)
}

// Get returns a tool by name
func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// All returns all registered tools
func (r *Registry) All() []Tool {
	result := make([]Tool, len(r.toolsSlice))
	copy(result, r.toolsSlice)
	return result
}

// ToLLMTools converts all tools to LLM ToolDefs
func (r *Registry) ToLLMTools() []llm.ToolDef {
	defs := make([]llm.ToolDef, len(r.toolsSlice))
	for i, t := range r.toolsSlice {
		defs[i] = t.ToLLMTool()
	}
	return defs
}

// ExecuteToolCall resolves and executes a tool call from the LLM
func (r *Registry) ExecuteToolCall(ctx context.Context, tc llm.ToolCall) (string, error) {
	t, ok := r.Get(tc.FunctionCall.Name)
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", tc.FunctionCall.Name)
	}
	result, err := t.Execute(ctx, tc.FunctionCall.Arguments)
	if err != nil {
		return "", fmt.Errorf("tool %s failed: %w", tc.FunctionCall.Name, err)
	}
	return result, nil
}

// ParseToolArgs parses the JSON arguments string into a map
func ParseToolArgs(argsJSON string) (map[string]interface{}, error) {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return nil, fmt.Errorf("invalid tool arguments: %w", err)
	}
	return args, nil
}

// ToolResult formats a tool result or error as a string for the LLM
func ToolResult(result interface{}, err error) string {
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	switch v := result.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		data, marshalErr := json.Marshal(v)
		if marshalErr != nil {
			return fmt.Sprintf("%v", result)
		}
		return string(data)
	}
}
