package jwtauth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func mintToken(t *testing.T, secret []byte, claims Claims, method jwt.SigningMethod, key any) string {
	t.Helper()
	token := jwt.NewWithClaims(method, claims)
	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("sign test token: %v", err)
	}
	return signed
}

func TestVerifier_Parse(t *testing.T) {
	secret := []byte("test-secret-do-not-use-in-real-deploys")
	now := time.Now()

	validClaims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "11111111-1111-4111-8111-111111111111",
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		PhoneNumber: "+6281234567890",
	}

	t.Run("valid token parses with correct claims", func(t *testing.T) {
		tokenString := mintToken(t, secret, validClaims, jwt.SigningMethodHS256, secret)
		v := NewVerifier(secret)

		claims, err := v.Parse(tokenString)
		if err != nil {
			t.Fatalf("Parse() error = %v, want nil", err)
		}
		if claims.Subject != validClaims.Subject {
			t.Fatalf("got subject %q, want %q", claims.Subject, validClaims.Subject)
		}
		if claims.PhoneNumber != validClaims.PhoneNumber {
			t.Fatalf("got phone number %q, want %q", claims.PhoneNumber, validClaims.PhoneNumber)
		}
	})

	t.Run("wrong secret is rejected", func(t *testing.T) {
		tokenString := mintToken(t, secret, validClaims, jwt.SigningMethodHS256, secret)
		v := NewVerifier([]byte("a-completely-different-secret"))

		if _, err := v.Parse(tokenString); err == nil {
			t.Fatalf("expected error for wrong secret, got nil")
		}
	})

	t.Run("expired token is rejected", func(t *testing.T) {
		expiredClaims := validClaims
		expiredClaims.ExpiresAt = jwt.NewNumericDate(now.Add(-1 * time.Minute))
		tokenString := mintToken(t, secret, expiredClaims, jwt.SigningMethodHS256, secret)
		v := NewVerifier(secret)

		if _, err := v.Parse(tokenString); err == nil {
			t.Fatalf("expected error for expired token, got nil")
		}
	})

	t.Run("none-algorithm token is rejected (algorithm confusion)", func(t *testing.T) {
		tokenString := mintToken(t, secret, validClaims, jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType)
		v := NewVerifier(secret)

		if _, err := v.Parse(tokenString); err == nil {
			t.Fatalf("expected error for none-algorithm token, got nil")
		}
	})

	t.Run("malformed token string is rejected", func(t *testing.T) {
		v := NewVerifier(secret)

		if _, err := v.Parse("not-a-jwt-at-all"); err == nil {
			t.Fatalf("expected error for malformed token, got nil")
		}
	})
}
