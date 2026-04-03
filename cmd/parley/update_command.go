package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bentos-lab/parley/adapter/inbound/cli"
)

// newUpdateCommand builds the Cobra parent command for update-related actions.
// Parameters: ctx is the base context passed to child commands.
// Returns: a configured Cobra command with the subcommands attached.
func newUpdateCommand(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Manage application updates",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("update requires a subcommand")
			}
			return fmt.Errorf("unknown update subcommand %q", args[0])
		},
	}
	cmd.AddCommand(newUpdateSelfCommand(ctx))
	return cmd
}

// newUpdateSelfCommand builds the Cobra command that replaces the running binary.
// Parameters: ctx is the fallback execution context used to cancel network requests.
// Returns: a configured Cobra command that replaces the current executable with the latest release.
func newUpdateSelfCommand(ctx context.Context) *cobra.Command {
	return &cobra.Command{
		Use:   "self",
		Short: "Update the current Parley binary",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("update self does not accept arguments")
			}
			return cli.UpdateSelf(commandContext(cmd, ctx), cmd.OutOrStdout())
		},
	}
}
