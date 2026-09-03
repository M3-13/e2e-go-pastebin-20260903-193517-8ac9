package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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

func TestCreatePasteReturns201WithMetadata(t *testing.T) {
	h := NewHandler(NewStore())

	rec := postPaste(t, h, `{"content":"hello world","language":"go","expires_in_seconds":3600}`, "application/json")

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not JSON: %v; body=%s", err, rec.Body.String())
	}
	for _, key := range []string{"id", "language", "created_at", "expires_at"} {
		if _, ok := body[key]; !ok {
			t.Errorf("response missing %q field: %s", key, rec.Body.String())
		}
	}
	if _, ok := body["content"]; ok {
		t.Errorf("response must not include content: %s", rec.Body.String())
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
