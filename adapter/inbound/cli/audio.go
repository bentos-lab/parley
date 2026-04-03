package cli

import (
	"context"
	"fmt"

	"github.com/bentos-lab/parley/core"
	"github.com/bentos-lab/parley/core/debate"
	"github.com/bentos-lab/parley/wiring"
)

// Audio generates the audio file for a debate.
// Parameters: ctx is the request context, usecases holds the debate usecases, debateID is the debate identifier,
// ttsProvider is the selected TTS provider.
// Returns: the generated audio path and an error if generation fails.
func Audio(ctx context.Context, usecases *wiring.Usecases, debateID string, ttsProvider string) (string, error) {
	if debateID == "" {
		return "", fmt.Errorf("id is required")
	}
	filename := debate.FilenameFromID(debateID)
	output, err := usecases.GenerateAudio.Execute(ctx, core.GenerateAudioInput{
		Filename:    filename,
		TTSProvider: ttsProvider,
	})
	if err != nil {
		return "", err
	}
	return output.Path, nil
}
