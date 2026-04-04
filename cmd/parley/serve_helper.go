package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/bentos-lab/parley/adapter/inbound/rest"
	"github.com/bentos-lab/parley/adapter/inbound/whatsapp"
	"github.com/bentos-lab/parley/config"
	"github.com/bentos-lab/parley/wiring"
)

const defaultHTTPAddr = "localhost:8080"

// serveEngine holds the shared HTTP server and optional WhatsApp listener.
type serveEngine struct {
	server           *http.Server
	whatsappListener *whatsapp.Listener
}

// newServeEngine creates the HTTP server and tries to initialize the WhatsApp listener.
// Parameters: ctx carries cancellation for listener initialization, usecases and cfg provide shared services, addr is the HTTP listen address.
// Returns: the ready-to-run engine instance.
func newServeEngine(ctx context.Context, usecases *wiring.Usecases, cfg config.Config, addr string) *serveEngine {
	server := rest.NewServer(addr)
	var whatsappListener *whatsapp.Listener
	if usecases != nil {
		l, err := whatsapp.NewListener(ctx, usecases, cfg)
		if err != nil {
			log.Printf("WhatsApp listener disabled")
		} else {
			whatsappListener = l
		}
	}
	return &serveEngine{
		server:           server,
		whatsappListener: whatsappListener,
	}
}

// startWhatsAppListener begins the WhatsApp listener if it was initialized.
// Parameters: ctx controls the listener lifecycle.
// Returns: nothing.
func (e *serveEngine) startWhatsAppListener(ctx context.Context) {
	if e == nil || e.whatsappListener == nil {
		return
	}
	e.whatsappListener.Start(ctx)
}

// runRestServer starts listening on the HTTP server and normalizes ErrServerClosed to nil.
// Parameters: none.
// Returns: any fatal listen error.
func (e *serveEngine) runRestServer() error {
	if e == nil || e.server == nil {
		return nil
	}
	err := e.server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// shutdown gracefully shuts the HTTP server down using the provided context.
// Parameters: ctx controls the shutdown deadline.
// Returns: any error returned by Server.Shutdown.
func (e *serveEngine) shutdown(ctx context.Context) error {
	if e == nil || e.server == nil {
		return nil
	}
	return e.server.Shutdown(ctx)
}
