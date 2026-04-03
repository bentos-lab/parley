package rest

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/bentos-lab/parley/core"
	"github.com/bentos-lab/parley/core/debate"
)

// getRoundAudio handles fetching the audio file for a debate round by index.
// Parameters: w is the response writer, r is the HTTP request.
// Returns: nothing.
func (h *Handler) getRoundAudio(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	indexParam := chi.URLParam(r, "index")
	index, err := strconv.Atoi(indexParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid round index")
		return
	}
	filename := debate.FilenameFromID(id)
	usecases, _, ok := h.loadUsecases(w)
	if !ok {
		return
	}
	output, err := usecases.GetRoundAudio.Execute(r.Context(), core.GetRoundAudioInput{
		Filename: filename,
		Index:    index,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	http.ServeFile(w, r, output.Path)
}
