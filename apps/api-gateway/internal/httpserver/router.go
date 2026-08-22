// Package httpserver assembles the gateway's HTTP route table.
package httpserver

import "net/http"

// UserHandler is the subset of handler.UserHandler the router depends on.
type UserHandler interface {
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
}

// NewRouter builds the gateway's public HTTP surface. This is the only
// place an HTTP path is mapped to a handler — the mapping from here to an
// internal gRPC call happens inside the handler itself, per service.
func NewRouter(userHandler UserHandler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/auth/register", userHandler.Register)
	mux.HandleFunc("POST /v1/auth/login", userHandler.Login)
	return mux
}
