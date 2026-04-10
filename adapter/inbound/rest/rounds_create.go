package rest

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/bentos-lab/parley/core"
	"github.com/bentos-lab/parley/core/debate"
)

// createRound handles adding a new round to a debate by ID.
// Parameters: w is the response writer, r is the HTTP request.
// Returns: nothing.
func (h *Handler) createRound(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	filename := debate.FilenameFromID(id)
	var req roundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	usecases, _, ok := h.loadUsecases(w)
	if !ok {
		return
	}
	output, err := usecases.CreateRound.Execute(r.Context(), core.CreateRoundInput{
		Filename:    filename,
		AgentID:     req.AgentID,
		Content:     req.Content,
		LLMProvider: "",
		LLMModel:    "",
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, roundResponse{
		AgentID:  output.Round.AgentID,
		Content:  output.Round.Message,
		Weakness: output.Round.Weakness,
		NewPoint: output.Round.NewPoint,
		Rebuttal: output.Round.Rebuttal,
		Summary:  output.Round.Summary,
	})
}
