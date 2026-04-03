package core

import (
	"context"
	"fmt"

	"github.com/bentos-lab/parley/core/contract"
	"github.com/bentos-lab/parley/core/debate"
)

// GetRoundAudioInput defines the inputs for fetching round audio.
type GetRoundAudioInput struct {
	Filename string
	Index    int
}

// GetRoundAudioOutput is the result of fetching round audio.
type GetRoundAudioOutput struct {
	Path string
}

// GetRoundAudioUsecase ensures round audio exists and returns its path.
type GetRoundAudioUsecase struct {
	TTSResolver        contract.Resolver[contract.TTS]
	VoiceAssn          contract.AssignVoices
	DefaultTTSProvider string // fallback provider when the stored debate value is empty.
}

// Execute ensures a single round has audio and returns the audio path.
func (u *GetRoundAudioUsecase) Execute(ctx context.Context, input GetRoundAudioInput) (GetRoundAudioOutput, error) {
	debateItem, err := debate.LoadDebate(input.Filename)
	if err != nil {
		return GetRoundAudioOutput{}, err
	}
	if input.Index < 0 || input.Index >= len(debateItem.Rounds) {
		return GetRoundAudioOutput{}, fmt.Errorf("round index out of range")
	}
	if u.TTSResolver == nil {
		return GetRoundAudioOutput{}, fmt.Errorf("tts resolver is required")
	}
	provider, err := resolveTTSProvider("", debateItem.TTSProvider, u.DefaultTTSProvider)
	if err != nil {
		return GetRoundAudioOutput{}, err
	}
	ttsClient, err := u.TTSResolver.Resolve(provider)
	if err != nil {
		return GetRoundAudioOutput{}, err
	}
	if err := u.assignVoicesIfNeeded(ctx, debateItem, ttsClient); err != nil {
		return GetRoundAudioOutput{}, err
	}
	path, err := debateItem.SynthesizeRound(ctx, ttsClient, input.Index)
	if err != nil {
		return GetRoundAudioOutput{}, err
	}
	return GetRoundAudioOutput{Path: path}, nil
}

// assignVoicesIfNeeded assigns voices using the configured voice assigner when available.
func (u *GetRoundAudioUsecase) assignVoicesIfNeeded(ctx context.Context, debateItem *debate.Debate, ttsClient contract.TTS) error {
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
