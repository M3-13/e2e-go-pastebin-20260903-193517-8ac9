package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// storeWithPaste returns a Store pre-populated with a single paste, so the
// delete handler can be exercised against an existing id without depending on
// the store's Create (owned by another ticket).
func storeWithPaste(id string) *Store {
	s := NewStore()
	s.mu.Lock()
	s.pastes[id] = Paste{
		ID:        id,
		Content:   "hello world",
		Language:  "text",
		CreatedAt: time.Now(),
	}
	s.mu.Unlock()
	return s
}

func TestDeletePasteExistingReturns204AndRemoves(t *testing.T) {
	s := storeWithPaste("abc123")
	h := NewHandler(s)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/pastes/abc123", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("DELETE existing returned %d, want 204", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Errorf("204 response should have no body, got %q", rec.Body.String())
	}

	s.mu.Lock()
	_, stillThere := s.pastes["abc123"]
	s.mu.Unlock()
	if stillThere {
		t.Error("paste should be removed from the store after DELETE")
	}
}

func TestGetAfterDeleteReturns404(t *testing.T) {
	s := storeWithPaste("abc123")
	h := NewHandler(s)

	del := httptest.NewRecorder()
	h.ServeHTTP(del, httptest.NewRequest(http.MethodDelete, "/pastes/abc123", nil))
	if del.Code != http.StatusNoContent {
		t.Fatalf("DELETE existing returned %d, want 204", del.Code)
	}

	get := httptest.NewRecorder()
	h.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/pastes/abc123", nil))
	if get.Code != http.StatusNotFound {
		t.Fatalf("GET after DELETE returned %d, want 404", get.Code)
	}
}

func TestDeletePasteTwiceReturns404(t *testing.T) {
	s := storeWithPaste("abc123")
	h := NewHandler(s)

	first := httptest.NewRecorder()
	h.ServeHTTP(first, httptest.NewRequest(http.MethodDelete, "/pastes/abc123", nil))
	if first.Code != http.StatusNoContent {
		t.Fatalf("first DELETE returned %d, want 204", first.Code)
	}

	second := httptest.NewRecorder()
	h.ServeHTTP(second, httptest.NewRequest(http.MethodDelete, "/pastes/abc123", nil))
	if second.Code != http.StatusNotFound {
		t.Fatalf("second DELETE returned %d, want 404", second.Code)
	}
}

func TestDeletePasteUnknownReturns404(t *testing.T) {
	s := NewStore()
	h := NewHandler(s)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/pastes/does-not-exist", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("DELETE unknown returned %d, want 404", rec.Code)
	}
}
