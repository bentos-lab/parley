package jsonoutput

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/bentos-lab/parley/adapter/inbound/cli"
	"github.com/stretchr/testify/require"
)

// TestDebateDetailsEmitsEnvelope verifies get output uses the type/data envelope and required keys.
// Parameters: t provides the test context.
// Returns: nothing.
func TestDebateDetailsEmitsEnvelope(t *testing.T) {
	output := New()
	var buffer bytes.Buffer
	err := output.DebateDetails(&buffer, cli.DebateDetailsOutput{
		Header: cli.DebateHeaderOutput{ID: "debate-1"},
		Topic:  "Topic",
		Name:   "Name",
		Agents: []cli.AgentRow{{ID: "agent-1", Name: "Alice"}},
		Rounds: []cli.DebateRoundOutput{{Number: 1, AgentID: "agent-1", AgentName: "Alice"}},
		Summary: cli.DebateSummaryDetail{
			Agents:     [][]string{{"Point"}},
			Conclusion: "Done",
		},
	})
	require.NoError(t, err)

	var envelope struct {
		Type string         `json:"type"`
		Data map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(buffer.Bytes(), &envelope))
	require.Equal(t, "debate_get", envelope.Type)
	require.Contains(t, envelope.Data, "header")
	require.Contains(t, envelope.Data, "topic")
	require.Contains(t, envelope.Data, "name")
	require.Contains(t, envelope.Data, "agents")
	require.Contains(t, envelope.Data, "rounds")
	require.Contains(t, envelope.Data, "summary")
}
