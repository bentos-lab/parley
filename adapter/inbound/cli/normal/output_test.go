package normal_test

import (
	"bytes"
	"testing"

	"github.com/bentos-lab/parley/adapter/inbound/cli"
	"github.com/bentos-lab/parley/adapter/inbound/cli/normal"
	"github.com/stretchr/testify/require"
)

// TestDebateDetailsIncludesStableSections verifies normal formatter prints required sections.
// Parameters: t provides the test context.
// Returns: nothing.
func TestDebateDetailsIncludesStableSections(t *testing.T) {
	output := normal.New()
	var buffer bytes.Buffer
	err := output.DebateDetails(&buffer, cli.DebateDetailsOutput{
		Header: cli.DebateHeaderOutput{
			ID:          "debate.2026-04-10-00-00-00",
			AppName:     "Debate",
			LLMProvider: "openai",
			LLMModel:    "gpt-5.4",
			TTSProvider: "native",
			TTSModel:    "say",
			AgentsCount: 1,
		},
		Topic: "Topic A",
		Name:  "Name A",
		Agents: []cli.AgentRow{
			{ID: "agent-1", Name: "Alice", Stance: "For", Voice: "voice-a"},
		},
		Rounds: []cli.DebateRoundOutput{
			{Number: 1, AgentID: "agent-1", AgentName: "Alice", Message: "Msg", Summary: "Sum"},
		},
		Summary: cli.DebateSummaryDetail{
			Agents:     [][]string{{"Point A"}},
			Conclusion: "Done",
		},
	})
	require.NoError(t, err)
	content := buffer.String()
	require.Contains(t, content, "Header:")
	require.Contains(t, content, "Topic:")
	require.Contains(t, content, "Name:")
	require.Contains(t, content, "Agents:")
	require.Contains(t, content, "Rounds:")
	require.Contains(t, content, "Summary:")
}
