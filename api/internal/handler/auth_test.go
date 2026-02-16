//go:build integration

package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cfegela/flyhalf/internal/auth"
	"github.com/cfegela/flyhalf/internal/config"
	"github.com/cfegela/flyhalf/internal/database"
	"github.com/cfegela/flyhalf/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test helpers to avoid import cycle with testutil

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()

	host := getEnv("TEST_DB_HOST", "localhost")
	port := getEnv("TEST_DB_PORT", "5432")
	user := getEnv("TEST_DB_USER", "flyhalf_test")
	password := getEnv("TEST_DB_PASSWORD", "test_password")
	dbname := getEnv("TEST_DB_NAME", "flyhalf_test")
	sslmode := getEnv("TEST_DB_SSLMODE", "disable")

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)

	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		t.Fatalf("failed to ping test database: %v", err)
	}

	if err := database.RunMigrations(context.Background(), pool); err != nil {
		pool.Close()
		t.Fatalf("failed to run migrations: %v", err)
	}

	cleanup := func() {
		pool.Close()
	}

	return pool, cleanup
}

func createTestUser(pool *pgxpool.Pool, role model.UserRole) (*model.User, string, error) {
	ctx := context.Background()
	auth.SetBcryptCost(4)

	email := fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8])
	password := "TestP@ssw0rd123!"

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return nil, "", err
	}

	user := &model.User{
		Email:        email,
		PasswordHash: hashedPassword,
		Role:         role,
		IsActive:     true,
	}

	query := `
		INSERT INTO users (email, password_hash, role, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, role, is_active, must_change_password, created_at, updated_at
	`

	err = pool.QueryRow(ctx, query, user.Email, user.PasswordHash, user.Role, user.IsActive).Scan(
		&user.ID, &user.Email, &user.Role, &user.IsActive, &user.MustChangePassword,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, "", err
	}

	return user, password, nil
}

func newTestJWTService() *auth.JWTService {
	cfg := &config.JWTConfig{
		AccessSecret:     "test-access-secret-key",
		RefreshSecret:    "test-refresh-secret-key",
		AccessExpiryMin:  15,
		RefreshExpiryDay: 7,
	}
	return auth.NewJWTService(cfg)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func TestLogin_Development_CookieSecureFalse(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := model.NewUserRepository(pool)
	jwtService := newTestJWTService()
	handler := NewAuthHandler(userRepo, jwtService, false) // isProduction = false

	// Create test user
	testUser, plainPassword, err := createTestUser(pool, model.RoleUser)
	require.NoError(t, err)

	// Prepare login request
	loginReq := LoginRequest{
		Email:    testUser.Email,
		Password: plainPassword,
	}
	body, _ := json.Marshal(loginReq)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Execute login
	handler.Login(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify cookie exists
	cookies := rr.Result().Cookies()
	require.NotEmpty(t, cookies, "refresh_token cookie should be set")

	var refreshCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "refresh_token" {
			refreshCookie = cookie
			break
		}
	}

	require.NotNil(t, refreshCookie, "refresh_token cookie not found")
	assert.NotEmpty(t, refreshCookie.Value, "cookie value should not be empty")
	assert.Equal(t, "/", refreshCookie.Path)
	assert.True(t, refreshCookie.HttpOnly, "cookie should be HttpOnly")
	assert.False(t, refreshCookie.Secure, "cookie Secure should be false in development")
	assert.Equal(t, http.SameSiteStrictMode, refreshCookie.SameSite)

	// Verify response body contains access token
	var resp LoginResponse
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotNil(t, resp.User)
	assert.Equal(t, testUser.Email, resp.User.Email)
}

func TestLogin_Production_CookieSecureTrue(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := model.NewUserRepository(pool)
	jwtService := newTestJWTService()
	handler := NewAuthHandler(userRepo, jwtService, true) // isProduction = true

	// Create test user
	testUser, plainPassword, err := createTestUser(pool, model.RoleUser)
	require.NoError(t, err)

	// Prepare login request
	loginReq := LoginRequest{
		Email:    testUser.Email,
		Password: plainPassword,
	}
	body, _ := json.Marshal(loginReq)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	// Execute login
	handler.Login(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify cookie Secure flag is true
	cookies := rr.Result().Cookies()
	require.NotEmpty(t, cookies)

	var refreshCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "refresh_token" {
			refreshCookie = cookie
			break
		}
	}

	require.NotNil(t, refreshCookie)
	assert.True(t, refreshCookie.Secure, "cookie Secure should be true in production")
}

func TestLogin_InvalidCredentials(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := model.NewUserRepository(pool)
	jwtService := newTestJWTService()
	handler := NewAuthHandler(userRepo, jwtService, false)

	// Create test user
	testUser, plainPassword, err := createTestUser(pool, model.RoleUser)
	require.NoError(t, err)

	tests := []struct {
		name     string
		email    string
		password string
	}{
		{
			name:     "wrong password",
			email:    testUser.Email,
			password: "WrongPassword123!",
		},
		{
			name:     "wrong email",
			email:    "nonexistent@example.com",
			password: plainPassword,
		},
		{
			name:     "empty email",
			email:    "",
			password: plainPassword,
		},
		{
			name:     "empty password",
			email:    testUser.Email,
			password: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loginReq := LoginRequest{
				Email:    tt.email,
				Password: tt.password,
			}
			body, _ := json.Marshal(loginReq)

			req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.Login(rr, req)

			assert.True(t, rr.Code == http.StatusBadRequest || rr.Code == http.StatusUnauthorized)

			// Should not set cookie on failed login
			cookies := rr.Result().Cookies()
			for _, cookie := range cookies {
				assert.NotEqual(t, "refresh_token", cookie.Name, "should not set refresh_token on failed login")
			}
		})
	}
}

func TestRefresh_Development_CookieSecureFalse(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := model.NewUserRepository(pool)
	jwtService := newTestJWTService()
	handler := NewAuthHandler(userRepo, jwtService, false) // isProduction = false

	// Create test user
	testUser, _, err := createTestUser(pool, model.RoleUser)
	require.NoError(t, err)

	// Generate initial token pair
	tokenPair, refreshTokenHash, err := jwtService.GenerateTokenPair(testUser)
	require.NoError(t, err)

	// Store refresh token in database
	refreshToken := &model.RefreshToken{
		UserID:    testUser.ID,
		TokenHash: refreshTokenHash,
		ExpiresAt: jwtService.RefreshTokenExpiry(),
	}
	err = userRepo.CreateRefreshToken(context.Background(), refreshToken)
	require.NoError(t, err)

	// Create request with refresh token cookie
	req := httptest.NewRequest("POST", "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: tokenPair.RefreshToken,
	})
	rr := httptest.NewRecorder()

	// Execute refresh
	handler.Refresh(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify new cookie has Secure=false
	cookies := rr.Result().Cookies()
	require.NotEmpty(t, cookies)

	var refreshCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "refresh_token" {
			refreshCookie = cookie
			break
		}
	}

	require.NotNil(t, refreshCookie)
	assert.NotEmpty(t, refreshCookie.Value)
	assert.NotEqual(t, tokenPair.RefreshToken, refreshCookie.Value, "should generate new refresh token")
	assert.False(t, refreshCookie.Secure, "cookie Secure should be false in development")

	// Verify response body
	var resp LoginResponse
	err = json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.AccessToken)
}

func TestRefresh_Production_CookieSecureTrue(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := model.NewUserRepository(pool)
	jwtService := newTestJWTService()
	handler := NewAuthHandler(userRepo, jwtService, true) // isProduction = true

	// Create test user
	testUser, _, err := createTestUser(pool, model.RoleUser)
	require.NoError(t, err)

	// Generate initial token pair
	tokenPair, refreshTokenHash, err := jwtService.GenerateTokenPair(testUser)
	require.NoError(t, err)

	// Store refresh token in database
	refreshToken := &model.RefreshToken{
		UserID:    testUser.ID,
		TokenHash: refreshTokenHash,
		ExpiresAt: jwtService.RefreshTokenExpiry(),
	}
	err = userRepo.CreateRefreshToken(context.Background(), refreshToken)
	require.NoError(t, err)

	// Create request with refresh token cookie
	req := httptest.NewRequest("POST", "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: tokenPair.RefreshToken,
	})
	rr := httptest.NewRecorder()

	// Execute refresh
	handler.Refresh(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify new cookie has Secure=true
	cookies := rr.Result().Cookies()
	require.NotEmpty(t, cookies)

	var refreshCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "refresh_token" {
			refreshCookie = cookie
			break
		}
	}

	require.NotNil(t, refreshCookie)
	assert.True(t, refreshCookie.Secure, "cookie Secure should be true in production")
}

func TestRefresh_InvalidToken(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := model.NewUserRepository(pool)
	jwtService := newTestJWTService()
	handler := NewAuthHandler(userRepo, jwtService, false)

	tests := []struct {
		name        string
		cookieValue string
	}{
		{
			name:        "invalid token",
			cookieValue: "invalid-token",
		},
		{
			name:        "missing cookie",
			cookieValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/auth/refresh", nil)
			if tt.cookieValue != "" {
				req.AddCookie(&http.Cookie{
					Name:  "refresh_token",
					Value: tt.cookieValue,
				})
			}
			rr := httptest.NewRecorder()

			handler.Refresh(rr, req)

			assert.Equal(t, http.StatusUnauthorized, rr.Code)
		})
	}
}

func TestLogout_Development_CookieCleared(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := model.NewUserRepository(pool)
	jwtService := newTestJWTService()
	handler := NewAuthHandler(userRepo, jwtService, false) // isProduction = false

	// Create test user
	testUser, _, err := createTestUser(pool, model.RoleUser)
	require.NoError(t, err)

	// Create authenticated request
	req := httptest.NewRequest("POST", "/auth/logout", nil)

	// Add user context (simulating auth middleware)
	ctx := auth.SetUserID(req.Context(), testUser.ID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	// Execute logout
	handler.Logout(rr, req)

	// Verify response
	assert.Equal(t, http.StatusNoContent, rr.Code)

	// Verify cookie is cleared
	cookies := rr.Result().Cookies()
	require.NotEmpty(t, cookies)

	var refreshCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "refresh_token" {
			refreshCookie = cookie
			break
		}
	}

	require.NotNil(t, refreshCookie)
	assert.Empty(t, refreshCookie.Value, "cookie value should be empty")
	assert.True(t, refreshCookie.Expires.Before(jwtService.RefreshTokenExpiry()), "cookie should be expired")
}

func TestLogout_Production_CookieCleared(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := model.NewUserRepository(pool)
	jwtService := newTestJWTService()
	handler := NewAuthHandler(userRepo, jwtService, true) // isProduction = true

	// Create test user
	testUser, _, err := createTestUser(pool, model.RoleUser)
	require.NoError(t, err)

	// Create authenticated request
	req := httptest.NewRequest("POST", "/auth/logout", nil)

	// Add user context (simulating auth middleware)
	ctx := auth.SetUserID(req.Context(), testUser.ID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	// Execute logout
	handler.Logout(rr, req)

	// Verify response
	assert.Equal(t, http.StatusNoContent, rr.Code)

	// Verify cookie is cleared with Secure=true in production
	cookies := rr.Result().Cookies()
	require.NotEmpty(t, cookies)

	var refreshCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "refresh_token" {
			refreshCookie = cookie
			break
		}
	}

	require.NotNil(t, refreshCookie)
	assert.Empty(t, refreshCookie.Value)
	assert.True(t, refreshCookie.Secure, "cookie Secure should be true in production even when clearing")
}

func TestLogout_Unauthorized(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := model.NewUserRepository(pool)
	jwtService := newTestJWTService()
	handler := NewAuthHandler(userRepo, jwtService, false)

	// Create request without user context
	req := httptest.NewRequest("POST", "/auth/logout", nil)
	rr := httptest.NewRecorder()

	handler.Logout(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestMe_Success(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := model.NewUserRepository(pool)
	jwtService := newTestJWTService()
	handler := NewAuthHandler(userRepo, jwtService, false)

	// Create test user
	testUser, _, err := createTestUser(pool, model.RoleUser)
	require.NoError(t, err)

	// Create authenticated request
	req := httptest.NewRequest("GET", "/auth/me", nil)
	ctx := auth.SetUserID(req.Context(), testUser.ID)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Me(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var user model.User
	err = json.NewDecoder(rr.Body).Decode(&user)
	require.NoError(t, err)
	assert.Equal(t, testUser.Email, user.Email)
	assert.Empty(t, user.PasswordHash, "password hash should not be returned")
}

func TestChangePassword_Success(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := model.NewUserRepository(pool)
	jwtService := newTestJWTService()
	handler := NewAuthHandler(userRepo, jwtService, false)

	// Create test user
	testUser, plainPassword, err := createTestUser(pool, model.RoleUser)
	require.NoError(t, err)

	// Prepare change password request
	changeReq := ChangePasswordRequest{
		CurrentPassword: plainPassword,
		NewPassword:     "NewP@ssw0rd456!",
	}
	body, _ := json.Marshal(changeReq)

	req := httptest.NewRequest("POST", "/auth/change-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := auth.SetUserID(req.Context(), testUser.ID)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ChangePassword(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)

	// Verify new password works
	updatedUser, err := userRepo.GetByID(req.Context(), testUser.ID)
	require.NoError(t, err)
	assert.True(t, auth.CheckPassword(changeReq.NewPassword, updatedUser.PasswordHash))
}

func TestChangePassword_InvalidCurrent(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := model.NewUserRepository(pool)
	jwtService := newTestJWTService()
	handler := NewAuthHandler(userRepo, jwtService, false)

	// Create test user
	testUser, _, err := createTestUser(pool, model.RoleUser)
	require.NoError(t, err)

	// Prepare change password request with wrong current password
	changeReq := ChangePasswordRequest{
		CurrentPassword: "WrongPassword123!",
		NewPassword:     "NewP@ssw0rd456!",
	}
	body, _ := json.Marshal(changeReq)

	req := httptest.NewRequest("POST", "/auth/change-password", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := auth.SetUserID(req.Context(), testUser.ID)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.ChangePassword(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestChangePassword_WeakPassword(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	userRepo := model.NewUserRepository(pool)
	jwtService := newTestJWTService()
	handler := NewAuthHandler(userRepo, jwtService, false)

	// Create test user
	testUser, plainPassword, err := createTestUser(pool, model.RoleUser)
	require.NoError(t, err)

	tests := []struct {
		name        string
		newPassword string
	}{
		{
			name:        "too short",
			newPassword: "Short1!",
		},
		{
			name:        "no uppercase",
			newPassword: "password123!",
		},
		{
			name:        "no lowercase",
			newPassword: "PASSWORD123!",
		},
		{
			name:        "no number",
			newPassword: "Password!",
		},
		{
			name:        "no special char",
			newPassword: "Password123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changeReq := ChangePasswordRequest{
				CurrentPassword: plainPassword,
				NewPassword:     tt.newPassword,
			}
			body, _ := json.Marshal(changeReq)

			req := httptest.NewRequest("POST", "/auth/change-password", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			ctx := auth.SetUserID(req.Context(), testUser.ID)
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			handler.ChangePassword(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)
		})
	}
}
