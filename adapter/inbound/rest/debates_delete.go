package rest

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/bentos-lab/parley/core"
	"github.com/bentos-lab/parley/core/debate"
)

// deleteDebate handles deleting a saved debate by ID.
// Parameters: w is the response writer, r is the HTTP request.
// Returns: nothing.
func (h *Handler) deleteDebate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	filename := debate.FilenameFromID(id)
	usecases, _, ok := h.loadUsecases(w)
	if !ok {
		return
	}
	if err := usecases.DeleteDebate.Execute(core.DeleteDebateInput{Filename: filename}); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
