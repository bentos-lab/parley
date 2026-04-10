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
	Anthropic     AnthropicConfig
	Gemini        GeminiConfig
	InworldAPIKey string
	InworldModel  string
}

type OpenAIConfig struct {
	BaseURL string
	APIKey  string
	Model   string
}

type AnthropicConfig struct {
	APIKey string
	Model  string
}

type GeminiConfig struct {
	APIKey string
	Model  string
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
		Anthropic struct {
			APIKey string `toml:"api_key"`
			Model  string `toml:"model"`
		} `toml:"anthropic"`
		Gemini struct {
			APIKey string `toml:"api_key"`
			Model  string `toml:"model"`
		} `toml:"gemini"`
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
		Anthropic: AnthropicConfig{
			APIKey: fileCfg.LLM.Anthropic.APIKey,
			Model:  fileCfg.LLM.Anthropic.Model,
		},
		Gemini: GeminiConfig{
			APIKey: fileCfg.LLM.Gemini.APIKey,
			Model:  fileCfg.LLM.Gemini.Model,
		},
		InworldAPIKey: fileCfg.TTS.Inworld.APIKey,
		InworldModel:  fileCfg.TTS.Inworld.Model,
	}
	if cfg.TTSProvider == "" {
		cfg.TTSProvider = defaultTTSProvider
	}
	if cfg.OpenAI.BaseURL == "" {
		cfg.OpenAI.BaseURL = defaultLLMBaseURL
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
	if value := os.Getenv("ANTHROPIC_API_KEY"); value != "" {
		cfg.Anthropic.APIKey = value
	}
	if value := os.Getenv("GEMINI_API_KEY"); value != "" {
		cfg.Gemini.APIKey = value
	}
	if value := os.Getenv("INWORLD_API_KEY"); value != "" {
		cfg.InworldAPIKey = value
	}
	if value := os.Getenv("INWORLD_MODEL"); value != "" {
		cfg.InworldModel = value
	}
	if cfg.LLMProvider == "" {
		if value := strings.TrimSpace(os.Getenv("LLM_PROVIDER")); value != "" {
			cfg.LLMProvider = value
		} else {
			cfg.LLMProvider = defaultLLMProvider
		}
	}
	if cfg.OpenAI.Model == "" {
		if value := os.Getenv("OPENAI_MODEL"); value != "" {
			cfg.OpenAI.Model = value
		} else {
			cfg.OpenAI.Model = defaultLLMModel
		}
	}
	if cfg.Anthropic.Model == "" {
		if value := os.Getenv("ANTHROPIC_MODEL"); value != "" {
			cfg.Anthropic.Model = value
		}
	}
	if cfg.Gemini.Model == "" {
		if value := os.Getenv("GEMINI_MODEL"); value != "" {
			cfg.Gemini.Model = value
		}
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
	switch c.LLMProvider {
	case "openai":
		if c.OpenAI.APIKey == "" {
			return fmt.Errorf("OPENAI_API_KEY is required")
		}
	case "anthropic":
		if c.Anthropic.APIKey == "" {
			return fmt.Errorf("ANTHROPIC_API_KEY is required")
		}
		if c.Anthropic.Model == "" {
			return fmt.Errorf("ANTHROPIC_MODEL is required")
		}
	case "gemini":
		if c.Gemini.APIKey == "" {
			return fmt.Errorf("GEMINI_API_KEY is required")
		}
		if c.Gemini.Model == "" {
			return fmt.Errorf("GEMINI_MODEL is required")
		}
	default:
		return fmt.Errorf("unsupported llm provider: %s", c.LLMProvider)
	}
	return nil
}

func isSupportedLLMProvider(provider string) bool {
	switch provider {
	case "openai", "anthropic", "gemini":
		return true
	default:
		return false
	}
}
