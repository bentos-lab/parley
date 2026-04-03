package rest

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/bentos-lab/parley/core"
	"github.com/bentos-lab/parley/core/debate"
)

// updateDebate handles updating a saved debate by ID.
// Parameters: w is the response writer, r is the HTTP request.
// Returns: nothing.
func (h *Handler) updateDebate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	filename := debate.FilenameFromID(id)
	var body debate.Debate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	usecases, _, ok := h.loadUsecases(w)
	if !ok {
		return
	}
	if err := usecases.UpdateDebate.Execute(core.UpdateDebateInput{
		Filename: filename,
		Debate:   &body,
	}); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
