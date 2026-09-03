package main

import "net/http"

// getPaste handles GET /pastes/{id}. Implemented by the GET /pastes/{id} ticket.
func (h *Handler) getPaste(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}
