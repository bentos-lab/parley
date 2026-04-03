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

// TestGetRoundByIndexReturnsRound verifies the round payload is returned for valid indexes.
// Parameters: t provides the test context.
func TestGetRoundByIndexReturnsRound(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	filename := "alpha.2026-03-31-22-34-54.json"
	debateItem := debate.Debate{
		Name:  "Alpha",
		Topic: "Topic A",
		Rounds: []debate.DebateRound{
			{
				AgentID:  "agent-1",
				Message:  "First round",
				Weakness: "Weakness A",
				NewPoint: "New point A",
				Rebuttal: "Rebuttal A",
				Summary:  "Summary A",
			},
		},
	}
	require.NoError(t, writeDebateWithRounds(tempHome, filename, debateItem))

	handler := NewHandler()
	router := chi.NewRouter()
	router.Route("/api", func(r chi.Router) {
		handler.Routes(r)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/debates/alpha.2026-03-31-22-34-54/rounds/0", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	var payload roundResponse
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&payload))
	require.Equal(t, roundResponse{
		AgentID:  "agent-1",
		Content:  "First round",
		Weakness: "Weakness A",
		NewPoint: "New point A",
		Rebuttal: "Rebuttal A",
		Summary:  "Summary A",
	}, payload)
}

// TestGetRoundByIndexMissingDebateReturnsNotFound verifies unknown debate IDs return 404.
// Parameters: t provides the test context.
func TestGetRoundByIndexMissingDebateReturnsNotFound(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	handler := NewHandler()
	router := chi.NewRouter()
	router.Route("/api", func(r chi.Router) {
		handler.Routes(r)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/debates/missing.2026-03-31-22-34-54/rounds/0", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusNotFound, recorder.Code)
}

// TestGetRoundByIndexOutOfRangeReturnsNotFound verifies out-of-range index returns 404.
// Parameters: t provides the test context.
func TestGetRoundByIndexOutOfRangeReturnsNotFound(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	filename := "alpha.2026-03-31-22-34-54.json"
	debateItem := debate.Debate{
		Name:  "Alpha",
		Topic: "Topic A",
		Rounds: []debate.DebateRound{
			{AgentID: "agent-1", Message: "Only round"},
		},
	}
	require.NoError(t, writeDebateWithRounds(tempHome, filename, debateItem))

	handler := NewHandler()
	router := chi.NewRouter()
	router.Route("/api", func(r chi.Router) {
		handler.Routes(r)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/debates/alpha.2026-03-31-22-34-54/rounds/1", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusNotFound, recorder.Code)
}

// TestGetRoundByIndexInvalidIndexReturnsBadRequest verifies invalid index returns 400.
// Parameters: t provides the test context.
func TestGetRoundByIndexInvalidIndexReturnsBadRequest(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	handler := NewHandler()
	router := chi.NewRouter()
	router.Route("/api", func(r chi.Router) {
		handler.Routes(r)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/debates/alpha.2026-03-31-22-34-54/rounds/not-a-number", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

// writeDebateWithRounds writes a debate file in the expected storage directory.
// Parameters: home is the base HOME directory, filename is the debate file name, payload is the debate to store.
// Returns: an error if writing fails.
func writeDebateWithRounds(home string, filename string, payload debate.Debate) error {
	dir := filepath.Join(home, ".bentos", "parley")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(dir, filename)
	return os.WriteFile(path, data, 0o644)
}
