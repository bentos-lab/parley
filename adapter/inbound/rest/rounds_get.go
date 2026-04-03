package rest

import (
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/bentos-lab/parley/core"
	"github.com/bentos-lab/parley/core/debate"
)

// getRound handles returning a single debate round by index.
// Parameters: w is the response writer, r is the HTTP request.
// Returns: nothing.
func (h *Handler) getRound(w http.ResponseWriter, r *http.Request) {
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
	output, err := usecases.LoadDebate.Execute(core.LoadDebateInput{Filename: filename})
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			writeError(w, http.StatusNotFound, "debate not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if index < 0 || index >= len(output.Debate.Rounds) {
		writeError(w, http.StatusNotFound, "round not found")
		return
	}
	round := output.Debate.Rounds[index]
	writeJSON(w, http.StatusOK, roundResponse{
		AgentID:  round.AgentID,
		Content:  round.Message,
		Weakness: round.Weakness,
		NewPoint: round.NewPoint,
		Rebuttal: round.Rebuttal,
		Summary:  round.Summary,
	})
}
