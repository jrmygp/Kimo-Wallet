package authtoken

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestMinter_Mint(t *testing.T) {
	secret := []byte("test-secret-do-not-use-in-real-deploys")
	minter := NewMinter(secret)

	before := time.Now()
	tokenString, err := minter.Mint("11111111-1111-4111-8111-111111111111", "+6281234567890")
	after := time.Now()
	if err != nil {
		t.Fatalf("Mint() error = %v, want nil", err)
	}
	if tokenString == "" {
		t.Fatalf("Mint() returned empty token")
	}

	var claims Claims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		return secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		t.Fatalf("parse minted token: %v", err)
	}
	if !token.Valid {
		t.Fatalf("minted token is not valid")
	}

	if claims.Subject != "11111111-1111-4111-8111-111111111111" {
		t.Fatalf("got subject %q, want the user id", claims.Subject)
	}
	if claims.PhoneNumber != "+6281234567890" {
		t.Fatalf("got phone_number %q, want +6281234567890", claims.PhoneNumber)
	}
	if claims.Issuer != issuer {
		t.Fatalf("got issuer %q, want %q", claims.Issuer, issuer)
	}

	// jwt.NumericDate truncates to whole seconds (jwt.TimePrecision), so
	// compare against second-truncated bounds with a second of slack rather
	// than the exact before/after instants.
	wantExpiryLower := before.Add(AccessTokenTTL).Truncate(time.Second).Add(-time.Second)
	wantExpiryUpper := after.Add(AccessTokenTTL).Truncate(time.Second).Add(time.Second)
	gotExpiry := claims.ExpiresAt.Time
	if gotExpiry.Before(wantExpiryLower) || gotExpiry.After(wantExpiryUpper) {
		t.Fatalf("got expiry %v, want between %v and %v", gotExpiry, wantExpiryLower, wantExpiryUpper)
	}
}

func TestMinter_Mint_WrongSecretFailsVerification(t *testing.T) {
	minter := NewMinter([]byte("correct-secret"))

	tokenString, err := minter.Mint("11111111-1111-4111-8111-111111111111", "+6281234567890")
	if err != nil {
		t.Fatalf("Mint() error = %v, want nil", err)
	}

	var claims Claims
	_, err = jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		return []byte("wrong-secret"), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err == nil {
		t.Fatalf("expected verification with the wrong secret to fail, it succeeded")
	}
}
