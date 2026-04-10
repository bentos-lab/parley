package openai

import (
	"context"
	"fmt"
	"strings"

	openaiSDK "github.com/openai/openai-go/v3"
	openaiOption "github.com/openai/openai-go/v3/option"
	openaiParam "github.com/openai/openai-go/v3/packages/param"
	openaiShared "github.com/openai/openai-go/v3/shared"

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
	client openaiSDK.Client
}

// NewClient creates a new OpenAI-compatible client using the provided configuration.
// Parameters:
// - config: OpenAI-compatible configuration values such as base URL, API key, and model.
// Returns:
// - *Client: a client instance ready for use.
func NewClient(config Config) *Client {
	options := []openaiOption.RequestOption{openaiOption.WithAPIKey(config.APIKey)}
	if config.BaseURL != "" {
		normalized := normalizeBaseURL(config.BaseURL)
		if normalized != "" {
			options = append(options, openaiOption.WithBaseURL(normalized))
		}
	}
	return &Client{
		config: config,
		client: openaiSDK.NewClient(options...),
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
	params, err := buildChatCompletionParams(c.config, req, nil)
	if err != nil {
		return "", err
	}
	resp, err := c.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty choices")
	}
	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
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
	responseFormat := buildResponseFormat(schema)
	params, err := buildChatCompletionParams(c.config, req, &responseFormat)
	if err != nil {
		return "", err
	}
	resp, err := c.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty choices")
	}
	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
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
	params, err := buildChatCompletionParams(c.config, req, nil)
	if err != nil {
		errCh <- err
		close(chunkCh)
		close(errCh)
		return chunkCh, errCh
	}
	go func() {
		defer close(chunkCh)
		defer close(errCh)
		stream := c.client.Chat.Completions.NewStreaming(ctx, params)
		defer stream.Close()
		for stream.Next() {
			chunk := stream.Current()
			if len(chunk.Choices) == 0 {
				continue
			}
			delta := chunk.Choices[0].Delta.Content
			if delta != "" {
				chunkCh <- contract.LLMStreamChunk{Content: delta}
			}
		}
		if err := stream.Err(); err != nil {
			errCh <- err
		}
	}()
	return chunkCh, errCh
}

// buildChatCompletionParams constructs chat completion parameters for an LLM request.
// Parameters:
// - config: default OpenAI client configuration values.
// - req: LLM request payload including prompts and generation settings.
// - responseFormat: optional response format override, nil to omit.
// Returns:
// - openaiSDK.ChatCompletionNewParams: parameters ready for the SDK request.
// - error: non-nil when message roles are invalid.
func buildChatCompletionParams(config Config, req contract.LLMRequest, responseFormat *openaiSDK.ChatCompletionNewParamsResponseFormatUnion) (openaiSDK.ChatCompletionNewParams, error) {
	model := config.Model
	if model == "" {
		return openaiSDK.ChatCompletionNewParams{}, fmt.Errorf("model is required")
	}
	messages, err := buildChatMessages(req)
	if err != nil {
		return openaiSDK.ChatCompletionNewParams{}, err
	}
	params := openaiSDK.ChatCompletionNewParams{
		Model:    openaiShared.ChatModel(model),
		Messages: messages,
	}
	if req.Temperature != 0 {
		params.Temperature = openaiParam.NewOpt(req.Temperature)
	}
	if req.MaxTokens > 0 {
		params.MaxTokens = openaiParam.NewOpt(int64(req.MaxTokens))
	}
	if responseFormat != nil {
		params.ResponseFormat = *responseFormat
	}
	return params, nil
}

// buildChatMessages converts the LLM request into OpenAI chat message parameters.
// Parameters:
// - req: LLM request payload including prompts and generation settings.
// Returns:
// - []openaiSDK.ChatCompletionMessageParamUnion: ordered chat messages for the SDK.
// - error: non-nil when message roles are invalid.
func buildChatMessages(req contract.LLMRequest) ([]openaiSDK.ChatCompletionMessageParamUnion, error) {
	messages := make([]openaiSDK.ChatCompletionMessageParamUnion, 0, len(req.Messages)+1)
	if req.SystemInstruction != "" {
		messages = append(messages, openaiSDK.SystemMessage(req.SystemInstruction))
	}
	for _, message := range req.Messages {
		switch message.Role {
		case "user":
			messages = append(messages, openaiSDK.UserMessage(message.Content))
		case "assistant":
			messages = append(messages, openaiSDK.AssistantMessage(message.Content))
		default:
			return nil, fmt.Errorf("invalid message role: %s", message.Role)
		}
	}
	return messages, nil
}

// buildResponseFormat creates a chat completion response_format payload.
// Parameters:
// - schema: JSON schema payload describing the expected response shape; nil selects json_object.
// Returns:
// - openaiSDK.ChatCompletionNewParamsResponseFormatUnion: response format union for the SDK.
func buildResponseFormat(schema *contract.LLMJSONSchema) openaiSDK.ChatCompletionNewParamsResponseFormatUnion {
	if schema == nil {
		objectFormat := openaiShared.NewResponseFormatJSONObjectParam()
		return openaiSDK.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONObject: &objectFormat,
		}
	}
	schemaName := schema.Name
	if schemaName == "" {
		schemaName = "response"
	}
	responseSchema := openaiShared.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:   schemaName,
		Schema: schema.Schema,
	}
	if schema.Strict != nil {
		responseSchema.Strict = openaiParam.NewOpt(*schema.Strict)
	}
	return openaiSDK.ChatCompletionNewParamsResponseFormatUnion{
		OfJSONSchema: &openaiShared.ResponseFormatJSONSchemaParam{JSONSchema: responseSchema},
	}
}

// normalizeBaseURL ensures the base URL includes an OpenAI-compatible API path.
// Parameters:
// - baseURL: raw base URL value from configuration.
// Returns:
// - string: normalized base URL with a versioned path when needed.
func normalizeBaseURL(baseURL string) string {
	trimmed := strings.TrimSuffix(strings.TrimSpace(baseURL), "/")
	if trimmed == "" {
		return ""
	}
	if strings.HasSuffix(trimmed, "/v1") || strings.HasSuffix(trimmed, "/openai") || strings.HasSuffix(trimmed, "/v1beta/openai") {
		return trimmed
	}
	return trimmed + "/v1"
}
