# AI Agent Go

Модульный AI-агент на Go с поддержкой инструментов (tools), streaming и реальной оркестрации через LLM.

## Архитектура

```
ai-agent-go/
├── main.go                  # CLI-входная точка
├── agent/
│   ├── agent.go             # Оркестратор: Execute / ExecuteStream
│   ├── core/                # Базовая структура агента, интерфейсы, конфиги
│   ├── llm/                 # LLM-провайдеры (OpenAI) с HTTP, streaming и tool calling
│   │   ├── mock.go          # MockLLMProvider для тестов
│   ├── filesystem/          # Работа с файловой системой (чтение, запись, список)
│   ├── database/            # SQLite-база данных (запросы, транзакции)
│   ├── web/                 # Веб-поиск (DuckDuckGo) и скрапинг (goquery)
│   └── tools/               # Система инструментов — мост между модулями и LLM
```

### Пакеты

| Пакет | Описание |
|-------|----------|
| `core` | Базовая структура `Agent`, `AgentInterface`, `AgentConfig`, `AgentStatus` |
| `llm` | Интерфейс `LLMProvider`, реализация `OpenAIProvider` с реальным HTTP, streaming (SSE) и function calling; `MockLLMProvider` для тестов |
| `filesystem` | Интерфейс `FileSystemInterface`, реализация `LocalFileSystem` с защитой от выхода за `BasePath` |
| `database` | Интерфейс `DatabaseInterface`, реализация `SQLiteDB` с пулом соединений и транзакциями |
| `web` | Интерфейс `WebSearchInterface`, реализация `WebSearch` — реальный поиск через DuckDuckGo API и HTML-парсинг через goquery |
| `tools` | Система инструментов: `Tool`, `Registry`, `ToolFunc` и билдеры для каждого модуля |

## Возможности

1. **Реальная интеграция с OpenAI** — HTTP-запросы к API с ретриами (5xx, 429) и экспоненциальной задержкой
2. **Tool Calling (Function Calling)** — LLM решает, какие инструменты вызвать; агент выполняет и возвращает результаты в цикл
3. **Streaming** — SSE-потоковый вывод ответов в CLI для мгновенной обратной связи (без дублирующего API-вызова)
4. **Безопасность** — деструктивные операции (`delete_file`, `exec_sql`, `write_file`) требуют подтверждения пользователя; `basePath`-песочница для файловой системы
5. **Потокобезопасность** — `sync.Mutex` защищает историю разговора (`conversation`) при конкурентном использовании
6. **Модульная архитектура** — каждый компонент завязан на интерфейс, легко мокать и заменять
7. **Расширяемость** — добавление нового инструмента — одна функция через `tools.NewToolFunc`

## Установка

```bash
go install github.com/PVMezencev/ai-agent-go@latest
```

**Требования:** Go 1.25+

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

### Локальные модели (Ollama, vLLM, LM Studio)

Агент поддерживает любые LLM-сервисы с OpenAI-совместимым HTTP API. Укажите кастомный эндпоинт и запустите без `OPENAI_API_KEY`:

```bash
# Ollama (по умолчанию на http://127.0.0.1:11434)
export OPENAI_API_KEY=""
export OPENAI_ENDPOINT="http://localhost:11435/v1"
export OPENAI_MODEL="llama3"

ai-agent "Расскажи о goroutines"

# LM Studio / vLLM / любой OpenAI-совместимый сервер
export OPENAI_ENDPOINT="http://localhost:8855/v1"
ai-agent "Свертай SELECT на Go"
```

**Важно:** при использовании кастомного эндпоинта `OPENAI_API_KEY` может быть пустым — агент передаст запрос без `Authorization` заголовка.

### Переменные окружения

| Переменная | Описание | По умолчанию |
|------------|----------|-------------|
| `OPENAI_API_KEY` | API-ключ OpenAI (обязательно) | — |
| `OPENAI_MODEL` | Модель для использования | `gpt-4o` |
| `OPENAI_ENDPOINT` | Кастумный API-эндпоинт (Ollama, vLLM, LM Studio и любые совместимые с OpenAI API) | `https://api.openai.com` |

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
    result, err := a.Execute(ctx, "Найди информацию о Go 1.25 и сохрани в файл news.md")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(result)

    // Streaming — вывод ответа по мере генерации (без дублирующего API-вызова)
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
| `web_search` | web | Поиск в интернете (DuckDuckGo) | — |
| `scrape_url` | web | Скрапинг веб-страницы (goquery) | — |

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

### Streaming без дублирующего запроса

`ExecuteStream` использует тот же ответ LLM из основного цикла — финальный ответ стримится из уже полученного контента через `streamContent()`, а не делает отдельный API-вызов.

## Тестирование

```bash
go test ./...
```

```bash
go test ./... -race    # проверка на data race
```

Включены тесты оркестрации с моком `MockLLMProvider`:
- `Execute` — прямой ответ, tool call → ответ, исчерпание раундов, ошибки
- `ExecuteStream` — стриминг ответа, tool call → стриминг
- Потокобезопасность (`ConversationIsThreadSafe`)
- Интеграционный тест реального веб-поиска (`TestWebSearch_RealSearch`)

Все интерфейсы спроектированы для простого мокирования.

## Примеры

В каталоге `example/` — минимальное веб-приложение, демонстрирующее использование пакета как библиотеки:

- **Поиск в интернете** через `web_search` tool агента
- **Запись результатов в SQLite** через `exec_sql` / `query_sql`
- **Логирование в файл** через кастомный `append_log` tool, зарегистрированный в `ToolRegistry`

Агент сам orchestrate-ит весь пайплайн: создаёт директорию, ищет информацию, сохраняет в БД, логирует каждый шаг.

```bash
cd example
export OPENAI_API_KEY="sk-..."
go run main.go
```

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

## Лицензия

MIT
