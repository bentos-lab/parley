package contract

import "context"

// LLMRequest defines the inputs for LLM generation.
type LLMRequest struct {
	SystemInstruction string
	Messages          []LLMMessage
	Temperature       float64
	Model             string
	MaxTokens         int
}

// LLMMessage represents a single role-based message for LLM input.
type LLMMessage struct {
	Role    string
	Content string
}

// LLMJSONSchema defines a JSON schema payload for structured LLM responses.
type LLMJSONSchema struct {
	Name   string
	Schema map[string]any
	Strict *bool
}

// LLMStreamChunk represents a chunk of streamed LLM content.
type LLMStreamChunk struct {
	Content string
}

// LLM defines a contract for large language model generation.
type LLM interface {
	// Generate produces free-form text from the input request.
	// Parameters:
	// - ctx: context for request cancellation and deadlines.
	// - req: LLM request payload including prompts and generation settings.
	// Returns:
	// - string: generated content.
	// - error: non-nil when generation fails.
	Generate(ctx context.Context, req LLMRequest) (string, error)
	// GenerateJSON produces JSON output from the input request.
	// Parameters:
	// - ctx: context for request cancellation and deadlines.
	// - req: LLM request payload including prompts and generation settings.
	// - schema: JSON schema payload describing the expected response shape.
	// Returns:
	// - string: generated JSON content.
	// - error: non-nil when generation fails.
	GenerateJSON(ctx context.Context, req LLMRequest, schema *LLMJSONSchema) (string, error)
	// GenerateStream produces streaming output from the input request.
	// Parameters:
	// - ctx: context for request cancellation and deadlines.
	// - req: LLM request payload including prompts and generation settings.
	// Returns:
	// - <-chan LLMStreamChunk: channel of streamed chunks.
	// - <-chan error: channel for stream errors.
	GenerateStream(ctx context.Context, req LLMRequest) (<-chan LLMStreamChunk, <-chan error)
}
