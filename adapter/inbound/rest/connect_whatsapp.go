package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// connectWhatsAppSession handles the WhatsApp connect status and pairing flow.
// Parameters: w is the response writer, r is the HTTP request.
// Returns: nothing.
func (h *Handler) connectWhatsAppSession(w http.ResponseWriter, r *http.Request) {
	connectParam := r.URL.Query().Get("connect")
	if connectParam != "true" {
		h.getWhatsAppConnectionStatus(w, r)
		return
	}

	existed, err := h.checkWhatsAppSession(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if existed {
		writeError(w, http.StatusBadRequest, "whatsapp session already exists")
		return
	}

	code, waitScan, finalize, err := h.connectWhatsApp(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}

	timeoutMS := code.Timeout.Milliseconds()
	if err := writeSSEData(w, flusher, map[string]any{
		"code":    code.Code,
		"timeout": timeoutMS,
	}); err != nil {
		return
	}

	scanCtx, scanCancel := context.WithTimeout(r.Context(), code.Timeout)
	defer scanCancel()

	if err := waitScan(scanCtx); err != nil {
		writeSSEData(w, flusher, map[string]string{"error": err.Error()})
		return
	}
	if scanCtx.Err() != nil {
		if errors.Is(scanCtx.Err(), context.DeadlineExceeded) {
			writeSSEData(w, flusher, map[string]string{"error": "timed out waiting for scan"})
		}
		return
	}

	if err := writeSSEData(w, flusher, map[string]bool{"scanned": true}); err != nil {
		return
	}

	if err := finalize(r.Context()); err != nil {
		writeSSEData(w, flusher, map[string]string{"error": err.Error()})
		return
	}
}

// deleteWhatsAppSession handles removing the WhatsApp session store.
// Parameters: w is the response writer, r is the HTTP request.
// Returns: nothing.
func (h *Handler) deleteWhatsAppSession(w http.ResponseWriter, r *http.Request) {
	if err := h.removeWhatsAppSession(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// getWhatsAppConnectionStatus reports whether a WhatsApp session exists.
// Parameters: w is the response writer, r is the HTTP request.
// Returns: nothing.
func (h *Handler) getWhatsAppConnectionStatus(w http.ResponseWriter, r *http.Request) {
	existed, err := h.checkWhatsAppSession(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"connected": existed})
}

// writeSSEData emits a JSON payload as a server-sent event and flushes it.
// Parameters: w is the response writer, flusher flushes streamed responses, payload is the JSON payload.
// Returns: any error returned while encoding or writing.
func writeSSEData(w http.ResponseWriter, flusher http.Flusher, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "encode event")
		return err
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", string(data)); err != nil {
		return err
	}
	flusher.Flush()
	return nil
}
