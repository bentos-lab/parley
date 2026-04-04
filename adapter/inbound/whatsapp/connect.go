package whatsapp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
)

const (
	connectDirName        = ".bentos/parley/connect"
	connectWhatsAppSubDir = "whatsapp"
	connectSessionName    = "session.json"
)

// CheckExisted reports whether a WhatsApp session already exists and propagates
// any error encountered while examining the filesystem within the provided
// context.
// Parameters: ctx provides the context used when accessing the session path.
// Returns: a boolean flag indicating if the session file exists and any
// error seen during path resolution or stat.
func CheckExisted(ctx context.Context) (bool, error) {
	path, err := sessionPath()
	if err != nil {
		return false, err
	}

	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, fmt.Errorf("stat whatsapp store: %w", err)
	}
}

// Connect initializes the WhatsApp pairing flow, returns the first QR code
// string with its expiration, and exposes finalize/cancel hooks that drive the
// rest of the pairing process.
//
// The returned finalize function blocks until the QR code is scanned (or the
// context is canceled, the process times out, or cancel is invoked) and
// performs the synchronization wait that previously relied on Ctrl+C handling.
// Cancel quickly interrupts finalize and causes it to return the supplied error.
func Connect(ctx context.Context) (
	code whatsmeow.QRChannelItem,
	waitScan func(context.Context) error,
	finalize func(context.Context) error,
	err error,
) {
	device, err := getConnectDevice(ctx)
	if err != nil {
		return whatsmeow.QRChannelItem{}, nil, nil, fmt.Errorf("read whatsapp device: %w", err)
	}
	client := whatsmeow.NewClient(device, nil)
	qrCh, err := client.GetQRChannel(ctx)
	if err != nil {
		return whatsmeow.QRChannelItem{}, nil, nil, fmt.Errorf("start qr channel: %w", err)
	}
	if err := client.Connect(); err != nil {
		return whatsmeow.QRChannelItem{}, nil, nil, fmt.Errorf("connect: %w", err)
	}

	waitScan = func(ctx context.Context) error {
		stop := false
		for !stop {
			select {
			case item, ok := <-qrCh:
				if !ok {
					stop = true
				} else {
					switch {
					case item == whatsmeow.QRChannelSuccess:
						return nil
					case item == whatsmeow.QRChannelTimeout:
						return fmt.Errorf("qr timeout")
					case item.Event == whatsmeow.QRChannelEventError:
						return fmt.Errorf("qr error: %w", item.Error)
					default:
					}
				}
			case <-ctx.Done():
				stop = true
			}
		}

		return nil
	}

	finalize = func(ctx context.Context) error {
		defer client.Disconnect()

		select {
		case <-time.After(5 * time.Minute):
			return fmt.Errorf("timed out waiting for QR scan")
		case <-ctx.Done():
			return nil
		}
	}

	for item := range qrCh {
		switch {
		case item.Event == whatsmeow.QRChannelEventCode:
			return item, waitScan, finalize, nil
		case item == whatsmeow.QRChannelTimeout:
			return whatsmeow.QRChannelItem{}, nil, nil, fmt.Errorf("qr timeout")
		case item.Event == whatsmeow.QRChannelEventError:
			return whatsmeow.QRChannelItem{}, nil, nil, fmt.Errorf("qr error: %w", item.Error)
		case item == whatsmeow.QRChannelSuccess:
			return whatsmeow.QRChannelItem{}, nil, nil, fmt.Errorf("qr success before code")
		default:
		}
	}

	return whatsmeow.QRChannelItem{}, nil, nil, fmt.Errorf("qr channel closed unexpectedly")
}

// getConnectDevice returns an existing session device or creates a new device for pairing.
// Parameters: ctx supplies the context for container access.
// Returns: the stored or newly created device along with any error.
func getConnectDevice(ctx context.Context) (*store.Device, error) {
	path, err := sessionPath()
	if err != nil {
		return nil, err
	}
	container, err := newSessionContainer(path)
	if err != nil {
		return nil, fmt.Errorf("open whatsapp store: %w", err)
	}
	device, err := container.GetFirstDevice(ctx)
	if err != nil {
		return nil, fmt.Errorf("read whatsapp device: %w", err)
	}
	return device, nil
}

// getDevice returns the WhatsApp device stored in the session file and errors when the session is missing.
// Parameters: ctx supplies the context for container access.
// Returns: the stored device or an error, which is wrapped os.ErrNotExist when no session exists.
func getDevice(ctx context.Context) (*store.Device, error) {
	path, err := sessionPath()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, os.ErrNotExist
		}
		return nil, fmt.Errorf("stat whatsapp store: %w", err)
	}
	container, err := newSessionContainer(path)
	if err != nil {
		return nil, fmt.Errorf("open whatsapp store: %w", err)
	}
	device, err := container.GetFirstDevice(ctx)
	if err != nil {
		return nil, fmt.Errorf("read whatsapp device: %w", err)
	}
	return device, nil
}

// RemoveSession removes the WhatsApp session JSON file and history cache before a fresh pairing attempt.
// Parameters: none.
// Returns: any error encountered while removing existing files (ignores missing files).
func RemoveSession() error {
	path, err := sessionPath()
	if err != nil {
		return err
	}

	history := filepath.Join(filepath.Dir(path), "whatsapp.history.json")
	files := []string{path, path + ".tmp", history}
	for _, file := range files {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove %s: %w", file, err)
		}
	}
	return nil
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
