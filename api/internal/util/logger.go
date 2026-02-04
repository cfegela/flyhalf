package util

import (
	"log"
	"net/http"

	"github.com/google/uuid"
)

// SecurityEvent represents different types of security events
type SecurityEvent string

const (
	EventLoginSuccess      SecurityEvent = "LOGIN_SUCCESS"
	EventLoginFailure      SecurityEvent = "LOGIN_FAILED"
	EventLogout            SecurityEvent = "LOGOUT"
	EventPasswordChange    SecurityEvent = "PASSWORD_CHANGED"
	EventTokenRefresh      SecurityEvent = "TOKEN_REFRESHED"
	EventPermissionDenied  SecurityEvent = "PERMISSION_DENIED"
	EventUserCreated       SecurityEvent = "USER_CREATED"
	EventUserUpdated       SecurityEvent = "USER_UPDATED"
	EventUserDeleted       SecurityEvent = "USER_DELETED"
	EventDemoReset         SecurityEvent = "DEMO_RESET"
	EventDemoReseed        SecurityEvent = "DEMO_RESEEDED"
	EventRateLimitExceeded SecurityEvent = "RATE_LIMIT_EXCEEDED"
)

// LogSecurityEvent logs a security-related event with structured data
func LogSecurityEvent(event SecurityEvent, userID *uuid.UUID, email string, ip string, details string) {
	userIDStr := "unknown"
	if userID != nil {
		userIDStr = userID.String()
	}

	if email == "" {
		email = "unknown"
	}

	log.Printf("[SECURITY] event=%s user_id=%s email=%s ip=%s details=%s",
		event, userIDStr, email, ip, details)
}

// GetIPFromRequest extracts the real IP address from the request
func GetIPFromRequest(r *http.Request) string {
	// Check X-Forwarded-For header (set by load balancers)
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}

	// Check X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
