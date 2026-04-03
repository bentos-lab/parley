package config

import (
	"fmt"
	"os"
	"strings"
)

// Config stores all configuration values for the application.
type Config struct {
	LLMProvider   string
	TTSProvider   string
	OpenAI        OpenAIConfig
	InworldAPIKey string
	InworldModel  string
}

type OpenAIConfig struct {
	BaseURL string
	APIKey  string
	Model   string
}

const (
	defaultLLMProvider  = "openai"
	defaultTTSProvider  = "native"
	defaultLLMBaseURL   = "https://api.openai.com/v1"
	defaultLLMModel     = "gpt-4.1-mini"
	defaultInworldModel = "inworld-tts-1.5-max"
)

type fileConfig struct {
	LLM struct {
		Provider string `toml:"provider"`
		OpenAI   struct {
			BaseURL string `toml:"base_url"`
			APIKey  string `toml:"api_key"`
			Model   string `toml:"model"`
		} `toml:"openai"`
	} `toml:"llm"`
	TTS struct {
		Provider string `toml:"provider"`
		Inworld  struct {
			APIKey string `toml:"api_key"`
			Model  string `toml:"model"`
		} `toml:"inworld"`
	} `toml:"tts"`
}

// Load reads configuration from a TOML config file with sane defaults,
// then applies environment overrides.
// It returns a Config object or an error if parsing fails.
func Load() (Config, error) {
	fileCfg, err := loadFileConfig()
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		LLMProvider: strings.TrimSpace(fileCfg.LLM.Provider),
		TTSProvider: strings.TrimSpace(fileCfg.TTS.Provider),
		OpenAI: OpenAIConfig{
			BaseURL: fileCfg.LLM.OpenAI.BaseURL,
			APIKey:  fileCfg.LLM.OpenAI.APIKey,
			Model:   fileCfg.LLM.OpenAI.Model,
		},
		InworldAPIKey: fileCfg.TTS.Inworld.APIKey,
		InworldModel:  fileCfg.TTS.Inworld.Model,
	}
	if cfg.LLMProvider == "" {
		cfg.LLMProvider = defaultLLMProvider
	}
	if cfg.TTSProvider == "" {
		cfg.TTSProvider = defaultTTSProvider
	}
	if cfg.OpenAI.BaseURL == "" {
		cfg.OpenAI.BaseURL = defaultLLMBaseURL
	}
	if cfg.OpenAI.Model == "" {
		cfg.OpenAI.Model = defaultLLMModel
	}
	if cfg.InworldModel == "" {
		cfg.InworldModel = defaultInworldModel
	}
	if value := os.Getenv("OPENAI_BASE_URL"); value != "" {
		cfg.OpenAI.BaseURL = value
	}
	if value := os.Getenv("OPENAI_API_KEY"); value != "" {
		cfg.OpenAI.APIKey = value
	}
	if value := os.Getenv("OPENAI_MODEL"); value != "" {
		cfg.OpenAI.Model = value
	}
	if value := os.Getenv("INWORLD_API_KEY"); value != "" {
		cfg.InworldAPIKey = value
	}
	if value := os.Getenv("INWORLD_MODEL"); value != "" {
		cfg.InworldModel = value
	}
	if !isSupportedLLMProvider(cfg.LLMProvider) {
		return Config{}, fmt.Errorf("unsupported llm provider: %s", cfg.LLMProvider)
	}
	switch cfg.TTSProvider {
	case "native", "inworld":
	default:
		return Config{}, fmt.Errorf("unsupported tts provider: %s", cfg.TTSProvider)
	}

	return cfg, nil
}

// Validate returns an error if required configuration is missing.
func (c Config) Validate() error {
	if c.OpenAI.APIKey == "" {
		return fmt.Errorf("OPENAI_API_KEY is required")
	}
	return nil
}

func isSupportedLLMProvider(provider string) bool {
	return provider == defaultLLMProvider
}
