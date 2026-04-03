package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bentos-lab/parley/adapter/inbound/cli"
	"github.com/bentos-lab/parley/wiring"
)

// newCreateCommand builds the Cobra command for creating debates.
// Parameters: ctx is the fallback execution context, usecases holds the debate usecases, runtime holds display info.
// Returns: the configured Cobra command.
func newCreateCommand(ctx context.Context, usecases *wiring.Usecases, runtime cli.RuntimeInfo) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new debate",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("topic is required")
			}
			output, err := outputForCommand(cmd)
			if err != nil {
				return err
			}
			topic := args[0]
			numAgents, err := cmd.Flags().GetInt("num-agents")
			if err != nil {
				return err
			}
			numRounds, err := cmd.Flags().GetInt("num-rounds")
			if err != nil {
				return err
			}
			ttsProvider, err := cmd.Flags().GetString("tts-provider")
			if err != nil {
				return err
			}
			return cli.Create(commandContext(cmd, ctx), usecases, output, runtime, topic, numAgents, numRounds, ttsProvider)
		},
	}
	cmd.Flags().Int("num-agents", 2, "number of agents to auto-generate")
	cmd.Flags().Int("num-rounds", 10, "maximum number of rounds")
	cmd.Flags().String("format", "pretty", "output format (pretty|json)")
	return cmd
}
