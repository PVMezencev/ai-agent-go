package tools_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/PVMezencev/ai-agent-go/agent/filesystem"
	"github.com/PVMezencev/ai-agent-go/agent/llm"
	"github.com/PVMezencev/ai-agent-go/agent/tools"
	"github.com/PVMezencev/ai-agent-go/agent/web"

	"github.com/stretchr/testify/assert"
)

func TestToolFunc(t *testing.T) {
	fn := tools.NewToolFunc(
		"test_tool",
		"A test tool",
		map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{"x": map[string]interface{}{"type": "number"}},
		},
		func(ctx context.Context, argsJSON string) (string, error) {
			return "ok", nil
		},
	)

	assert.Equal(t, "test_tool", fn.Name())
	assert.Equal(t, "A test tool", fn.Description())

	result, err := fn.Execute(context.Background(), `{"x": 1}`)
	assert.NoError(t, err)
	assert.Equal(t, "ok", result)

	toolDef := fn.ToLLMTool()
	assert.Equal(t, "function", toolDef.Type)
	assert.Equal(t, "test_tool", toolDef.Function.Name)
}

func TestRegistry(t *testing.T) {
	reg := tools.NewRegistry()

	t1 := tools.NewToolFunc("tool1", "desc1", nil, func(ctx context.Context, argsJSON string) (string, error) {
		return "r1", nil
	})
	t2 := tools.NewToolFunc("tool2", "desc2", nil, func(ctx context.Context, argsJSON string) (string, error) {
		return "r2", nil
	})

	reg.Register(t1)
	reg.Register(t2)

	assert.Len(t, reg.All(), 2)

	got, ok := reg.Get("tool1")
	assert.True(t, ok)
	assert.Equal(t, "tool1", got.Name())

	_, ok = reg.Get("nonexistent")
	assert.False(t, ok)

	defs := reg.ToLLMTools()
	assert.Len(t, defs, 2)
}

func TestParseToolArgs(t *testing.T) {
	args, err := tools.ParseToolArgs(`{"name": "Alice", "age": 30}`)
	assert.NoError(t, err)
	assert.Equal(t, "Alice", args["name"])
	assert.Equal(t, float64(30), args["age"])

	_, err = tools.ParseToolArgs(`invalid json`)
	assert.Error(t, err)
}

func TestToolResult(t *testing.T) {
	assert.Equal(t, "success", tools.ToolResult("success", nil))
	assert.Contains(t, tools.ToolResult("", assert.AnError), "Error:")
}

func TestBuildFileSystemTools_Nil(t *testing.T) {
	result := tools.BuildFileSystemTools(nil)
	assert.Nil(t, result)
}

func TestBuildDatabaseTools_Nil(t *testing.T) {
	result := tools.BuildDatabaseTools(nil)
	assert.Nil(t, result)
}

func TestBuildWebTools_Nil(t *testing.T) {
	result := tools.BuildWebTools(nil)
	assert.Nil(t, result)
}

func TestBuildAllTools(t *testing.T) {
	reg := tools.BuildAllTools(
		filesystem.NewLocalFileSystem(filesystem.FileSystemConfig{BasePath: "/tmp"}),
		nil, // no database
		func() web.WebSearchInterface {
			w, _ := web.NewWebSearch(web.WebConfig{Timeout: 5 * time.Second})
			return w
		}(),
	)

	all := reg.All()
	// filesystem tools + web tools, no database tools
	assert.GreaterOrEqual(t, len(all), 5) // at least 5 fs tools + 2 web tools would be 7, but db is nil
}

func TestRegistry_ExecuteToolCall(t *testing.T) {
	reg := tools.NewRegistry()
	reg.Register(tools.NewToolFunc(
		"add",
		"Add two numbers",
		map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{
				"a": map[string]interface{}{"type": "number"},
				"b": map[string]interface{}{"type": "number"},
			},
		},
		func(ctx context.Context, argsJSON string) (string, error) {
			args, err := tools.ParseToolArgs(argsJSON)
			if err != nil {
				return "", err
			}
			a := args["a"].(float64)
			b := args["b"].(float64)
			result := fmt.Sprintf("%d", int(a+b))
			return result, nil
		},
	))

	result, err := reg.ExecuteToolCall(context.Background(), llm.ToolCall{
		ID:   "call_1",
		Type: "function",
		FunctionCall: llm.FunctionCall{
			Name:      "add",
			Arguments: `{"a": 3, "b": 4}`,
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, "7", result)
}
