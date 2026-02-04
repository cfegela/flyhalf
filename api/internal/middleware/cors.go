package middleware

import (
	"context"
	"net/http"
	"time"
)

// CORSConfig defines the configuration for Cross-Origin Resource Sharing (CORS).
// CORS is a security feature that allows the API to specify which domains can access it
// from a browser context. This prevents unauthorized websites from making requests to the API
// on behalf of users.
//
// Configuration:
//   - AllowedOrigins: List of domains that are permitted to make cross-origin requests
//     Example: ["https://demo.flyhalf.app", "http://localhost:3000"]
//   - Methods: Standard REST methods (GET, POST, PUT, PATCH, DELETE) plus OPTIONS for preflight
//   - Headers: Content-Type for JSON bodies, Authorization for JWT tokens
//   - Credentials: true - allows cookies and authorization headers to be sent with requests
//   - Max-Age: 3600 seconds (1 hour) - browsers cache preflight responses to reduce overhead
//
// Security Notes:
//   - Never use "*" for AllowedOrigins in production when credentials are enabled
//   - Only add trusted domains to AllowedOrigins
//   - The middleware validates the Origin header against the allowlist on each request
type CORSConfig struct {
	AllowedOrigins []string
}

// CORS returns a middleware that handles Cross-Origin Resource Sharing (CORS).
// It validates the request origin against the configured allowed origins and sets
// appropriate CORS headers if the origin is permitted.
func CORS(cfg *CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Validate origin against allowlist
			allowed := false
			for _, allowedOrigin := range cfg.AllowedOrigins {
				if origin == allowedOrigin {
					allowed = true
					break
				}
			}

			// Only set CORS headers if origin is explicitly allowed
			if allowed {
				// Echo back the specific origin (more secure than "*" when using credentials)
				w.Header().Set("Access-Control-Allow-Origin", origin)

				// Allow cookies and authorization headers to be sent with requests
				w.Header().Set("Access-Control-Allow-Credentials", "true")

				// Define which HTTP methods the API supports
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

				// Define which request headers are allowed (Content-Type for JSON, Authorization for JWT)
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

				// Cache preflight responses for 1 hour to reduce overhead
				w.Header().Set("Access-Control-Max-Age", "3600")
			}

			// Handle preflight requests (OPTIONS) - browser sends these before actual requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// CSP removed - if needed, configure properly with all required sources:
		// script-src, style-src, img-src, connect-src, etc.

		next.ServeHTTP(w, r)
	})
}

// RequestSizeLimit limits the size of incoming request bodies
func RequestSizeLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Limit the request body size
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// Timeout adds a timeout to the request context
func Timeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
