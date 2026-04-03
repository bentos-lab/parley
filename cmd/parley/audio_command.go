package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/bentos-lab/parley/adapter/inbound/cli"
	"github.com/bentos-lab/parley/wiring"
)

// newAudioCommand builds the Cobra command for generating debate audio.
// Parameters: ctx is the fallback execution context, usecases holds the debate usecases.
// Returns: the configured Cobra command.
func newAudioCommand(ctx context.Context, usecases *wiring.Usecases) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "audio",
		Short: "Generate audio for a debate",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("id is required")
			}
			id := args[0]
			ttsProvider, err := cmd.Flags().GetString("tts-provider")
			if err != nil {
				return err
			}
			if !cmd.Flags().Changed("tts-provider") {
				ttsProvider = ""
			}
			path, err := cli.Audio(commandContext(cmd, ctx), usecases, id, ttsProvider)
			if err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, "Audio Generated")
			fmt.Fprintln(os.Stdout, "Audio generated successfully.")
			fmt.Fprintf(os.Stdout, "Path: %s\n", path)
			return nil
		},
	}
	return cmd
}
