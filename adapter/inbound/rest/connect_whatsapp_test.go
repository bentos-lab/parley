package rest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"go.mau.fi/whatsmeow"
)

// TestConnectWhatsAppStatusFalse verifies the status endpoint reports no session.
// Parameters: t provides the test context.
func TestConnectWhatsAppStatusFalse(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	handler := NewHandler()
	router := chi.NewRouter()
	router.Route("/api", func(r chi.Router) {
		handler.Routes(r)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/connect/whatsapp", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), "\"connected\":false")
}

// TestConnectWhatsAppStatusTrue verifies the status endpoint reports an existing session.
// Parameters: t provides the test context.
func TestConnectWhatsAppStatusTrue(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	createWhatsAppSessionFile(t, tempHome)

	handler := NewHandler()
	router := chi.NewRouter()
	router.Route("/api", func(r chi.Router) {
		handler.Routes(r)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/connect/whatsapp", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Body.String(), "\"connected\":true")
}

// TestDeleteWhatsAppSessionRemovesFiles verifies deleting the session removes the stored file.
// Parameters: t provides the test context.
func TestDeleteWhatsAppSessionRemovesFiles(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	sessionPath := createWhatsAppSessionFile(t, tempHome)

	handler := NewHandler()
	router := chi.NewRouter()
	router.Route("/api", func(r chi.Router) {
		handler.Routes(r)
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/connect/whatsapp", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusNoContent, recorder.Code)
	_, err := os.Stat(sessionPath)
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))
}

// TestConnectWhatsAppWithExistingSessionReturnsBadRequest verifies connect is blocked when a session exists.
// Parameters: t provides the test context.
func TestConnectWhatsAppWithExistingSessionReturnsBadRequest(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	createWhatsAppSessionFile(t, tempHome)

	handler := NewHandler()
	router := chi.NewRouter()
	router.Route("/api", func(r chi.Router) {
		handler.Routes(r)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/connect/whatsapp?connect=true", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Contains(t, recorder.Body.String(), "whatsapp session already exists")
}

// TestConnectWhatsAppSSEEmitsCodeAndScan verifies the SSE flow emits code and scanned events.
// Parameters: t provides the test context.
func TestConnectWhatsAppSSEEmitsCodeAndScan(t *testing.T) {
	connectFn := func(ctx context.Context) (whatsmeow.QRChannelItem, func(context.Context) error, func(context.Context) error, error) {
		code := whatsmeow.QRChannelItem{
			Code:    "test-qr",
			Timeout: 2 * time.Second,
		}
		waitScan := func(ctx context.Context) error {
			return nil
		}
		finalize := func(ctx context.Context) error {
			return nil
		}
		return code, waitScan, finalize, nil
	}

	handler := NewHandler(
		WithCheckWhatsAppSession(func(ctx context.Context) (bool, error) {
			return false, nil
		}),
		WithConnectWhatsApp(connectFn),
	)
	router := chi.NewRouter()
	router.Route("/api", func(r chi.Router) {
		handler.Routes(r)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/connect/whatsapp?connect=true", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	body := recorder.Body.String()
	require.Equal(t, http.StatusOK, recorder.Code)
	require.True(t, strings.Contains(body, "\"timeout\":2000"))
	require.True(t, strings.Contains(body, "\"scanned\":true"))
	require.True(t, strings.Contains(body, "\"code\":\""))
	require.False(t, strings.Contains(body, "\"code\":\"\""))
}

// createWhatsAppSessionFile writes the WhatsApp session file under the expected home directory.
// Parameters: t provides the test context, home is the base HOME directory.
// Returns: the path to the session file.
func createWhatsAppSessionFile(t *testing.T, home string) string {
	t.Helper()
	dir := filepath.Join(home, ".bentos", "parley", "connect", "whatsapp")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	path := filepath.Join(dir, "session.json")
	require.NoError(t, os.WriteFile(path, []byte("{}"), 0o600))
	return path
}
