package api

import (
	"crypto/rand"
	"encoding/hex"
)

// newID returns a random hex id with the given prefix.
func newID(prefix string) string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return prefix + "_0000000000000000"
	}
	return prefix + "_" + hex.EncodeToString(b)
}

// codeAlphabet omits I, O, 0 and 1 so codes stay unambiguous when read aloud or
// re-typed. Its length (32) divides 256, so index = byte % 32 stays uniform.
const codeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

const codeLength = 6

// genCode returns a uniformly random invite code. Callers must handle the
// (rare) collision by retrying, since codes are unique per community.
func genCode() (string, error) {
	b := make([]byte, codeLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	out := make([]byte, codeLength)
	for i, v := range b {
		out[i] = codeAlphabet[int(v)%len(codeAlphabet)]
	}
	return string(out), nil
}
