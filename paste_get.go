package main

import (
	"net/http"
	"strings"
)

// getPaste handles GET /pastes/{id}. It reads the id from the path, looks the
// paste up in the store and answers 200 with the full paste, or 404 with a JSON
// error when the id is unknown or the paste has expired.
//
// The id is taken from the path tail. The router registers the handler under
// the "/" catch-all and routes /pastes/{id} here manually, so r.PathValue is
// never populated; the tail is the id.
func (h *Handler) getPaste(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/pastes/")

	paste, ok := h.store.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "paste not found")
		return
	}

	writeJSON(w, http.StatusOK, paste)
}
