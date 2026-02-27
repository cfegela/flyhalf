package util

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "trims whitespace",
			input:    "  hello world  ",
			expected: "hello world",
		},
		{
			name:     "preserves HTML (escaping handled on frontend)",
			input:    "<script>alert('xss')</script>",
			expected: "<script>alert('xss')</script>",
		},
		{
			name:     "preserves HTML and trims",
			input:    "  <b>bold</b>  ",
			expected: "<b>bold</b>",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: "",
		},
		{
			name:     "normal text",
			input:    "normal text",
			expected: "normal text",
		},
		{
			name:     "preserves quotes",
			input:    `"quoted" text`,
			expected: `"quoted" text`,
		},
		{
			name:     "preserves ampersand",
			input:    "A & B",
			expected: "A & B",
		},
		{
			name:     "preserves apostrophes",
			input:    "don't",
			expected: "don't",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeStrings(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "sanitizes multiple strings",
			input:    []string{"  hello  ", "<script>", "normal"},
			expected: []string{"hello", "<script>", "normal"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "single string",
			input:    []string{"  test  "},
			expected: []string{"test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeStrings(tt.input...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name      string
		email     string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid email",
			email:     "user@example.com",
			expectErr: false,
		},
		{
			name:      "valid email with plus",
			email:     "user+tag@example.com",
			expectErr: false,
		},
		{
			name:      "valid email with subdomain",
			email:     "user@mail.example.com",
			expectErr: false,
		},
		{
			name:      "valid email with hyphen",
			email:     "user-name@example.com",
			expectErr: false,
		},
		{
			name:      "valid email with underscore",
			email:     "user_name@example.com",
			expectErr: false,
		},
		{
			name:      "valid email with numbers",
			email:     "user123@example.com",
			expectErr: false,
		},
		{
			name:      "empty email",
			email:     "",
			expectErr: true,
			errMsg:    "email is required",
		},
		{
			name:      "whitespace only",
			email:     "   ",
			expectErr: true,
			errMsg:    "email is required",
		},
		{
			name:      "missing @",
			email:     "userexample.com",
			expectErr: true,
			errMsg:    "invalid email format",
		},
		{
			name:      "missing domain",
			email:     "user@",
			expectErr: true,
			errMsg:    "invalid email format",
		},
		{
			name:      "missing local part",
			email:     "@example.com",
			expectErr: true,
			errMsg:    "invalid email format",
		},
		{
			name:      "missing TLD",
			email:     "user@example",
			expectErr: true,
			errMsg:    "invalid email format",
		},
		{
			name:      "too long (>254 chars)",
			email:     "verylongemailverylongemailverylongemailverylongemailverylongemailverylongemailverylongemailverylongemailverylongemailverylongemailverylongemailverylongemailverylongemailverylongemailverylongemailverylongemailverylongemailverylongemailverylongemailverylongemail@example.com",
			expectErr: true,
			errMsg:    "email must not exceed 254 characters",
		},
		{
			name:      "invalid characters",
			email:     "user name@example.com",
			expectErr: true,
			errMsg:    "invalid email format",
		},
		{
			name:      "trims and lowercases",
			email:     "  USER@EXAMPLE.COM  ",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if tt.expectErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetIPFromRequest(t *testing.T) {
	tests := []struct {
		name             string
		xForwardedFor    string
		xRealIP          string
		remoteAddr       string
		expectedContains string
	}{
		{
			name:             "X-Forwarded-For takes priority",
			xForwardedFor:    "203.0.113.1",
			xRealIP:          "203.0.113.2",
			remoteAddr:       "203.0.113.3:12345",
			expectedContains: "203.0.113.1",
		},
		{
			name:             "X-Real-IP when no X-Forwarded-For",
			xForwardedFor:    "",
			xRealIP:          "203.0.113.2",
			remoteAddr:       "203.0.113.3:12345",
			expectedContains: "203.0.113.2",
		},
		{
			name:             "RemoteAddr when no headers",
			xForwardedFor:    "",
			xRealIP:          "",
			remoteAddr:       "203.0.113.3:12345",
			expectedContains: "203.0.113.3",
		},
		{
			name:             "RemoteAddr with IPv6",
			xForwardedFor:    "",
			xRealIP:          "",
			remoteAddr:       "[::1]:12345",
			expectedContains: "[::1]:12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a minimal http.Request with headers
			req := &http.Request{
				Header:     make(http.Header),
				RemoteAddr: tt.remoteAddr,
			}
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			result := GetIPFromRequest(req)
			assert.Contains(t, result, tt.expectedContains)
		})
	}
}
