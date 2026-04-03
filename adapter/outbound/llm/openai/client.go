package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/bentos-lab/parley/core/contract"
)

// Config holds OpenAI-compatible configuration values.
type Config struct {
	BaseURL string
	APIKey  string
	Model   string
}

// Client implements the LLM contract using OpenAI-compatible APIs.
type Client struct {
	config Config
	http   *http.Client
}

// NewClient creates a new OpenAI-compatible client using the provided configuration.
// Parameters:
// - config: OpenAI-compatible configuration values such as base URL, API key, and model.
// Returns:
// - *Client: a client instance ready for use.
func NewClient(config Config) *Client {
	return &Client{
		config: config,
		http:   &http.Client{},
	}
}

// Generate produces free-form text using chat completions.
// Parameters:
// - ctx: context for request cancellation and deadlines.
// - req: LLM request payload including prompts and generation settings.
// Returns:
// - string: generated content.
// - error: non-nil when request or decoding fails.
func (c *Client) Generate(ctx context.Context, req contract.LLMRequest) (string, error) {
	payload, err := buildRequest(c.config, req, false, false, nil)
	if err != nil {
		return "", err
	}
	body, err := c.doRequest(ctx, payload)
	if err != nil {
		return "", err
	}
	var parsed chatCompletionResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("empty choices")
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}

// GenerateJSON produces JSON output using chat completions with response_format when supported.
// Parameters:
// - ctx: context for request cancellation and deadlines.
// - req: LLM request payload including prompts and generation settings.
// - schema: JSON schema payload describing the expected response shape.
// Returns:
// - string: JSON content returned by the model.
// - error: non-nil when request or decoding fails, or when JSON schema is invalid.
func (c *Client) GenerateJSON(ctx context.Context, req contract.LLMRequest, schema *contract.LLMJSONSchema) (string, error) {
	if schema != nil && schema.Schema == nil {
		return "", fmt.Errorf("json schema is required when schema is provided")
	}
	payload, err := buildRequest(c.config, req, true, false, schema)
	if err != nil {
		return "", err
	}
	body, err := c.doRequest(ctx, payload)
	if err != nil {
		return "", err
	}
	var parsed chatCompletionResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("empty choices")
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}

// GenerateStream produces streaming output using chat completions.
// Parameters:
// - ctx: context for request cancellation and deadlines.
// - req: LLM request payload including prompts and generation settings.
// Returns:
// - <-chan contract.LLMStreamChunk: channel of streamed chunks.
// - <-chan error: channel for stream errors.
func (c *Client) GenerateStream(ctx context.Context, req contract.LLMRequest) (<-chan contract.LLMStreamChunk, <-chan error) {
	chunkCh := make(chan contract.LLMStreamChunk)
	errCh := make(chan error, 1)
	payload, err := buildRequest(c.config, req, false, true, nil)
	if err != nil {
		errCh <- err
		close(chunkCh)
		close(errCh)
		return chunkCh, errCh
	}
	go func() {
		defer close(chunkCh)
		defer close(errCh)
		url := chatCompletionsURL(c.config.BaseURL)
		data, err := json.Marshal(payload)
		if err != nil {
			errCh <- fmt.Errorf("encode request: %w", err)
			return
		}
		reqHTTP, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
		if err != nil {
			errCh <- fmt.Errorf("create request: %w", err)
			return
		}
		reqHTTP.Header.Set("Content-Type", "application/json")
		reqHTTP.Header.Set("Authorization", "Bearer "+c.config.APIKey)
		resp, err := c.http.Do(reqHTTP)
		if err != nil {
			errCh <- fmt.Errorf("do request: %w", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			body, _ := io.ReadAll(resp.Body)
			errCh <- fmt.Errorf("llm error %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
			return
		}
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || !strings.HasPrefix(line, "data:") {
				continue
			}
			payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if payload == "[DONE]" {
				return
			}
			var event streamResponse
			if err := json.Unmarshal([]byte(payload), &event); err != nil {
				continue
			}
			if len(event.Choices) == 0 {
				continue
			}
			delta := event.Choices[0].Delta.Content
			if delta != "" {
				chunkCh <- contract.LLMStreamChunk{Content: delta}
			}
		}
		if err := scanner.Err(); err != nil {
			errCh <- fmt.Errorf("stream read: %w", err)
		}
	}()
	return chunkCh, errCh
}

// buildRequest constructs a chat completion request payload.
// Parameters:
// - config: default OpenAI client configuration values.
// - req: LLM request payload including prompts and generation settings.
// - jsonMode: whether JSON response formatting should be requested.
// - stream: whether streaming should be enabled for the request.
// - schema: JSON schema payload describing the expected response shape.
// Returns:
// - chatCompletionRequest: the payload ready to be sent to the API.
// - error: non-nil when message roles are invalid.
func buildRequest(config Config, req contract.LLMRequest, jsonMode bool, stream bool, schema *contract.LLMJSONSchema) (chatCompletionRequest, error) {
	model := req.Model
	if model == "" {
		model = config.Model
	}
	messages := make([]chatMessage, 0, len(req.Messages)+1)
	if req.SystemInstruction != "" {
		messages = append(messages, chatMessage{Role: "system", Content: req.SystemInstruction})
	}
	for _, message := range req.Messages {
		if message.Role != "user" && message.Role != "assistant" {
			return chatCompletionRequest{}, fmt.Errorf("invalid message role: %s", message.Role)
		}
		messages = append(messages, chatMessage{Role: message.Role, Content: message.Content})
	}
	payload := chatCompletionRequest{
		Model:       model,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Stream:      stream,
		Messages:    messages,
	}
	if jsonMode {
		if schema != nil {
			schemaName := schema.Name
			if schemaName == "" {
				schemaName = "response"
			}
			payload.ResponseFormat = &responseFormat{
				Type: "json_schema",
				JSONSchema: &responseSchema{
					Name:   schemaName,
					Schema: schema.Schema,
					Strict: schema.Strict,
				},
			}
		} else {
			payload.ResponseFormat = &responseFormat{Type: "json_object"}
		}
	}
	return payload, nil
}

// doRequest performs the HTTP request to the OpenAI-compatible API.
// Parameters:
// - ctx: context for request cancellation and deadlines.
// - payload: JSON-serializable request body.
// Returns:
// - []byte: raw response body.
// - error: non-nil when request or decoding fails.
func (c *Client) doRequest(ctx context.Context, payload chatCompletionRequest) ([]byte, error) {
	url := chatCompletionsURL(c.config.BaseURL)
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode request: %w", err)
	}
	reqHTTP, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	reqHTTP.Header.Set("Content-Type", "application/json")
	reqHTTP.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	resp, err := c.http.Do(reqHTTP)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("llm error %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return body, nil
}

func chatCompletionsURL(baseURL string) string {
	base := strings.TrimSuffix(baseURL, "/")
	if strings.HasSuffix(base, "/v1") || strings.HasSuffix(base, "/openai") || strings.HasSuffix(base, "/v1beta/openai") {
		return base + "/chat/completions"
	}
	return base + "/v1/chat/completions"
}

type chatCompletionRequest struct {
	Model          string          `json:"model"`
	Messages       []chatMessage   `json:"messages"`
	Temperature    float64         `json:"temperature,omitempty"`
	MaxTokens      int             `json:"max_tokens,omitempty"`
	Stream         bool            `json:"stream,omitempty"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

type responseFormat struct {
	Type       string          `json:"type"`
	JSONSchema *responseSchema `json:"json_schema,omitempty"`
}

type responseSchema struct {
	Name   string         `json:"name"`
	Schema map[string]any `json:"schema"`
	Strict *bool          `json:"strict,omitempty"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

type streamResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}
