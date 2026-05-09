// Package middleware provides HTTP middleware including rate limiting.
package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RateLimiter implements token bucket rate limiting per client.
type RateLimiter struct {
	clients map[string]*bucket
	mu      sync.RWMutex
	limit   int
	window  time.Duration
}

type bucket struct {
	tokens   int
	lastSeen time.Time
	mu       sync.Mutex
}

// NewRateLimiter creates a rate limiter with the given requests per window.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*bucket),
		limit:   limit,
		window:  window,
	}
	go rl.cleanup()
	return rl
}

// Limit is HTTP middleware that rate-limits requests.
func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := rl.clientKey(r)

		rl.mu.Lock()
		b, exists := rl.clients[key]
		if !exists {
			b = &bucket{tokens: rl.limit, lastSeen: time.Now()}
			rl.clients[key] = b
		}
		rl.mu.Unlock()

		b.mu.Lock()
		now := time.Now()
		elapsed := now.Sub(b.lastSeen)
		refill := int(elapsed / rl.window * time.Duration(rl.limit))
		if refill > 0 {
			b.tokens = min(b.tokens+refill, rl.limit)
			b.lastSeen = now
		}

		if b.tokens <= 0 {
			b.mu.Unlock()
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.limit))
			w.Header().Set("X-RateLimit-Window", rl.window.String())
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"rate limit exceeded"}`))
			return
		}

		b.tokens--
		remaining := b.tokens
		b.mu.Unlock()

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(rl.limit))
		w.Header().Set("X-RateLimit-Window", rl.window.String())
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) clientKey(r *http.Request) string {
	// Prefer API key or authenticated user
	apiKey := r.Header.Get("X-API-Key")
	if apiKey != "" {
		return "api:" + apiKey
	}

	// Fall back to IP + path pattern
	ip := r.RemoteAddr
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		ip = strings.Split(fwd, ",")[0]
	}
	return "ip:" + ip
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, b := range rl.clients {
			b.mu.Lock()
			if now.Sub(b.lastSeen) > rl.window*2 {
				delete(rl.clients, key)
			}
			b.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
