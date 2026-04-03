//go:build windows

package native

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Synthesize generates WAV audio bytes for the given text using Windows System.Speech.
// Parameters: ctx controls cancellation, text is the content to synthesize, voiceName is ignored for native TTS.
// Returns: the synthesized WAV bytes or an error if synthesis fails.
func (c *Client) Synthesize(ctx context.Context, text string, voiceName string) ([]byte, error) {
	_ = voiceName
	normalized := NormalizeText(text)
	wavFile, err := os.CreateTemp("", "bentos-tts-*.wav")
	if err != nil {
		return nil, fmt.Errorf("native wav temp file: %w", err)
	}
	wavPath := wavFile.Name()
	if err := wavFile.Close(); err != nil {
		return nil, fmt.Errorf("native wav temp close: %w", err)
	}
	defer os.Remove(wavPath)

	script := buildPowerShellScript(wavPath, normalized)
	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", script)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("native windows synthesize: %w: %s", err, string(output))
	}
	bytes, err := os.ReadFile(wavPath)
	if err != nil {
		return nil, fmt.Errorf("native wav read: %w", err)
	}
	return bytes, nil
}

func buildPowerShellScript(wavPath string, text string) string {
	path := escapePowerShellString(wavPath)

	encoded := base64.StdEncoding.EncodeToString([]byte(text))

	return strings.Join([]string{
		"Add-Type -AssemblyName System.Speech",
		"$synth = New-Object System.Speech.Synthesis.SpeechSynthesizer",
		fmt.Sprintf("$synth.SetOutputToWaveFile('%s')", path),

		// decode base64 trong PowerShell
		fmt.Sprintf("$text = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String('%s'))", encoded),

		"$synth.Speak($text)",
		"$synth.SetOutputToNull()",
		"$synth.Dispose()",
	}, "; ")
}

func escapePowerShellString(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}
