package server

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// rateLimiter implements a simple in-memory token-bucket rate limiter per IP.
type rateLimiter struct {
	mu       sync.RWMutex
	buckets  map[string]*bucket
	limit    int
	window   time.Duration
	cleanFreq time.Duration
}

// bucket represents a token bucket for a single IP.
type bucket struct {
	tokens    int
	lastSeen  time.Time
	resetAt   time.Time
}

// newRateLimiter creates a rate limiter with the given request limit per time window.
func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		buckets:   make(map[string]*bucket),
		limit:     limit,
		window:    window,
		cleanFreq: 5 * time.Minute,
	}
	go rl.cleanupLoop()
	return rl
}

// allow checks if the given IP is allowed to proceed. Returns true if the
// request is within the rate limit, false otherwise. When false, it also
// returns the number of seconds until the bucket resets.
func (rl *rateLimiter) allow(ip string) (bool, int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, exists := rl.buckets[ip]
	if !exists || now.After(b.resetAt) {
		// First request or window expired — create new bucket
		rl.buckets[ip] = &bucket{
			tokens:   rl.limit - 1,
			lastSeen: now,
			resetAt:  now.Add(rl.window),
		}
		return true, 0
	}

	b.lastSeen = now
	if b.tokens > 0 {
		b.tokens--
		return true, 0
	}

	// Bucket exhausted — compute retry after
	retryAfter := int(b.resetAt.Sub(now).Seconds())
	if retryAfter < 1 {
		retryAfter = 1
	}
	return false, retryAfter
}

// cleanupLoop periodically removes stale entries to prevent memory growth.
func (rl *rateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanFreq)
	defer ticker.Stop()
	for range ticker.C {
		rl.cleanup()
	}
}

// cleanup removes buckets that have been inactive for more than 2 windows.
func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	threshold := time.Now().Add(-2 * rl.window)
	for ip, b := range rl.buckets {
		if b.lastSeen.Before(threshold) {
			delete(rl.buckets, ip)
		}
	}
}

// clientIP extracts the client IP from the request, preferring X-Forwarded-For
// but falling back to RemoteAddr. Only the first (leftmost) IP in
// X-Forwarded-For is used, as the rest may be attacker-controlled.
func clientIP(r *http.Request) string {
	fwd := r.Header.Get("X-Forwarded-For")
	if fwd != "" {
		// Take the first IP which is the original client
		if idx := 0; idx >= 0 {
			// fwd could be "client, proxy1, proxy2"
			// We take only the first one as it's the only somewhat-trusted one
			var first string
			for i := 0; i < len(fwd); i++ {
				if fwd[i] == ',' {
					first = fwd[:i]
					break
				}
			}
			if first == "" {
				first = fwd
			}
			first = stringTrimSpace(first)
			if first != "" {
				return first
			}
		}
	}

	if xri := r.Header.Get("X-Real-Ip"); xri != "" {
		return xri
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func stringTrimSpace(s string) string {
	var start, end int
	for start < len(s) && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	end = len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

// withRateLimit wraps a handler with rate limiting (5 requests per IP per minute).
func (s *Server) withRateLimit(next http.HandlerFunc) http.HandlerFunc {
	// 5 login attempts per IP per minute
	limiter := newRateLimiter(5, 1*time.Minute)
	return func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		allowed, retryAfter := limiter.allow(ip)
		if !allowed {
			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"too many requests","retry_after":` + strconv.Itoa(retryAfter) + `}`))
			return
		}
		next(w, r)
	}
}
