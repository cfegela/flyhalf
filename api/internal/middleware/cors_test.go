package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCORSAllowedOrigin(t *testing.T) {
	cfg := &CORSConfig{
		AllowedOrigins: []string{"https://example.com", "http://localhost:3000"},
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name   string
		origin string
		expect bool
	}{
		{
			name:   "allowed origin https://example.com",
			origin: "https://example.com",
			expect: true,
		},
		{
			name:   "allowed origin http://localhost:3000",
			origin: "http://localhost:3000",
			expect: true,
		},
		{
			name:   "disallowed origin",
			origin: "https://evil.com",
			expect: false,
		},
		{
			name:   "no origin header",
			origin: "",
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			rr := httptest.NewRecorder()

			handler := CORS(cfg)(testHandler)
			handler.ServeHTTP(rr, req)

			if tt.expect {
				assert.Equal(t, tt.origin, rr.Header().Get("Access-Control-Allow-Origin"))
				assert.Equal(t, "true", rr.Header().Get("Access-Control-Allow-Credentials"))
				assert.Equal(t, "GET, POST, PUT, PATCH, DELETE, OPTIONS", rr.Header().Get("Access-Control-Allow-Methods"))
				assert.Equal(t, "Content-Type, Authorization", rr.Header().Get("Access-Control-Allow-Headers"))
				assert.Equal(t, "3600", rr.Header().Get("Access-Control-Max-Age"))
			} else {
				assert.Empty(t, rr.Header().Get("Access-Control-Allow-Origin"))
			}
		})
	}
}

func TestCORSPreflightRequest(t *testing.T) {
	cfg := &CORSConfig{
		AllowedOrigins: []string{"https://example.com"},
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called for OPTIONS request")
	})

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	rr := httptest.NewRecorder()

	handler := CORS(cfg)(testHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Equal(t, "https://example.com", rr.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSPreflightDisallowedOrigin(t *testing.T) {
	cfg := &CORSConfig{
		AllowedOrigins: []string{"https://example.com"},
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called for OPTIONS request")
	})

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "https://evil.com")
	rr := httptest.NewRecorder()

	handler := CORS(cfg)(testHandler)
	handler.ServeHTTP(rr, req)

	// Should still return 204 for OPTIONS, but no CORS headers
	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Empty(t, rr.Header().Get("Access-Control-Allow-Origin"))
}

func TestSecurityHeaders(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler := SecurityHeaders(testHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "nosniff", rr.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", rr.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", rr.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "strict-origin-when-cross-origin", rr.Header().Get("Referrer-Policy"))
}

func TestRequestSizeLimit(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to read body
		buf := make([]byte, 100)
		_, _ = r.Body.Read(buf)
		// We don't check error here as the limit is enforced at the reader level
		w.WriteHeader(http.StatusOK)
	})

	// Test with size under limit
	req := httptest.NewRequest("POST", "/test", nil)
	rr := httptest.NewRecorder()

	handler := RequestSizeLimit(1024)(testHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestTimeout(t *testing.T) {
	// Create a handler that checks if deadline is set
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deadline, ok := r.Context().Deadline()
		assert.True(t, ok, "context should have a deadline")
		assert.True(t, time.Until(deadline) > 0, "deadline should be in the future")
		assert.True(t, time.Until(deadline) <= 5*time.Second, "deadline should be within 5 seconds")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler := Timeout(5 * time.Second)(testHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestTimeoutContextCancellation(t *testing.T) {
	// Create a handler that checks if context can be canceled
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		select {
		case <-ctx.Done():
			t.Fatal("context should not be done immediately")
		default:
			// Context is not done yet, which is expected
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler := Timeout(5 * time.Second)(testHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestTimeoutInheritance(t *testing.T) {
	// Test that timeout middleware properly sets context
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify context is not the background context
		assert.NotEqual(t, context.Background(), r.Context())

		// Verify we can get a deadline
		_, ok := r.Context().Deadline()
		assert.True(t, ok)

		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler := Timeout(1 * time.Second)(testHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
