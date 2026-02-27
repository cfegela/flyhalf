package util

import (
	"fmt"
	"regexp"
	"strings"
)

// SanitizeString trims whitespace from input
// Note: HTML escaping is handled on the frontend when rendering to prevent XSS
// Escaping on input causes double-escaping issues when editing existing data
func SanitizeString(s string) string {
	// Trim whitespace
	return strings.TrimSpace(s)
}

// SanitizeStrings sanitizes multiple strings
func SanitizeStrings(strings ...string) []string {
	result := make([]string, len(strings))
	for i, s := range strings {
		result[i] = SanitizeString(s)
	}
	return result
}

// Email validation regex (RFC 5322 simplified)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail checks if an email address is valid
func ValidateEmail(email string) error {
	email = strings.TrimSpace(strings.ToLower(email))

	if email == "" {
		return fmt.Errorf("email is required")
	}

	if len(email) > 254 {
		return fmt.Errorf("email must not exceed 254 characters")
	}

	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}
