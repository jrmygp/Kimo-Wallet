package middleware

import "net/http"

// CORS allows cross-origin requests from allowedOrigin. Needed because the
// frontend runs on a different origin (e.g. http://localhost:3000) than
// this gateway (e.g. http://localhost:8080) and calls it directly from the
// browser — without this, every request is blocked by the browser's own
// preflight check before it ever reaches a handler (see
// docs/agent-logs/2026-08-23 for how this was found: a real 405 on the
// browser's OPTIONS preflight, not a bug in the POST handler itself).
//
// Wraps the whole router (not a single route) so an OPTIONS preflight to
// ANY path gets a response — http.ServeMux's method-specific patterns
// (e.g. "POST /v1/auth/login") don't match OPTIONS at all, so this has to
// short-circuit before the mux's own routing runs.
//
// Only one origin is supported for now — this project has exactly one
// frontend. A real multi-origin allowlist can replace the single string
// if that's ever needed; not built speculatively.
func CORS(allowedOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
