package rest

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/bentos-lab/parley/core"
	"github.com/bentos-lab/parley/core/debate"
)

// getAudio handles fetching the generated audio file for a debate by ID.
// Parameters: w is the response writer, r is the HTTP request.
// Returns: nothing.
func (h *Handler) getAudio(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	filename := debate.FilenameFromID(id)
	usecases, _, ok := h.loadUsecases(w)
	if !ok {
		return
	}
	output, err := usecases.GetDebateAudio.Execute(r.Context(), core.GetDebateAudioInput{
		Filename: filename,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	http.ServeFile(w, r, output.Path)
}
