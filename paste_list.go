package main

import (
	"net/http"
	"time"
)

// pasteListItem is the list representation of a paste: metadata only, no content.
type pasteListItem struct {
	ID        string     `json:"id"`
	Language  string     `json:"language"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// pasteListItems converts stored pastes into their metadata-only list form,
// preserving the order List() returns them in (newest first).
func pasteListItems(pastes []Paste) []pasteListItem {
	items := make([]pasteListItem, 0, len(pastes))
	for _, p := range pastes {
		items = append(items, pasteListItem{
			ID:        p.ID,
			Language:  p.Language,
			CreatedAt: p.CreatedAt,
			ExpiresAt: p.ExpiresAt,
		})
	}
	return items
}

// listPastes handles GET /pastes, answering 200 with the metadata (no content)
// of all valid pastes, newest first.
func (h *Handler) listPastes(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, pasteListItems(h.store.List()))
}
