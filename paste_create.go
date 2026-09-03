package main

import "net/http"

// createPaste handles POST /pastes. Implemented by the POST /pastes ticket.
func (h *Handler) createPaste(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}
