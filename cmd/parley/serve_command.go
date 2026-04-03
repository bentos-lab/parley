package main

import (
	"context"
	"log"

	"github.com/spf13/cobra"

	"github.com/bentos-lab/parley/config"
	"github.com/bentos-lab/parley/wiring"
)

// newServeCommand builds the Cobra command for running the HTTP server.
// Parameters: ctx is the base context, usecases contains the debate usecases, cfg carries configured providers.
// Returns: the configured Cobra command.
func newServeCommand(usecases *wiring.Usecases, cfg config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the HTTP server",
		RunE: func(cmd *cobra.Command, args []string) error {
			httpAddr, err := cmd.Flags().GetString("addr")
			if err != nil {
				return err
			}
			reqCtx := commandContext(cmd, context.Background())
			engine := newServeEngine(reqCtx, usecases, cfg, httpAddr)
			engine.startListener(reqCtx)
			log.Printf("HTTP server listening on http://%s", httpAddr)
			return engine.runServer()
		},
	}
	cmd.Flags().String("addr", "localhost:8080", "HTTP listen address")
	return cmd
}
