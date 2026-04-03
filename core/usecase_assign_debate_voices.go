package core

import (
	"context"
	"fmt"

	"github.com/bentos-lab/parley/core/contract"
	"github.com/bentos-lab/parley/core/debate"
)

// AssignDebateVoicesInput defines inputs for assigning debate voices.
type AssignDebateVoicesInput struct {
	Agents      []debate.DebateAgent
	TTSProvider string
	AgentVoices map[string]string
}

// AssignDebateVoicesOutput is the result of assigning debate voices.
type AssignDebateVoicesOutput struct {
	Agents []debate.DebateAgent
}

// AssignDebateVoicesUsecase applies explicit voices and auto-assigns when needed.
type AssignDebateVoicesUsecase struct {
	TTSResolver contract.Resolver[contract.TTS]
	VoiceAssn   contract.AssignVoices
}

// Execute applies voice overrides and auto-assignment when required.
func (u *AssignDebateVoicesUsecase) Execute(ctx context.Context, input AssignDebateVoicesInput) (AssignDebateVoicesOutput, error) {
	if len(input.Agents) == 0 {
		return AssignDebateVoicesOutput{}, fmt.Errorf("agents are required")
	}
	agents := append([]debate.DebateAgent(nil), input.Agents...)
	assignMissingAgentIDs(agents)
	if err := u.applyVoices(agents, input); err != nil {
		return AssignDebateVoicesOutput{}, err
	}
	if input.TTSProvider != "" {
		if u.TTSResolver == nil {
			return AssignDebateVoicesOutput{}, fmt.Errorf("tts resolver is required")
		}
		ttsClient, err := u.TTSResolver.Resolve(input.TTSProvider)
		if err != nil {
			return AssignDebateVoicesOutput{}, err
		}
		if err := u.assignVoicesIfNeeded(ctx, agents, ttsClient); err != nil {
			return AssignDebateVoicesOutput{}, err
		}
	}
	return AssignDebateVoicesOutput{Agents: agents}, nil
}

func (u *AssignDebateVoicesUsecase) applyVoices(agents []debate.DebateAgent, input AssignDebateVoicesInput) error {
	if input.TTSProvider != "" {
		if u.TTSResolver == nil {
			return fmt.Errorf("tts resolver is required")
		}
		if _, err := u.TTSResolver.Resolve(input.TTSProvider); err != nil {
			return err
		}
	}
	if len(input.AgentVoices) == 0 && !hasAnyVoiceName(agents) {
		return nil
	}
	provider := input.TTSProvider
	if provider == "" {
		return fmt.Errorf("tts_provider is required when setting agent voices")
	}
	if u.TTSResolver == nil {
		return fmt.Errorf("tts resolver is required")
	}
	ttsClient, err := u.TTSResolver.Resolve(provider)
	if err != nil {
		return err
	}
	voices := ttsClient.AgentVoices()
	if len(voices) == 0 {
		return fmt.Errorf("tts provider does not support agent voices")
	}
	if err := validateAgentVoices(voices, input.AgentVoices); err != nil {
		return err
	}
	if err := applyAgentVoices(agents, input.AgentVoices); err != nil {
		return err
	}
	if err := validateAgentsVoiceNames(voices, agents); err != nil {
		return err
	}
	return nil
}

// assignVoicesIfNeeded assigns voices using the configured voice assigner when available.
func (u *AssignDebateVoicesUsecase) assignVoicesIfNeeded(ctx context.Context, agents []debate.DebateAgent, ttsClient contract.TTS) error {
	if u.VoiceAssn == nil {
		return nil
	}
	voices := ttsClient.AgentVoices()
	if len(voices) == 0 {
		return nil
	}
	if !needsVoiceAssignment(voices, agents) {
		return nil
	}
	agentSpecs := buildAgentSpecMap(agents)
	assigned, err := u.VoiceAssn.AssignVoices(ctx, voices, agentSpecs)
	if err != nil {
		return err
	}
	applyAssignedVoices(agents, assigned)
	return nil
}
