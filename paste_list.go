package main

import "net/http"

// listPastes handles GET /pastes. Implemented by the GET /pastes list ticket.
func (h *Handler) listPastes(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}
