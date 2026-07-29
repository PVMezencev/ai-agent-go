package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/PVMezencev/ai-agent-go/agent/database"
	"github.com/PVMezencev/ai-agent-go/agent/filesystem"
	"github.com/PVMezencev/ai-agent-go/agent/web"
)

// BuildFileSystemTools creates tools for the filesystem module
func BuildFileSystemTools(fs filesystem.FileSystemInterface) []Tool {
	if fs == nil {
		return nil
	}

	return []Tool{
		NewToolFunc(
			"read_file",
			"Read the contents of a file. Returns the file content as text.",
			map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the file to read (relative to base path)",
					},
				},
				"required": []string{"path"},
			},
			func(ctx context.Context, argsJSON string) (string, error) {
				args, err := ParseToolArgs(argsJSON)
				if err != nil {
					return "", err
				}
				path, ok := args["path"].(string)
				if !ok || path == "" {
					return "", fmt.Errorf("path is required")
				}
				data, err := fs.ReadFile(ctx, path)
				if err != nil {
					return "", fmt.Errorf("failed to read file %s: %w", path, err)
				}
				return string(data), nil
			},
		),
		NewToolFunc(
			"write_file",
			"Write content to a file. Creates the file or overwrites it if it exists.",
			map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the file (relative to base path)",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Content to write to the file",
					},
				},
				"required": []string{"path", "content"},
			},
			func(ctx context.Context, argsJSON string) (string, error) {
				args, err := ParseToolArgs(argsJSON)
				if err != nil {
					return "", err
				}
				path, ok := args["path"].(string)
				if !ok || path == "" {
					return "", fmt.Errorf("path is required")
				}
				content, ok := args["content"].(string)
				if !ok {
					return "", fmt.Errorf("content is required")
				}
				if err := fs.WriteFile(ctx, path, []byte(content), 0644); err != nil {
					return "", fmt.Errorf("failed to write file %s: %w", path, err)
				}
				return fmt.Sprintf("File %s written successfully (%d bytes)", path, len(content)), nil
			},
		),
		NewToolFunc(
			"list_files",
			"List files and directories in a given path.",
			map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Directory path to list (relative to base path, empty for root)",
					},
				},
				"required": []string{"path"},
			},
			func(ctx context.Context, argsJSON string) (string, error) {
				args, err := ParseToolArgs(argsJSON)
				if err != nil {
					return "", err
				}
				path, _ := args["path"].(string)
				files, err := fs.ListFiles(ctx, path)
				if err != nil {
					return "", fmt.Errorf("failed to list files in %s: %w", path, err)
				}
				if len(files) == 0 {
					return "Directory is empty", nil
				}
				var lines []string
				for _, f := range files {
					prefix := "-"
					if f.IsDir {
						prefix = "d"
					}
					lines = append(lines, fmt.Sprintf("%s %s (%d bytes, modified: %s)", prefix, f.Name, f.Size, f.ModTime.Format("2006-01-02 15:04:05")))
				}
				return strings.Join(lines, "\n"), nil
			},
		),
		NewToolFunc(
			"delete_file",
			"Delete a file from the file system. This operation cannot be undone.",
			map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the file to delete (relative to base path)",
					},
				},
				"required": []string{"path"},
			},
			func(ctx context.Context, argsJSON string) (string, error) {
				args, err := ParseToolArgs(argsJSON)
				if err != nil {
					return "", err
				}
				path, ok := args["path"].(string)
				if !ok || path == "" {
					return "", fmt.Errorf("path is required")
				}
				if err := fs.DeleteFile(ctx, path); err != nil {
					return "", fmt.Errorf("failed to delete file %s: %w", path, err)
				}
				return fmt.Sprintf("File %s deleted successfully", path), nil
			},
		),
		NewToolFunc(
			"create_directory",
			"Create a new directory (and any parent directories as needed).",
			map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the directory to create (relative to base path)",
					},
				},
				"required": []string{"path"},
			},
			func(ctx context.Context, argsJSON string) (string, error) {
				args, err := ParseToolArgs(argsJSON)
				if err != nil {
					return "", err
				}
				path, ok := args["path"].(string)
				if !ok || path == "" {
					return "", fmt.Errorf("path is required")
				}
				if err := fs.CreateDir(ctx, path, 0755); err != nil {
					return "", fmt.Errorf("failed to create directory %s: %w", path, err)
				}
				return fmt.Sprintf("Directory %s created successfully", path), nil
			},
		),
	}
}

// BuildDatabaseTools creates tools for the database module
func BuildDatabaseTools(db database.DatabaseInterface) []Tool {
	if db == nil {
		return nil
	}

	return []Tool{
		NewToolFunc(
			"query_sql",
			"Execute a SQL query on the database. Use for SELECT statements to read data.",
			map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "SQL query to execute (e.g., SELECT * FROM users WHERE active = 1)",
					},
				},
				"required": []string{"query"},
			},
			func(ctx context.Context, argsJSON string) (string, error) {
				args, err := ParseToolArgs(argsJSON)
				if err != nil {
					return "", err
				}
				query, ok := args["query"].(string)
				if !ok || query == "" {
					return "", fmt.Errorf("query is required")
				}
				rows, err := db.Query(ctx, query)
				if err != nil {
					return "", fmt.Errorf("query failed: %w", err)
				}
				defer rows.Close()

				columns, err := rows.Columns()
				if err != nil {
					return "", fmt.Errorf("failed to get columns: %w", err)
				}

				var results []string
				results = append(results, strings.Join(columns, " | "))
				for rows.Next() {
					values := make([]interface{}, len(columns))
					valuePtrs := make([]interface{}, len(columns))
					for i := range values {
						valuePtrs[i] = &values[i]
					}
					if err := rows.Scan(valuePtrs...); err != nil {
						return "", fmt.Errorf("scan error: %w", err)
					}
					strValues := make([]string, len(values))
					for i, v := range values {
						if v == nil {
							strValues[i] = "NULL"
						} else {
							strValues[i] = fmt.Sprintf("%v", v)
						}
					}
					results = append(results, strings.Join(strValues, " | "))
				}
				if len(results) == 1 {
					return "Query returned 0 rows", nil
				}
				return strings.Join(results, "\n"), nil
			},
		),
		NewToolFunc(
			"exec_sql",
			"Execute a SQL statement that modifies data (INSERT, UPDATE, DELETE, CREATE). Use with caution.",
			map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "SQL statement to execute (INSERT, UPDATE, DELETE, CREATE TABLE, etc.)",
					},
				},
				"required": []string{"query"},
			},
			func(ctx context.Context, argsJSON string) (string, error) {
				args, err := ParseToolArgs(argsJSON)
				if err != nil {
					return "", err
				}
				query, ok := args["query"].(string)
				if !ok || query == "" {
					return "", fmt.Errorf("query is required")
				}
				result, err := db.Exec(ctx, query)
				if err != nil {
					return "", fmt.Errorf("exec failed: %w", err)
				}
				rowsAffected, _ := result.RowsAffected()
				return fmt.Sprintf("Statement executed successfully. Rows affected: %d", rowsAffected), nil
			},
		),
	}
}

// BuildWebTools creates tools for the web module
func BuildWebTools(ws web.WebSearchInterface) []Tool {
	if ws == nil {
		return nil
	}

	return []Tool{
		NewToolFunc(
			"web_search",
			"Search the web for information. Returns search results with titles, URLs, and snippets.",
			map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query string",
					},
					"max_results": map[string]interface{}{
						"type":        "number",
						"description": "Maximum number of results to return (default: 5)",
					},
				},
				"required": []string{"query"},
			},
			func(ctx context.Context, argsJSON string) (string, error) {
				args, err := ParseToolArgs(argsJSON)
				if err != nil {
					return "", err
				}
				query, ok := args["query"].(string)
				if !ok || query == "" {
					return "", fmt.Errorf("query is required")
				}
				maxResults := 5
				if mr, ok := args["max_results"]; ok {
					if f, ok := mr.(float64); ok {
						maxResults = int(f)
					}
				}

				results, err := ws.Search(ctx, query, web.SearchOptions{
					MaxResults: maxResults,
				})
				if err != nil {
					return "", fmt.Errorf("search failed: %w", err)
				}

				var lines []string
				for i, r := range results.Results {
					lines = append(lines, fmt.Sprintf("%d. %s\n   URL: %s\n   %s", i+1, r.Title, r.URL, r.Description))
				}
				if len(lines) == 0 {
					return "No search results found", nil
				}
				return strings.Join(lines, "\n\n"), nil
			},
		),
		NewToolFunc(
			"scrape_url",
			"Scrape a web page and return its content as text.",
			map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{
					"url": map[string]interface{}{
						"type":        "string",
						"description": "URL of the page to scrape",
					},
				},
				"required": []string{"url"},
			},
			func(ctx context.Context, argsJSON string) (string, error) {
				args, err := ParseToolArgs(argsJSON)
				if err != nil {
					return "", err
				}
				url, ok := args["url"].(string)
				if !ok || url == "" {
					return "", fmt.Errorf("url is required")
				}
				result, err := ws.Scrape(ctx, url, web.ScrapeOptions{})
				if err != nil {
					return "", fmt.Errorf("scrape failed: %w", err)
				}
				return fmt.Sprintf("Title: %s\n\n%s", result.Title, result.Content), nil
			},
		),
	}
}

// BuildAllTools creates tools from all available modules
func BuildAllTools(fs filesystem.FileSystemInterface, db database.DatabaseInterface, ws web.WebSearchInterface) *Registry {
	reg := NewRegistry()

	for _, t := range BuildFileSystemTools(fs) {
		reg.Register(t)
	}
	for _, t := range BuildDatabaseTools(db) {
		reg.Register(t)
	}
	for _, t := range BuildWebTools(ws) {
		reg.Register(t)
	}

	return reg
}
