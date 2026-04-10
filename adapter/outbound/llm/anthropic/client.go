package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	anthropicSDK "github.com/anthropics/anthropic-sdk-go"
	anthropicOption "github.com/anthropics/anthropic-sdk-go/option"
	anthropicParam "github.com/anthropics/anthropic-sdk-go/packages/param"

	"github.com/bentos-lab/parley/core/contract"
)

// Config holds Anthropic configuration values.
type Config struct {
	APIKey string
	Model  string
}

// Client implements the LLM contract using the Anthropic Messages API.
type Client struct {
	config Config
	client anthropicSDK.Client
}

// NewClient creates a new Anthropic client using the provided configuration.
// Parameters:
// - config: Anthropic configuration values such as API key and default model.
// Returns:
// - *Client: a client instance ready for use.
func NewClient(config Config) *Client {
	return &Client{
		config: config,
		client: anthropicSDK.NewClient(
			anthropicOption.WithAPIKey(config.APIKey),
		),
	}
}

// Generate produces free-form text using the Messages API.
// Parameters:
// - ctx: context for request cancellation and deadlines.
// - req: LLM request payload including prompts and generation settings.
// Returns:
// - string: generated content.
// - error: non-nil when request or decoding fails.
func (c *Client) Generate(ctx context.Context, req contract.LLMRequest) (string, error) {
	params, err := buildMessageParams(c.config, req, nil)
	if err != nil {
		return "", err
	}
	resp, err := c.client.Messages.New(ctx, params)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(extractTextBlocks(resp.Content)), nil
}

// GenerateJSON produces JSON output by forcing a tool call and returning the tool input as JSON.
// Parameters:
// - ctx: context for request cancellation and deadlines.
// - req: LLM request payload including prompts and generation settings.
// - schema: JSON schema payload describing the expected response shape.
// Returns:
// - string: generated JSON content.
// - error: non-nil when request or decoding fails, or when JSON schema is invalid.
func (c *Client) GenerateJSON(ctx context.Context, req contract.LLMRequest, schema *contract.LLMJSONSchema) (string, error) {
	tool := buildJSONTool(schema)
	toolChoice := anthropicSDK.ToolChoiceParamOfTool(tool.Name)
	params, err := buildMessageParams(c.config, req, []anthropicSDK.ToolUnionParam{{OfTool: &tool}})
	if err != nil {
		return "", err
	}
	params.ToolChoice = toolChoice

	resp, err := c.client.Messages.New(ctx, params)
	if err != nil {
		return "", err
	}
	jsonBytes, ok := extractToolInput(resp.Content, tool.Name)
	if !ok {
		return "", fmt.Errorf("anthropic did not return required tool input")
	}
	var validate any
	if err := json.Unmarshal(jsonBytes, &validate); err != nil {
		return "", fmt.Errorf("invalid json response: %w", err)
	}
	return strings.TrimSpace(string(jsonBytes)), nil
}

// GenerateStream produces streaming output using the Messages streaming API.
// Parameters:
// - ctx: context for request cancellation and deadlines.
// - req: LLM request payload including prompts and generation settings.
// Returns:
// - <-chan contract.LLMStreamChunk: channel of streamed chunks.
// - <-chan error: channel for stream errors.
func (c *Client) GenerateStream(ctx context.Context, req contract.LLMRequest) (<-chan contract.LLMStreamChunk, <-chan error) {
	chunkCh := make(chan contract.LLMStreamChunk)
	errCh := make(chan error, 1)
	params, err := buildMessageParams(c.config, req, nil)
	if err != nil {
		errCh <- err
		close(chunkCh)
		close(errCh)
		return chunkCh, errCh
	}
	go func() {
		defer close(chunkCh)
		defer close(errCh)

		stream := c.client.Messages.NewStreaming(ctx, params)
		for stream.Next() {
			event := stream.Current()
			if event.Type != "content_block_delta" {
				continue
			}
			if event.Delta.Type != "text_delta" {
				continue
			}
			if event.Delta.Text == "" {
				continue
			}
			chunkCh <- contract.LLMStreamChunk{Content: event.Delta.Text}
		}
		if err := stream.Err(); err != nil {
			errCh <- err
		}
	}()
	return chunkCh, errCh
}

// buildMessageParams constructs Anthropic message parameters for an LLM request.
// Parameters:
// - config: default Anthropic client configuration values.
// - req: LLM request payload including prompts and generation settings.
// - tools: optional tool definitions to include in the request.
// Returns:
// - anthropicSDK.MessageNewParams: parameters ready for the SDK request.
// - error: non-nil when message roles are invalid.
func buildMessageParams(config Config, req contract.LLMRequest, tools []anthropicSDK.ToolUnionParam) (anthropicSDK.MessageNewParams, error) {
	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = strings.TrimSpace(config.Model)
	}
	if model == "" {
		return anthropicSDK.MessageNewParams{}, fmt.Errorf("model is required")
	}

	messages, err := buildAnthropicMessages(req.Messages)
	if err != nil {
		return anthropicSDK.MessageNewParams{}, err
	}
	params := anthropicSDK.MessageNewParams{
		Model:    anthropicSDK.Model(model),
		Messages: messages,
		MaxTokens: func() int64 {
			if req.MaxTokens > 0 {
				return int64(req.MaxTokens)
			}
			return 1024
		}(),
	}
	if strings.TrimSpace(req.SystemInstruction) != "" {
		params.System = []anthropicSDK.TextBlockParam{{Text: req.SystemInstruction}}
	}
	if req.Temperature != 0 {
		params.Temperature = anthropicParam.NewOpt(req.Temperature)
	}
	if len(tools) > 0 {
		params.Tools = tools
	}
	return params, nil
}

// buildAnthropicMessages converts contract messages into Anthropic message params.
// Parameters:
// - messages: contract message list.
// Returns:
// - []anthropicSDK.MessageParam: Anthropic message parameters.
// - error: non-nil when a role is invalid.
func buildAnthropicMessages(messages []contract.LLMMessage) ([]anthropicSDK.MessageParam, error) {
	result := make([]anthropicSDK.MessageParam, 0, len(messages))
	for _, message := range messages {
		switch message.Role {
		case "user":
			result = append(result, anthropicSDK.NewUserMessage(anthropicSDK.NewTextBlock(message.Content)))
		case "assistant":
			result = append(result, anthropicSDK.NewAssistantMessage(anthropicSDK.NewTextBlock(message.Content)))
		default:
			return nil, fmt.Errorf("invalid message role: %s", message.Role)
		}
	}
	return result, nil
}

// extractTextBlocks collects "text" content blocks into a single string.
// Parameters:
// - blocks: the Anthropic content blocks.
// Returns:
// - string: concatenated text content.
func extractTextBlocks(blocks []anthropicSDK.ContentBlockUnion) string {
	var builder strings.Builder
	for _, block := range blocks {
		if block.Type != "text" {
			continue
		}
		builder.WriteString(block.Text)
	}
	return builder.String()
}

// buildJSONTool constructs a strict JSON tool definition from a JSON schema.
// Parameters:
// - schema: optional JSON schema; when nil, a permissive object schema is used.
// Returns:
// - anthropicSDK.ToolParam: tool definition.
func buildJSONTool(schema *contract.LLMJSONSchema) anthropicSDK.ToolParam {
	name := "response"
	if schema != nil && strings.TrimSpace(schema.Name) != "" {
		name = strings.TrimSpace(schema.Name)
	}

	inputSchema := anthropicSDK.ToolInputSchemaParam{
		Properties: map[string]any{},
		Required:   []string{},
	}
	inputSchema.ExtraFields = map[string]any{}

	if schema != nil && schema.Schema != nil {
		// Map commonly used JSON Schema fields into the Anthropic tool input schema.
		if properties, ok := schema.Schema["properties"]; ok {
			inputSchema.Properties = properties
		}
		if required, ok := schema.Schema["required"]; ok {
			if list, ok := required.([]string); ok {
				inputSchema.Required = list
			} else if list, ok := required.([]any); ok {
				converted := make([]string, 0, len(list))
				for _, item := range list {
					value, ok := item.(string)
					if !ok {
						continue
					}
					converted = append(converted, value)
				}
				inputSchema.Required = converted
			}
		}
		for key, value := range schema.Schema {
			if key == "properties" || key == "required" || key == "type" {
				continue
			}
			inputSchema.ExtraFields[key] = value
		}
	} else {
		inputSchema.ExtraFields["additionalProperties"] = true
	}

	tool := anthropicSDK.ToolParam{
		Name:        name,
		InputSchema: inputSchema,
		Strict:      anthropicParam.NewOpt(true),
	}
	return tool
}

// extractToolInput finds the tool input JSON for a named tool.
// Parameters:
// - blocks: content blocks returned by Anthropic.
// - toolName: tool name to match.
// Returns:
// - []byte: tool input JSON bytes.
// - bool: true when a matching tool block is found.
func extractToolInput(blocks []anthropicSDK.ContentBlockUnion, toolName string) ([]byte, bool) {
	for _, block := range blocks {
		if block.Type != "tool_use" {
			continue
		}
		if block.Name != toolName {
			continue
		}
		if len(block.Input) == 0 {
			return []byte("{}"), true
		}
		return []byte(block.Input), true
	}
	return nil, false
}
