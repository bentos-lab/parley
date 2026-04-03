package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"time"

	"github.com/bentos-lab/parley/config"
	"github.com/bentos-lab/parley/wiring"
	webview "github.com/webview/webview_go"
)

var desktopOpenURL = fmt.Sprintf("http://%s", defaultHTTPAddr)

const (
	launcherWidth  = 300
	launcherHeight = 150
)

const launcherHTML = `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <title>Parley desktop</title>
    <style>
      body {
        margin: 0;
        font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
        background: #f8fafc;
        color: #0f172a;
      }
      .shell {
        display: flex;
        flex-direction: column;
        justify-content: center;
        align-items: center;
        height: 100vh;
        gap: 1rem;
      }
      .buttons {
        display: flex;
        gap: 0.5rem;
      }
      button {
        border: none;
        border-radius: 6px;
        padding: 0.5rem 1.2rem;
        font-size: 0.9rem;
        font-weight: 600;
        cursor: pointer;
      }
      button.primary {
        background: #2563eb;
        color: white;
      }
      button.secondary {
        background: #e2e8f0;
        color: #0f172a;
      }
    </style>
  </head>
  <body>
    <div class="shell">
      <div>Parley server is running.</div>
      <div class="buttons">
        <button class="primary" onclick="desktopAction('open')">Open</button>
        <button class="secondary" onclick="desktopAction('exit')">Exit</button>
      </div>
    </div>
  </body>
</html>`

// runDesktopLauncher starts the shared serve engine and serves a minimal WebView control panel.
// Parameters: ctx is the parent context, usecases and cfg configure services.
// Returns: any fatal error from the HTTP server startup or UI run.
func runDesktopLauncher(ctx context.Context, usecases *wiring.Usecases, cfg config.Config) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	engine := newServeEngine(ctx, usecases, cfg, defaultHTTPAddr)
	go engine.startListener(ctx)

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- engine.runServer()
	}()

	w := webview.New(false)
	defer w.Destroy()
	w.SetTitle("Parley")
	w.SetSize(launcherWidth, launcherHeight, webview.Hint(webview.HintFixed))
	if err := w.Bind("desktopAction", func(action string) error {
		switch action {
		case "open":
			if err := openBrowser(desktopOpenURL); err != nil {
				log.Printf("desktop: open browser failed: %v", err)
			}
		case "exit":
			cancel()
			w.Terminate()
		}
		return nil
	}); err != nil {
		return fmt.Errorf("desktop: bind failed: %w", err)
	}

	w.Navigate(launcherURL())
	w.Run()
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := engine.shutdown(shutdownCtx); err != nil {
		log.Printf("desktop: shutdown failed: %v", err)
	}

	return <-serverErrCh
}

// launcherURL builds a data URL encoding the launcher HTML so no filesystem assets are required.
func launcherURL() string {
	encoded := base64.StdEncoding.EncodeToString([]byte(launcherHTML))
	return "data:text/html;base64," + encoded
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
