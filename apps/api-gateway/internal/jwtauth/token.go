// Package jwtauth verifies JWT access tokens minted by user-service. It
// only parses/validates — it never signs a token. Issuing tokens is
// user-service's job (it owns identity); this gateway's job is to check
// the signature and expiry on the way in, per its documented
// "authenticate incoming requests" responsibility.
package jwtauth

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

// Claims mirrors the exact claim shape user-service's internal/authtoken
// package mints. This is not shared code across the two modules — they are
// separate Go modules with no shared internal package, same reasoning as
// each service already keeping its own copy of the generated protobuf
// code. If user-service's claim shape changes, update this by hand to match.
type Claims struct {
	jwt.RegisteredClaims
	PhoneNumber string `json:"phone_number"`
}

// ErrInvalidToken covers every verification failure — bad signature,
// expired, malformed, wrong algorithm. Deliberately not distinguished
// further: none of those reasons should be told apart for an
// unauthenticated caller (see docs/CLAUDE.md §6 on not leaking internals).
var ErrInvalidToken = errors.New("invalid or expired token")

type Verifier struct {
	secret []byte
}

func NewVerifier(secret []byte) *Verifier {
	return &Verifier{secret: secret}
}

// Parse validates the token's signature and expiry and returns its claims.
func (v *Verifier) Parse(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		return v.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
