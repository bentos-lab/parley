package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bentos-lab/parley/adapter/inbound/cli"
	"github.com/bentos-lab/parley/wiring"
)

// newDeleteCommand builds the Cobra command for deleting debates by ID.
// Parameters: ctx is the fallback execution context, usecases contains the debate usecases.
// Returns: the configured Cobra command.
func newDeleteCommand(ctx context.Context, usecases *wiring.Usecases) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a saved debate",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("id is required")
			}
			return cli.Delete(commandContext(cmd, ctx), usecases, args[0])
		},
	}
	return cmd
}
