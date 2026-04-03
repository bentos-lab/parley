package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/bentos-lab/parley/config"
	"github.com/bentos-lab/parley/wiring"
)

var desktopOpenURL = fmt.Sprintf("http://%s", defaultHTTPAddr)

// runDesktopLauncher starts the shared serve engine and displays the Fyne launcher window.
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

	desktopApp := app.NewWithID("com.bentos.parley")
	window := desktopApp.NewWindow("Parley")
	window.SetCloseIntercept(func() {
		cancel()
		window.Close()
	})

	openButton := widget.NewButton("Open", func() {
		if err := openBrowser(desktopOpenURL); err != nil {
			log.Printf("desktop: open browser failed: %v", err)
		}
	})
	exitButton := widget.NewButton("Exit", func() {
		cancel()
		window.Close()
	})

	window.SetContent(container.NewVBox(
		widget.NewLabel("Parley server running"),
		container.NewHBox(openButton, exitButton),
	))
	window.Resize(fyne.NewSize(300, 150))
	window.CenterOnScreen()

	desktopApp.Run()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := engine.shutdown(shutdownCtx); err != nil {
		log.Printf("desktop: shutdown failed: %v", err)
	}

	return <-serverErrCh
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
