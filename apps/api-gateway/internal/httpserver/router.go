// Package httpserver assembles the gateway's HTTP route table.
package httpserver

import "net/http"

// UserHandler is the subset of handler.UserHandler the router depends on.
type UserHandler interface {
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	GetUserByID(w http.ResponseWriter, r *http.Request)
}

// NewRouter builds the gateway's public HTTP surface. This is the only
// place an HTTP path is mapped to a handler — the mapping from here to an
// internal gRPC call happens inside the handler itself, per service.
//
// requireAuth guards routes that need a valid access token; Register and
// Login stay unguarded since you can't authenticate to get a token in the
// first place with one already required.
func NewRouter(userHandler UserHandler, requireAuth func(http.Handler) http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/auth/register", userHandler.Register)
	mux.HandleFunc("POST /v1/auth/login", userHandler.Login)
	mux.Handle("GET /v1/user/{kimoId}", requireAuth(http.HandlerFunc(userHandler.GetUserByID)))
	return mux
}
