package llm

import (
	"fmt"

	"github.com/bentos-lab/parley/adapter/outbound/llm/anthropic"
	"github.com/bentos-lab/parley/adapter/outbound/llm/gemini"
	"github.com/bentos-lab/parley/adapter/outbound/llm/openai"
	"github.com/bentos-lab/parley/core/contract"
)

const (
	// ProviderOpenAI is the OpenAI provider name.
	ProviderOpenAI = "openai"
	// ProviderAnthropic is the Anthropic provider name.
	ProviderAnthropic = "anthropic"
	// ProviderGemini is the Gemini provider name.
	ProviderGemini = "gemini"
)

// Config holds the LLM resolver configuration.
type Config struct {
	OpenAI    openai.Config
	Anthropic anthropic.Config
	Gemini    gemini.Config
}

// Resolver resolves LLM providers by name.
type Resolver struct {
	config Config
}

// NewResolver creates a new LLM resolver.
// Parameters:
// - config: resolver configuration with provider settings.
// Returns:
// - *Resolver: initialized resolver instance.
func NewResolver(config Config) *Resolver {
	return &Resolver{config: config}
}

// Resolve returns the LLM client for the named provider.
// Parameters:
// - provider: provider name to resolve.
// - model: model override for the provider.
// Returns:
// - contract.LLM: resolved LLM client.
// - error: non-nil when the provider is unsupported or configuration is invalid.
func (r *Resolver) Resolve(provider string, model string) (contract.LLM, error) {
	switch provider {
	case ProviderOpenAI:
		config := r.config.OpenAI
		if model != "" {
			config.Model = model
		}
		if config.APIKey == "" {
			return nil, fmt.Errorf("OPENAI_API_KEY is required when using the openai provider")
		}
		if config.Model == "" {
			return nil, fmt.Errorf("OPENAI_MODEL is required when using the openai provider")
		}
		return openai.NewClient(config), nil
	case ProviderAnthropic:
		config := r.config.Anthropic
		if model != "" {
			config.Model = model
		}
		if config.APIKey == "" {
			return nil, fmt.Errorf("ANTHROPIC_API_KEY is required when using the anthropic provider")
		}
		if config.Model == "" {
			return nil, fmt.Errorf("ANTHROPIC_MODEL is required when using the anthropic provider")
		}
		return anthropic.NewClient(config), nil
	case ProviderGemini:
		config := r.config.Gemini
		if model != "" {
			config.Model = model
		}
		if config.APIKey == "" {
			return nil, fmt.Errorf("GEMINI_API_KEY is required when using the gemini provider")
		}
		if config.Model == "" {
			return nil, fmt.Errorf("GEMINI_MODEL is required when using the gemini provider")
		}
		return gemini.NewClient(config), nil
	default:
		return nil, fmt.Errorf("unknown llm provider: %s", provider)
	}
}
