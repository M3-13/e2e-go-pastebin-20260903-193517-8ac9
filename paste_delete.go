package main

import "net/http"

// deletePaste handles DELETE /pastes/{id}. Implemented by the DELETE /pastes/{id} ticket.
func (h *Handler) deletePaste(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "not implemented")
}
