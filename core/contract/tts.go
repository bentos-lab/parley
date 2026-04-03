package contract

import "context"

// TTS defines a text-to-speech contract.
type TTS interface {
	// Synthesize converts text into WAV audio bytes.
	// Parameters:
	// - ctx: context for request cancellation and deadlines.
	// - text: content to synthesize.
	// - voiceName: optional voice identifier for providers that support multiple voices.
	// Returns:
	// - []byte: synthesized WAV bytes.
	// - error: non-nil when synthesis fails.
	Synthesize(ctx context.Context, text string, voiceName string) ([]byte, error)
	// AgentVoices returns the available voice catalog keyed by voice name.
	// Parameters: none.
	// Returns:
	// - map[string]string: mapping of voice name to voice description.
	AgentVoices() map[string]string
}
