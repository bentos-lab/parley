package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"time"

	"github.com/bentos-lab/parley/config"
	"github.com/bentos-lab/parley/wiring"
)

var desktopOpenURL = fmt.Sprintf("http://%s", defaultHTTPAddr)

// runDesktopLauncher starts the HTTP server with a CLI-friendly workflow for double-click launches.
// Parameters: ctx is the parent context, usecases configures services, and cfg provides global configuration.
// Returns: any fatal error occurred while running the HTTP server or managing the PID file.
func runDesktopLauncher(ctx context.Context, usecases *wiring.Usecases, cfg config.Config) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	engine := newServeEngine(ctx, usecases, cfg, defaultHTTPAddr)
	go engine.startListener(ctx)
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		if err := engine.shutdown(shutdownCtx); err != nil {
			log.Printf("desktop: shutdown failed: %v", err)
		}
		shutdownCancel()
	}()

	fmt.Printf("Parley server listening on http://%s\n", defaultHTTPAddr)

	go func() {
		if err := openBrowser(desktopOpenURL); err != nil {
			log.Printf("desktop: open browser failed: %v", err)
		}
	}()

	return engine.runServer()
}

// openBrowser launches the provided URL using a platform-appropriate command.
// Parameters: url is the destination that should open in the user\'s default browser.
// Returns: any error encountered while starting the helper process.
func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/C", "start", "", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
