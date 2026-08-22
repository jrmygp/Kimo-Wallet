// Package middleware provides HTTP middleware for the gateway — currently
// just JWT verification. Per docs/CLAUDE.md §4, the gateway does not decide
// business rules; RequireAuth only decides whether a request carries a
// valid token, not whether the caller is allowed to do anything specific
// with it.
package middleware

import (
	"net/http"
	"strings"

	"github.com/jrmygp/kimo-wallet/apps/api-gateway/internal/httpresponse"
	"github.com/jrmygp/kimo-wallet/apps/api-gateway/internal/jwtauth"
)

// tokenVerifier is the subset of jwtauth.Verifier this middleware depends
// on, so it can be tested with a stub instead of a real secret/signature.
type tokenVerifier interface {
	Parse(tokenString string) (*jwtauth.Claims, error)
}

// RequireAuth rejects any request without a valid "Authorization: Bearer
// <token>" header with 401, and injects the token's claims into the
// request context for downstream handlers via jwtauth.ClaimsFromContext.
func RequireAuth(verifier tokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			token, ok := strings.CutPrefix(authHeader, "Bearer ")
			if !ok || token == "" {
				httpresponse.WriteError(w, http.StatusUnauthorized, "missing or malformed Authorization header")
				return
			}

			claims, err := verifier.Parse(token)
			if err != nil {
				httpresponse.WriteError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := jwtauth.WithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
