package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCreateAndGetRoundtrip(t *testing.T) {
	s := NewStore()
	s.idGen = func() (string, error) { return "id-1", nil }

	p, err := s.Create("hello", "text", 0)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if p.ID != "id-1" {
		t.Fatalf("Create id = %q, want %q", p.ID, "id-1")
	}
	if p.Content != "hello" || p.Language != "text" {
		t.Fatalf("Create returned wrong content/language: %+v", p)
	}
	if p.CreatedAt.IsZero() {
		t.Fatal("CreatedAt should be set")
	}
	if p.ExpiresAt != nil {
		t.Fatalf("Create with 0 expiry should have nil ExpiresAt, got %v", p.ExpiresAt)
	}

	got, ok := s.Get("id-1")
	if !ok {
		t.Fatal("Get should find the created paste")
	}
	if got.Content != "hello" || got.Language != "text" {
		t.Fatalf("Get returned wrong paste: %+v", got)
	}
}

func TestCreateWithExpiry(t *testing.T) {
	s := NewStore()
	s.idGen = func() (string, error) { return "id-exp", nil }

	p, err := s.Create("x", "text", 60)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if p.ExpiresAt == nil {
		t.Fatal("Create with expiry should set ExpiresAt")
	}
	want := p.CreatedAt.Add(60 * time.Second)
	if !p.ExpiresAt.Equal(want) {
		t.Fatalf("ExpiresAt = %v, want %v", p.ExpiresAt, want)
	}
}

func TestGenerateIDIs32HexChars(t *testing.T) {
	id, err := GenerateID()
	if err != nil {
		t.Fatalf("GenerateID returned error: %v", err)
	}
	if len(id) != 32 {
		t.Fatalf("GenerateID length = %d, want 32", len(id))
	}
	for _, c := range id {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Fatalf("GenerateID returned non-hex char %q in %q", c, id)
		}
	}
}

func TestGenerateIDIsUnique(t *testing.T) {
	seen := make(map[string]struct{})
	for i := 0; i < 1000; i++ {
		id, err := GenerateID()
		if err != nil {
			t.Fatalf("GenerateID returned error: %v", err)
		}
		if _, ok := seen[id]; ok {
			t.Fatalf("GenerateID returned duplicate id %q", id)
		}
		seen[id] = struct{}{}
	}
}

func TestExpiredPasteIsNotReturned(t *testing.T) {
	s := NewStore()
	s.idGen = func() (string, error) { return "id-1", nil }

	if _, err := s.Create("x", "text", 1); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	s.mu.Lock()
	p := s.pastes["id-1"]
	past := time.Now().UTC().Add(-time.Second)
	p.ExpiresAt = &past
	s.pastes["id-1"] = p
	s.mu.Unlock()

	if _, ok := s.Get("id-1"); ok {
		t.Fatal("Get should return false for an expired paste")
	}
	if _, ok := s.pastes["id-1"]; ok {
		t.Fatal("expired paste should be lazily removed from the map")
	}
	if got := s.List(); len(got) != 0 {
		t.Fatalf("List should be empty after expiry, got %d entries", len(got))
	}
}

func TestGetUnknownReturnsFalse(t *testing.T) {
	s := NewStore()
	if _, ok := s.Get("nope"); ok {
		t.Fatal("Get for unknown id should return false")
	}
}

func TestListSortsNewestFirst(t *testing.T) {
	s := NewStore()
	ids := []string{"a", "b", "c"}
	i := 0
	s.idGen = func() (string, error) {
		id := ids[i]
		i++
		return id, nil
	}

	for range ids {
		if _, err := s.Create("x", "text", 0); err != nil {
			t.Fatalf("Create returned error: %v", err)
		}
	}

	base := time.Now().UTC()
	s.mu.Lock()
	for idx, id := range ids {
		p := s.pastes[id]
		p.CreatedAt = base.Add(time.Duration(idx) * time.Second)
		s.pastes[id] = p
	}
	s.mu.Unlock()

	got := s.List()
	if len(got) != 3 {
		t.Fatalf("List returned %d entries, want 3", len(got))
	}
	if got[0].ID != "c" || got[1].ID != "b" || got[2].ID != "a" {
		t.Fatalf("List not sorted newest first: got %q, %q, %q", got[0].ID, got[1].ID, got[2].ID)
	}
}

func TestListExcludesExpiredKeepsValid(t *testing.T) {
	s := NewStore()
	ids := []string{"valid", "expired"}
	i := 0
	s.idGen = func() (string, error) {
		id := ids[i]
		i++
		return id, nil
	}

	if _, err := s.Create("v", "text", 0); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if _, err := s.Create("e", "text", 60); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	s.mu.Lock()
	p := s.pastes["expired"]
	past := time.Now().UTC().Add(-time.Second)
	p.ExpiresAt = &past
	s.pastes["expired"] = p
	s.mu.Unlock()

	got := s.List()
	if len(got) != 1 {
		t.Fatalf("List should contain only valid pastes, got %d entries", len(got))
	}
	if got[0].ID != "valid" {
		t.Fatalf("List should contain %q, got %q", "valid", got[0].ID)
	}
}

func TestDeleteIdempotent(t *testing.T) {
	s := NewStore()
	s.idGen = func() (string, error) { return "id-1", nil }

	if _, err := s.Create("x", "text", 0); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if !s.Delete("id-1") {
		t.Fatal("first Delete should return true")
	}
	if s.Delete("id-1") {
		t.Fatal("second Delete should return false")
	}
	if _, ok := s.Get("id-1"); ok {
		t.Fatal("paste should be gone after Delete")
	}
}

func TestDeleteUnknownReturnsFalse(t *testing.T) {
	s := NewStore()
	if s.Delete("nope") {
		t.Fatal("Delete for unknown id should return false")
	}
}

func TestConcurrentAccess(t *testing.T) {
	s := NewStore()
	var counter int64
	s.idGen = func() (string, error) {
		n := atomic.AddInt64(&counter, 1)
		return fmt.Sprintf("id-%d", n), nil
	}

	const goroutines = 20
	const perGoroutine = 50

	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < perGoroutine; i++ {
				if _, err := s.Create("content", "text", 0); err != nil {
					t.Errorf("Create error: %v", err)
					return
				}
				s.List()
				s.Get("missing")
				s.Delete("missing")
			}
		}()
	}
	wg.Wait()

	if got := len(s.List()); got != goroutines*perGoroutine {
		t.Fatalf("expected %d pastes after concurrent access, got %d", goroutines*perGoroutine, got)
	}
}
