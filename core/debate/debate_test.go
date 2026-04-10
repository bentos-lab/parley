package debate_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bentos-lab/parley/core/debate"
	"github.com/stretchr/testify/require"
)

// writeDebateFile writes a minimal debate JSON file under the configured HOME directory.
// Parameters: t provides the test context, home is the fake HOME directory, filename is the debate filename to create.
// Returns: nothing.
func writeDebateFile(t *testing.T, home string, filename string) {
	t.Helper()

	dir := filepath.Join(home, ".bentos", "parley")
	require.NoError(t, os.MkdirAll(dir, 0o755))

	payload := []byte(`{"name":"Test Debate","topic":"Test Topic","agents":[],"rounds":[],"tts_provider":"","llm_provider":"","llm_model":""}`)
	require.NoError(t, os.WriteFile(filepath.Join(dir, filename), payload, 0o644))
}

// TestGetAllDebateSortsByTimestampInID verifies GetAllDebate orders debates by the timestamp embedded in the ID.
// Parameters: t provides the test context.
func TestGetAllDebateSortsByTimestampInID(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	writeDebateFile(t, home, "alpha.2026-04-10-10-00-00.json")
	writeDebateFile(t, home, "beta.2026-04-09-10-00-00.json")
	writeDebateFile(t, home, "legacy.2026-04-08.json")
	writeDebateFile(t, home, "gamma.extra.2026-04-11-01-02.json")
	writeDebateFile(t, home, "notimestamp.json")

	items, err := debate.GetAllDebate()
	require.NoError(t, err)
	require.Equal(t, []string{
		"gamma.extra.2026-04-11-01-02",
		"alpha.2026-04-10-10-00-00",
		"beta.2026-04-09-10-00-00",
		"legacy.2026-04-08",
		"notimestamp",
	}, collectIDs(items))
}

// collectIDs extracts the ID fields from debate summaries.
// Parameters: items holds the summaries returned by GetAllDebate.
// Returns: the list of IDs in order.
func collectIDs(items []debate.DebateSummary) []string {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids
}
