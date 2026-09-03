package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestListPastesReturnsMetadataNewestFirst(t *testing.T) {
	now := time.Now()
	expired := now.Add(-time.Hour)
	older := now.Add(-2 * time.Minute)
	newer := now.Add(-time.Minute)

	s := NewStore()
	s.pastes["older"] = Paste{ID: "older", Content: "secret-older", Language: "text", CreatedAt: older}
	s.pastes["newer"] = Paste{ID: "newer", Content: "secret-newer", Language: "go", CreatedAt: newer}
	s.pastes["expired"] = Paste{ID: "expired", Content: "secret-expired", Language: "text", CreatedAt: now, ExpiresAt: &expired}

	h := NewHandler(s)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/pastes", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /pastes returned %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}

	var items []pasteListItem
	if err := json.Unmarshal(rec.Body.Bytes(), &items); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}

	if len(items) != 2 {
		t.Fatalf("got %d items, want 2 (expired paste must be excluded)", len(items))
	}
	if items[0].ID != "newer" {
		t.Errorf("first item = %q, want %q (newest first)", items[0].ID, "newer")
	}
	if items[1].ID != "older" {
		t.Errorf("second item = %q, want %q", items[1].ID, "older")
	}

	var raw []map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("response is not a JSON array: %v", err)
	}
	for i, m := range raw {
		if _, ok := m["content"]; ok {
			t.Errorf("item %d leaks the content field", i)
		}
		for _, key := range []string{"id", "language", "created_at", "expires_at"} {
			if _, ok := m[key]; !ok {
				t.Errorf("item %d is missing field %q", i, key)
			}
		}
	}
}

func TestListPastesEmptyStoreReturnsEmptyArray(t *testing.T) {
	h := NewHandler(NewStore())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/pastes", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /pastes returned %d, want 200", rec.Code)
	}

	var raw []map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("empty list is not a JSON array: %v (body %q)", err, rec.Body.String())
	}
	if raw == nil {
		t.Fatal("empty list serialized as null, want []")
	}
	if len(raw) != 0 {
		t.Fatalf("empty list has %d items, want 0", len(raw))
	}
}
