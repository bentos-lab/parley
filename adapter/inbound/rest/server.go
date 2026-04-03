package rest

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"github.com/bentos-lab/parley/adapter/inbound/rest/middleware"
)

// NewServer constructs an HTTP server with debate routes wired.
// Parameters: usecases holds the debate usecases, addr is the listen address,
// defaultTTSProvider is used when no provider is specified.
// Returns: the configured HTTP server.
func NewServer(addr string) *http.Server {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logging)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
		MaxAge:         300,
	}))

	handler := NewHandler()
	router.Route("/api", func(r chi.Router) {
		handler.Routes(r)
	})
	RegisterStaticAssets(router)
	return &http.Server{
		Addr:    addr,
		Handler: router,
	}
}
