package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mdp/qrterminal/v3"
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
			existed, err := whatsapp.CheckExisted(ctx)
			if err != nil {
				return fmt.Errorf("check whatsapp existed session: %w", err)
			}

			if existed {
				confirmed, err := promptSessionRemoval(os.Stdin, os.Stdout)
				if err != nil {
					return fmt.Errorf("confirm existing session: %w", err)
				}
				if !confirmed {
					return nil
				}
				if err := whatsapp.RemoveSession(); err != nil {
					return fmt.Errorf("cleanup existing session: %w", err)
				}
			}

			code, waitScan, finalize, err := whatsapp.Connect(ctx)
			if err != nil {
				return err
			}

			qrterminal.GenerateHalfBlock(code.Code, qrterminal.L, os.Stdout)
			fmt.Printf("Scan QR code (expires in %s)\n", code.Timeout)

			scanCtx, scanCancel := context.WithTimeout(ctx, code.Timeout)
			defer scanCancel()

			waitScan(scanCtx)

			fmt.Println("QR scanned — waiting for synchronization...")
			fmt.Println("Press enter once your device is fully connected.")

			ctx, cancel := context.WithCancel(ctx)

			go func() {
				bufio.NewReader(os.Stdin).ReadBytes('\n')
				cancel()
			}()

			if err := finalize(ctx); err != nil {
				return err
			}

			fmt.Println("Device connection completed.")
			return nil
		},
	}
}

func promptSessionRemoval(reader io.Reader, writer io.Writer) (bool, error) {
	prompt := "An existing WhatsApp session was detected. Remove it and continue? [y/N]: "
	if _, err := writer.Write([]byte(prompt)); err != nil {
		return false, fmt.Errorf("write confirmation prompt: %w", err)
	}
	line, err := bufio.NewReader(reader).ReadString('\n')
	if err != nil && err != io.EOF {
		return false, fmt.Errorf("read confirmation response: %w", err)
	}
	response := strings.TrimSpace(strings.ToLower(line))
	return response == "y" || response == "yes", nil
}
