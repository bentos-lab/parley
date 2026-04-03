package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bentos-lab/parley/adapter/inbound/cli"
	"github.com/bentos-lab/parley/wiring"
)

// newListCommand builds the Cobra command for listing debates by ID.
// Parameters: usecases contains the debate usecases.
// Returns: the configured Cobra command.
func newListCommand(usecases *wiring.Usecases) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List saved debates",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("list does not accept arguments")
			}
			output, err := listOutputForCommand(cmd)
			if err != nil {
				return err
			}
			return cli.List(usecases, output)
		},
	}
	cmd.Flags().String("format", "pretty", "output format (pretty|json)")
	return cmd
}
