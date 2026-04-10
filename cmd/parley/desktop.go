package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"time"

	"github.com/bentos-lab/parley/adapter/inbound/rest"
)

var desktopOpenURL = fmt.Sprintf("http://%s", defaultHTTPAddr)

// runDesktopLauncher starts the HTTP server with a CLI-friendly workflow for double-click launches.
// Parameters: ctx is the parent context, usecases configures services, and cfg provides global configuration.
// Returns: any fatal error occurred while running the HTTP server or managing the PID file.
func runDesktopLauncher(_ context.Context) error {
	go func() {
		time.Sleep(2 * time.Second)
		if err := openBrowser(desktopOpenURL); err != nil {
			log.Printf("desktop: open browser failed: %v", err)
		}
	}()

	server := rest.NewServer(defaultHTTPAddr)
	log.Printf("HTTP server listening on http://%s", defaultHTTPAddr)

	return server.ListenAndServe()
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
