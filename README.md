# AI Agent Module for Go

This is a modular AI agent implementation for Go applications, designed to provide a flexible and extensible foundation for building AI-powered applications.

## Architecture

The agent follows a modular architecture with the following key components:

### Core Package
- `core/agent.go`: Main agent structure and interfaces
- Defines the core agent functionality and interfaces

### LLM Package
- `llm/llm.go`: LLM provider interfaces and data structures
- `llm/openai.go`: OpenAI API implementation

### Filesystem Package
- `filesystem/filesystem.go`: File system interfaces
- `filesystem/local.go`: Local file system implementation

### Database Package
- `database/database.go`: Database interfaces and data structures
- `database/sqlite.go`: SQLite database implementation

### Web Package
- `web/web.go`: Web search and scraping interfaces
- `web/search.go`: Web search implementation

## Features

1. **Modular Design**: Each component is a separate package with clear interfaces
2. **SOLID Principles**: Follows SOLID design principles with interface-based design
3. **Extensible**: Easy to add new providers or implementations
4. **Configurable**: All components can be configured via configuration structures
5. **Testable**: Interfaces make it easy to mock components for testing

## Usage

### Basic Agent Creation

```go
import "ai-agent-go/agent"

// Create configuration
config := agent.AgentConfig{
    AgentConfig: core.AgentConfig{
        ID: "my-agent",
        Name: "My AI Agent",
        Enabled: true,
    },
    LLMConfig: llm.LLMConfig{
        APIKey: "your-openai-api-key",
        Model: "gpt-4",
    },
    FileSystemConfig: filesystem.FileSystemConfig{
        BasePath: "/path/to/data",
    },
    DatabaseConfig: database.DatabaseConfig{
        DSN: "file:mydb.sqlite?cache=shared",
    },
    WebConfig: web.WebConfig{
        Timeout: 30 * time.Second,
    },
}

// Create agent
agent, err := agent.NewAgent(config)
if err != nil {
    log.Fatal(err)
}

// Start agent
ctx := context.Background()
err = agent.Start(ctx)
if err != nil {
    log.Fatal(err)
}

// Execute a task
result, err := agent.Execute(ctx, "Hello, world!")
if err != nil {
    log.Fatal(err)
}

fmt.Println(result)
```

### Module Usage

Each module can be used independently:

```go
// LLM usage
llmProvider, _ := llm.NewOpenAIProvider(llmConfig)
response, err := llmProvider.ChatCompletion(ctx, chatRequest)

// File system usage
fs := filesystem.NewLocalFileSystem(fsConfig)
content, err := fs.ReadFile(ctx, "data.txt")

// Database usage
db, _ := database.NewSQLiteDB(dbConfig)
err = db.Connect(ctx)
rows, err := db.Query(ctx, "SELECT * FROM users")

// Web search usage
webSearch, _ := web.NewWebSearch(webConfig)
results, err := webSearch.Search(ctx, "Go programming", web.SearchOptions{})
```

## Configuration

The agent can be configured through:
- Configuration files (JSON/YAML)
- Environment variables
- Constructor parameters

## Extending the Agent

To extend the agent with new functionality:

1. Create a new package in the `agent/` directory
2. Define interfaces for your new functionality
3. Implement the interfaces with concrete types
4. Add the new module to the agent initialization
5. Update the agent's Execute method to use the new functionality

## Testing

All interfaces are designed to be easily mockable for testing purposes. You can create mock implementations for testing without requiring actual external services.

## License

MIT