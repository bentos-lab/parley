//go:build darwin

package native

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// Synthesize generates WAV audio bytes for the given text using macOS say + afconvert.
// Parameters: ctx controls cancellation, text is the content to synthesize, voiceName is ignored for native TTS.
// Returns: the synthesized WAV bytes or an error if synthesis fails.
func (c *Client) Synthesize(ctx context.Context, text string, voiceName string) ([]byte, error) {
	_ = voiceName
	normalized := NormalizeText(text)
	aiffFile, err := os.CreateTemp("", "bentos-tts-*.aiff")
	if err != nil {
		return nil, fmt.Errorf("native say temp file: %w", err)
	}
	aiffPath := aiffFile.Name()
	if err := aiffFile.Close(); err != nil {
		return nil, fmt.Errorf("native say temp close: %w", err)
	}
	defer os.Remove(aiffPath)

	wavFile, err := os.CreateTemp("", "bentos-tts-*.wav")
	if err != nil {
		return nil, fmt.Errorf("native wav temp file: %w", err)
	}
	wavPath := wavFile.Name()
	if err := wavFile.Close(); err != nil {
		return nil, fmt.Errorf("native wav temp close: %w", err)
	}
	defer os.Remove(wavPath)

	sayCmd := exec.CommandContext(ctx, "say", "-o", aiffPath, normalized)
	if output, err := sayCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("native say synthesize: %w: %s", err, string(output))
	}
	convertCmd := exec.CommandContext(ctx, "afconvert", "-f", "WAVE", "-d", "LEI16@22050", aiffPath, wavPath)
	if output, err := convertCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("native afconvert: %w: %s", err, string(output))
	}
	bytes, err := os.ReadFile(wavPath)
	if err != nil {
		return nil, fmt.Errorf("native wav read: %w", err)
	}
	return bytes, nil
}
