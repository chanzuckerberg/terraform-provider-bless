package util

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"strings"

	"github.com/pkg/errors"
)

// HashForState hashes the state
func HashForState(value string) string {
	if value == "" {
		return ""
	}
	hash := sha1.Sum([]byte(strings.TrimSpace(value)))
	return hex.EncodeToString(hash[:])
}

// GenerateRandomBytes generates random bytes
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, errors.Wrap(err, "Could not generate password")
	}
	return b, nil
}
