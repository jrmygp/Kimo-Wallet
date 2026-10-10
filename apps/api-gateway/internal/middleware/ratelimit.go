package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/jrmygp/kimo-wallet/apps/api-gateway/internal/httpresponse"
	"github.com/jrmygp/kimo-wallet/apps/api-gateway/internal/jwtauth"
)

// KeyFunc derives the bucket key a request is rate-limited under.
type KeyFunc func(r *http.Request) string

// KeyByIP buckets by the caller's remote address — used for routes with no
// authenticated identity yet (register, login), per docs/CLAUDE.md §6
// ("rate-limit auth attempts"). RemoteAddr is the TCP peer address as seen
// by this process; there is no reverse proxy in front of the gateway yet,
// so an X-Forwarded-For header is not trusted here (a later deployment that
// adds one must set this from the proxy's trusted header instead).
func KeyByIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// KeyByAuthenticatedUser buckets by the caller's verified user id, so one
// user's financial-mutation attempts can't be throttled by another's
// traffic from the same IP (docs/CLAUDE.md §6: "rate-limit ... financial
// mutations"). Must only be used downstream of RequireAuth, which puts
// claims on the context; if claims are somehow absent, it falls back to
// KeyByIP rather than panicking or exempting the request from any limit.
func KeyByAuthenticatedUser(r *http.Request) string {
	claims, ok := jwtauth.ClaimsFromContext(r.Context())
	if !ok {
		return "ip:" + KeyByIP(r)
	}
	return "user:" + claims.Subject
}

type bucket struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter buckets requests per key with a token-bucket limit: `burst`
// requests are allowed immediately, refilling at `limit` events/sec after
// that. Buckets untouched for longer than idleTTL are swept, so a flood of
// distinct keys (e.g. many IPs) can't grow the map without bound.
//
// This is in-memory and per-process — it does not coordinate across
// multiple gateway instances. Fine for this project's current single
// api-gateway process; a multi-instance deployment would need a shared
// store (e.g. Redis) instead, which is a real architectural change, not
// made speculatively here (docs/CLAUDE.md §4.6).
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	limit   rate.Limit
	burst   int
	idleTTL time.Duration

	stopOnce sync.Once
	stop     chan struct{}
}

func NewRateLimiter(limit rate.Limit, burst int, idleTTL time.Duration) *RateLimiter {
	rl := &RateLimiter{
		buckets: make(map[string]*bucket),
		limit:   limit,
		burst:   burst,
		idleTTL: idleTTL,
		stop:    make(chan struct{}),
	}
	go rl.sweepLoop()
	return rl
}

func (rl *RateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, ok := rl.buckets[key]
	if !ok {
		b = &bucket{limiter: rate.NewLimiter(rl.limit, rl.burst)}
		rl.buckets[key] = b
	}
	b.lastSeen = time.Now()
	return b.limiter.Allow()
}

func (rl *RateLimiter) sweepLoop() {
	ticker := time.NewTicker(rl.idleTTL)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			cutoff := time.Now().Add(-rl.idleTTL)
			rl.mu.Lock()
			for key, b := range rl.buckets {
				if b.lastSeen.Before(cutoff) {
					delete(rl.buckets, key)
				}
			}
			rl.mu.Unlock()
		case <-rl.stop:
			return
		}
	}
}

// Close stops the background sweep goroutine. Call once at shutdown.
func (rl *RateLimiter) Close() {
	rl.stopOnce.Do(func() { close(rl.stop) })
}

// Middleware rejects a request with 429 once its key (as derived by keyFn)
// has exceeded the configured rate, otherwise passes it through unchanged.
// This is a liveness guard only, never a business rule — it has no opinion
// on *why* a caller is retrying quickly, only on how often (docs/CLAUDE.md
// §4.2: the gateway holds no domain logic).
func (rl *RateLimiter) Middleware(keyFn KeyFunc) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !rl.allow(keyFn(r)) {
				httpresponse.WriteError(w, http.StatusTooManyRequests, "too many requests, please slow down and try again")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
