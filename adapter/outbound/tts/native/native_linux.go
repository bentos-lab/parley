//go:build linux

package native

import (
	"context"
	"fmt"
	"os/exec"
)

// Synthesize generates WAV audio bytes for the given text using espeak.
// Parameters: ctx controls cancellation, text is the content to synthesize, voiceName is ignored for native TTS.
// Returns: the synthesized WAV bytes or an error if synthesis fails.
func (c *Client) Synthesize(ctx context.Context, text string, voiceName string) ([]byte, error) {
	_ = voiceName
	normalized := NormalizeText(text)
	cmd := exec.CommandContext(ctx, "espeak", "--stdout", normalized)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("native espeak synthesize: %w", err)
	}
	return output, nil
}
