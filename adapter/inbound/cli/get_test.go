package cli_test

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	clicore "github.com/bentos-lab/parley/adapter/inbound/cli"
	"github.com/bentos-lab/parley/core"
	"github.com/bentos-lab/parley/core/debate"
	"github.com/bentos-lab/parley/wiring"
	"github.com/stretchr/testify/require"
)

type captureGetOutput struct {
	called  bool
	details clicore.DebateDetailsOutput
}

// DebateDetails captures get output payloads for assertions.
// Parameters: writer is ignored, details is the captured payload.
// Returns: nil.
func (o *captureGetOutput) DebateDetails(writer io.Writer, details clicore.DebateDetailsOutput) error {
	o.called = true
	o.details = details
	return nil
}

// TestGetLoadsDebateAndBuildsDetails verifies get loads debate data and emits all required sections.
// Parameters: t provides the test context.
// Returns: nothing.
func TestGetLoadsDebateAndBuildsDetails(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	debateID := "alpha.2026-04-10-11-20-30"

	writeDebateForGet(t, home, debate.FilenameFromID(debateID), debate.Debate{
		Name:  "City Transit",
		Topic: "Should downtown ban private cars?",
		Agents: []debate.DebateAgent{
			{ID: "agent-1", Name: "A", Stance: "For", VoiceName: "voice-a"},
			{ID: "agent-2", Name: "B", Stance: "Against", VoiceName: "voice-b"},
		},
		Rounds: []debate.DebateRound{
			{AgentID: "agent-1", Message: "Opening", Summary: "S1", Weakness: "W1", NewPoint: "N1", Rebuttal: "R1"},
			{AgentID: "", Message: "User follow-up", Summary: "S2"},
		},
		Summary: &debate.DebateSummaryDetail{
			Agents:     [][]string{{"Point A"}, {"Point B"}},
			Conclusion: "Final",
		},
		TTSProvider: "native",
		LLMProvider: "openai",
		LLMModel:    "gpt-5.4",
	})

	output := &captureGetOutput{}
	usecases := &wiring.Usecases{LoadDebate: &core.LoadDebateUsecase{}}
	runtime := clicore.RuntimeInfo{
		LLMProvider:    "openai",
		OpenAIModel:    "gpt-5.4",
		AnthropicModel: "claude-3-7-sonnet",
		GeminiModel:    "gemini-2.5-pro",
		TTSModel:       "inworld-voice-v1",
	}

	err := clicore.Get(context.Background(), usecases, output, runtime, debateID)
	require.NoError(t, err)
	require.True(t, output.called)

	require.Equal(t, debateID, output.details.Header.ID)
	require.Equal(t, "Debate", output.details.Header.AppName)
	require.Equal(t, "Should downtown ban private cars?", output.details.Topic)
	require.Equal(t, "City Transit", output.details.Name)
	require.Len(t, output.details.Agents, 2)
	require.Len(t, output.details.Rounds, 2)
	require.Equal(t, 1, output.details.Rounds[0].Number)
	require.Equal(t, "A", output.details.Rounds[0].AgentName)
	require.Equal(t, "User", output.details.Rounds[1].AgentName)
	require.Equal(t, "Final", output.details.Summary.Conclusion)
}

// TestGetMissingDebateReturnsError verifies get returns the storage error when debate is missing.
// Parameters: t provides the test context.
// Returns: nothing.
func TestGetMissingDebateReturnsError(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	output := &captureGetOutput{}
	usecases := &wiring.Usecases{LoadDebate: &core.LoadDebateUsecase{}}

	err := clicore.Get(context.Background(), usecases, output, clicore.RuntimeInfo{}, "missing.2026-04-10-00-00-00")
	require.Error(t, err)
	require.ErrorIs(t, err, os.ErrNotExist)
	require.False(t, output.called)
}

// TestGetUsesSummaryPlaceholdersWhenMissing verifies get emits read-only summary placeholders.
// Parameters: t provides the test context.
// Returns: nothing.
func TestGetUsesSummaryPlaceholdersWhenMissing(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	debateID := "nosummary.2026-04-10-00-00-00"
	writeDebateForGet(t, home, debate.FilenameFromID(debateID), debate.Debate{
		Name:  "No Summary Debate",
		Topic: "Topic",
		Agents: []debate.DebateAgent{
			{ID: "agent-1", Name: "A", Stance: "For"},
			{ID: "agent-2", Name: "B", Stance: "Against"},
		},
		Rounds: []debate.DebateRound{
			{AgentID: "agent-1", Message: "Hello"},
		},
	})

	output := &captureGetOutput{}
	usecases := &wiring.Usecases{LoadDebate: &core.LoadDebateUsecase{}}
	err := clicore.Get(context.Background(), usecases, output, clicore.RuntimeInfo{}, debateID)
	require.NoError(t, err)
	require.True(t, output.called)
	require.Len(t, output.details.Summary.Agents, 2)
	require.Empty(t, output.details.Summary.Conclusion)
}

// writeDebateForGet writes debate fixture JSON in the expected storage path.
// Parameters: t provides the test context, home is the fake HOME directory,
// filename is the target debate file name, payload is the debate fixture.
// Returns: nothing.
func writeDebateForGet(t *testing.T, home string, filename string, payload debate.Debate) {
	t.Helper()
	dir := filepath.Join(home, ".bentos", "parley")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	data, err := json.MarshalIndent(payload, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, filename), data, 0o644))
}
