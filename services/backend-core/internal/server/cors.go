package server

<<<<<<< HEAD
import "net/http"

func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Tenant-ID, X-Request-ID")
		w.Header().Set("Access-Control-Max-Age", "86400")
=======
import (
	"net/http"
	"os"
	"strings"
)

// corsAllowedOrigins returns the configured list of allowed origins.
// Defaults to local development URLs if CORS_ALLOWED_ORIGINS is not set.
func corsAllowedOrigins() []string {
	env := os.Getenv("CORS_ALLOWED_ORIGINS")
	if env == "" {
		return []string{"http://localhost:3000", "http://localhost:5173"}
	}
	var origins []string
	for _, o := range strings.Split(env, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			origins = append(origins, o)
		}
	}
	return origins
}

// isOriginAllowed checks if the given origin is in the allowlist.
func isOriginAllowed(origin string, allowed []string) bool {
	if origin == "" {
		return false
	}
	for _, a := range allowed {
		if a == origin {
			return true
		}
	}
	return false
}

// withCORS wraps a handler with configurable origin allowlist CORS support.
func withCORS(next http.HandlerFunc) http.HandlerFunc {
	allowed := corsAllowedOrigins()
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Only set CORS headers if the origin is in our allowlist
		if isOriginAllowed(origin, allowed) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Tenant-ID, X-Request-ID")
			w.Header().Set("Access-Control-Max-Age", "86400")
		}
>>>>>>> phase1/security-stability

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}
