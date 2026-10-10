package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/time/rate"

	"github.com/jrmygp/kimo-wallet/apps/api-gateway/internal/jwtauth"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

// Every request in these tests uses the same immediately-consumed burst, so
// no real time needs to pass for the "then blocks" assertion — the limiter
// starts each new key's bucket full, and burst+1 synchronous Allow() calls
// deterministically exhausts it without any sleep (docs/CLAUDE.md §5.5).
func TestRateLimiter_Middleware_AllowsBurstThenBlocks(t *testing.T) {
	rl := NewRateLimiter(rate.Every(time.Hour), 2, time.Hour)
	defer rl.Close()

	handler := rl.Middleware(func(r *http.Request) string { return "fixed-key" })(okHandler())

	statuses := make([]int, 0, 3)
	for range 3 {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/v1/transaction", nil))
		statuses = append(statuses, rec.Code)
	}

	want := []int{http.StatusOK, http.StatusOK, http.StatusTooManyRequests}
	for i, got := range statuses {
		if got != want[i] {
			t.Fatalf("request %d: status = %d, want %d (all statuses: %v)", i, got, want[i], statuses)
		}
	}
}

func TestRateLimiter_Middleware_DifferentKeysAreIndependent(t *testing.T) {
	rl := NewRateLimiter(rate.Every(time.Hour), 1, time.Hour)
	defer rl.Close()

	handler := rl.Middleware(func(r *http.Request) string { return r.Header.Get("X-Test-Key") })(okHandler())

	for _, key := range []string{"key-a", "key-b"} {
		req := httptest.NewRequest(http.MethodPost, "/v1/transaction", nil)
		req.Header.Set("X-Test-Key", key)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("first request for key %q: status = %d, want %d", key, rec.Code, http.StatusOK)
		}
	}

	// key-a's single-token bucket is now empty; a second request on the
	// same key is rejected regardless of key-b's independent bucket.
	req := httptest.NewRequest(http.MethodPost, "/v1/transaction", nil)
	req.Header.Set("X-Test-Key", "key-a")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("second request for key-a: status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}
}

func TestKeyByIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		want       string
	}{
		{name: "host:port is split", remoteAddr: "203.0.113.5:54321", want: "203.0.113.5"},
		{name: "no port falls back to the raw value", remoteAddr: "not-a-host-port", want: "not-a-host-port"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", nil)
			req.RemoteAddr = tt.remoteAddr
			if got := KeyByIP(req); got != tt.want {
				t.Fatalf("KeyByIP(%q) = %q, want %q", tt.remoteAddr, got, tt.want)
			}
		})
	}
}

func TestKeyByAuthenticatedUser(t *testing.T) {
	t.Run("uses the authenticated user id when claims are present", func(t *testing.T) {
		claims := &jwtauth.Claims{}
		claims.Subject = "11111111-1111-4111-8111-111111111111"

		req := httptest.NewRequest(http.MethodPost, "/v1/transaction", nil)
		req = req.WithContext(jwtauth.WithClaims(req.Context(), claims))

		want := "user:11111111-1111-4111-8111-111111111111"
		if got := KeyByAuthenticatedUser(req); got != want {
			t.Fatalf("KeyByAuthenticatedUser() = %q, want %q", got, want)
		}
	})

	t.Run("falls back to IP when claims are absent", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/v1/transaction", nil)
		req.RemoteAddr = "203.0.113.5:54321"

		want := "ip:203.0.113.5"
		if got := KeyByAuthenticatedUser(req); got != want {
			t.Fatalf("KeyByAuthenticatedUser() = %q, want %q", got, want)
		}
	})
}
