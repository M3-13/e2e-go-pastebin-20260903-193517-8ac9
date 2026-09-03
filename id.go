package main

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateID returns a new unique identifier for a paste: 16 random bytes
// hex-encoded into 32 hex characters.
func GenerateID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
