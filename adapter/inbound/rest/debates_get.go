package rest

import (
	"errors"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/bentos-lab/parley/core"
	"github.com/bentos-lab/parley/core/debate"
)

// getDebate handles returning a saved debate by ID.
// Parameters: w is the response writer, r is the HTTP request.
// Returns: nothing.
func (h *Handler) getDebate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	filename := debate.FilenameFromID(id)
	usecases, _, ok := h.loadUsecases(w)
	if !ok {
		return
	}
	output, err := usecases.LoadDebate.Execute(core.LoadDebateInput{Filename: filename})
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			writeError(w, http.StatusNotFound, "debate not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, output.Debate)
}
