package rest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/bentos-lab/parley/core/debate"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

// TestGetDebateByIDReturnsDebate verifies IDs resolve to filenames with a .json suffix.
// Parameters: t provides the test context.
func TestGetDebateByIDReturnsDebate(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	filename := "alpha.2026-03-31-22-34-54.json"
	require.NoError(t, writeDebateJSON(tempHome, filename, "Alpha", "Topic A"))

	handler := NewHandler()
	router := chi.NewRouter()
	router.Route("/api", func(r chi.Router) {
		handler.Routes(r)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/debates/alpha.2026-03-31-22-34-54", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
}

// TestGetDebateByIDWithJsonSuffixTreatsAsLiteral verifies .json is treated as part of the ID.
// Parameters: t provides the test context.
func TestGetDebateByIDWithJsonSuffixTreatsAsLiteral(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	filename := "alpha.2026-03-31-22-34-54.json"
	require.NoError(t, writeDebateJSON(tempHome, filename, "Alpha", "Topic A"))

	handler := NewHandler()
	router := chi.NewRouter()
	router.Route("/api", func(r chi.Router) {
		handler.Routes(r)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/debates/alpha.2026-03-31-22-34-54.json", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusNotFound, recorder.Code)
}

// writeDebateJSON writes a debate file in the expected storage directory.
// Parameters: home is the base HOME directory, filename is the debate file name,
// name is the debate display name, topic is the debate topic.
// Returns: an error if writing fails.
func writeDebateJSON(home string, filename string, name string, topic string) error {
	dir := filepath.Join(home, ".bentos", "parley")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	payload := debate.Debate{
		Name:  name,
		Topic: topic,
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(dir, filename)
	return os.WriteFile(path, data, 0o644)
}
