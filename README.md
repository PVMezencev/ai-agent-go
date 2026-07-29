# AI Agent Go

Модульный AI-агент на Go с поддержкой инструментов (tools), streaming и реальной оркестрации через LLM.

## Архитектура

```
agent/
├── core/          # Базовая структура агента, интерфейсы, конфиги
├── llm/           # LLM-провайдеры (OpenAI) с HTTP, streaming и tool calling
├── filesystem/    # Работа с файловой системой (чтение, запись, список)
├── database/      # SQLite-база данных (запросы, транзакции)
├── web/           # Веб-поиск и скрапинг страниц
└── tools/         # Система инструментов — мост между модулями и LLM
```

### Пакеты

| Пакет | Описание |
|-------|----------|
| `core` | Базовая структура `Agent`, `AgentInterface`, `AgentConfig`, `AgentStatus` |
| `llm` | Интерфейс `LLMProvider`, реализация `OpenAIProvider` с реальным HTTP, streaming (SSE) и function calling |
| `filesystem` | Интерфейс `FileSystemInterface`, реализация `LocalFileSystem` с защитой от выхода за `BasePath` |
| `database` | Интерфейс `DatabaseInterface`, реализация `SQLiteDB` с пулом соединений и транзакциями |
| `web` | Интерфейс `WebSearchInterface`, реализация `WebSearch` для поиска и скрапинга |
| `tools` | Система инструментов: `Tool`, `Registry`, `ToolFunc` и билдеры для каждого модуля |

## Возможности

1. **Реальная интеграция с OpenAI** — HTTP-запросы к API с ретриями (5xx, 429) и экспоненциальной задержкой
2. **Tool Calling (Function Calling)** — LLM решает, какие инструменты вызвать; агент выполняет и возвращает результаты в цикл
3. **Streaming** — SSE-потоковый вывод ответов в CLI для мгновенной обратной связи
4. **Безопасность** — деструктивные операции (`delete_file`, `exec_sql`, `write_file`) требуют подтверждения пользователя
5. **Модульная архитектура** — каждый компонент закреплён на интерфейсе, легко мокать и заменять
6. **Расширяемость** — добавление нового инструмента — одна функция через `tools.NewToolFunc`

## Установка

```bash
go install github.com/PVMezencev/ai-agent-go@latest
```

## Быстрый старт

### CLI

```bash
export OPENAI_API_KEY="sk-..."

# Базовый запрос
ai-agent "Напиши краткий обзор goroutines в Go"

# С инструментами
ai-agent "Прочитай файл go.mod и объясни зависимости"
ai-agent "Поискай в интернете последнюю версию Go"
```

### Переменные окружения

| Переменная | Описание | По умолчанию |
|------------|----------|-------------|
| `OPENAI_API_KEY` | API-ключ OpenAI (обязательно) | — |
| `OPENAI_MODEL` | Модель для использования | `gpt-4o` |
| `OPENAI_ENDPOINT` | Кастомный API-эндпоинт | `https://api.openai.com` |

### Как библиотека

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/PVMezencev/ai-agent-go/agent"
    "github.com/PVMezencev/ai-agent-go/agent/core"
    "github.com/PVMezencev/ai-agent-go/agent/llm"
    "github.com/PVMezencev/ai-agent-go/agent/filesystem"
    "github.com/PVMezencev/ai-agent-go/agent/database"
    "github.com/PVMezencev/ai-agent-go/agent/web"
)

func main() {
    config := agent.AgentConfig{
        AgentConfig: core.AgentConfig{
            ID:       "my-agent",
            Name:     "My AI Agent",
            Enabled:  true,
            Timeout:  120 * time.Second,
        },
        LLMConfig: llm.LLMConfig{
            APIKey:     "sk-...",
            Model:      "gpt-4o",
            Timeout:    120 * time.Second,
            MaxRetries: 3,
        },
        FileSystemConfig: filesystem.FileSystemConfig{
            BasePath: "/data",
            Timeout:  10 * time.Second,
        },
        DatabaseConfig: database.DatabaseConfig{
            DSN:             "file:app.db?cache=shared",
            MaxOpenConns:    10,
            MaxIdleConns:    5,
            ConnMaxLifetime: 5 * time.Minute,
            Timeout:         10 * time.Second,
        },
        WebConfig: web.WebConfig{
            Timeout:   30 * time.Second,
            MaxRetries: 3,
            UserAgent: "MyAgent/1.0",
        },
        MaxToolRounds: 10,
    }

    a, err := agent.NewAgent(config)
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    if err := a.Start(ctx); err != nil {
        log.Fatal(err)
    }
    defer a.Stop(ctx)

    // Execute — отправляет задачу в LLM с tools,
    // обрабатывает tool calls и возвращает финальный ответ
    result, err := a.Execute(ctx, "Найди информацию о Go 1.24 и сохрани в файл news.md")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(result)

    // Streaming — вывод ответа по мере генерации
    err = a.ExecuteStream(ctx, "Расскажи о паттернах concurrency в Go", func(chunk llm.ChatStreamChunk) error {
        fmt.Print(chunk.Content)
        return nil
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

## Доступные инструменты

Агент автоматически регистрирует инструменты из доступных модулей:

| Инструмент | Модуль | Описание | Безопасность |
|------------|--------|----------|-------------|
| `read_file` | filesystem | Чтение содержимого файла | — |
| `write_file` | filesystem | Запись содержимого в файл | Требует подтверждения |
| `list_files` | filesystem | Список файлов в директории | — |
| `delete_file` | filesystem | Удаление файла | Требует подтверждения |
| `create_directory` | filesystem | Создание директории | — |
| `query_sql` | database | Выполнение SELECT-запросов | — |
| `exec_sql` | database | Выполнение SQL-операций (INSERT/UPDATE/DELETE) | Требует подтверждения |
| `web_search` | web | Поиск в интернете | — |
| `scrape_url` | web | Скрапинг веб-страницы | — |

## Как работает оркестрация

```
Пользователь → Execute(task)
                    ↓
           [LLM + tools definitions]
                    ↓
           LLM решает: ответ или tool call?
                    ↓
        ┌───────────┴───────────┐
        │                        │
    Текст ответа            Tool Call(s)
        │                        ↓
        │                 Выполнение инструментов
        │                        ↓
        │                 Результат → conversation
        │                        ↓
        │                 Снова к LLM (round + 1)
        │                        ↓
        └───────────←───────────┘
                    ↓
           Результат пользователю
```

Цикл повторяется до `MaxToolRounds` (по умолчанию 10) или пока LLM не вернёт текстовый ответ без tool calls.

## Добавление нового инструмента

```go
import "github.com/PVMezencev/ai-agent-go/agent/tools"

registry := tools.NewRegistry()

registry.Register(tools.NewToolFunc(
    "calculate",
    "Perform a mathematical calculation",
    map[string]interface{}{
        "type":       "object",
        "properties": map[string]interface{}{
            "expression": map[string]interface{}{
                "type":        "string",
                "description": "Math expression to evaluate",
            },
        },
        "required": []string{"expression"},
    },
    func(ctx context.Context, argsJSON string) (string, error) {
        args, _ := tools.ParseToolArgs(argsJSON)
        expr := args["expression"].(string)
        // evaluate...
        return fmt.Sprintf("Result: %s", expr), nil
    },
))
```

## Тестирование

```bash
go test ./...
```

Все интерфейсы спроектированы для простого мокирования. Модули можно заменить на тестовые реализации без подключения к внешним сервисам.

## Лицензия

MIT
