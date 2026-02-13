package auth

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	// Set a lower cost for faster tests
	originalCost := bcryptCost
	SetBcryptCost(4)
	t.Cleanup(func() {
		SetBcryptCost(originalCost)
	})

	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "simple password",
			password: "password123",
		},
		{
			name:     "complex password",
			password: "P@ssw0rd!Complex#2024",
		},
		{
			name:     "unicode password",
			password: "パスワード123!",
		},
		{
			name:     "very long password",
			password: strings.Repeat("a", 72), // bcrypt max length
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			assert.NoError(t, err)
			assert.NotEmpty(t, hash)
			assert.NotEqual(t, tt.password, hash)
			assert.True(t, strings.HasPrefix(hash, "$2a$")) // bcrypt format
		})
	}
}

func TestHashPasswordUniqueSalts(t *testing.T) {
	// Set a lower cost for faster tests
	originalCost := bcryptCost
	SetBcryptCost(4)
	t.Cleanup(func() {
		SetBcryptCost(originalCost)
	})

	password := "testpassword"
	hash1, err1 := HashPassword(password)
	hash2, err2 := HashPassword(password)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotEqual(t, hash1, hash2, "same password should produce different hashes due to salt")
}

func TestCheckPassword(t *testing.T) {
	// Set a lower cost for faster tests
	originalCost := bcryptCost
	SetBcryptCost(4)
	t.Cleanup(func() {
		SetBcryptCost(originalCost)
	})

	password := "correctPassword123!"
	hash, err := HashPassword(password)
	assert.NoError(t, err)

	tests := []struct {
		name     string
		password string
		hash     string
		expected bool
	}{
		{
			name:     "correct password",
			password: password,
			hash:     hash,
			expected: true,
		},
		{
			name:     "wrong password",
			password: "wrongPassword",
			hash:     hash,
			expected: false,
		},
		{
			name:     "empty password",
			password: "",
			hash:     hash,
			expected: false,
		},
		{
			name:     "case sensitive",
			password: "CORRECTPASSWORD123!",
			hash:     hash,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckPassword(tt.password, tt.hash)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		expectErr bool
		errMsg    string
	}{
		{
			name:      "valid password",
			password:  "SecureP@ss123",
			expectErr: false,
		},
		{
			name:      "valid complex password",
			password:  "MyP@ssw0rd!2024",
			expectErr: false,
		},
		{
			name:      "too short",
			password:  "Sh0rt!",
			expectErr: true,
			errMsg:    "at least 8 characters",
		},
		{
			name:      "too long (>72 chars)",
			password:  "VeryLongPassword123!" + strings.Repeat("a", 60) + "A!",
			expectErr: true,
			errMsg:    "must not exceed 72 characters",
		},
		{
			name:      "missing uppercase",
			password:  "lowercase123!",
			expectErr: true,
			errMsg:    "uppercase letter",
		},
		{
			name:      "missing lowercase",
			password:  "UPPERCASE123!",
			expectErr: true,
			errMsg:    "lowercase letter",
		},
		{
			name:      "missing digit",
			password:  "NoDigitsHere!",
			expectErr: true,
			errMsg:    "digit",
		},
		{
			name:      "missing special character",
			password:  "NoSpecial123",
			expectErr: true,
			errMsg:    "special character",
		},
		{
			name:      "common password - password",
			password:  "password",
			expectErr: true,
			errMsg:    "too common",
		},
		{
			name:      "common password - password123",
			password:  "password123",
			expectErr: true,
			errMsg:    "too common",
		},
		{
			name:      "common password - 12345678",
			password:  "12345678",
			expectErr: true,
			errMsg:    "too common",
		},
		{
			name:      "common password - qwerty (too short)",
			password:  "qwerty",
			expectErr: true,
			errMsg:    "at least 8 characters", // Fails length check before common check
		},
		{
			name:      "common password - admin (too short)",
			password:  "admin",
			expectErr: true,
			errMsg:    "at least 8 characters", // Fails length check before common check
		},
		{
			name:      "common password - qwerty123",
			password:  "qwerty123",
			expectErr: true,
			errMsg:    "too common",
		},
		{
			name:      "common password - admin123",
			password:  "admin123",
			expectErr: true,
			errMsg:    "too common",
		},
		{
			name:      "common password case insensitive",
			password:  "PASSWORD",
			expectErr: true,
			errMsg:    "too common",
		},
		{
			name:      "has all requirements with symbols",
			password:  "P@ssw0rd#2024",
			expectErr: false,
		},
		{
			name:      "has all requirements with punctuation",
			password:  "MyP@ss.word1",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
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

func TestSetBcryptCost(t *testing.T) {
	originalCost := bcryptCost
	t.Cleanup(func() {
		SetBcryptCost(originalCost)
	})

	SetBcryptCost(10)
	assert.Equal(t, 10, bcryptCost)

	SetBcryptCost(8)
	assert.Equal(t, 8, bcryptCost)
}

func TestBcryptCostAffectsHash(t *testing.T) {
	originalCost := bcryptCost
	t.Cleanup(func() {
		SetBcryptCost(originalCost)
	})

	password := "testPassword123!"

	// Hash with cost 4
	SetBcryptCost(4)
	hash4, err := HashPassword(password)
	assert.NoError(t, err)
	assert.Contains(t, hash4, "$2a$04$") // Cost 4 in hash

	// Hash with cost 6
	SetBcryptCost(6)
	hash6, err := HashPassword(password)
	assert.NoError(t, err)
	assert.Contains(t, hash6, "$2a$06$") // Cost 6 in hash

	// Both should validate the password
	assert.True(t, CheckPassword(password, hash4))
	assert.True(t, CheckPassword(password, hash6))
}
