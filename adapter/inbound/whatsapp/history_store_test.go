package whatsapp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bentos-lab/parley/core"
)

func historyFilePathForTest(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	return filepath.Join(dir, connectDirName, "whatsapp.history.json")
}

// Test_historyStore_loadFilters ensures the loader strips out entries that do not start with /parley or [parley].
func Test_historyStore_loadFilters(t *testing.T) {
	t.Helper()
	path := historyFilePathForTest(t)
	raw := map[string][]core.ParleyCommandHistoryMessage{
		"chat-1": {
			{Role: "user", Content: "/parley start"},
			{Role: "assistant", Content: "[parley] OK"},
			{Role: "user", Content: "hello"},
			{Role: "assistant", Content: "[other] ignore"},
			{Role: "assistant", Content: " [parley] spaced"},
		},
	}
	data, err := json.MarshalIndent(raw, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, data, 0o644))

	store, err := newHistoryStore()
	require.NoError(t, err)

	entries := store.snapshot("chat-1")
	require.Len(t, entries, 3)
	require.Equal(t, "/parley start", entries[0].Content)
	require.Equal(t, "[parley] OK", entries[1].Content)
	require.Equal(t, "[parley] spaced", entries[2].Content)
}

// Test_historyStore_addEnforcesCapacity verifies that add keeps only the latest historyCapacity entries and persists them.
func Test_historyStore_addEnforcesCapacity(t *testing.T) {
	t.Helper()
	path := historyFilePathForTest(t)
	store, err := newHistoryStore()
	require.NoError(t, err)

	const iterations = historyCapacity + 5
	chatID := "chat-2"
	for i := 0; i < iterations; i++ {
		require.NoError(t, store.add(chatID, core.ParleyCommandHistoryMessage{
			Role:    "user",
			Content: fmt.Sprintf("/parley u%d", i),
		}))
		require.NoError(t, store.add(chatID, core.ParleyCommandHistoryMessage{
			Role:    "assistant",
			Content: fmt.Sprintf("[parley] a%d", i),
		}))
	}

	entries := store.snapshot(chatID)
	require.Len(t, entries, historyCapacity)
	require.Equal(t, "/parley u10", entries[0].Content)
	require.Equal(t, fmt.Sprintf("[parley] a%d", iterations-1), entries[len(entries)-1].Content)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	var persisted map[string][]core.ParleyCommandHistoryMessage
	require.NoError(t, json.Unmarshal(data, &persisted))

	persistedEntries := persisted[chatID]
	require.Len(t, persistedEntries, historyCapacity)
	require.Equal(t, entries, persistedEntries)
}
