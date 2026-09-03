package main

import (
	"encoding/json"
	"errors"
	"mime"
	"net/http"
)

// maxPasteBodyBytes is the upper bound for a POST /pastes request body (1 MiB).
const maxPasteBodyBytes = 1 << 20

// createPasteRequest is the JSON body accepted by POST /pastes.
// ExpiresInSeconds is a pointer so that an absent field (no expiry) can be told
// apart from an explicit value, which must then be > 0.
type createPasteRequest struct {
	Content          string `json:"content"`
	Language         string `json:"language"`
	ExpiresInSeconds *int   `json:"expires_in_seconds"`
}

// createPaste handles POST /pastes. It enforces the Content-Type, limits the
// request body to maxPasteBodyBytes, decodes and validates the JSON payload and
// delegates persistence to the store, answering 201 with the created paste's
// metadata.
func (h *Handler) createPaste(w http.ResponseWriter, r *http.Request) {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeError(w, http.StatusBadRequest, "content type must be application/json")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxPasteBodyBytes)

	var req createPasteRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			writeError(w, http.StatusBadRequest, "request body too large")
		} else {
			writeError(w, http.StatusBadRequest, "invalid JSON")
		}
		return
	}

	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "content is required")
		return
	}

	language := req.Language
	if language == "" {
		language = "text"
	}

	expiresInSeconds := 0
	if req.ExpiresInSeconds != nil {
		if *req.ExpiresInSeconds <= 0 {
			writeError(w, http.StatusBadRequest, "expires_in_seconds must be greater than 0")
			return
		}
		expiresInSeconds = *req.ExpiresInSeconds
	}

	paste, err := h.store.Create(req.Content, language, expiresInSeconds)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":         paste.ID,
		"language":   paste.Language,
		"created_at": paste.CreatedAt,
		"expires_at": paste.ExpiresAt,
	})
}
