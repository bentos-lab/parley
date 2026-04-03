package core

import (
	"context"
	"testing"

	"github.com/bentos-lab/parley/core/contract"
	"github.com/bentos-lab/parley/core/debate"
	"github.com/stretchr/testify/require"
)

type stubTTS struct {
	voices map[string]string
}

func (s *stubTTS) Synthesize(ctx context.Context, text string, voiceName string) ([]byte, error) {
	return nil, nil
}

func (s *stubTTS) AgentVoices() map[string]string {
	return s.voices
}

type stubVoiceAssn struct {
	assigned map[string]string
}

func (s *stubVoiceAssn) AssignVoices(ctx context.Context, voices map[string]string, agents map[string]contract.AgentSpec) (map[string]string, error) {
	return s.assigned, nil
}

func TestAssignDebateVoicesNoopWithoutProvider(t *testing.T) {
	t.Parallel()
	usecase := &AssignDebateVoicesUsecase{}
	agents := []debate.DebateAgent{{ID: "a1", Name: "Alex"}}
	output, err := usecase.Execute(context.Background(), AssignDebateVoicesInput{
		Agents: agents,
	})
	require.NoError(t, err)
	require.Equal(t, agents, output.Agents)
}

func TestAssignDebateVoicesRequiresProviderForOverrides(t *testing.T) {
	t.Parallel()
	usecase := &AssignDebateVoicesUsecase{}
	_, err := usecase.Execute(context.Background(), AssignDebateVoicesInput{
		Agents: []debate.DebateAgent{{ID: "a1", Name: "Alex"}},
		AgentVoices: map[string]string{
			"a1": "voice-1",
		},
	})
	require.Error(t, err)
}

func TestAssignDebateVoicesAutoAssigns(t *testing.T) {
	t.Parallel()
	tts := &stubTTS{voices: map[string]string{"voice-1": "desc"}}
	usecase := &AssignDebateVoicesUsecase{
		TTSResolver: contract.ResolverFunc[contract.TTS](func(name string) (contract.TTS, error) {
			return tts, nil
		}),
		VoiceAssn: &stubVoiceAssn{assigned: map[string]string{"a1": "voice-1"}},
	}
	output, err := usecase.Execute(context.Background(), AssignDebateVoicesInput{
		Agents:      []debate.DebateAgent{{ID: "a1", Name: "Alex"}},
		TTSProvider: "native",
	})
	require.NoError(t, err)
	require.Equal(t, "voice-1", output.Agents[0].VoiceName)
}
