package core

import (
	"context"
	"fmt"
	"strings"
)

// LLMDefaults defines default provider/model values for LLM selection.
type LLMDefaults struct {
	Provider       string
	OpenAIModel    string
	AnthropicModel string
	GeminiModel    string
}

// ResolveLLMSelection resolves the effective LLM provider and model based on inputs and defaults.
// Parameters:
// - inputProvider: provider override passed by the caller (highest precedence).
// - inputModel: model override passed by the caller (highest precedence).
// - storedProvider: provider stored in the debate payload (middle precedence).
// - storedModel: model stored in the debate payload (middle precedence).
// - defaults: application defaults used as a fallback (lowest precedence).
// Returns:
// - provider: resolved provider string.
// - model: resolved model string.
// - error: non-nil when the provider is missing/unsupported or model is required but empty.
func ResolveLLMSelection(inputProvider string, inputModel string, storedProvider string, storedModel string, defaults LLMDefaults) (provider string, model string, err error) {
	provider = strings.TrimSpace(inputProvider)
	if provider == "" {
		provider = strings.TrimSpace(storedProvider)
	}
	if provider == "" {
		provider = strings.TrimSpace(defaults.Provider)
	}
	if provider == "" {
		return "", "", fmt.Errorf("llm provider is required")
	}

	model = strings.TrimSpace(inputModel)
	if model == "" {
		model = strings.TrimSpace(storedModel)
	}
	if model == "" {
		model = strings.TrimSpace(defaultModelForProvider(provider, defaults))
	}
	if model == "" {
		return "", "", fmt.Errorf("llm model is required for provider %q", provider)
	}
	return provider, model, nil
}

// defaultModelForProvider selects the default model for a provider.
// Parameters:
// - provider: the resolved provider name.
// - defaults: defaults that may contain a provider-specific model.
// Returns:
// - string: the provider-specific default model, or empty when unknown.
func defaultModelForProvider(provider string, defaults LLMDefaults) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "openai":
		return defaults.OpenAIModel
	case "anthropic":
		return defaults.AnthropicModel
	case "gemini":
		return defaults.GeminiModel
	default:
		return ""
	}
}

type llmSelectionContextKey struct{}

type llmSelection struct {
	Provider string
	Model    string
}

// WithLLMSelection attaches LLM provider/model overrides to the context.
// Parameters:
// - ctx: the parent context.
// - provider: optional provider override.
// - model: optional model override.
// Returns:
// - context.Context: derived context containing the overrides.
func WithLLMSelection(ctx context.Context, provider string, model string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, llmSelectionContextKey{}, llmSelection{
		Provider: strings.TrimSpace(provider),
		Model:    strings.TrimSpace(model),
	})
}

// LLMSelectionFromContext reads LLM provider/model overrides from the context.
// Parameters:
// - ctx: context potentially containing LLM selection overrides.
// Returns:
// - provider: provider override (empty when unset).
// - model: model override (empty when unset).
func LLMSelectionFromContext(ctx context.Context) (provider string, model string) {
	if ctx == nil {
		return "", ""
	}
	value, ok := ctx.Value(llmSelectionContextKey{}).(llmSelection)
	if !ok {
		return "", ""
	}
	return value.Provider, value.Model
}
