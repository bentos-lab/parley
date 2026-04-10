package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/bentos-lab/parley/adapter/outbound/tts"
	"github.com/bentos-lab/parley/adapter/outbound/tts/native"
	"github.com/bentos-lab/parley/config"
)

// ConfigRequest describes the shape of the config update payload.
type ConfigRequest struct {
	LLM *LLMConfigRequest `json:"llm,omitempty"`
	TTS *TTSConfigRequest `json:"tts,omitempty"`
}

// LLMConfigRequest captures optional LLM configuration overrides.
type LLMConfigRequest struct {
	Provider  *string                 `json:"provider,omitempty"`
	OpenAI    *OpenAIConfigRequest    `json:"openai,omitempty"`
	Anthropic *AnthropicConfigRequest `json:"anthropic,omitempty"`
	Gemini    *GeminiConfigRequest    `json:"gemini,omitempty"`
}

// OpenAIConfigRequest exposes settings for the OpenAI-compatible LLM provider.
type OpenAIConfigRequest struct {
	BaseURL *string `json:"base_url,omitempty"`
	APIKey  *string `json:"api_key,omitempty"`
	Model   *string `json:"model,omitempty"`
}

// AnthropicConfigRequest exposes settings for the Anthropic LLM provider.
type AnthropicConfigRequest struct {
	APIKey *string `json:"api_key,omitempty"`
	Model  *string `json:"model,omitempty"`
}

// GeminiConfigRequest exposes settings for the Gemini LLM provider.
type GeminiConfigRequest struct {
	APIKey *string `json:"api_key,omitempty"`
	Model  *string `json:"model,omitempty"`
}

// TTSConfigRequest captures optional TTS configuration overrides.
type TTSConfigRequest struct {
	Provider *string               `json:"provider,omitempty"`
	Inworld  *InworldConfigRequest `json:"inworld,omitempty"`
}

// InworldConfigRequest defines the Inworld-specific TTS credentials.
type InworldConfigRequest struct {
	APIKey *string `json:"api_key,omitempty"`
	Model  *string `json:"model,omitempty"`
}

// ConfigResponse mirrors the runtime configuration returned by GET /api/config.
type ConfigResponse struct {
	LLM LLMConfigResponse `json:"llm"`
	TTS TTSConfigResponse `json:"tts"`
}

// LLMConfigResponse describes the LLM portion of the config response.
type LLMConfigResponse struct {
	Provider  string                  `json:"provider"`
	OpenAI    OpenAIConfigResponse    `json:"openai"`
	Anthropic AnthropicConfigResponse `json:"anthropic"`
	Gemini    GeminiConfigResponse    `json:"gemini"`
}

// OpenAIConfigResponse exposes the fields returned for OpenAI in the config response.
type OpenAIConfigResponse struct {
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
	Model   string `json:"model"`
}

// AnthropicConfigResponse exposes fields returned for Anthropic in the config response.
type AnthropicConfigResponse struct {
	APIKey string `json:"api_key"`
	Model  string `json:"model"`
}

// GeminiConfigResponse exposes fields returned for Gemini in the config response.
type GeminiConfigResponse struct {
	APIKey string `json:"api_key"`
	Model  string `json:"model"`
}

// TTSConfigResponse describes the TTS portion of the config response.
type TTSConfigResponse struct {
	Provider string                `json:"provider"`
	Inworld  InworldConfigResponse `json:"inworld"`
}

// InworldConfigResponse exposes the fields returned for Inworld in the config response.
type InworldConfigResponse struct {
	APIKey string `json:"api_key"`
	Model  string `json:"model"`
}

// getConfig responds with the current runtime configuration.
// Parameters: w is the response writer, r is the incoming HTTP request.
// Returns: nothing.
func (h *Handler) getConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.Load()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, newConfigResponse(cfg))
}

// updateConfig applies the provided config overrides to the persistent file.
// Parameters: w is the response writer, r is the incoming HTTP request.
// Returns: nothing.
func (h *Handler) updateConfig(w http.ResponseWriter, r *http.Request) {
	var req ConfigRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid config payload")
		return
	}
	if req.LLM == nil && req.TTS == nil {
		writeError(w, http.StatusBadRequest, "payload must include llm or tts")
		return
	}
	cfg, err := config.Load()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	cfgMap, err := config.ReadFileMap()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	finalTTSProvider := cfg.TTSProvider
	needsWrite := false
	if req.LLM != nil {
		if req.LLM.Provider != nil {
			provider := strings.TrimSpace(*req.LLM.Provider)
			switch provider {
			case "openai", "anthropic", "gemini":
				config.SetNestedValue(cfgMap, []string{"llm"}, "provider", provider)
				needsWrite = true
				if provider == "anthropic" {
					if req.LLM.Anthropic == nil || req.LLM.Anthropic.APIKey == nil || strings.TrimSpace(*req.LLM.Anthropic.APIKey) == "" {
						writeError(w, http.StatusBadRequest, "ANTHROPIC_API_KEY is required when enabling the anthropic provider")
						return
					}
					if req.LLM.Anthropic.Model == nil || strings.TrimSpace(*req.LLM.Anthropic.Model) == "" {
						writeError(w, http.StatusBadRequest, "ANTHROPIC_MODEL is required when enabling the anthropic provider")
						return
					}
				}
				if provider == "gemini" {
					if req.LLM.Gemini == nil || req.LLM.Gemini.APIKey == nil || strings.TrimSpace(*req.LLM.Gemini.APIKey) == "" {
						writeError(w, http.StatusBadRequest, "GEMINI_API_KEY is required when enabling the gemini provider")
						return
					}
					if req.LLM.Gemini.Model == nil || strings.TrimSpace(*req.LLM.Gemini.Model) == "" {
						writeError(w, http.StatusBadRequest, "GEMINI_MODEL is required when enabling the gemini provider")
						return
					}
				}
			default:
				writeError(w, http.StatusBadRequest, fmt.Sprintf("unsupported llm provider: %s", provider))
				return
			}
		}
		if req.LLM.OpenAI != nil {
			if req.LLM.OpenAI.BaseURL != nil {
				config.SetNestedValue(cfgMap, []string{"llm", "openai"}, "base_url", *req.LLM.OpenAI.BaseURL)
				needsWrite = true
			}
			if req.LLM.OpenAI.APIKey != nil {
				config.SetNestedValue(cfgMap, []string{"llm", "openai"}, "api_key", *req.LLM.OpenAI.APIKey)
				needsWrite = true
			}
			if req.LLM.OpenAI.Model != nil {
				config.SetNestedValue(cfgMap, []string{"llm", "openai"}, "model", *req.LLM.OpenAI.Model)
				needsWrite = true
			}
		}
		if req.LLM.Anthropic != nil {
			if req.LLM.Anthropic.APIKey != nil {
				config.SetNestedValue(cfgMap, []string{"llm", "anthropic"}, "api_key", *req.LLM.Anthropic.APIKey)
				needsWrite = true
			}
			if req.LLM.Anthropic.Model != nil {
				config.SetNestedValue(cfgMap, []string{"llm", "anthropic"}, "model", *req.LLM.Anthropic.Model)
				needsWrite = true
			}
		}
		if req.LLM.Gemini != nil {
			if req.LLM.Gemini.APIKey != nil {
				config.SetNestedValue(cfgMap, []string{"llm", "gemini"}, "api_key", *req.LLM.Gemini.APIKey)
				needsWrite = true
			}
			if req.LLM.Gemini.Model != nil {
				config.SetNestedValue(cfgMap, []string{"llm", "gemini"}, "model", *req.LLM.Gemini.Model)
				needsWrite = true
			}
		}
	}
	if req.TTS != nil {
		if req.TTS.Provider != nil {
			switch strings.TrimSpace(*req.TTS.Provider) {
			case tts.ProviderNative:
				finalTTSProvider = tts.ProviderNative
				config.SetNestedValue(cfgMap, []string{"tts"}, "provider", tts.ProviderNative)
				needsWrite = true
			case tts.ProviderInworld:
				finalTTSProvider = tts.ProviderInworld
				config.SetNestedValue(cfgMap, []string{"tts"}, "provider", tts.ProviderInworld)
				needsWrite = true
				if req.TTS.Inworld == nil || req.TTS.Inworld.APIKey == nil || strings.TrimSpace(*req.TTS.Inworld.APIKey) == "" {
					writeError(w, http.StatusBadRequest, "INWORLD_API_KEY is required when enabling the inworld provider")
					return
				}
			default:
				writeError(w, http.StatusBadRequest, fmt.Sprintf("unsupported tts provider: %s", *req.TTS.Provider))
				return
			}
		}
		if req.TTS.Inworld != nil {
			if req.TTS.Inworld.APIKey != nil {
				config.SetNestedValue(cfgMap, []string{"tts", "inworld"}, "api_key", *req.TTS.Inworld.APIKey)
				needsWrite = true
			}
			if req.TTS.Inworld.Model != nil {
				config.SetNestedValue(cfgMap, []string{"tts", "inworld"}, "model", *req.TTS.Inworld.Model)
				needsWrite = true
			}
		}
	}
	if !needsWrite {
		writeError(w, http.StatusBadRequest, "no configurable keys provided")
		return
	}
	if req.TTS != nil && finalTTSProvider == tts.ProviderNative {
		if err := h.ensureNativeTool(); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	if err := config.WriteFileMap(cfgMap); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	updatedCfg, err := config.Load()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, newConfigResponse(updatedCfg))
}

// newConfigResponse builds the API response object from the loaded configuration.
// Parameters: cfg is the loaded configuration.
// Returns: the config response payload.
func newConfigResponse(cfg config.Config) ConfigResponse {
	return ConfigResponse{
		LLM: LLMConfigResponse{
			Provider: cfg.LLMProvider,
			OpenAI: OpenAIConfigResponse{
				BaseURL: cfg.OpenAI.BaseURL,
				APIKey:  cfg.OpenAI.APIKey,
				Model:   cfg.OpenAI.Model,
			},
			Anthropic: AnthropicConfigResponse{
				APIKey: cfg.Anthropic.APIKey,
				Model:  cfg.Anthropic.Model,
			},
			Gemini: GeminiConfigResponse{
				APIKey: cfg.Gemini.APIKey,
				Model:  cfg.Gemini.Model,
			},
		},
		TTS: TTSConfigResponse{
			Provider: cfg.TTSProvider,
			Inworld: InworldConfigResponse{
				APIKey: cfg.InworldAPIKey,
				Model:  cfg.InworldModel,
			},
		},
	}
}

// ensureNativeTool verifies the native TTS executable is available.
// Parameters: none (uses handler hooks and native info).
// Returns: an error if lookup or installation fails.
func (h *Handler) ensureNativeTool() error {
	info := native.CurrentInfo()
	if info.Executable == "" {
		return nil
	}
	if _, err := h.lookPath(info.Executable); err == nil {
		return nil
	}
	if info.InstallCommand == "" {
		return fmt.Errorf("native TTS tool %s is unavailable", info.Label)
	}
	if err := h.runInstall(info.InstallCommand); err != nil {
		return fmt.Errorf("install native TTS tool: %w", err)
	}
	if _, err := h.lookPath(info.Executable); err != nil {
		return fmt.Errorf("native TTS tool %s is still unavailable", info.Label)
	}
	return nil
}
