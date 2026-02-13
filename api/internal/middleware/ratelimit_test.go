package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestRateLimiterAllow(t *testing.T) {
	// Create rate limiter: 10 requests per second, burst of 5
	rl := NewRateLimiter(10, 5)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := rl.Limit(testHandler)

	// First 5 requests should succeed (within burst)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code, "request %d should succeed", i+1)
	}
}

func TestRateLimiterExceed(t *testing.T) {
	// Create rate limiter: 1 request per second, burst of 2
	rl := NewRateLimiter(1, 2)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := rl.Limit(testHandler)

	// First 2 requests should succeed (within burst)
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code, "request %d should succeed", i+1)
	}

	// Third request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusTooManyRequests, rr.Code)
	assert.Contains(t, rr.Body.String(), "rate limit exceeded")
}

func TestRateLimiterPerIP(t *testing.T) {
	// Create rate limiter: 1 request per second, burst of 1
	rl := NewRateLimiter(1, 1)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := rl.Limit(testHandler)

	// First IP: first request succeeds
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)
	assert.Equal(t, http.StatusOK, rr1.Code)

	// Second IP: should also succeed (different IP)
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.2:12345"
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusOK, rr2.Code)

	// First IP: second request should be rate limited
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "192.168.1.1:12345"
	rr3 := httptest.NewRecorder()
	handler.ServeHTTP(rr3, req3)
	assert.Equal(t, http.StatusTooManyRequests, rr3.Code)

	// Second IP: second request should also be rate limited
	req4 := httptest.NewRequest("GET", "/test", nil)
	req4.RemoteAddr = "192.168.1.2:12345"
	rr4 := httptest.NewRecorder()
	handler.ServeHTTP(rr4, req4)
	assert.Equal(t, http.StatusTooManyRequests, rr4.Code)
}

func TestRateLimiterIPHeaders(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name             string
		xForwardedFor    string
		xRealIP          string
		remoteAddr       string
		expectedIP       string
	}{
		{
			name:             "X-Forwarded-For takes priority",
			xForwardedFor:    "203.0.113.1",
			xRealIP:          "203.0.113.2",
			remoteAddr:       "203.0.113.3:12345",
			expectedIP:       "203.0.113.1",
		},
		{
			name:             "X-Real-IP when no X-Forwarded-For",
			xForwardedFor:    "",
			xRealIP:          "203.0.113.2",
			remoteAddr:       "203.0.113.3:12345",
			expectedIP:       "203.0.113.2",
		},
		{
			name:             "RemoteAddr when no headers",
			xForwardedFor:    "",
			xRealIP:          "",
			remoteAddr:       "203.0.113.3:12345",
			expectedIP:       "203.0.113.3:12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use a fresh rate limiter for each test
			rl := NewRateLimiter(1, 1)
			handler := rl.Limit(testHandler)

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}
			req.RemoteAddr = tt.remoteAddr

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			// First request should succeed
			assert.Equal(t, http.StatusOK, rr.Code)

			// Second request with same IP should be rate limited
			req2 := httptest.NewRequest("GET", "/test", nil)
			if tt.xForwardedFor != "" {
				req2.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req2.Header.Set("X-Real-IP", tt.xRealIP)
			}
			req2.RemoteAddr = tt.remoteAddr

			rr2 := httptest.NewRecorder()
			handler.ServeHTTP(rr2, req2)
			assert.Equal(t, http.StatusTooManyRequests, rr2.Code)
		})
	}
}

func TestRateLimiterRecovery(t *testing.T) {
	// Create rate limiter: 10 requests per second, burst of 1
	rl := NewRateLimiter(rate.Limit(10), 1)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := rl.Limit(testHandler)

	// First request succeeds
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)
	assert.Equal(t, http.StatusOK, rr1.Code)

	// Second request immediately fails
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.1:12345"
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusTooManyRequests, rr2.Code)

	// Wait for rate limiter to recover (100ms at 10 req/s)
	time.Sleep(150 * time.Millisecond)

	// Third request should succeed
	req3 := httptest.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "192.168.1.1:12345"
	rr3 := httptest.NewRecorder()
	handler.ServeHTTP(rr3, req3)
	assert.Equal(t, http.StatusOK, rr3.Code)
}

func TestGetVisitor(t *testing.T) {
	rl := NewRateLimiter(10, 5)

	// Get visitor for IP1
	limiter1a := rl.getVisitor("192.168.1.1")
	assert.NotNil(t, limiter1a)

	// Get visitor for same IP should return same limiter (pointer equality)
	limiter1b := rl.getVisitor("192.168.1.1")
	assert.Same(t, limiter1a, limiter1b, "should return the same limiter instance for the same IP")

	// Get visitor for different IP should return different limiter
	limiter2 := rl.getVisitor("192.168.1.2")
	assert.NotNil(t, limiter2)
	assert.NotSame(t, limiter1a, limiter2, "should return different limiter instances for different IPs")
}

func TestRateLimiterBurstCapacity(t *testing.T) {
	// Create rate limiter with burst of 3
	rl := NewRateLimiter(1, 3)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := rl.Limit(testHandler)

	// Send 3 requests immediately (should all succeed due to burst)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code, "request %d should succeed", i+1)
	}

	// 4th request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusTooManyRequests, rr.Code)
}
