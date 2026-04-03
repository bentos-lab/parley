package main

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/bentos-lab/parley/adapter/inbound/whatsapp"
)

// newConnectWhatsappCommand builds the command that pairs WhatsApp via QR.
func newConnectWhatsappCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "connect whatsapp",
		Short: "Pair WhatsApp via QR to enable /parley DM commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := commandContext(cmd, context.Background())
			return whatsapp.Connect(ctx)
		},
	}
}
