package core

import "fmt"

// resolveTTSProvider determines which provider to use for TTS operations.
// Parameters:
// - override: explicit provider supplied by the caller.
// - stored: provider persisted with the debate (may be empty).
// - defaultProvider: configuration default injector used when no other value exists.
// Returns:
// - string: the resolved provider name when available.
// - error: non-nil when all inputs are empty.
func resolveTTSProvider(override string, stored string, defaultProvider string) (string, error) {
	if override != "" {
		return override, nil
	}
	if stored != "" {
		return stored, nil
	}
	if defaultProvider != "" {
		return defaultProvider, nil
	}
	return "", fmt.Errorf("tts_provider is required")
}
