package rest

import "net/http"

// listDebates handles listing all debates.
// Parameters: w is the response writer, r is the HTTP request.
// Returns: nothing.
func (h *Handler) listDebates(w http.ResponseWriter, r *http.Request) {
	usecases, _, ok := h.loadUsecases(w)
	if !ok {
		return
	}
	output, err := usecases.ListDebates.Execute()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, output.Items)
}
