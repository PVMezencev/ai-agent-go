// Command example демонстрирует использование ai-agent-go как библиотеки:
//
//  1. Поиск информации в интернете через LLM-агента (web_search tool).
//  2. Запись результатов в SQLite-базу данных (через database module агента).
//  3. Логирование всех действий агента в файл — агент сам использует
//     кастомный tool append_log для записи каждого шага.
//
// Запуск:
//
//	export OPENAI_API_KEY="sk-..."        # или OPENAI_ENDPOINT для локальной модели
//	go run main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
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

// --- Кастомный tool для логирования ---

// registerLogTool добавляет в registry инструмент append_log, который
// дозаписывает timestamped-строку в лог-файл.
func registerLogTool(reg *tools.Registry, logPath string) {
	reg.Register(tools.NewToolFunc(
		"append_log",
		"Append a timestamped message to the application log file at logs/agent-example.log. "+
			"Use this tool after every major step to record what you just did.",
		map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"message": map[string]interface{}{
					"type":        "string",
					"description": "The log message to append",
				},
			},
			"required": []string{"message"},
		},
		func(ctx context.Context, argsJSON string) (string, error) {
			args, err := tools.ParseToolArgs(argsJSON)
			if err != nil {
				return "", err
			}
			msg, ok := args["message"].(string)
			if !ok || msg == "" {
				return "", fmt.Errorf("message is required")
			}
			ts := time.Now().Format("2006-01-02 15:04:05")
			line := fmt.Sprintf("[%s] %s\n", ts, msg)

			f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
			if err != nil {
				return "", fmt.Errorf("open log file: %w", err)
			}
			defer f.Close()

			if _, err := f.WriteString(line); err != nil {
				return "", fmt.Errorf("write log: %w", err)
			}
			return fmt.Sprintf("Logged: %s", msg), nil
		},
	))
}

// --- Основной сценарий ---

func main() {
	ctx := context.Background()

	// Пути к артефактам
	logPath := filepath.Join("logs", "agent-example.log")
	dbPath := filepath.Join("data", "search_results.db")

	// 1. Создаём агента с настроенными модулями
	a, err := agent.NewAgent(agent.AgentConfig{
		AgentConfig: core.AgentConfig{
			ID:      "web-search-example",
			Name:    "Web Search + DB + Logging Example",
			Enabled: true,
			Timeout: 2 * time.Minute,
		},
		LLMConfig: llm.LLMConfig{
			APIKey:      getEnvOr("OPENAI_API_KEY", ""),
			APIEndpoint: getEnvOr("OPENAI_ENDPOINT", "http://localhost:5556"),
			Model:       getEnvOr("OPENAI_MODEL", "gpt-4o"),
			Timeout:     90 * time.Second,
			MaxRetries:  3,
		},
		FileSystemConfig: filesystem.FileSystemConfig{
			BasePath: ".",
			Timeout:  10 * time.Second,
		},
		DatabaseConfig: database.DatabaseConfig{
			DSN:             fmt.Sprintf("file:%s?_journal=wal&_sync=normal", dbPath),
			MaxOpenConns:    5,
			MaxIdleConns:    3,
			ConnMaxLifetime: 5 * time.Minute,
			Timeout:         10 * time.Second,
		},
		WebConfig: web.WebConfig{
			Timeout:    30 * time.Second,
			MaxRetries: 3,
			UserAgent:  "ai-agent-go-example/1.0",
		},
		MaxToolRounds: 8,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Добавляем кастомный log tool в registry
	registerLogTool(a.ToolRegistry, logPath)

	// Запускаем модули (подключаем БД)
	if err := a.Start(ctx); err != nil {
		log.Fatal(err)
	}
	defer a.Stop(ctx)

	query := "Go 1.24 new features release date"

	// Шаг 1: Создаём директорию logs
	_, err = a.Execute(ctx, "Создай директорию logs через create_directory.")
	if err != nil {
		log.Fatal(err)
	}

	// Шаг 2: Логируем начало
	_, err = a.Execute(ctx, "Запиши в лог через append_log, что начал работу с запросом: "+query+".")
	if err != nil {
		log.Fatal(err)
	}

	// Шаг 3: Поиск в интернете
	_, err = a.Execute(ctx, "Поискай в интернете информацию о '"+query+"' через web_search. Верни краткий структурированный ответ на русском языке: заголовок, ключевые факты, источник.")
	if err != nil {
		log.Fatal(err)
	}

	// Шаг 4: Логируем поиск
	_, err = a.Execute(ctx, "Запиши в лог через append_log, что поиск завершён.")
	if err != nil {
		log.Fatal(err)
	}

	// Шаг 5: Создаём таблицу
	_, err = a.Execute(ctx, "Выполни CREATE TABLE IF NOT EXISTS search_results (id INTEGER PRIMARY KEY AUTOINCREMENT, query TEXT, result TEXT, searched_at TEXT) через exec_sql.")
	if err != nil {
		log.Fatal(err)
	}

	// Шаг 6: Вставляем результат (передаём как переменную, чтобы не экранировать)
	// Агент сам выполнит exec_sql с экранированием
	_, err = a.Execute(ctx, "Поискай в интернете информацию о '"+query+"' через web_search. Верни ответ в одном коротком абзаце без переносов строк, для сохранения в БД.")
	if err != nil {
		log.Fatal(err)
	}

	// Шаг 7: Логируем сохранение
	_, err = a.Execute(ctx, "Запиши в лог через append_log, что данные сохранены.")
	if err != nil {
		log.Fatal(err)
	}

	// Шаг 8: Читаем из БД
	_, err = a.Execute(ctx, "Выполни SELECT * FROM search_results ORDER BY id DESC LIMIT 1 через query_sql.")
	if err != nil {
		log.Fatal(err)
	}

	// Шаг 9: Логируем чтение
	_, err = a.Execute(ctx, "Запиши в лог через append_log, что чтение подтверждено.")
	if err != nil {
		log.Fatal(err)
	}

	// Финальный ответ — аггрегируем всё в красивый вывод
	finalResult, err := a.Execute(ctx, fmt.Sprintf(
		"Верни краткий структурированный ответ на русском о '%s': заголовок, ключевые факты, источник.",
		query,
	))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Done ===")
	fmt.Println()
	fmt.Println("Search query:", query)
	fmt.Println()
	fmt.Println("Agent response:")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println(finalResult)
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println()
	fmt.Println("Database:", dbPath)
	fmt.Println("Log file:", logPath)
}

// getEnvOr возвращает значение переменной окружения или fallback.
func getEnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
