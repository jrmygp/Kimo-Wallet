package jwtauth

import "context"

type contextKey int

const claimsContextKey contextKey = iota

// WithClaims returns a copy of ctx carrying claims, retrievable via
// ClaimsFromContext. Called by middleware.RequireAuth after a successful
// Parse.
func WithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey, claims)
}

// ClaimsFromContext retrieves the claims WithClaims stored on the context.
// Returns false if the request never went through RequireAuth.
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(*Claims)
	return claims, ok
}
