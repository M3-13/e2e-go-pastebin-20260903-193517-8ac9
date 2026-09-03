package main

import (
	"sort"
	"sync"
	"time"
)

// Paste is a single stored paste. ExpiresAt is nil when the paste never expires.
type Paste struct {
	ID        string     `json:"id"`
	Content   string     `json:"content"`
	Language  string     `json:"language"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// Store is a thread-safe in-memory store of pastes.
type Store struct {
	mu     sync.Mutex
	pastes map[string]Paste
	idGen  func() (string, error)
}

// NewStore returns an initialized Store.
func NewStore() *Store {
	return &Store{
		pastes: make(map[string]Paste),
		idGen:  GenerateID,
	}
}

// Create stores a new paste and returns it. expiresInSeconds <= 0 means no expiry.
func (s *Store) Create(content, language string, expiresInSeconds int) (Paste, error) {
	id, err := s.idGen()
	if err != nil {
		return Paste{}, err
	}

	now := time.Now().UTC()
	p := Paste{
		ID:        id,
		Content:   content,
		Language:  language,
		CreatedAt: now,
	}
	if expiresInSeconds > 0 {
		exp := now.Add(time.Duration(expiresInSeconds) * time.Second)
		p.ExpiresAt = &exp
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.pastes[id] = p
	return p, nil
}

// Get returns the paste for id. The second result is false when the paste is
// unknown or has expired (lazy removal).
func (s *Store) Get(id string) (Paste, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.pastes[id]
	if !ok {
		return Paste{}, false
	}
	if expired(p) {
		delete(s.pastes, id)
		return Paste{}, false
	}
	return p, true
}

// List returns the metadata of all valid pastes, newest first. Expired pastes
// are removed from the map while iterating.
func (s *Store) List() []Paste {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Paste, 0, len(s.pastes))
	for id, p := range s.pastes {
		if expired(p) {
			delete(s.pastes, id)
			continue
		}
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out
}

// Delete removes the paste for id and reports whether it existed.
func (s *Store) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.pastes[id]; !ok {
		return false
	}
	delete(s.pastes, id)
	return true
}

// expired reports whether the paste has an expiry time that is already past.
func expired(p Paste) bool {
	return p.ExpiresAt != nil && !p.ExpiresAt.After(time.Now().UTC())
}
