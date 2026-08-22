// Package authtoken mints short-lived JWT access tokens for authenticated
// users. user-service owns identity, so it is the only thing that signs a
// token; verification happens in api-gateway (internal/jwtauth there),
// which never signs one.
package authtoken

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	issuer = "kimo-wallet-user-service"

	// AccessTokenTTL is deliberately short-lived with no refresh-token
	// rotation yet — see docs/agent-logs for the known-gap note. When a
	// token expires the client must call Login again.
	AccessTokenTTL = 15 * time.Minute
)

// Claims are the JWT claims minted for an authenticated user. api-gateway's
// internal/jwtauth.Claims mirrors this exact shape by hand — the two
// modules don't share Go code across the service boundary (same reasoning
// as each service having its own copy of the generated protobuf code), so
// if this shape changes, update jwtauth.Claims to match.
type Claims struct {
	jwt.RegisteredClaims
	PhoneNumber string `json:"phone_number"`
}

type Minter struct {
	secret []byte
}

func NewMinter(secret []byte) *Minter {
	return &Minter{secret: secret}
}

// Mint signs a new access token for the given user, valid for AccessTokenTTL.
func (m *Minter) Mint(userID, phoneNumber string) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenTTL)),
		},
		PhoneNumber: phoneNumber,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}

	return signed, nil
}
