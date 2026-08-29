package idgen

import (
	"crypto/rand"
	"fmt"
)

// kimoIDAlphabet is every character a KimoID can contain: uppercase
// alphanumeric only, per the product spec.
const kimoIDAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

const kimoIDLength = 12

// NewKimoID returns a random 12-character, uppercase alphanumeric KimoID
// (A-Z0-9) — the public-facing identifier users search each other by,
// as opposed to the internal UUID from NewV4. 36^12 (~62 bits) of keyspace
// makes an accidental collision vanishingly unlikely, but nowhere near as
// unlikely as a 122-bit UUID's — callers that create a user are expected
// to retry on a unique-constraint violation (see
// internal/service/user_service.go's Register), the same belt-and-suspenders
// approach this file already has no need for on the UUID side.
//
// Uses crypto/rand directly rather than a UUID/nanoid dependency, same
// reasoning as NewV4: a small, fixed amount of code. The `%len(alphabet)`
// reduction introduces a slight modulo bias (256 is not evenly divisible
// by 36), which is irrelevant here — a KimoID is a public lookup handle,
// not a secret or a security token.
func NewKimoID() (string, error) {
	raw := make([]byte, kimoIDLength)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate kimo id: %w", err)
	}

	id := make([]byte, kimoIDLength)
	for i, b := range raw {
		id[i] = kimoIDAlphabet[int(b)%len(kimoIDAlphabet)]
	}

	return string(id), nil
}
