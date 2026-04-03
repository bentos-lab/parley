package audio_test

import (
	"testing"

	"github.com/bentos-lab/parley/shared/audio"
	"github.com/stretchr/testify/require"
)

// testWAVBytes builds a small WAV payload for test cases.
// Parameters: none.
// Returns: a WAV byte slice and the original sample data.
func testWAVBytes() ([]byte, []byte) {
	samples := []byte{0x01, 0x02, 0x03, 0x04}
	wav := audio.WAV{
		SampleRate:    8000,
		Channels:      1,
		BitsPerSample: 16,
		Data:          samples,
	}
	return wav.Bytes(), samples
}

// TestParseWAVReadsValidData verifies ParseWAV parses intact WAV data.
// Parameters: t provides the test context.
func TestParseWAVReadsValidData(t *testing.T) {
	t.Parallel()

	wavBytes, samples := testWAVBytes()
	parsed, err := audio.ParseWAV(wavBytes)
	require.NoError(t, err)
	require.Equal(t, 8000, parsed.SampleRate)
	require.Equal(t, 1, parsed.Channels)
	require.Equal(t, 16, parsed.BitsPerSample)
	require.Equal(t, samples, parsed.Data)
}

// TestParseWAVAllowsTruncatedData verifies ParseWAV tolerates truncated data chunks.
// Parameters: t provides the test context.
func TestParseWAVAllowsTruncatedData(t *testing.T) {
	t.Parallel()

	wavBytes, _ := testWAVBytes()
	truncated := wavBytes[:len(wavBytes)-2]
	parsed, err := audio.ParseWAV(truncated)
	require.NoError(t, err)
	require.Equal(t, 2, len(parsed.Data))
	require.Equal(t, []byte{0x01, 0x02}, parsed.Data)
}
