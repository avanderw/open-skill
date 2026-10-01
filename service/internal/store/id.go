package store

import (
	"crypto/rand"
	"encoding/hex"
)

// newID returns a random hex id with the given prefix, e.g. "m_ab12...".
func newID(prefix string) string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// Fallback (should not happen): use a time-derived value.
		return prefix + "_0000000000000000"
	}
	return prefix + "_" + hex.EncodeToString(b)
}
