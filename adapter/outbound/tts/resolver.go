package tts

import (
	"fmt"

	"github.com/bentos-lab/parley/adapter/outbound/tts/inworld"
	"github.com/bentos-lab/parley/adapter/outbound/tts/native"
	"github.com/bentos-lab/parley/core/contract"
)

const (
	// ProviderNative is the native TTS provider name.
	ProviderNative = "native"
	// ProviderInworld is the Inworld TTS provider name.
	ProviderInworld = "inworld"
)

// Config holds resolver configuration for TTS providers.
type Config struct {
	InworldAPIKey string
	InworldModel  string
}

// Resolver resolves TTS providers by name.
type Resolver struct {
	config Config
}

// NewResolver creates a new TTS resolver.
func NewResolver(config Config) *Resolver {
	return &Resolver{config: config}
}

// Resolve returns the TTS client for the named provider.
func (r *Resolver) Resolve(name string) (contract.TTS, error) {
	switch name {
	case ProviderNative:
		return native.NewClient(), nil
	case ProviderInworld:
		if r.config.InworldAPIKey == "" {
			return nil, fmt.Errorf("INWORLD_API_KEY is required when using the inworld TTS provider")
		}
		const inworldSampleRateHz = 22050
		const inworldTemperature = 1.0
		return inworld.NewClient(inworld.Config{
			APIKey:     r.config.InworldAPIKey,
			ModelID:    r.config.InworldModel,
			SampleRate: inworldSampleRateHz,
			Temp:       inworldTemperature,
		}), nil
	default:
		return nil, fmt.Errorf("unknown TTS provider: %s", name)
	}
}
