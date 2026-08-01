package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

// OpenAIProvider implements the LLMProvider interface for OpenAI
type OpenAIProvider struct {
	config LLMConfig
	client *http.Client
}

// NewOpenAIProvider creates a new OpenAI provider instance
func NewOpenAIProvider(config LLMConfig) (*OpenAIProvider, error) {
	if config.APIKey == "" && config.APIEndpoint == "" {
		return nil, fmt.Errorf("either APIKey or APIEndpoint must be set")
	}
	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}

	return &OpenAIProvider{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}, nil
}

// ChatCompletion implements the LLMProvider interface
func (p *OpenAIProvider) ChatCompletion(ctx context.Context, request ChatRequest) (*ChatResponse, error) {
	if request.Model == "" {
		request.Model = p.config.Model
	}

	var lastErr error
	for attempt := 0; attempt <= p.config.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(int(math.Pow(2, float64(attempt))) * 1000) * time.Millisecond
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return nil, ctx.Err()
			case <-timer.C:
			}
		}

		response, err := p.doChatRequest(ctx, request)
		if err != nil {
			lastErr = err
			if !isRetryableError(err) {
				return nil, err
			}
			continue
		}
		return response, nil
	}

	return nil, fmt.Errorf("OpenAI ChatCompletion failed after %d retries: %w", p.config.MaxRetries, lastErr)
}

// ChatCompletionStream streams chat completion responses via SSE
func (p *OpenAIProvider) ChatCompletionStream(ctx context.Context, request ChatRequest, handler func(chunk ChatStreamChunk) error) error {
	if request.Model == "" {
		request.Model = p.config.Model
	}

	var lastErr error
	for attempt := 0; attempt <= p.config.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(int(math.Pow(2, float64(attempt))) * 1000) * time.Millisecond
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}

		err := p.doStreamRequest(ctx, request, handler)
		if err != nil {
			lastErr = err
			if !isRetryableError(err) {
				return err
			}
			continue
		}
		return nil
	}

	return fmt.Errorf("OpenAI ChatCompletionStream failed after %d retries: %w", p.config.MaxRetries, lastErr)
}

// Embeddings implements the LLMProvider interface
func (p *OpenAIProvider) Embeddings(ctx context.Context, request EmbeddingRequest) (*EmbeddingResponse, error) {
	if request.Model == "" {
		request.Model = p.config.Model
	}

	payload := struct {
		Model string   `json:"model"`
		Input []string `json:"input"`
	}{
		Model: request.Model,
		Input: request.Input,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}

	endpoint := p.config.APIEndpoint
	if endpoint == "" {
		endpoint = "https://api.openai.com"
	}
	url := endpoint + "/v1/embeddings"

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if p.config.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("LLM API error (status %d): %s", resp.StatusCode, string(errBody))
	}

	var result openaiEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	data := make([]EmbeddingData, len(result.Data))
	for i, d := range result.Data {
		data[i] = EmbeddingData{
			Object:    d.Object,
			Embedding: d.Embedding,
			Index:     d.Index,
		}
	}

	return &EmbeddingResponse{
		Model: result.Model,
		Data:  data,
	}, nil
}

// GetModelInfo implements the LLMProvider interface
func (p *OpenAIProvider) GetModelInfo() ModelInfo {
	return ModelInfo{
		Name:        p.config.Model,
		Description: "OpenAI " + p.config.Model,
		MaxTokens:   128000,
		Provider:    "OpenAI",
	}
}

// SetConfig implements the LLMProvider interface
func (p *OpenAIProvider) SetConfig(config LLMConfig) error {
	p.config = config
	if config.Timeout > 0 {
		p.client.Timeout = config.Timeout
	}
	return nil
}

// --- Internal helpers ---

func (p *OpenAIProvider) doChatRequest(ctx context.Context, request ChatRequest) (*ChatResponse, error) {
	messages := make([]openAIMessage, len(request.Messages))
	for i, m := range request.Messages {
		msg := openAIMessage{
			Role:    m.Role,
			Content: m.Content,
		}
		if m.Role == "assistant" && len(m.ToolCalls) > 0 {
			msg.ToolCalls = make([]openAIToolCall, len(m.ToolCalls))
			for j, tc := range m.ToolCalls {
				msg.ToolCalls[j] = openAIToolCall{
					ID:   tc.ID,
					Type: tc.Type,
					Function: openAIFunctionCall{
						Name:      tc.FunctionCall.Name,
						Arguments: tc.FunctionCall.Arguments,
					},
				}
			}
		}
		messages[i] = msg
	}

	// Convert tools to API format
	var tools []openAIToolDef
	for _, td := range request.Tools {
		tools = append(tools, openAIToolDef{
			Type: "function",
			Function: openAIFunctionDef{
				Name:        td.Function.Name,
				Description: td.Function.Description,
				Parameters:  mustRawJSON(td.Function.Parameters),
			},
		})
	}

	payload := struct {
		Model       string          `json:"model"`
		Messages    []openAIMessage `json:"messages"`
		Tools       []openAIToolDef `json:"tools,omitempty"`
		Stream      bool            `json:"stream"`
		Temperature float64         `json:"temperature,omitempty"`
		MaxTokens   *int            `json:"max_tokens,omitempty"`
	}{
		Model:    request.Model,
		Messages: messages,
		Tools:    tools,
		Stream:   false,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := p.config.APIEndpoint
	if endpoint == "" {
		endpoint = "https://api.openai.com"
	}
	url := endpoint + "/v1/chat/completions"

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if p.config.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return nil, newAPIError(resp.StatusCode, string(errBody))
	}

	var result openAIChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	choices := make([]Choice, len(result.Choices))
	for i, c := range result.Choices {
		msg := Message{
			Role:    c.Message.Role,
			Content: c.Message.Content,
		}
		if c.Message.ToolCalls != nil {
			msg.ToolCalls = make([]ToolCall, len(c.Message.ToolCalls))
			for j, tc := range c.Message.ToolCalls {
				msg.ToolCalls[j] = ToolCall{
					ID:   tc.ID,
					Type: tc.Type,
					FunctionCall: FunctionCall{
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				}
			}
		}
		choices[i] = Choice{
			Index:        c.Index,
			Message:      msg,
			FinishReason: c.FinishReason,
		}
	}

	return &ChatResponse{
		ID:      result.ID,
		Model:   result.Model,
		Created: time.Unix(result.Created, 0),
		Choices: choices,
	}, nil
}

func (p *OpenAIProvider) doStreamRequest(ctx context.Context, request ChatRequest, handler func(chunk ChatStreamChunk) error) error {
	messages := make([]openAIMessage, len(request.Messages))
	for i, m := range request.Messages {
		messages[i] = openAIMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	payload := struct {
		Model    string          `json:"model"`
		Messages []openAIMessage `json:"messages"`
		Stream   bool            `json:"stream"`
	}{
		Model:    request.Model,
		Messages: messages,
		Stream:   true,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := p.config.APIEndpoint
	if endpoint == "" {
		endpoint = "https://api.openai.com"
	}
	url := endpoint + "/v1/chat/completions"

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if p.config.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		return newAPIError(resp.StatusCode, string(errBody))
	}

	// Parse SSE stream
	scanner := newSSEScanner(resp.Body)
	for scanner.hasNext() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		event, data := scanner.next()
		if event == "done" {
			break
		}

		if strings.TrimSpace(data) == "[DONE]" {
			break
		}

		var chunk openAIChatResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		for _, c := range chunk.Choices {
			streamChunk := ChatStreamChunk{
				Index:        c.Index,
				Content:      c.Message.Content,
				FinishReason: c.FinishReason,
			}
			if err := handler(streamChunk); err != nil {
				return err
			}
		}
	}

	return nil
}

// isRetryableError checks if the error is retryable
func isRetryableError(err error) bool {
	if err == context.Canceled || err == context.DeadlineExceeded {
		return false
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode >= 500 || apiErr.StatusCode == 429
	}
	return true // network errors, timeouts, etc.
}

// newAPIError creates a structured API error
func newAPIError(status int, body string) *APIError {
	return &APIError{
		StatusCode: status,
		Body:       strings.TrimSpace(body),
	}
}

// APIError represents an error returned by the OpenAI API
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("OpenAI API error (status %d): %s", e.StatusCode, e.Body)
}

// ChatStreamChunk represents a single chunk in a streamed response
type ChatStreamChunk struct {
	Index        int
	Content      string
	FinishReason string
}

// SSEScanner parses Server-Sent Events stream
type sseScanner struct {
	reader   *io.LimitedReader
	buf      []byte
	line     []byte
	linePos  int
	eof      bool
}

func newSSEScanner(body io.Reader) *sseScanner {
	return &sseScanner{
		reader: &io.LimitedReader{R: body, N: math.MaxInt64},
		buf:    make([]byte, 0, 4096),
	}
}

func (s *sseScanner) hasNext() bool {
	if s.eof {
		// Check if there's buffered data left
		return len(s.buf) > 0 || len(s.line) > 0
	}

	// Read until we have a complete line or EOF
	for len(s.line) == 0 || s.line[len(s.line)-1] != '\n' {
		if cap(s.buf)-len(s.buf) < 512 {
			s.buf = append(s.buf, make([]byte, 4096)...)
		}
		n, err := s.reader.Read(s.buf[len(s.buf):cap(s.buf)])
		if n > 0 {
			s.buf = s.buf[:len(s.buf)+n]
		}
		if err != nil {
			s.eof = true
			break
		}
	}

	// Extract a line from buffer
	idx := strings.Index(string(s.buf), "\n")
	if idx >= 0 {
		s.line = []byte(s.buf[:idx])
		s.buf = s.buf[idx+1:]
		return true
	}

	if len(s.buf) > 0 {
		s.line = s.buf
		s.buf = nil
		return true
	}

	s.eof = true
	return false
}

func (s *sseScanner) next() (string, string) {
	lineStr := strings.TrimSpace(string(s.line))
	s.line = nil

	if lineStr == "" {
		// Empty line means end of event — return accumulated (not implemented here, simplified)
		return "", ""
	}

	if strings.HasPrefix(lineStr, "data: ") {
		data := strings.TrimPrefix(lineStr, "data: ")
		return "data", data
	}

	if strings.HasPrefix(lineStr, "event: ") {
		event := strings.TrimPrefix(lineStr, "event: ")
		return event, ""
	}

	return "data", lineStr
}

// --- JSON types for OpenAI API ---

type openAIChatResponse struct {
	ID      string               `json:"id"`
	Model   string               `json:"model"`
	Created int64                `json:"created"`
	Choices []openAIChoice       `json:"choices"`
}

type openAIChoice struct {
	Index        int            `json:"index"`
	Message      openAIMessage  `json:"message"`
	FinishReason string         `json:"finish_reason"`
}

type openAIMessage struct {
	Role       string         `json:"role"`
	Content    string         `json:"content"`
	ToolCalls  []openAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
}

// MarshalJSON implements custom JSON marshaling for openAIMessage
// to handle tool messages properly (tool role needs tool_call_id, not content at top level)
func (m openAIMessage) MarshalJSON() ([]byte, error) {
	type alias openAIMessage
	var fields map[string]interface{}
	_ = json.Unmarshal(marshalStruct(alias(m)), &fields)

	// Build the map manually for precise control
	result := map[string]interface{}{
		"role":    m.Role,
		"content": m.Content,
	}
	if m.ToolCalls != nil {
		data, err := json.Marshal(m.ToolCalls)
		if err != nil {
			return nil, err
		}
		result["tool_calls"] = json.RawMessage(data)
	}
	if m.ToolCallID != "" {
		result["tool_call_id"] = m.ToolCallID
	}

	return json.Marshal(result)
}

func marshalStruct(v interface{}) []byte {
	b, _ := json.Marshal(v)
	return b
}

// mustRawJSON converts a map/slice to json.RawMessage
func mustRawJSON(v interface{}) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage("{}")
	}
	return b
}

type openAIToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function openAIFunctionCall `json:"function"`
}

type openAIFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openAIToolDef struct {
	Type     string           `json:"type"`
	Function openAIFunctionDef `json:"function"`
}

type openAIFunctionDef struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type openaiEmbeddingResponse struct {
	Model string                 `json:"model"`
	Data  []openaiEmbeddingData  `json:"data"`
}

type openaiEmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}
