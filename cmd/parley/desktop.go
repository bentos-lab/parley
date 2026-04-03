package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
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

	pidPath, err := desktopPIDPath()
	if err != nil {
		return fmt.Errorf("desktop: pid path: %w", err)
	}

	stopped, err := stopExistingInstance(pidPath)
	if err != nil {
		return fmt.Errorf("desktop: stop check: %w", err)
	}
	if stopped {
		fmt.Println("Stopped. Run the binary again to start the server.")
		return nil
	}

	engine := newServeEngine(ctx, usecases, cfg, defaultHTTPAddr)
	go engine.startListener(ctx)
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		if err := engine.shutdown(shutdownCtx); err != nil {
			log.Printf("desktop: shutdown failed: %v", err)
		}
		shutdownCancel()
	}()

	if err := writeDesktopPID(pidPath); err != nil {
		return fmt.Errorf("desktop: write pid: %w", err)
	}
	defer removeDesktopPID(pidPath)

	fmt.Printf("Parley server listening on http://%s\n", defaultHTTPAddr)
	fmt.Println("Double-click the binary again to stop the server.")

	go func() {
		time.Sleep(5 * time.Second)
		if err := openBrowser(desktopOpenURL); err != nil {
			log.Printf("desktop: open browser failed: %v", err)
		}
	}()

	return engine.runServer()
}

// desktopPIDPath returns the path to the PID file stored in the user config directory,
// and falls back to the temporary directory when the config directory cannot be used.
func desktopPIDPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = os.TempDir()
	}
	path := filepath.Join(dir, "parley")
	if err := os.MkdirAll(path, 0o755); err != nil {
		path = filepath.Join(os.TempDir(), "parley")
		if err := os.MkdirAll(path, 0o755); err != nil {
			return "", err
		}
	}
	return filepath.Join(path, "parley.pid"), nil
}

// writeDesktopPID records the current process PID in the designated file.
// Parameters: pidPath is the file that should store the PID.
// Returns: any error encountered while creating or writing the file.
func writeDesktopPID(pidPath string) error {
	pid := os.Getpid()
	return os.WriteFile(pidPath, []byte(strconv.Itoa(pid)), 0o644)
}

// removeDesktopPID attempts to delete the PID file.
// Parameters: pidPath is the previously recorded PID file path.
// Returns: nothing, errors are ignored because removal is best-effort.
func removeDesktopPID(pidPath string) {
	_ = os.Remove(pidPath)
}

// stopExistingInstance kills an already running Parley process recorded in the PID file.
// Parameters: pidPath is the PID file to inspect.
// Returns: stopped=true when a running instance was terminated, or false when no stale PID was found.
func stopExistingInstance(pidPath string) (bool, error) {
	pid, err := readDesktopPID(pidPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		removeDesktopPID(pidPath)
		return false, nil
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		removeDesktopPID(pidPath)
		return false, nil
	}
	if err := proc.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return false, err
	}
	removeDesktopPID(pidPath)
	return true, nil
}

// readDesktopPID parses the PID stored in the PID file.
// Parameters: pidPath is the file path that should contain the numeric PID.
// Returns: the parsed PID or an error when the file is missing or malformed.
func readDesktopPID(pidPath string) (int, error) {
	data, err := os.ReadFile(pidPath)
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, err
	}
	return pid, nil
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
