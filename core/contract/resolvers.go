package contract

// LLMResolver resolves an LLM client for a provider/model selection.
type LLMResolver interface {
	// Resolve returns an LLM client for the named provider with an optional model override.
	// Parameters:
	// - provider: provider name to resolve (e.g. "openai", "anthropic", "gemini").
	// - model: optional model override; empty uses the resolver's configured default.
	// Returns:
	// - LLM: resolved LLM client.
	// - error: non-nil when the provider is unsupported or resolution fails.
	Resolve(provider string, model string) (LLM, error)
}

// LLMResolverFunc adapts a function to the LLMResolver interface.
type LLMResolverFunc func(provider string, model string) (LLM, error)

// Resolve returns an LLM client for the named provider.
// Parameters:
// - provider: provider name to resolve.
// - model: optional model override.
// Returns:
// - LLM: resolved LLM client.
// - error: non-nil when resolution fails.
func (f LLMResolverFunc) Resolve(provider string, model string) (LLM, error) {
	return f(provider, model)
}

// TTSResolver resolves a TTS client for a provider selection.
type TTSResolver interface {
	// Resolve returns a TTS client for the named provider.
	// Parameters:
	// - provider: provider name to resolve (e.g. "native", "inworld").
	// Returns:
	// - TTS: resolved TTS client.
	// - error: non-nil when the provider is unsupported or resolution fails.
	Resolve(provider string) (TTS, error)
}

// TTSResolverFunc adapts a function to the TTSResolver interface.
type TTSResolverFunc func(provider string) (TTS, error)

// Resolve returns a TTS client for the named provider.
// Parameters:
// - provider: provider name to resolve.
// Returns:
// - TTS: resolved TTS client.
// - error: non-nil when resolution fails.
func (f TTSResolverFunc) Resolve(provider string) (TTS, error) {
	return f(provider)
}
