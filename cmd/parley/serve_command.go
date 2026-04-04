package main

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/bentos-lab/parley/adapter/inbound/rest"
)

const defaultHTTPAddr = "localhost:8080"

// newServeCommand builds the Cobra command for running the HTTP server.
// Parameters: ctx is the base context, usecases contains the debate usecases, cfg carries configured providers.
// Returns: the configured Cobra command.
func newServeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Run the HTTP server",
		RunE: func(cmd *cobra.Command, args []string) error {
			httpAddr, err := cmd.Flags().GetString("addr")
			if err != nil {
				return err
			}

			server := rest.NewServer(httpAddr)
			log.Printf("HTTP server listening on http://%s", httpAddr)
			return server.ListenAndServe()
		},
	}
	cmd.Flags().String("addr", defaultHTTPAddr, "HTTP listen address")
	return cmd
}
