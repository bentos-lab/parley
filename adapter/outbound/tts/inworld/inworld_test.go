package inworld_test

import (
	"testing"

	"github.com/bentos-lab/parley/adapter/outbound/tts/inworld"
	"github.com/stretchr/testify/require"
)

// TestNormalizeTextReplacesPauseTags verifies pause tags map to SSML breaks.
// Parameters: t provides the test context.
func TestNormalizeTextReplacesPauseTags(t *testing.T) {
	t.Parallel()

	input := "Hi <pause300> there <pause500> now <pause1000> ok <pause250>"
	expected := "Hi <break time=\"300ms\" /> there <break time=\"500ms\" /> now <break time=\"1s\" /> ok <break time=\"500ms\" />"
	result := inworld.NormalizeText(input)
	require.Equal(t, expected, result)
}

// TestNormalizeTextReplacesNewlines verifies newlines become spaces.
// Parameters: t provides the test context.
func TestNormalizeTextReplacesNewlines(t *testing.T) {
	t.Parallel()

	input := "Hello\nworld\r\nfrom\nInworld"
	expected := "Hello world from Inworld"
	result := inworld.NormalizeText(input)
	require.Equal(t, expected, result)
}
