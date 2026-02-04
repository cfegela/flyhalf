package auth

import (
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// Default bcrypt cost
const defaultBcryptCost = 12

// Global bcrypt cost (set via SetBcryptCost)
var bcryptCost = defaultBcryptCost

// SetBcryptCost sets the bcrypt cost for password hashing
func SetBcryptCost(cost int) {
	bcryptCost = cost
}

// Common weak passwords to reject
var commonPasswords = map[string]bool{
	"password":   true,
	"password1":  true,
	"password123": true,
	"12345678":   true,
	"123456789":  true,
	"qwerty":     true,
	"qwerty123":  true,
	"admin":      true,
	"admin123":   true,
	"welcome":    true,
	"welcome1":   true,
	"letmein":    true,
	"monkey":     true,
	"dragon":     true,
	"master":     true,
	"sunshine":   true,
	"princess":   true,
	"football":   true,
	"iloveyou":   true,
	"trustno1":   true,
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePassword checks if a password meets security requirements
func ValidatePassword(password string) error {
	// Check minimum length
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	// Check maximum length (bcrypt has a 72 byte limit)
	if len(password) > 72 {
		return fmt.Errorf("password must not exceed 72 characters")
	}

	// Check for common weak passwords
	if commonPasswords[strings.ToLower(password)] {
		return fmt.Errorf("password is too common, please choose a stronger password")
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}
