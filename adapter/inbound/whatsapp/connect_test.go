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
	base := filepath.Join(dir, "whatsapp.db")
	for _, variant := range []string{"", "-wal", "-shm", "-journal"} {
		require.NoError(t, os.WriteFile(base+variant, []byte("data"), 0o644))
	}
	require.NoError(t, cleanupSessionFiles(base))
	for _, variant := range []string{"", "-wal", "-shm", "-journal"} {
		_, err := os.Stat(base + variant)
		require.True(t, os.IsNotExist(err))
	}
}

func TestCleanupSessionFilesMissingVariants(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "whatsapp.db")
	require.NoError(t, os.WriteFile(base, []byte("data"), 0o644))
	require.NoError(t, cleanupSessionFiles(base))
	_, err := os.Stat(base)
	require.True(t, os.IsNotExist(err))
}
