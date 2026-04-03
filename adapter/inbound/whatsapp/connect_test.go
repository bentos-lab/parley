package whatsapp

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfirmSessionRemoval(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected bool
	}{
		{name: "yes", input: "yes\n", expected: true},
		{name: "y uppercase", input: "Y\n", expected: true},
		{name: "no", input: "n\n", expected: false},
		{name: "empty", input: "\n", expected: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			result, err := confirmSessionRemoval(strings.NewReader(tc.input), &out)
			require.NoError(t, err)
			require.Equal(t, tc.expected, result)
			require.Contains(t, out.String(), "Remove it and continue?")
		})
	}
}

func TestCleanupSessionFiles(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "session.json")
	history := filepath.Join(dir, "whatsapp.history.json")
	require.NoError(t, os.WriteFile(base, []byte("data"), 0o644))
	require.NoError(t, os.WriteFile(base+".tmp", []byte("tmp"), 0o644))
	require.NoError(t, os.WriteFile(history, []byte("history"), 0o644))
	require.NoError(t, cleanupSessionFiles(base))
	for _, file := range []string{base, base + ".tmp", history} {
		_, err := os.Stat(file)
		require.True(t, os.IsNotExist(err))
	}
}

func TestCleanupSessionFilesMissingVariants(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "session.json")
	require.NoError(t, os.WriteFile(base, []byte("data"), 0o644))
	require.NoError(t, cleanupSessionFiles(base))
	_, err := os.Stat(base)
	require.True(t, os.IsNotExist(err))
}
