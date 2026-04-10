package wiring

import (
	"github.com/bentos-lab/parley/adapter/outbound/llm"
	llmanthropic "github.com/bentos-lab/parley/adapter/outbound/llm/anthropic"
	llmgemini "github.com/bentos-lab/parley/adapter/outbound/llm/gemini"
	llmopenai "github.com/bentos-lab/parley/adapter/outbound/llm/openai"
	"github.com/bentos-lab/parley/adapter/outbound/tts"
	"github.com/bentos-lab/parley/config"
	"github.com/bentos-lab/parley/core"
)

// Usecases bundles all debate usecases.
type Usecases struct {
	CreateDebate          *core.CreateDebateUsecase
	GenerateDebateName    *core.GenerateDebateNameUsecase
	GenerateDebateAgents  *core.GenerateAgentsUsecase
	GenerateDebateSummary *core.GenerateDebateSummaryUsecase
	AssignDebateVoices    *core.AssignDebateVoicesUsecase
	LoadDebate            *core.LoadDebateUsecase
	ListDebates           *core.ListDebatesUsecase
	UpdateDebate          *core.UpdateDebateUsecase
	DeleteDebate          *core.DeleteDebateUsecase
	CreateRound           *core.CreateRoundUsecase
	GenerateAudio         *core.GenerateAudioUsecase
	GetRoundAudio         *core.GetRoundAudioUsecase
	GetDebateAudio        *core.GetDebateAudioUsecase
	ParseParleyCommand    *core.ParseParleyCommandUsecase
}

// BuildUsecases constructs the debate usecases with outbound adapters.
// Parameters: cfg is the application config.
// Returns: the configured usecases or an error.
func BuildUsecases(cfg config.Config) (*Usecases, error) {
	llmResolver := llm.NewResolver(llm.Config{
		OpenAI: llmopenai.Config{
			BaseURL: cfg.OpenAI.BaseURL,
			APIKey:  cfg.OpenAI.APIKey,
			Model:   cfg.OpenAI.Model,
		},
		Anthropic: llmanthropic.Config{
			APIKey: cfg.Anthropic.APIKey,
			Model:  cfg.Anthropic.Model,
		},
		Gemini: llmgemini.Config{
			APIKey: cfg.Gemini.APIKey,
			Model:  cfg.Gemini.Model,
		},
	})
	ttsResolver := tts.NewResolver(tts.Config{
		InworldAPIKey: cfg.InworldAPIKey,
		InworldModel:  cfg.InworldModel,
	})
	llmDefaults := core.LLMDefaults{
		Provider:       cfg.LLMProvider,
		OpenAIModel:    cfg.OpenAI.Model,
		AnthropicModel: cfg.Anthropic.Model,
		GeminiModel:    cfg.Gemini.Model,
	}
	voiceAssigner := core.NewVoiceAssigner(llmResolver, llmDefaults)
	usecases := &Usecases{
		CreateDebate: &core.CreateDebateUsecase{
			DefaultTTSProvider: cfg.TTSProvider,
			LLMDefaults:        llmDefaults,
		},
		GenerateDebateName: &core.GenerateDebateNameUsecase{
			LLMResolver: llmResolver,
			Defaults:    llmDefaults,
		},
		GenerateDebateAgents: &core.GenerateAgentsUsecase{
			LLMResolver: llmResolver,
			Defaults:    llmDefaults,
		},
		GenerateDebateSummary: &core.GenerateDebateSummaryUsecase{
			LLMResolver: llmResolver,
			Defaults:    llmDefaults,
		},
		AssignDebateVoices: &core.AssignDebateVoicesUsecase{
			TTSResolver: ttsResolver,
			VoiceAssn:   voiceAssigner,
			Defaults:    llmDefaults,
		},
		LoadDebate:   &core.LoadDebateUsecase{},
		ListDebates:  &core.ListDebatesUsecase{},
		UpdateDebate: &core.UpdateDebateUsecase{},
		DeleteDebate: &core.DeleteDebateUsecase{},
		CreateRound: &core.CreateRoundUsecase{
			LLMResolver: llmResolver,
			Defaults:    llmDefaults,
		},
		GenerateAudio: &core.GenerateAudioUsecase{
			TTSResolver:        ttsResolver,
			VoiceAssn:          voiceAssigner,
			DefaultTTSProvider: cfg.TTSProvider,
			Defaults:           llmDefaults,
		},
		GetRoundAudio: &core.GetRoundAudioUsecase{
			TTSResolver:        ttsResolver,
			VoiceAssn:          voiceAssigner,
			DefaultTTSProvider: cfg.TTSProvider,
			Defaults:           llmDefaults,
		},
		GetDebateAudio: &core.GetDebateAudioUsecase{
			TTSResolver:        ttsResolver,
			VoiceAssn:          voiceAssigner,
			DefaultTTSProvider: cfg.TTSProvider,
			Defaults:           llmDefaults,
		},
		ParseParleyCommand: &core.ParseParleyCommandUsecase{
			LLMResolver: llmResolver,
			Defaults:    llmDefaults,
		},
	}
	return usecases, nil
}

// LoadUsecases loads the runtime configuration and usecases for request handlers.
// Parameters: w is the response writer used for emitting error responses.
// Returns: the usecases, the loaded config, and a boolean indicating success.
func LoadUsecases() (*Usecases, config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, config.Config{}, err
	}
	usecases, err := BuildUsecases(cfg)
	if err != nil {
		return nil, config.Config{}, err
	}
	return usecases, cfg, nil
}
