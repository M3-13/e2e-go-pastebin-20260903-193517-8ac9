package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newStoreWithPastes returns a Store pre-populated with the given pastes.
// This lets the handler tests control the store state directly without going
// through Create, so an already-expired paste can be set up as well.
func newStoreWithPastes(pastes map[string]Paste) *Store {
	s := NewStore()
	s.pastes = pastes
	return s
}

// getPasteRequest performs a GET /pastes/{id} against a handler backed by s and
// returns the recorded response.
func getPasteRequest(t *testing.T, s *Store, id string) *httptest.ResponseRecorder {
	t.Helper()
	h := NewHandler(s)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/pastes/"+id, nil)
	h.ServeHTTP(rec, req)
	return rec
}

// assertJSONError checks that the recorded response carries an "error" field.
func assertJSONError(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("error response is not valid JSON: %v (body: %s)", err, rec.Body.String())
	}
	if body["error"] == "" {
		t.Errorf("error response should contain a non-empty \"error\" field: %s", rec.Body.String())
	}
}

func TestGetPasteExistingReturns200WithAllFields(t *testing.T) {
	id := "abc123"
	created := time.Now().UTC().Truncate(time.Second)
	expires := time.Now().UTC().Add(time.Hour).Truncate(time.Second)
	s := newStoreWithPastes(map[string]Paste{
		id: {
			ID:        id,
			Content:   "hello world",
			Language:  "text",
			CreatedAt: created,
			ExpiresAt: &expires,
		},
	})

	rec := getPasteRequest(t, s, id)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	body := rec.Body.String()
	for _, field := range []string{"id", "content", "language", "created_at", "expires_at"} {
		if !strings.Contains(body, `"`+field+`"`) {
			t.Errorf("response missing field %q: %s", field, body)
		}
	}

	var got Paste
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if got.ID != id {
		t.Errorf("id = %q, want %q", got.ID, id)
	}
	if got.Content != "hello world" {
		t.Errorf("content = %q, want %q", got.Content, "hello world")
	}
	if got.Language != "text" {
		t.Errorf("language = %q, want %q", got.Language, "text")
	}
	if !got.CreatedAt.Equal(created) {
		t.Errorf("created_at = %v, want %v", got.CreatedAt, created)
	}
	if got.ExpiresAt == nil || !got.ExpiresAt.Equal(expires) {
		t.Errorf("expires_at = %v, want %v", got.ExpiresAt, expires)
	}
}

func TestGetPasteUnknownIDReturns404(t *testing.T) {
	s := newStoreWithPastes(map[string]Paste{})

	rec := getPasteRequest(t, s, "nonexistent")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	assertJSONError(t, rec)
}

func TestGetPasteExpiredReturns404(t *testing.T) {
	expired := time.Now().Add(-time.Hour)
	s := newStoreWithPastes(map[string]Paste{
		"expired-id": {
			ID:        "expired-id",
			Content:   "secret",
			Language:  "text",
			CreatedAt: time.Now().Add(-2 * time.Hour),
			ExpiresAt: &expired,
		},
	})

	rec := getPasteRequest(t, s, "expired-id")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	assertJSONError(t, rec)
}

func TestGetPasteContentEscapedForHTML(t *testing.T) {
	s := newStoreWithPastes(map[string]Paste{
		"html": {
			ID:        "html",
			Content:   "<script>alert(1)</script>",
			Language:  "html",
			CreatedAt: time.Now(),
			ExpiresAt: nil,
		},
	})

	rec := getPasteRequest(t, s, "html")

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if strings.Contains(body, "<script>") {
		t.Errorf("response contains raw HTML, want escaped: %s", body)
	}
	if !strings.Contains(body, `\u003c`) {
		t.Errorf("response should escape < as \\u003c, got: %s", body)
	}
}
