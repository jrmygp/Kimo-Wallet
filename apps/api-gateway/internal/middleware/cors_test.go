package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS(t *testing.T) {
	var nextCalled bool
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})
	handler := CORS("http://localhost:3000")(next)

	t.Run("OPTIONS preflight is answered directly, not passed to next", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest(http.MethodOptions, "/v1/auth/login", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
		if nextCalled {
			t.Fatalf("expected next not to be called for an OPTIONS preflight")
		}
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
			t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, "http://localhost:3000")
		}
	})

	t.Run("real request gets CORS headers and reaches next", func(t *testing.T) {
		nextCalled = false
		req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if !nextCalled {
			t.Fatalf("expected next to be called for a non-OPTIONS request")
		}
		if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
			t.Fatalf("Access-Control-Allow-Origin = %q, want %q", got, "http://localhost:3000")
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})
}
