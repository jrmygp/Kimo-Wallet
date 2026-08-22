package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jrmygp/kimo-wallet/apps/api-gateway/internal/jwtauth"
)

type stubVerifier struct {
	claims *jwtauth.Claims
	err    error
}

func (s stubVerifier) Parse(tokenString string) (*jwtauth.Claims, error) {
	return s.claims, s.err
}

func TestRequireAuth(t *testing.T) {
	validClaims := &jwtauth.Claims{PhoneNumber: "+6281234567890"}
	validClaims.Subject = "11111111-1111-4111-8111-111111111111"

	tests := []struct {
		name           string
		authHeader     string
		verifier       stubVerifier
		wantStatus     int
		wantNextCalled bool
	}{
		{
			name:           "missing Authorization header rejected with 401",
			authHeader:     "",
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "Authorization header without Bearer prefix rejected with 401",
			authHeader:     "sometoken",
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "Bearer prefix with empty token rejected with 401",
			authHeader:     "Bearer ",
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "verifier rejects the token with 401",
			authHeader:     "Bearer bad.token.here",
			verifier:       stubVerifier{err: jwtauth.ErrInvalidToken},
			wantStatus:     http.StatusUnauthorized,
			wantNextCalled: false,
		},
		{
			name:           "valid token calls next with claims in context",
			authHeader:     "Bearer good.token.here",
			verifier:       stubVerifier{claims: validClaims},
			wantStatus:     http.StatusOK,
			wantNextCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var nextCalled bool
			var gotClaims *jwtauth.Claims

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				gotClaims, _ = jwtauth.ClaimsFromContext(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			handler := RequireAuth(tt.verifier)(next)

			req := httptest.NewRequest(http.MethodGet, "/v1/users/me", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if nextCalled != tt.wantNextCalled {
				t.Fatalf("next called = %v, want %v", nextCalled, tt.wantNextCalled)
			}
			if tt.wantNextCalled {
				if gotClaims == nil || gotClaims.Subject != validClaims.Subject {
					t.Fatalf("expected claims to be propagated into context, got %+v", gotClaims)
				}
			}
		})
	}
}
