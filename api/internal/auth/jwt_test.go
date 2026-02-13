package auth

import (
	"testing"
	"time"

	"github.com/cfegela/flyhalf/internal/config"
	"github.com/cfegela/flyhalf/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGenerateAccessToken(t *testing.T) {
	jwtService := NewJWTService(&config.JWTConfig{
		AccessSecret:     "test-secret",
		AccessExpiryMin:  15,
		RefreshExpiryDay: 7,
	})

	user := &model.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  model.RoleUser,
	}

	token, err := jwtService.GenerateAccessToken(user)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate the token
	claims, err := jwtService.ValidateAccessToken(token)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, user.Role, claims.Role)
}

func TestGenerateAccessTokenClaims(t *testing.T) {
	jwtService := NewJWTService(&config.JWTConfig{
		AccessSecret:     "test-secret",
		AccessExpiryMin:  15,
		RefreshExpiryDay: 7,
	})

	user := &model.User{
		ID:    uuid.New(),
		Email: "user@example.com",
		Role:  model.RoleAdmin,
	}

	token, err := jwtService.GenerateAccessToken(user)
	assert.NoError(t, err)

	claims, err := jwtService.ValidateAccessToken(token)
	assert.NoError(t, err)

	// Verify all claims
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, model.RoleAdmin, claims.Role)
	assert.Equal(t, user.ID.String(), claims.Subject)
	assert.NotNil(t, claims.IssuedAt)
	assert.NotNil(t, claims.ExpiresAt)

	// Verify expiry is ~15 minutes from now
	expectedExpiry := time.Now().Add(15 * time.Minute)
	assert.WithinDuration(t, expectedExpiry, claims.ExpiresAt.Time, 5*time.Second)
}

func TestValidateAccessTokenExpired(t *testing.T) {
	jwtService := NewJWTService(&config.JWTConfig{
		AccessSecret:     "test-secret",
		AccessExpiryMin:  -1, // Already expired
		RefreshExpiryDay: 7,
	})

	user := &model.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  model.RoleUser,
	}

	token, err := jwtService.GenerateAccessToken(user)
	assert.NoError(t, err)

	// Wait a moment to ensure expiry
	time.Sleep(10 * time.Millisecond)

	_, err = jwtService.ValidateAccessToken(token)
	assert.Error(t, err)
}

func TestValidateAccessTokenInvalid(t *testing.T) {
	jwtService := NewJWTService(&config.JWTConfig{
		AccessSecret:     "test-secret",
		AccessExpiryMin:  15,
		RefreshExpiryDay: 7,
	})

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "malformed token",
			token: "not.a.valid.token",
		},
		{
			name:  "random string",
			token: "randomstring",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := jwtService.ValidateAccessToken(tt.token)
			assert.Error(t, err)
		})
	}
}

func TestValidateAccessTokenWrongSecret(t *testing.T) {
	// Create token with one secret
	jwtService1 := NewJWTService(&config.JWTConfig{
		AccessSecret:     "secret-1",
		AccessExpiryMin:  15,
		RefreshExpiryDay: 7,
	})

	user := &model.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  model.RoleUser,
	}

	token, err := jwtService1.GenerateAccessToken(user)
	assert.NoError(t, err)

	// Try to validate with different secret
	jwtService2 := NewJWTService(&config.JWTConfig{
		AccessSecret:     "secret-2",
		AccessExpiryMin:  15,
		RefreshExpiryDay: 7,
	})

	_, err = jwtService2.ValidateAccessToken(token)
	assert.Error(t, err)
}

func TestValidateAccessTokenWrongAlgorithm(t *testing.T) {
	jwtService := NewJWTService(&config.JWTConfig{
		AccessSecret:     "test-secret",
		AccessExpiryMin:  15,
		RefreshExpiryDay: 7,
	})

	// Create a token with a non-HMAC algorithm (RS256 - RSA)
	// This requires a private key, so we'll just test with a malformed token instead
	// The validation code checks for HMAC-based algorithms

	// Use None algorithm which should be rejected
	token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"user_id": uuid.New().String(),
		"email":   "test@example.com",
	})
	tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	assert.NoError(t, err)

	// Should fail due to algorithm mismatch (None is not HMAC)
	_, err = jwtService.ValidateAccessToken(tokenString)
	assert.Error(t, err)
}

func TestGenerateRefreshToken(t *testing.T) {
	jwtService := NewJWTService(&config.JWTConfig{
		AccessSecret:     "test-secret",
		AccessExpiryMin:  15,
		RefreshExpiryDay: 7,
	})

	token1, err := jwtService.GenerateRefreshToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, token1)

	token2, err := jwtService.GenerateRefreshToken()
	assert.NoError(t, err)
	assert.NotEmpty(t, token2)

	// Tokens should be unique
	assert.NotEqual(t, token1, token2)
}

func TestHashRefreshToken(t *testing.T) {
	jwtService := NewJWTService(&config.JWTConfig{
		AccessSecret:     "test-secret",
		AccessExpiryMin:  15,
		RefreshExpiryDay: 7,
	})

	token := "test-refresh-token"
	hash1 := jwtService.HashRefreshToken(token)
	hash2 := jwtService.HashRefreshToken(token)

	// Same token should produce same hash (deterministic)
	assert.Equal(t, hash1, hash2)
	assert.NotEqual(t, token, hash1)
	assert.NotEmpty(t, hash1)
}

func TestHashRefreshTokenDifferentTokens(t *testing.T) {
	jwtService := NewJWTService(&config.JWTConfig{
		AccessSecret:     "test-secret",
		AccessExpiryMin:  15,
		RefreshExpiryDay: 7,
	})

	token1 := "refresh-token-1"
	token2 := "refresh-token-2"

	hash1 := jwtService.HashRefreshToken(token1)
	hash2 := jwtService.HashRefreshToken(token2)

	// Different tokens should produce different hashes
	assert.NotEqual(t, hash1, hash2)
}

func TestGenerateTokenPair(t *testing.T) {
	jwtService := NewJWTService(&config.JWTConfig{
		AccessSecret:     "test-secret",
		AccessExpiryMin:  15,
		RefreshExpiryDay: 7,
	})

	user := &model.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  model.RoleUser,
	}

	tokenPair, refreshTokenHash, err := jwtService.GenerateTokenPair(user)
	assert.NoError(t, err)
	assert.NotNil(t, tokenPair)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
	assert.NotEmpty(t, refreshTokenHash)

	// Verify access token is valid
	claims, err := jwtService.ValidateAccessToken(tokenPair.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, claims.UserID)

	// Verify refresh token hash matches
	expectedHash := jwtService.HashRefreshToken(tokenPair.RefreshToken)
	assert.Equal(t, expectedHash, refreshTokenHash)
}

func TestRefreshTokenExpiry(t *testing.T) {
	jwtService := NewJWTService(&config.JWTConfig{
		AccessSecret:     "test-secret",
		AccessExpiryMin:  15,
		RefreshExpiryDay: 7,
	})

	expiry := jwtService.RefreshTokenExpiry()
	expectedExpiry := time.Now().Add(7 * 24 * time.Hour)

	// Should be ~7 days from now
	assert.WithinDuration(t, expectedExpiry, expiry, 5*time.Second)
}

func TestRefreshTokenExpiryCustomDuration(t *testing.T) {
	jwtService := NewJWTService(&config.JWTConfig{
		AccessSecret:     "test-secret",
		AccessExpiryMin:  15,
		RefreshExpiryDay: 30, // 30 days
	})

	expiry := jwtService.RefreshTokenExpiry()
	expectedExpiry := time.Now().Add(30 * 24 * time.Hour)

	// Should be ~30 days from now
	assert.WithinDuration(t, expectedExpiry, expiry, 5*time.Second)
}
