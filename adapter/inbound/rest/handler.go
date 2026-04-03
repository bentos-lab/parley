package rest

import (
	"encoding/json"
	"net/http"
	"os/exec"

	"github.com/go-chi/chi/v5"

	"github.com/bentos-lab/parley/config"
	"github.com/bentos-lab/parley/shared/install"
	"github.com/bentos-lab/parley/wiring"
)

// Handler exposes HTTP endpoints for debates.
type Handler struct {
	lookPath   func(string) (string, error)
	runInstall func(string) error
}

// HandlerOption configures optional handler behavior used in tests.
// Parameters: handler is the instance being configured.
// Returns: nothing.
type HandlerOption func(*Handler)

// WithLookPath installs a custom executable lookup hook.
// Parameters: fn is the lookup function.
// Returns: the option for chaining.
func WithLookPath(fn func(string) (string, error)) HandlerOption {
	return func(h *Handler) {
		h.lookPath = fn
	}
}

// WithRunInstall installs a custom installer hook.
// Parameters: fn is the installer function.
// Returns: the option for chaining.
func WithRunInstall(fn func(string) error) HandlerOption {
	return func(h *Handler) {
		h.runInstall = fn
	}
}

// NewHandler creates a new HTTP handler instance.
// Parameters: options override handler dependencies for testing.
// Returns: a configured handler instance.
func NewHandler(options ...HandlerOption) *Handler {
	h := &Handler{
		lookPath:   exec.LookPath,
		runInstall: install.Run,
	}
	for _, option := range options {
		option(h)
	}
	return h
}

// Routes registers HTTP routes on the provided router.
// Parameters: router is the chi router to register routes on.
// Returns: nothing.
func (h *Handler) Routes(router chi.Router) {
	router.Get("/config", h.getConfig)
	router.Put("/config", h.updateConfig)

	router.Post("/debates", h.createDebate)
	router.Get("/debates", h.listDebates)
	router.Get("/debates/{id}", h.getDebate)
	router.Put("/debates/{id}", h.updateDebate)
	router.Delete("/debates/{id}", h.deleteDebate)
	router.Post("/debates/{id}/rounds", h.createRound)
	router.Get("/debates/{id}/rounds/{index}", h.getRound)
	router.Get("/debates/{id}/rounds/{index}/audio", h.getRoundAudio)
	router.Get("/debates/{id}/rounds/sse", h.streamRounds)
	router.Get("/debates/{id}/audio", h.getAudio)
}

func (h *Handler) loadUsecases(w http.ResponseWriter) (*wiring.Usecases, config.Config, bool) {
	cfg, err := config.Load()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return nil, config.Config{}, false
	}
	usecases, err := wiring.BuildUsecases(cfg)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return nil, config.Config{}, false
	}
	return usecases, cfg, true
}

// writeJSON writes a JSON response with the provided status code.
// Parameters: w is the response writer, status is the HTTP status code, payload is the response body.
// Returns: nothing.
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// writeError writes a JSON error response with the provided status code and message.
// Parameters: w is the response writer, status is the HTTP status code, message is the error message.
// Returns: nothing.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
