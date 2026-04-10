package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genai"

	"github.com/bentos-lab/parley/core/contract"
)

// Config holds Gemini configuration values.
type Config struct {
	APIKey string
	Model  string
}

// Client implements the LLM contract using the Google Gen AI Gemini API.
type Client struct {
	config Config
	client *genai.Client
}

// NewClient creates a new Gemini client using the provided configuration.
// Parameters:
// - config: Gemini configuration values such as API key and default model.
// Returns:
// - *Client: a client instance ready for use.
func NewClient(config Config) *Client {
	return &Client{config: config}
}

// Generate produces free-form text using GenerateContent.
// Parameters:
// - ctx: context for request cancellation and deadlines.
// - req: LLM request payload including prompts and generation settings.
// Returns:
// - string: generated content.
// - error: non-nil when request or decoding fails.
func (c *Client) Generate(ctx context.Context, req contract.LLMRequest) (string, error) {
	client, err := c.ensureClient(ctx)
	if err != nil {
		return "", err
	}
	model := resolveModel(req.Model, c.config.Model)
	if model == "" {
		return "", fmt.Errorf("model is required")
	}
	contents := buildGeminiContents(req.Messages)
	config := buildGeminiConfig(req, nil)
	resp, err := client.Models.GenerateContent(ctx, model, contents, config)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(extractGeminiText(resp)), nil
}

// GenerateJSON produces JSON output using response schema enforcement when supported.
// Parameters:
// - ctx: context for request cancellation and deadlines.
// - req: LLM request payload including prompts and generation settings.
// - schema: JSON schema payload describing the expected response shape.
// Returns:
// - string: generated JSON content.
// - error: non-nil when request or decoding fails, or when JSON schema is invalid.
func (c *Client) GenerateJSON(ctx context.Context, req contract.LLMRequest, schema *contract.LLMJSONSchema) (string, error) {
	client, err := c.ensureClient(ctx)
	if err != nil {
		return "", err
	}
	model := resolveModel(req.Model, c.config.Model)
	if model == "" {
		return "", fmt.Errorf("model is required")
	}
	contents := buildGeminiContents(req.Messages)

	var responseSchema *genai.Schema
	if schema != nil && schema.Schema != nil {
		encoded, err := json.Marshal(schema.Schema)
		if err != nil {
			return "", fmt.Errorf("marshal json schema: %w", err)
		}
		var parsed genai.Schema
		if err := json.Unmarshal(encoded, &parsed); err != nil {
			return "", fmt.Errorf("parse json schema: %w", err)
		}
		responseSchema = &parsed
	}

	config := buildGeminiConfig(req, responseSchema)
	if config == nil {
		config = &genai.GenerateContentConfig{}
	}
	config.ResponseMIMEType = "application/json"
	resp, err := client.Models.GenerateContent(ctx, model, contents, config)
	if err != nil {
		return "", err
	}
	text := strings.TrimSpace(extractGeminiText(resp))
	var validate any
	if err := json.Unmarshal([]byte(text), &validate); err != nil {
		return "", fmt.Errorf("invalid json response: %w", err)
	}
	return text, nil
}

// GenerateStream produces streaming output using GenerateContentStream.
// Parameters:
// - ctx: context for request cancellation and deadlines.
// - req: LLM request payload including prompts and generation settings.
// Returns:
// - <-chan contract.LLMStreamChunk: channel of streamed chunks.
// - <-chan error: channel for stream errors.
func (c *Client) GenerateStream(ctx context.Context, req contract.LLMRequest) (<-chan contract.LLMStreamChunk, <-chan error) {
	chunkCh := make(chan contract.LLMStreamChunk)
	errCh := make(chan error, 1)
	client, err := c.ensureClient(ctx)
	if err != nil {
		errCh <- err
		close(chunkCh)
		close(errCh)
		return chunkCh, errCh
	}
	model := resolveModel(req.Model, c.config.Model)
	if model == "" {
		errCh <- fmt.Errorf("model is required")
		close(chunkCh)
		close(errCh)
		return chunkCh, errCh
	}
	contents := buildGeminiContents(req.Messages)
	config := buildGeminiConfig(req, nil)
	go func() {
		defer close(chunkCh)
		defer close(errCh)
		for resp, err := range client.Models.GenerateContentStream(ctx, model, contents, config) {
			if err != nil {
				errCh <- err
				return
			}
			text := strings.TrimSpace(extractGeminiText(resp))
			if text == "" {
				continue
			}
			chunkCh <- contract.LLMStreamChunk{Content: text}
		}
	}()
	return chunkCh, errCh
}

// ensureClient lazily constructs the underlying GenAI client.
// Parameters:
// - ctx: context used to initialize the client.
// Returns:
// - *genai.Client: initialized client.
// - error: non-nil when the client cannot be created.
func (c *Client) ensureClient(ctx context.Context) (*genai.Client, error) {
	if c.client != nil {
		return c.client, nil
	}
	if strings.TrimSpace(c.config.APIKey) == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required when using the gemini provider")
	}
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  c.config.APIKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}
	c.client = client
	return c.client, nil
}

// resolveModel selects the effective model name.
// Parameters:
// - override: request override model.
// - fallback: configured default model.
// Returns:
// - string: resolved model name.
func resolveModel(override string, fallback string) string {
	if strings.TrimSpace(override) != "" {
		return strings.TrimSpace(override)
	}
	return strings.TrimSpace(fallback)
}

// buildGeminiContents converts contract messages into Gemini contents.
// Parameters:
// - messages: contract message list.
// Returns:
// - []*genai.Content: Gemini contents suitable for GenerateContent.
func buildGeminiContents(messages []contract.LLMMessage) []*genai.Content {
	contents := make([]*genai.Content, 0, len(messages))
	for _, message := range messages {
		role := "user"
		if message.Role == "assistant" {
			role = "model"
		}
		contents = append(contents, &genai.Content{
			Role:  role,
			Parts: []*genai.Part{{Text: message.Content}},
		})
	}
	return contents
}

// buildGeminiConfig converts a contract request into a Gemini GenerateContentConfig.
// Parameters:
// - req: contract LLM request.
// - schema: optional response schema for JSON enforcement.
// Returns:
// - *genai.GenerateContentConfig: Gemini generation config (nil when all fields are empty).
func buildGeminiConfig(req contract.LLMRequest, schema *genai.Schema) *genai.GenerateContentConfig {
	cfg := &genai.GenerateContentConfig{}
	changed := false
	if strings.TrimSpace(req.SystemInstruction) != "" {
		cfg.SystemInstruction = &genai.Content{Parts: []*genai.Part{{Text: req.SystemInstruction}}}
		changed = true
	}
	if req.Temperature != 0 {
		value := req.Temperature
		cfg.Temperature = &value
		changed = true
	}
	if req.MaxTokens > 0 {
		value := int64(req.MaxTokens)
		cfg.MaxOutputTokens = &value
		changed = true
	}
	if schema != nil {
		cfg.ResponseSchema = schema
		changed = true
	}
	if !changed {
		return nil
	}
	return cfg
}

// extractGeminiText extracts response text from the first candidate.
// Parameters:
// - resp: Gemini GenerateContent response.
// Returns:
// - string: extracted text content.
func extractGeminiText(resp *genai.GenerateContentResponse) string {
	if resp == nil || len(resp.Candidates) == 0 {
		return ""
	}
	candidate := resp.Candidates[0]
	if candidate == nil || candidate.Content == nil {
		return ""
	}
	var builder strings.Builder
	for _, part := range candidate.Content.Parts {
		if part == nil {
			continue
		}
		if part.Text == "" {
			continue
		}
		builder.WriteString(part.Text)
	}
	return builder.String()
}
