package whatsapp

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
)

const (
	connectDirName        = ".bentos/parley/connect"
	connectWhatsAppSubDir = "whatsapp"
	connectSessionName    = "session.json"
)

// Connect prints a WhatsApp QR code, drives the pairing flow, and blocks until
// the provided context is canceled or the QR code is scanned, returning any
// error encountered along the way.
func Connect(ctx context.Context) error {
	path, err := sessionPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create connect dir: %w", err)
	}
	if exists, err := sessionExists(path); err != nil {
		return err
	} else if exists {
		confirmed, err := confirmSessionRemoval(os.Stdin, os.Stdout)
		if err != nil {
			return fmt.Errorf("confirm existing session: %w", err)
		}
		if !confirmed {
			return fmt.Errorf("connect aborted: existing session retained")
		}
		if err := cleanupSessionFiles(path); err != nil {
			return fmt.Errorf("cleanup existing session: %w", err)
		}
	}
	container, err := newSessionContainer(path)
	if err != nil {
		return fmt.Errorf("open whatsapp store: %w", err)
	}
	device, err := container.GetFirstDevice(ctx)
	if err != nil {
		return fmt.Errorf("read whatsapp device: %w", err)
	}
	client := whatsmeow.NewClient(device, nil)
	qrCh, err := client.GetQRChannel(ctx)
	if err != nil {
		return fmt.Errorf("start qr channel: %w", err)
	}
	defer client.Disconnect()

	successCh := make(chan struct{}, 1)
	errCh := make(chan error, 1)
	go func() {
		for item := range qrCh {
			switch {
			case item.Event == whatsmeow.QRChannelEventCode:
				qrterminal.GenerateHalfBlock(item.Code, qrterminal.L, os.Stdout)
				fmt.Printf("Scan QR code (expires in %s)\n", item.Timeout)
			case item == whatsmeow.QRChannelSuccess:
				fmt.Println("QR scanned — pairing complete.")
				select {
				case successCh <- struct{}{}:
				default:
				}
				return
			case item == whatsmeow.QRChannelTimeout:
				errCh <- fmt.Errorf("qr timeout")
				return
			case item.Event == whatsmeow.QRChannelEventError:
				errCh <- fmt.Errorf("qr error: %w", item.Error)
				return
			default:
				fmt.Printf("QR status: %s\n", item.Event)
			}
		}
		errCh <- fmt.Errorf("qr channel closed unexpectedly")
	}()

	if err := client.Connect(); err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	select {
	case err := <-errCh:
		return err
	case <-successCh:
		fmt.Println("Waiting for synchronization...")
		fmt.Println("Press Ctrl+C once your device is fully connected.")
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		defer signal.Stop(sigCh)
		timer := time.NewTimer(1 * time.Minute)
		defer timer.Stop()
		select {
		case <-timer.C:
		case <-ctx.Done():
			return ctx.Err()
		case <-sigCh:
			fmt.Println("Device connection completed.")
		}
		return nil
	case <-time.After(5 * time.Minute):
		return fmt.Errorf("timed out waiting for QR scan")
	case <-ctx.Done():
		return ctx.Err()
	}
}

// sessionPath returns the full path for the WhatsApp session JSON file.
// Parameters: none.
// Returns: the path and any error encountered while resolving the home directory.
func sessionPath() (string, error) {
	dir, err := whatsappConnectDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, connectSessionName), nil
}

// whatsappConnectDir returns the absolute directory that stores WhatsApp session data.
// Parameters: none.
// Returns: the WhatsApp session directory path and any error encountered while resolving the home directory.
func whatsappConnectDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}
	return filepath.Join(home, connectDirName, connectWhatsAppSubDir), nil
}

// sessionExists reports whether the WhatsApp session database already exists.
// Parameters: path is the fully qualified database location.
// Returns: true if the database file exists, false if it is missing, and any
// error encountered while stat-ing the file.
func sessionExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, fmt.Errorf("stat whatsapp store: %w", err)
	}
}

// confirmSessionRemoval writes a prompt about overwriting the existing session
// store and parses the user response.
// Parameters: reader supplies the user input, writer receives the prompt text.
// Returns: true when the user explicitly consents, false otherwise, plus any
// error encountered while prompting.
func confirmSessionRemoval(reader io.Reader, writer io.Writer) (bool, error) {
	prompt := "An existing WhatsApp session was detected. Remove it and continue? [y/N]: "
	if _, err := writer.Write([]byte(prompt)); err != nil {
		return false, fmt.Errorf("write confirmation prompt: %w", err)
	}
	line, err := bufio.NewReader(reader).ReadString('\n')
	if err != nil && err != io.EOF {
		return false, fmt.Errorf("read confirmation response: %w", err)
	}
	response := strings.ToLower(strings.TrimSpace(line))
	return response == "y" || response == "yes", nil
}

// cleanupSessionFiles removes the WhatsApp session JSON file and history cache before a fresh pairing attempt.
// Parameters: path is the session filename.
// Returns: any error encountered while removing existing files (ignores missing files).
func cleanupSessionFiles(path string) error {
	history := filepath.Join(filepath.Dir(path), "whatsapp.history.json")
	files := []string{path, path + ".tmp", history}
	for _, file := range files {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove %s: %w", file, err)
		}
	}
	return nil
}
