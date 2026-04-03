package llm

import (
	"fmt"

	"github.com/bentos-lab/parley/adapter/outbound/llm/openai"
	"github.com/bentos-lab/parley/core/contract"
)

// ProviderOpenAI is the only supported LLM provider name.
const ProviderOpenAI = "openai"

// Config holds the LLM resolver configuration.
type Config struct {
	OpenAI openai.Config
}

// Resolver resolves LLM providers by name.
type Resolver struct {
	config Config
}

// NewResolver creates a new LLM resolver.
func NewResolver(config Config) *Resolver {
	return &Resolver{config: config}
}

// Resolve returns the LLM client for the named provider.
func (r *Resolver) Resolve(name string) (contract.LLM, error) {
	if name != ProviderOpenAI {
		return nil, fmt.Errorf("unknown llm provider: %s", name)
	}
	return openai.NewClient(r.config.OpenAI), nil
}
