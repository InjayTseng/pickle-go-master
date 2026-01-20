package shortcode

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
)

const (
	// DefaultLength is the default length of generated short codes
	DefaultLength = 8
	// Alphabet is the set of characters used in short codes
	// Using URL-safe characters: a-z, A-Z, 0-9 (no confusing chars like 0/O, 1/l/I)
	Alphabet = "23456789abcdefghjkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"
)

// Generate creates a new random short code
func Generate() string {
	return GenerateWithLength(DefaultLength)
}

// GenerateWithLength creates a new random short code with specified length
func GenerateWithLength(length int) string {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		// Fallback to base64 if crypto/rand fails
		b := make([]byte, length)
		rand.Read(b)
		return base64.URLEncoding.EncodeToString(b)[:length]
	}

	alphabetLen := len(Alphabet)
	result := make([]byte, length)
	for i, b := range bytes {
		result[i] = Alphabet[int(b)%alphabetLen]
	}
	return string(result)
}

// IsValid checks if a short code is valid
func IsValid(code string) bool {
	if len(code) < 6 || len(code) > 12 {
		return false
	}
	for _, c := range code {
		if !strings.ContainsRune(Alphabet, c) {
			return false
		}
	}
	return true
}
