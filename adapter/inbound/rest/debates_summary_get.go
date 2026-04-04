package rest

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/bentos-lab/parley/core"
	"github.com/bentos-lab/parley/core/debate"
)

// getDebateSummary handles returning a debate summary by ID.
// Parameters: w is the response writer, r is the HTTP request.
// Returns: nothing.
func (h *Handler) getDebateSummary(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	filename := debate.FilenameFromID(id)
	usecases, _, ok := h.loadUsecases(w)
	if !ok {
		return
	}
	forceNew := strings.EqualFold(r.URL.Query().Get("new"), "true")
	if usecases.GenerateDebateSummary == nil {
		writeError(w, http.StatusBadRequest, "summary generator is required")
		return
	}
	output, err := usecases.GenerateDebateSummary.Execute(r.Context(), core.GenerateDebateSummaryInput{
		Filename: filename,
		ForceNew: forceNew,
	})
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			writeError(w, http.StatusNotFound, "debate not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, output.Summary)
}
