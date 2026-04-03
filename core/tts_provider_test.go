package core

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveTTSProviderPrioritizesOverride(t *testing.T) {
	t.Parallel()
	provider, err := resolveTTSProvider("override", "stored", "default")
	require.NoError(t, err)
	require.Equal(t, "override", provider)
}

func TestResolveTTSProviderFallsBackToStored(t *testing.T) {
	t.Parallel()
	provider, err := resolveTTSProvider("", "stored", "default")
	require.NoError(t, err)
	require.Equal(t, "stored", provider)
}

func TestResolveTTSProviderUsesDefault(t *testing.T) {
	t.Parallel()
	provider, err := resolveTTSProvider("", "", "default")
	require.NoError(t, err)
	require.Equal(t, "default", provider)
}

func TestResolveTTSProviderErrorsWhenMissing(t *testing.T) {
	t.Parallel()
	_, err := resolveTTSProvider("", "", "")
	require.Error(t, err)
}
