// Package httpserver assembles the gateway's HTTP route table.
package httpserver

import "net/http"

// UserHandler is the subset of handler.UserHandler the router depends on.
type UserHandler interface {
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	GetUserByID(w http.ResponseWriter, r *http.Request)
}

type TransactionHandler interface {
	CreateTransaction(w http.ResponseWriter, r *http.Request)
}

// NewRouter builds the gateway's public HTTP surface. This is the only
// place an HTTP path is mapped to a handler — the mapping from here to an
// internal gRPC call happens inside the handler itself, per service.
//
// requireAuth guards routes that need a valid access token; Register and
// Login stay unguarded since you can't authenticate to get a token in the
// first place with one already required.
//
// authRateLimit and transactionRateLimit are separate limiters (docs/CLAUDE.md
// §6: rate-limit auth attempts and financial mutations) because they key on
// different things: authRateLimit runs on the unauthenticated auth routes
// and buckets by caller IP, while transactionRateLimit runs *inside*
// requireAuth on the money-moving route so it can bucket by the
// authenticated user id instead.
func NewRouter(
	userHandler UserHandler,
	transactionHandler TransactionHandler,
	requireAuth func(http.Handler) http.Handler,
	authRateLimit func(http.Handler) http.Handler,
	transactionRateLimit func(http.Handler) http.Handler,
) http.Handler {
	mux := http.NewServeMux()
	// Auth routes
	mux.Handle("POST /v1/auth/register", authRateLimit(http.HandlerFunc(userHandler.Register)))
	mux.Handle("POST /v1/auth/login", authRateLimit(http.HandlerFunc(userHandler.Login)))
	mux.Handle("GET /v1/user/{kimoId}", requireAuth(http.HandlerFunc(userHandler.GetUserByID)))

	// Transaction routes
	mux.Handle("POST /v1/transaction", requireAuth(transactionRateLimit(http.HandlerFunc(transactionHandler.CreateTransaction))))
	return mux
}
