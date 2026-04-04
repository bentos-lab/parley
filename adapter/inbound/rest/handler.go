package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"os/exec"

	"github.com/go-chi/chi/v5"
	"go.mau.fi/whatsmeow"

	"github.com/bentos-lab/parley/adapter/inbound/whatsapp"
	"github.com/bentos-lab/parley/config"
	"github.com/bentos-lab/parley/shared/install"
	"github.com/bentos-lab/parley/wiring"
)

// Handler exposes HTTP endpoints for debates.
type Handler struct {
	lookPath              func(string) (string, error)
	runInstall            func(string) error
	checkWhatsAppSession  func(context.Context) (bool, error)
	removeWhatsAppSession func() error
	connectWhatsApp       func(context.Context) (whatsmeow.QRChannelItem, func(context.Context) error, func(context.Context) error, error)
	loadUsecasesHook      func(http.ResponseWriter) (*wiring.Usecases, config.Config, bool)
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

// WithCheckWhatsAppSession installs a custom hook for checking WhatsApp sessions.
// Parameters: fn is the session check function.
// Returns: the option for chaining.
func WithCheckWhatsAppSession(fn func(context.Context) (bool, error)) HandlerOption {
	return func(h *Handler) {
		h.checkWhatsAppSession = fn
	}
}

// WithRemoveWhatsAppSession installs a custom hook for removing WhatsApp sessions.
// Parameters: fn is the session removal function.
// Returns: the option for chaining.
func WithRemoveWhatsAppSession(fn func() error) HandlerOption {
	return func(h *Handler) {
		h.removeWhatsAppSession = fn
	}
}

// WithConnectWhatsApp installs a custom hook for connecting WhatsApp sessions.
// Parameters: fn is the connect function.
// Returns: the option for chaining.
func WithConnectWhatsApp(fn func(context.Context) (whatsmeow.QRChannelItem, func(context.Context) error, func(context.Context) error, error)) HandlerOption {
	return func(h *Handler) {
		h.connectWhatsApp = fn
	}
}

// WithUsecasesLoader installs a custom usecase loader hook for testing.
// Parameters: fn is the loader function.
// Returns: the option for chaining.
func WithUsecasesLoader(fn func(http.ResponseWriter) (*wiring.Usecases, config.Config, bool)) HandlerOption {
	return func(h *Handler) {
		h.loadUsecasesHook = fn
	}
}

// NewHandler creates a new HTTP handler instance.
// Parameters: options override handler dependencies for testing.
// Returns: a configured handler instance.
func NewHandler(options ...HandlerOption) *Handler {
	h := &Handler{
		lookPath:              exec.LookPath,
		runInstall:            install.Run,
		checkWhatsAppSession:  whatsapp.CheckExisted,
		removeWhatsAppSession: whatsapp.RemoveSession,
		connectWhatsApp:       whatsapp.Connect,
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

	router.Get("/connect/whatsapp", h.connectWhatsAppSession)
	router.Delete("/connect/whatsapp", h.deleteWhatsAppSession)

	router.Post("/debates", h.createDebate)
	router.Get("/debates", h.listDebates)
	router.Get("/debates/{id}", h.getDebate)
	router.Get("/debates/{id}/summary", h.getDebateSummary)
	router.Put("/debates/{id}", h.updateDebate)
	router.Delete("/debates/{id}", h.deleteDebate)
	router.Post("/debates/{id}/rounds", h.createRound)
	router.Get("/debates/{id}/rounds/{index}", h.getRound)
	router.Get("/debates/{id}/rounds/{index}/audio", h.getRoundAudio)
	router.Get("/debates/{id}/rounds/sse", h.streamRounds)
	router.Get("/debates/{id}/audio", h.getAudio)
}

// loadUsecases loads the runtime configuration and usecases for request handlers.
// Parameters: w is the response writer used for emitting error responses.
// Returns: the usecases, the loaded config, and a boolean indicating success.
func (h *Handler) loadUsecases(w http.ResponseWriter) (*wiring.Usecases, config.Config, bool) {
	if h.loadUsecasesHook != nil {
		return h.loadUsecasesHook(w)
	}

	usecases, cfg, err := wiring.LoadUsecases()
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
