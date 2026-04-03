package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bentos-lab/parley/adapter/inbound/cli"
	"github.com/bentos-lab/parley/wiring"
)

// newResumeCommand builds the Cobra command for resuming debates.
// Parameters: ctx is the fallback execution context, usecases holds the debate usecases, runtime holds display info.
// Returns: the configured Cobra command.
func newResumeCommand(ctx context.Context, usecases *wiring.Usecases, runtime cli.RuntimeInfo) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume",
		Short: "Resume an existing debate",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("id is required")
			}
			output, err := outputForCommand(cmd)
			if err != nil {
				return err
			}
			id := args[0]
			numRounds, err := cmd.Flags().GetInt("num-rounds")
			if err != nil {
				return err
			}
			return cli.Resume(commandContext(cmd, ctx), usecases, output, runtime, id, numRounds)
		},
	}
	cmd.Flags().Int("num-rounds", 10, "maximum number of rounds")
	cmd.Flags().String("format", "pretty", "output format (pretty|json)")
	return cmd
}
