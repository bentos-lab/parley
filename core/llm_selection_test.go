package core

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestResolveEffectiveLLMSelectionContextOverridesStored verifies inbound overrides have highest priority.
// Parameters: t is the test context.
func TestResolveEffectiveLLMSelectionContextOverridesStored(t *testing.T) {
	t.Parallel()
	defaults := LLMDefaults{
		Provider:       "openai",
		OpenAIModel:    "default-openai",
		AnthropicModel: "default-anthropic",
		GeminiModel:    "default-gemini",
	}
	ctx := WithLLMSelection(context.Background(), "gemini", "ctx-model")
	provider, model, err := ResolveEffectiveLLMSelection(ctx, "anthropic", "stored-model", defaults)
	require.NoError(t, err)
	require.Equal(t, "gemini", provider)
	require.Equal(t, "ctx-model", model)
}

// TestResolveEffectiveLLMSelectionStoredOverridesDefaults verifies stored debate values win over defaults.
// Parameters: t is the test context.
func TestResolveEffectiveLLMSelectionStoredOverridesDefaults(t *testing.T) {
	t.Parallel()
	defaults := LLMDefaults{
		Provider:       "openai",
		OpenAIModel:    "default-openai",
		AnthropicModel: "default-anthropic",
		GeminiModel:    "default-gemini",
	}
	provider, model, err := ResolveEffectiveLLMSelection(context.Background(), "anthropic", "stored-model", defaults)
	require.NoError(t, err)
	require.Equal(t, "anthropic", provider)
	require.Equal(t, "stored-model", model)
}

// TestResolveEffectiveLLMSelectionUsesDefaults verifies defaults are used when no override or stored value exists.
// Parameters: t is the test context.
func TestResolveEffectiveLLMSelectionUsesDefaults(t *testing.T) {
	t.Parallel()
	defaults := LLMDefaults{
		Provider:       "openai",
		OpenAIModel:    "default-openai",
		AnthropicModel: "default-anthropic",
		GeminiModel:    "default-gemini",
	}
	provider, model, err := ResolveEffectiveLLMSelection(context.Background(), "", "", defaults)
	require.NoError(t, err)
	require.Equal(t, "openai", provider)
	require.Equal(t, "default-openai", model)
}
