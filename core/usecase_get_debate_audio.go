package core

import (
	"context"
	"fmt"

	"github.com/bentos-lab/parley/core/contract"
	"github.com/bentos-lab/parley/core/debate"
)

// GetDebateAudioInput defines the inputs for fetching debate audio.
type GetDebateAudioInput struct {
	Filename string
}

// GetDebateAudioOutput is the result of fetching debate audio.
type GetDebateAudioOutput struct {
	Path string
}

// GetDebateAudioUsecase ensures debate audio exists and returns its path.
type GetDebateAudioUsecase struct {
	TTSResolver        contract.Resolver[contract.TTS]
	VoiceAssn          contract.AssignVoices
	DefaultTTSProvider string // fallback provider when the stored debate value is empty.
}

// Execute ensures audio exists for the debate and returns the file path.
func (u *GetDebateAudioUsecase) Execute(ctx context.Context, input GetDebateAudioInput) (GetDebateAudioOutput, error) {
	debateItem, err := debate.LoadDebate(input.Filename)
	if err != nil {
		return GetDebateAudioOutput{}, err
	}
	if len(debateItem.Rounds) == 0 {
		return GetDebateAudioOutput{}, fmt.Errorf("no rounds to synthesize")
	}
	if u.TTSResolver == nil {
		return GetDebateAudioOutput{}, fmt.Errorf("tts resolver is required")
	}
	provider, err := resolveTTSProvider("", debateItem.TTSProvider, u.DefaultTTSProvider)
	if err != nil {
		return GetDebateAudioOutput{}, err
	}
	ttsClient, err := u.TTSResolver.Resolve(provider)
	if err != nil {
		return GetDebateAudioOutput{}, err
	}
	if err := u.assignVoicesIfNeeded(ctx, debateItem, ttsClient); err != nil {
		return GetDebateAudioOutput{}, err
	}
	expectedPath, err := debateItem.DebateAudioPath()
	if err != nil {
		return GetDebateAudioOutput{}, err
	}
	exists, err := fileExists(expectedPath)
	if err != nil {
		return GetDebateAudioOutput{}, err
	}
	if exists {
		return GetDebateAudioOutput{Path: expectedPath}, nil
	}
	path, err := debateItem.Synthesize(ctx, ttsClient)
	if err != nil {
		return GetDebateAudioOutput{}, err
	}
	return GetDebateAudioOutput{Path: path}, nil
}

// assignVoicesIfNeeded assigns voices using the configured voice assigner when available.
func (u *GetDebateAudioUsecase) assignVoicesIfNeeded(ctx context.Context, debateItem *debate.Debate, ttsClient contract.TTS) error {
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
