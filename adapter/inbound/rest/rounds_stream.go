package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/bentos-lab/parley/core"
	"github.com/bentos-lab/parley/core/debate"
)

// streamRounds handles streaming generated rounds for a debate by ID via SSE.
// Parameters: w is the response writer, r is the HTTP request.
// Returns: nothing.
func (h *Handler) streamRounds(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	filename := debate.FilenameFromID(id)
	count := 0
	if value := r.URL.Query().Get("n"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid n")
			return
		}
		count = parsed
	}
	if count <= 0 {
		count = 1
	}
	usecases, _, ok := h.loadUsecases(w)
	if !ok {
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}
	for i := 0; i < count; i++ {
		select {
		case <-r.Context().Done():
			return
		default:
		}
		output, err := usecases.CreateRound.Execute(r.Context(), core.CreateRoundInput{
			Filename: filename,
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		payload := roundResponse{
			AgentID:  output.Round.AgentID,
			Content:  output.Round.Message,
			Weakness: output.Round.Weakness,
			NewPoint: output.Round.NewPoint,
			Rebuttal: output.Round.Rebuttal,
			Summary:  output.Round.Summary,
		}
		data, err := json.Marshal(payload)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "encode event")
			return
		}
		fmt.Fprintf(w, "data: %s\n\n", string(data))
		flusher.Flush()
	}
}
