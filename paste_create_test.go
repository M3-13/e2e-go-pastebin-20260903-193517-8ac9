package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// pasteResponse is the JSON body of a successful POST /pastes response.
type pasteResponse struct {
	ID        string     `json:"id"`
	Language  string     `json:"language"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at"`
}

func postPaste(t *testing.T, h http.Handler, body, contentType string) *httptest.ResponseRecorder {
	t.Helper()
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(http.MethodPost, "/pastes", nil)
	} else {
		req = httptest.NewRequest(http.MethodPost, "/pastes", strings.NewReader(body))
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func assertErrorBody(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int) {
	t.Helper()
	if rec.Code != wantStatus {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, wantStatus, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("error body is not JSON: %v; body=%s", err, rec.Body.String())
	}
	msg, ok := body["error"]
	if !ok || msg == "" {
		t.Fatalf("error body missing non-empty 'error' field: %s", rec.Body.String())
	}
	if len(body) != 1 {
		t.Errorf("error body should contain only 'error', got %d fields: %s", len(body), rec.Body.String())
	}
}

func TestCreatePasteReturns201AndPersistsToStore(t *testing.T) {
	store := NewStore()
	h := NewHandler(store)

	rec := postPaste(t, h, `{"content":"hello world","language":"go","expires_in_seconds":3600}`, "application/json")

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var resp pasteResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not JSON: %v; body=%s", err, rec.Body.String())
	}
	if resp.ID == "" {
		t.Fatalf("response id is empty: %s", rec.Body.String())
	}
	if resp.Language != "go" {
		t.Errorf("response language = %q, want %q", resp.Language, "go")
	}
	if resp.ExpiresAt == nil {
		t.Errorf("response expires_at is nil, want a timestamp for expires_in_seconds=3600")
	}
	if resp.CreatedAt.IsZero() {
		t.Errorf("response created_at is zero")
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err == nil {
		if _, ok := raw["content"]; ok {
			t.Errorf("response must not include content: %s", rec.Body.String())
		}
	}

	got, ok := store.Get(resp.ID)
	if !ok {
		t.Fatalf("paste %q not retrievable from store after POST", resp.ID)
	}
	if got.Content != "hello world" {
		t.Errorf("stored content = %q, want %q", got.Content, "hello world")
	}
	if got.Language != "go" {
		t.Errorf("stored language = %q, want %q", got.Language, "go")
	}
	if got.ID != resp.ID {
		t.Errorf("stored id = %q, want %q", got.ID, resp.ID)
	}
	if got.ExpiresAt == nil {
		t.Errorf("stored paste has nil expires_at, want a timestamp")
	}
}

func TestCreatePasteDefaultsLanguageToText(t *testing.T) {
	store := NewStore()
	h := NewHandler(store)

	rec := postPaste(t, h, `{"content":"no language"}`, "application/json")

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	var resp pasteResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not JSON: %v; body=%s", err, rec.Body.String())
	}
	if resp.Language != "text" {
		t.Errorf("response language = %q, want %q (default)", resp.Language, "text")
	}
	if resp.ExpiresAt != nil {
		t.Errorf("response expires_at = %v, want nil when no expiry", resp.ExpiresAt)
	}

	got, ok := store.Get(resp.ID)
	if !ok {
		t.Fatalf("paste %q not retrievable from store", resp.ID)
	}
	if got.Language != "text" {
		t.Errorf("stored language = %q, want %q", got.Language, "text")
	}
	if got.ExpiresAt != nil {
		t.Errorf("stored paste has expires_at = %v, want nil", got.ExpiresAt)
	}
}

func TestCreatePasteMissingContentReturns400(t *testing.T) {
	h := NewHandler(NewStore())
	rec := postPaste(t, h, `{"language":"go"}`, "application/json")
	assertErrorBody(t, rec, http.StatusBadRequest)
}

func TestCreatePasteEmptyContentReturns400(t *testing.T) {
	h := NewHandler(NewStore())
	rec := postPaste(t, h, `{"content":""}`, "application/json")
	assertErrorBody(t, rec, http.StatusBadRequest)
}

func TestCreatePasteExpiresInSecondsZeroReturns400(t *testing.T) {
	h := NewHandler(NewStore())
	rec := postPaste(t, h, `{"content":"x","expires_in_seconds":0}`, "application/json")
	assertErrorBody(t, rec, http.StatusBadRequest)
}

func TestCreatePasteExpiresInSecondsNegativeReturns400(t *testing.T) {
	h := NewHandler(NewStore())
	rec := postPaste(t, h, `{"content":"x","expires_in_seconds":-5}`, "application/json")
	assertErrorBody(t, rec, http.StatusBadRequest)
}

func TestCreatePasteBrokenJSONReturns400(t *testing.T) {
	h := NewHandler(NewStore())
	rec := postPaste(t, h, `{"content":"x",`, "application/json")
	assertErrorBody(t, rec, http.StatusBadRequest)
}

func TestCreatePasteWrongContentTypeReturns400(t *testing.T) {
	h := NewHandler(NewStore())
	rec := postPaste(t, h, `{"content":"x"}`, "text/plain")
	assertErrorBody(t, rec, http.StatusBadRequest)
}

func TestCreatePasteMissingContentTypeReturns400(t *testing.T) {
	h := NewHandler(NewStore())
	rec := postPaste(t, h, `{"content":"x"}`, "")
	assertErrorBody(t, rec, http.StatusBadRequest)
}

func TestCreatePasteBodyTooLargeReturns400(t *testing.T) {
	h := NewHandler(NewStore())
	big := `{"content":"` + strings.Repeat("a", maxPasteBodyBytes) + `"}`
	rec := postPaste(t, h, big, "application/json")
	assertErrorBody(t, rec, http.StatusBadRequest)
}

func TestCreatePasteWrongMethodReturns405(t *testing.T) {
	h := NewHandler(NewStore())
	req := httptest.NewRequest(http.MethodPut, "/pastes", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405; body=%s", rec.Code, rec.Body.String())
	}
	if allow := rec.Header().Get("Allow"); allow != "POST, GET" {
		t.Errorf("Allow = %q, want %q", allow, "POST, GET")
	}
}
