package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bentos-lab/parley/config"
	"github.com/bentos-lab/parley/core"
	"github.com/bentos-lab/parley/core/contract"
	"github.com/bentos-lab/parley/core/debate"
	"github.com/bentos-lab/parley/wiring"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

type stubSummaryLLM struct {
	jsonResponse       string
	jsonResponses      []string
	generateJSONCalled bool
	generateJSONCalls  int
}

// Generate returns a placeholder response for non-JSON calls.
// Parameters: ctx is the request context, req is the LLM request payload.
// Returns: an empty string and nil error.
func (s *stubSummaryLLM) Generate(ctx context.Context, req contract.LLMRequest) (string, error) {
	return "", nil
}

// GenerateJSON returns the stubbed JSON response and records invocation.
// Parameters: ctx is the request context, req is the LLM request payload, schema is the expected schema.
// Returns: the JSON response string and nil error.
func (s *stubSummaryLLM) GenerateJSON(ctx context.Context, req contract.LLMRequest, schema *contract.LLMJSONSchema) (string, error) {
	s.generateJSONCalled = true
	s.generateJSONCalls++
	if len(s.jsonResponses) > 0 {
		response := s.jsonResponses[0]
		s.jsonResponses = s.jsonResponses[1:]
		return response, nil
	}
	return s.jsonResponse, nil
}

// GenerateStream returns closed channels to satisfy the interface.
// Parameters: ctx is the request context, req is the LLM request payload.
// Returns: closed channels for stream chunks and errors.
func (s *stubSummaryLLM) GenerateStream(ctx context.Context, req contract.LLMRequest) (<-chan contract.LLMStreamChunk, <-chan error) {
	chunkCh := make(chan contract.LLMStreamChunk)
	errCh := make(chan error, 1)
	close(chunkCh)
	close(errCh)
	return chunkCh, errCh
}

// TestGetDebateSummaryReturnsStoredSummary verifies stored summaries are returned when new is not set.
// Parameters: t provides the test context.
func TestGetDebateSummaryReturnsStoredSummary(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	debateItem := &debate.Debate{
		Name:  "Alpha",
		Topic: "Topic A",
		Agents: []debate.DebateAgent{
			{ID: "agent-1", Name: "Alex", Stance: "pro"},
		},
		Rounds: []debate.DebateRound{
			{AgentID: "agent-1", Message: "Round 1"},
		},
		Summary: &debate.DebateSummaryDetail{
			Agents:     [][]string{{"Point A"}},
			Conclusion: "Conclusion A",
		},
	}
	filename := "alpha.2026-03-31-22-34-54.json"
	require.NoError(t, debateItem.SaveAs(filename))

	handler := NewHandler(WithUsecasesLoader(func(w http.ResponseWriter) (*wiring.Usecases, config.Config, bool) {
		usecases := &wiring.Usecases{
			GenerateDebateSummary: &core.GenerateDebateSummaryUsecase{},
		}
		return usecases, config.Config{}, true
	}))
	router := chi.NewRouter()
	router.Route("/api", func(r chi.Router) {
		handler.Routes(r)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/debates/alpha.2026-03-31-22-34-54/summary", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	var payload debate.DebateSummaryDetail
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&payload))
	require.Equal(t, debateItem.Summary, &payload)
}

// TestGetDebateSummaryMissingDebateReturnsNotFound verifies missing debates return 404.
// Parameters: t provides the test context.
func TestGetDebateSummaryMissingDebateReturnsNotFound(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	handler := NewHandler(WithUsecasesLoader(func(w http.ResponseWriter) (*wiring.Usecases, config.Config, bool) {
		usecases := &wiring.Usecases{
			GenerateDebateSummary: &core.GenerateDebateSummaryUsecase{},
		}
		return usecases, config.Config{}, true
	}))
	router := chi.NewRouter()
	router.Route("/api", func(r chi.Router) {
		handler.Routes(r)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/debates/missing.2026-03-31-22-34-54/summary", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusNotFound, recorder.Code)
}

// TestGetDebateSummaryNoRoundsReturnsBadRequest verifies empty debates return 400.
// Parameters: t provides the test context.
func TestGetDebateSummaryNoRoundsReturnsBadRequest(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	debateItem := &debate.Debate{
		Name:  "Alpha",
		Topic: "Topic A",
	}
	filename := "alpha.2026-03-31-22-34-54.json"
	require.NoError(t, debateItem.SaveAs(filename))

	handler := NewHandler(WithUsecasesLoader(func(w http.ResponseWriter) (*wiring.Usecases, config.Config, bool) {
		usecases := &wiring.Usecases{
			GenerateDebateSummary: &core.GenerateDebateSummaryUsecase{},
		}
		return usecases, config.Config{}, true
	}))
	router := chi.NewRouter()
	router.Route("/api", func(r chi.Router) {
		handler.Routes(r)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/debates/alpha.2026-03-31-22-34-54/summary", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

// TestGetDebateSummaryNewForcesRegeneration verifies new=true regenerates the summary.
// Parameters: t provides the test context.
func TestGetDebateSummaryNewForcesRegeneration(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	debateItem := &debate.Debate{
		Name:  "Alpha",
		Topic: "Topic A",
		Agents: []debate.DebateAgent{
			{ID: "agent-1", Name: "Alex", Stance: "pro"},
		},
		Rounds: []debate.DebateRound{
			{AgentID: "agent-1", Message: "Round 1"},
		},
		Summary: &debate.DebateSummaryDetail{
			Agents:     [][]string{{"Old point"}},
			Conclusion: "Old conclusion",
		},
	}
	filename := "alpha.2026-03-31-22-34-54.json"
	require.NoError(t, debateItem.SaveAs(filename))

	llm := &stubSummaryLLM{
		jsonResponses: []string{
			`{"points":["New point"]}`,
			`{"final_conclusion":"New conclusion"}`,
		},
	}
	handler := NewHandler(WithUsecasesLoader(func(w http.ResponseWriter) (*wiring.Usecases, config.Config, bool) {
		usecases := &wiring.Usecases{
			GenerateDebateSummary: &core.GenerateDebateSummaryUsecase{
				LLMResolver: contract.LLMResolverFunc(func(provider string, model string) (contract.LLM, error) {
					return llm, nil
				}),
				LLMProvider: "test",
				Model:       "model",
			},
		}
		return usecases, config.Config{}, true
	}))
	router := chi.NewRouter()
	router.Route("/api", func(r chi.Router) {
		handler.Routes(r)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/debates/alpha.2026-03-31-22-34-54/summary?new=true", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.True(t, llm.generateJSONCalled)
	require.Equal(t, 2, llm.generateJSONCalls)
	var payload debate.DebateSummaryDetail
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&payload))
	require.Equal(t, "New conclusion", payload.Conclusion)
	require.Equal(t, [][]string{{"New point"}}, payload.Agents)
}
