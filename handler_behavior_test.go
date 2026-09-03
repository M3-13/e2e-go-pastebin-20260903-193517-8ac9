package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestCreateListDeleteRoutesRegistered verifies, end-to-end through the handler,
// that the create, list and delete routes are all reachable. The delete subtest
// drives the real POST -> DELETE flow: it creates a paste, deletes exactly that
// id (204) and then deletes it again (404), which proves the route is registered
// and the paste was permanently removed — instead of sending DELETE against an
// id that was never created and misreading the resulting 404 as a missing route.
func TestCreateListDeleteRoutesRegistered(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		h := NewHandler(NewStore())

		req := httptest.NewRequest(http.MethodPost, "/pastes", strings.NewReader(`{"content":"hello","language":"text"}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("POST /pastes answered %d, want 201; route must be registered", rec.Code)
		}
		var resp pasteResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("POST /pastes response is not JSON: %v; body=%s", err, rec.Body.String())
		}
		if resp.ID == "" {
			t.Fatalf("POST /pastes response missing id: %s", rec.Body.String())
		}
	})

	t.Run("list", func(t *testing.T) {
		h := NewHandler(NewStore())

		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/pastes", nil))

		if rec.Code != http.StatusOK {
			t.Fatalf("GET /pastes answered %d, want 200; route must be registered", rec.Code)
		}
	})

	t.Run("delete", func(t *testing.T) {
		h := NewHandler(NewStore())

		create := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/pastes", strings.NewReader(`{"content":"hello","language":"text"}`))
		req.Header.Set("Content-Type", "application/json")
		h.ServeHTTP(create, req)

		if create.Code != http.StatusCreated {
			t.Fatalf("POST /pastes answered %d, want 201; body=%s", create.Code, create.Body.String())
		}
		var resp pasteResponse
		if err := json.Unmarshal(create.Body.Bytes(), &resp); err != nil {
			t.Fatalf("POST /pastes response is not JSON: %v; body=%s", err, create.Body.String())
		}
		if resp.ID == "" {
			t.Fatalf("POST /pastes response missing id: %s", create.Body.String())
		}

		del := httptest.NewRecorder()
		h.ServeHTTP(del, httptest.NewRequest(http.MethodDelete, "/pastes/"+resp.ID, nil))
		if del.Code != http.StatusNoContent {
			t.Fatalf("DELETE /pastes/%s answered %d, want 204; route must be registered", resp.ID, del.Code)
		}
		if del.Body.Len() != 0 {
			t.Errorf("204 response should have no body, got %q", del.Body.String())
		}

		second := httptest.NewRecorder()
		h.ServeHTTP(second, httptest.NewRequest(http.MethodDelete, "/pastes/"+resp.ID, nil))
		if second.Code != http.StatusNotFound {
			t.Fatalf("second DELETE /pastes/%s answered %d, want 404; paste must be gone", resp.ID, second.Code)
		}
	})
}
