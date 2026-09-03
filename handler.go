package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

// Handler routes HTTP requests to the paste endpoints.
type Handler struct {
	store *Store
}

// NewHandler returns a Handler backed by the given store.
func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

// ServeHTTP routes by method for /pastes and /pastes/{id}. Any defined path with
// an unsupported method answers 405 with an Allow header; unknown paths answer 404.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/pastes":
		switch r.Method {
		case http.MethodPost:
			h.createPaste(w, r)
		case http.MethodGet:
			h.listPastes(w, r)
		default:
			h.methodNotAllowed(w, "POST, GET")
		}
	default:
		if strings.HasPrefix(r.URL.Path, "/pastes/") && len(r.URL.Path) > len("/pastes/") {
			switch r.Method {
			case http.MethodGet:
				h.getPaste(w, r)
			case http.MethodDelete:
				h.deletePaste(w, r)
			default:
				h.methodNotAllowed(w, "GET, DELETE")
			}
		} else {
			writeError(w, http.StatusNotFound, "not found")
		}
	}
}

// methodNotAllowed sets the Allow header and answers 405 with a JSON error.
func (h *Handler) methodNotAllowed(w http.ResponseWriter, allow string) {
	w.Header().Set("Allow", allow)
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

// writeJSON serializes v as JSON with Content-Type application/json.
// encoding/json escapes HTML by default, keeping paste content XSS-safe.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError answers with a JSON body containing only the "error" field.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
