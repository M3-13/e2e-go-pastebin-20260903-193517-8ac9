package main

import (
	"net/http"
	"strings"
)

// deletePaste handles DELETE /pastes/{id}. It removes the paste with the id from
// the path and answers 204 on success, or 404 with a JSON error when the id is
// unknown. The id is read from the request path rather than r.PathValue, because
// the router registers the handler under the catch-all pattern "/" (see
// main.go), which carries no named path wildcards.
func (h *Handler) deletePaste(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/pastes/")
	if h.store.Delete(id) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	writeError(w, http.StatusNotFound, "not found")
}
