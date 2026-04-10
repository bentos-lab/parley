package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bentos-lab/parley/adapter/inbound/cli"
	"github.com/bentos-lab/parley/wiring"
)

// newGetCommand builds the Cobra command for fetching saved debate details by ID.
// Parameters: ctx is the fallback execution context, usecases holds debate usecases,
// runtime carries model metadata used for fallback display.
// Returns: the configured Cobra command.
func newGetCommand(ctx context.Context, usecases *wiring.Usecases, runtime cli.RuntimeInfo) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Show details of a saved debate",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("id is required")
			}
			output, err := getOutputForCommand(cmd)
			if err != nil {
				return err
			}
			return cli.Get(commandContext(cmd, ctx), usecases, output, runtime, args[0])
		},
	}
	cmd.Flags().String("format", "pretty", "output format (pretty|normal|json)")
	return cmd
}
