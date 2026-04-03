package core

import (
	"context"
	"fmt"

	"github.com/bentos-lab/parley/core/contract"
	"github.com/bentos-lab/parley/core/debate"
)

// GenerateAudioInput defines overrides for audio generation.
type GenerateAudioInput struct {
	Filename    string
	TTSProvider string
	AgentVoices map[string]string
}

// GenerateAudioOutput is the result of audio generation.
type GenerateAudioOutput struct {
	Path string
}

// GenerateAudioUsecase generates audio with optional overrides.
type GenerateAudioUsecase struct {
	TTSResolver        contract.Resolver[contract.TTS]
	VoiceAssn          contract.AssignVoices
	DefaultTTSProvider string // default provider used when debate lacks one and override is absent.
}

// Execute produces and saves audio for a debate using optional overrides.
func (u *GenerateAudioUsecase) Execute(ctx context.Context, input GenerateAudioInput) (GenerateAudioOutput, error) {
	debateItem, err := debate.LoadDebate(input.Filename)
	if err != nil {
		return GenerateAudioOutput{}, err
	}
	provider, err := resolveTTSProvider(input.TTSProvider, debateItem.TTSProvider, u.DefaultTTSProvider)
	if err != nil {
		return GenerateAudioOutput{}, err
	}
	if u.TTSResolver == nil {
		return GenerateAudioOutput{}, fmt.Errorf("tts resolver is required")
	}
	ttsClient, err := u.TTSResolver.Resolve(provider)
	if err != nil {
		return GenerateAudioOutput{}, err
	}
	working := cloneDebate(debateItem)
	working.TTSProvider = provider
	if input.TTSProvider != "" && input.TTSProvider != debateItem.TTSProvider {
		clearAgentVoices(working)
	}
	if err := u.applyAudioOverrides(working, ttsClient, input.AgentVoices); err != nil {
		return GenerateAudioOutput{}, err
	}
	if err := u.assignVoicesIfNeeded(ctx, working, ttsClient); err != nil {
		return GenerateAudioOutput{}, err
	}
	path, err := working.Synthesize(ctx, ttsClient)
	if err != nil {
		return GenerateAudioOutput{}, err
	}
	return GenerateAudioOutput{Path: path}, nil
}

// applyAudioOverrides validates and applies voice overrides in memory for audio generation.
func (u *GenerateAudioUsecase) applyAudioOverrides(debateItem *debate.Debate, ttsClient contract.TTS, overrides map[string]string) error {
	if len(overrides) == 0 {
		return nil
	}
	voices := ttsClient.AgentVoices()
	if len(voices) == 0 {
		return fmt.Errorf("tts provider does not support agent voices")
	}
	if err := validateAgentVoices(voices, overrides); err != nil {
		return err
	}
	if err := applyAgentVoices(debateItem.Agents, overrides); err != nil {
		return err
	}
	return nil
}

// assignVoicesIfNeeded assigns voices using the configured voice assigner when available.
func (u *GenerateAudioUsecase) assignVoicesIfNeeded(ctx context.Context, debateItem *debate.Debate, ttsClient contract.TTS) error {
	if u.VoiceAssn == nil {
		return nil
	}
	voices := ttsClient.AgentVoices()
	if len(voices) == 0 {
		return nil
	}
	if !needsVoiceAssignment(voices, debateItem.Agents) {
		return nil
	}
	agents := buildAgentSpecMap(debateItem.Agents)
	assigned, err := u.VoiceAssn.AssignVoices(ctx, voices, agents)
	if err != nil {
		return err
	}
	applyAssignedVoices(debateItem.Agents, assigned)
	return nil
}
