package foundation

import (
	"crypto/rand"
	"fmt"
	"io"
)

// NewUUID generates a v4 pseudo-random UUID.
// It relies on crypto/rand for entropy.
func NewUUID() (string, error) {
	var uuid [16]byte
	if _, err := io.ReadFull(rand.Reader, uuid[:]); err != nil {
		return "", fmt.Errorf("foundation: failed to read random bytes: %w", err)
	}

	// Set version (4) and variant (RFC4122) bits
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4],
		uuid[4:6],
		uuid[6:8],
		uuid[8:10],
		uuid[10:16]), nil
}
