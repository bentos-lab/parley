package main

import (
	"context"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"

	"github.com/bentos-lab/parley/adapter/inbound/cli"
	"github.com/bentos-lab/parley/config"
	"github.com/bentos-lab/parley/wiring"
)

// version holds the build-time application version value. It defaults to
// "dev" so locally built binaries still report a sensible placeholder.
// The build pipeline overrides it with git metadata via `-ldflags`.
var version = "dev"

// main loads environment configuration, builds services, and dispatches
// to server or CLI execution based on command-line arguments.
// Parameters: none.
// Returns: nothing.
func main() {
	godotenv.Load()
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	usecases, err := wiring.BuildUsecases(cfg)
	if err != nil {
		log.Fatalf("build usecases: %v", err)
	}
	runtime := cli.RuntimeInfo{
		LLMBaseURL: cfg.OpenAI.BaseURL,
		LLMModel:   cfg.OpenAI.Model,
		TTSModel:   cfg.InworldModel,
	}
	rootCmd := newRootCommand(ctx, usecases, runtime, cfg.TTSProvider, cfg)
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		log.Fatalf("cli error: %v", err)
	}
}

// newRootCommand builds the Cobra root command for the CLI.
// Parameters: ctx is the fallback execution context, service is the debate service handler,
// runtime carries display information, defaultTTSProvider and cfg provide shared preferences.
// Returns: the configured Cobra root command.
func newRootCommand(ctx context.Context, usecases *wiring.Usecases, runtime cli.RuntimeInfo, defaultTTSProvider string, cfg config.Config) *cobra.Command {
	root := &cobra.Command{
		Use:          "parley",
		Version:      version,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return fmt.Errorf("unknown command: %s", args[0])
			}

			return runDesktopLauncher(commandContext(cmd, ctx), usecases, cfg)
		},
	}
	root.SetVersionTemplate("version: {{.Version}}\n")
	root.PersistentFlags().String("tts-provider", defaultTTSProvider, "override the TTS provider")
	root.AddCommand(newServeCommand(usecases, cfg))
	root.AddCommand(newConfigCommand())
	root.AddCommand(newLoginCommand())
	root.AddCommand(newConnectWhatsappCommand())
	root.AddCommand(newCreateCommand(ctx, usecases, runtime))
	root.AddCommand(newResumeCommand(ctx, usecases, runtime))
	root.AddCommand(newListCommand(usecases))
	root.AddCommand(newDeleteCommand(ctx, usecases))
	root.AddCommand(newAudioCommand(ctx, usecases))
	root.AddCommand(newUpdateCommand(ctx))
	return root
}

// commandContext selects the command context when available or falls back to the provided context.
// Parameters: cmd is the executing Cobra command, fallback is the context to use when the command has none.
// Returns: the context that should be used for service calls.
func commandContext(cmd *cobra.Command, fallback context.Context) context.Context {
	if cmd == nil {
		return fallback
	}
	cmdCtx := cmd.Context()
	if cmdCtx != nil {
		return cmdCtx
	}
	return fallback
}
