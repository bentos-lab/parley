package rest

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"

	"github.com/bentos-lab/parley/adapter/inbound/rest/middleware"
	"github.com/bentos-lab/parley/adapter/inbound/whatsapp"
)

var resetWhatsappListenerCh = make(chan struct{}, 1)

type Server struct {
	server *http.Server
}

// NewServer constructs an HTTP server with debate routes wired.
// Parameters: usecases holds the debate usecases, addr is the listen address,
// defaultTTSProvider is used when no provider is specified.
// Returns: the configured HTTP server.
func NewServer(addr string) *Server {
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

	return &Server{
		&http.Server{
			Addr:    addr,
			Handler: router,
		},
	}
}

func (s *Server) ListenAndServe() error {
	go startWhatsappListener()
	return s.server.ListenAndServe()
}

func startWhatsappListener() {
	for {
		listenerCtx, cancel := context.WithCancel(context.Background())

		listener, err := whatsapp.NewListener(listenerCtx)
		if err != nil {
			select {
			case <-time.After(time.Minute):
			case <-resetWhatsappListenerCh:
			}
			continue
		}

		listener.Start(listenerCtx) // non-blocking

		<-resetWhatsappListenerCh
		fmt.Println("Restarting listener...")
		cancel()
	}
}
