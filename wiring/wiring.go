package wiring

import (
	"github.com/bentos-lab/parley/adapter/outbound/llm"
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
	})
	ttsResolver := tts.NewResolver(tts.Config{
		InworldAPIKey: cfg.InworldAPIKey,
		InworldModel:  cfg.InworldModel,
	})
	voiceAssigner := core.NewVoiceAssigner(llmResolver, cfg.LLMProvider, cfg.OpenAI.Model)
	usecases := &Usecases{
		CreateDebate: &core.CreateDebateUsecase{DefaultTTSProvider: cfg.TTSProvider},
		GenerateDebateName: &core.GenerateDebateNameUsecase{
			LLMResolver: llmResolver,
			LLMProvider: cfg.LLMProvider,
			Model:       cfg.OpenAI.Model,
		},
		GenerateDebateAgents: &core.GenerateAgentsUsecase{
			LLMResolver: llmResolver,
			LLMProvider: cfg.LLMProvider,
			Model:       cfg.OpenAI.Model,
		},
		GenerateDebateSummary: &core.GenerateDebateSummaryUsecase{
			LLMResolver: llmResolver,
			LLMProvider: cfg.LLMProvider,
			Model:       cfg.OpenAI.Model,
		},
		AssignDebateVoices: &core.AssignDebateVoicesUsecase{
			TTSResolver: ttsResolver,
			VoiceAssn:   voiceAssigner,
		},
		LoadDebate:   &core.LoadDebateUsecase{},
		ListDebates:  &core.ListDebatesUsecase{},
		UpdateDebate: &core.UpdateDebateUsecase{},
		DeleteDebate: &core.DeleteDebateUsecase{},
		CreateRound: &core.CreateRoundUsecase{
			LLMResolver: llmResolver,
			LLMProvider: cfg.LLMProvider,
			Model:       cfg.OpenAI.Model,
		},
		GenerateAudio: &core.GenerateAudioUsecase{
			TTSResolver:        ttsResolver,
			VoiceAssn:          voiceAssigner,
			DefaultTTSProvider: cfg.TTSProvider,
		},
		GetRoundAudio: &core.GetRoundAudioUsecase{
			TTSResolver:        ttsResolver,
			VoiceAssn:          voiceAssigner,
			DefaultTTSProvider: cfg.TTSProvider,
		},
		GetDebateAudio: &core.GetDebateAudioUsecase{
			TTSResolver:        ttsResolver,
			VoiceAssn:          voiceAssigner,
			DefaultTTSProvider: cfg.TTSProvider,
		},
		ParseParleyCommand: &core.ParseParleyCommandUsecase{
			LLMResolver: llmResolver,
			LLMProvider: cfg.LLMProvider,
			Model:       cfg.OpenAI.Model,
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
