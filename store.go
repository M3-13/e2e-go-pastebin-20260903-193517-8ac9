package main

import (
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
	return Paste{}, nil
}

// Get returns the paste for id. The second result is false when the paste is
// unknown or has expired (lazy removal).
func (s *Store) Get(id string) (Paste, bool) {
	return Paste{}, false
}

// List returns the metadata of all valid pastes, newest first.
func (s *Store) List() []Paste {
	return nil
}

// Delete removes the paste for id and reports whether it existed.
func (s *Store) Delete(id string) bool {
	return false
}
