package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cfegela/flyhalf/internal/config"
	"github.com/cfegela/flyhalf/internal/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticateValidToken(t *testing.T) {
	jwtService := NewJWTService(&config.JWTConfig{
		AccessSecret:     "test-secret",
		AccessExpiryMin:  15,
		RefreshExpiryDay: 7,
	})
	middleware := NewAuthMiddleware(jwtService)

	user := &model.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  model.RoleUser,
	}

	token, err := jwtService.GenerateAccessToken(user)
	assert.NoError(t, err)

	// Create a test handler that checks if context has user info
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		assert.True(t, ok)
		assert.Equal(t, user.ID, userID)

		userRole, ok := GetUserRole(r.Context())
		assert.True(t, ok)
		assert.Equal(t, user.Role, userRole)

		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	middleware.Authenticate(testHandler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestAuthenticateMissingHeader(t *testing.T) {
	jwtService := NewJWTService(&config.JWTConfig{
		AccessSecret:     "test-secret",
		AccessExpiryMin:  15,
		RefreshExpiryDay: 7,
	})
	middleware := NewAuthMiddleware(jwtService)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	middleware.Authenticate(testHandler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "missing authorization header")
}

func TestAuthenticateInvalidFormat(t *testing.T) {
	jwtService := NewJWTService(&config.JWTConfig{
		AccessSecret:     "test-secret",
		AccessExpiryMin:  15,
		RefreshExpiryDay: 7,
	})
	middleware := NewAuthMiddleware(jwtService)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	tests := []struct {
		name   string
		header string
	}{
		{
			name:   "missing Bearer prefix",
			header: "token-without-bearer",
		},
		{
			name:   "wrong prefix",
			header: "Basic sometoken",
		},
		{
			name:   "only Bearer",
			header: "Bearer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tt.header)
			rr := httptest.NewRecorder()

			middleware.Authenticate(testHandler).ServeHTTP(rr, req)

			assert.Equal(t, http.StatusUnauthorized, rr.Code)
			assert.Contains(t, rr.Body.String(), "invalid authorization header format")
		})
	}
}

func TestAuthenticateInvalidToken(t *testing.T) {
	jwtService := NewJWTService(&config.JWTConfig{
		AccessSecret:     "test-secret",
		AccessExpiryMin:  15,
		RefreshExpiryDay: 7,
	})
	middleware := NewAuthMiddleware(jwtService)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr := httptest.NewRecorder()

	middleware.Authenticate(testHandler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid or expired token")
}

func TestAuthenticateExpiredToken(t *testing.T) {
	jwtService := NewJWTService(&config.JWTConfig{
		AccessSecret:     "test-secret",
		AccessExpiryMin:  -1, // Already expired
		RefreshExpiryDay: 7,
	})
	middleware := NewAuthMiddleware(jwtService)

	user := &model.User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  model.RoleUser,
	}

	token, err := jwtService.GenerateAccessToken(user)
	assert.NoError(t, err)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	middleware.Authenticate(testHandler).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "invalid or expired token")
}

func TestRequireRoleAllowed(t *testing.T) {
	jwtService := NewJWTService(&config.JWTConfig{
		AccessSecret:     "test-secret",
		AccessExpiryMin:  15,
		RefreshExpiryDay: 7,
	})
	middleware := NewAuthMiddleware(jwtService)

	user := &model.User{
		ID:    uuid.New(),
		Email: "admin@example.com",
		Role:  model.RoleAdmin,
	}

	token, err := jwtService.GenerateAccessToken(user)
	assert.NoError(t, err)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	// First authenticate, then check role
	handler := middleware.Authenticate(middleware.RequireRole(model.RoleAdmin)(testHandler))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestRequireRoleForbidden(t *testing.T) {
	jwtService := NewJWTService(&config.JWTConfig{
		AccessSecret:     "test-secret",
		AccessExpiryMin:  15,
		RefreshExpiryDay: 7,
	})
	middleware := NewAuthMiddleware(jwtService)

	user := &model.User{
		ID:    uuid.New(),
		Email: "user@example.com",
		Role:  model.RoleUser,
	}

	token, err := jwtService.GenerateAccessToken(user)
	assert.NoError(t, err)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	// Regular user trying to access admin-only resource
	handler := middleware.Authenticate(middleware.RequireRole(model.RoleAdmin)(testHandler))
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
	assert.Contains(t, rr.Body.String(), "insufficient permissions")
}

func TestRequireRoleMultipleRoles(t *testing.T) {
	jwtService := NewJWTService(&config.JWTConfig{
		AccessSecret:     "test-secret",
		AccessExpiryMin:  15,
		RefreshExpiryDay: 7,
	})
	middleware := NewAuthMiddleware(jwtService)

	userUser := &model.User{
		ID:    uuid.New(),
		Email: "user@example.com",
		Role:  model.RoleUser,
	}

	adminUser := &model.User{
		ID:    uuid.New(),
		Email: "admin@example.com",
		Role:  model.RoleAdmin,
	}

	userToken, _ := jwtService.GenerateAccessToken(userUser)
	adminToken, _ := jwtService.GenerateAccessToken(adminUser)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Test user can access
	req1 := httptest.NewRequest("GET", "/resource", nil)
	req1.Header.Set("Authorization", "Bearer "+userToken)
	rr1 := httptest.NewRecorder()

	handler := middleware.Authenticate(middleware.RequireRole(model.RoleUser, model.RoleAdmin)(testHandler))
	handler.ServeHTTP(rr1, req1)
	assert.Equal(t, http.StatusOK, rr1.Code)

	// Test admin can also access
	req2 := httptest.NewRequest("GET", "/resource", nil)
	req2.Header.Set("Authorization", "Bearer "+adminToken)
	rr2 := httptest.NewRecorder()

	handler.ServeHTTP(rr2, req2)
	assert.Equal(t, http.StatusOK, rr2.Code)
}

func TestRequireRoleNoContext(t *testing.T) {
	jwtService := NewJWTService(&config.JWTConfig{
		AccessSecret:     "test-secret",
		AccessExpiryMin:  15,
		RefreshExpiryDay: 7,
	})
	middleware := NewAuthMiddleware(jwtService)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})

	req := httptest.NewRequest("GET", "/admin", nil)
	rr := httptest.NewRecorder()

	// Call RequireRole without Authenticate first (no context set)
	handler := middleware.RequireRole(model.RoleAdmin)(testHandler)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusForbidden, rr.Code)
}

func TestGetUserID(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name      string
		ctx       context.Context
		expectOK  bool
		expectID  uuid.UUID
	}{
		{
			name:      "user ID in context",
			ctx:       context.WithValue(context.Background(), UserIDKey, userID),
			expectOK:  true,
			expectID:  userID,
		},
		{
			name:      "no user ID in context",
			ctx:       context.Background(),
			expectOK:  false,
			expectID:  uuid.Nil,
		},
		{
			name:      "wrong type in context",
			ctx:       context.WithValue(context.Background(), UserIDKey, "not-a-uuid"),
			expectOK:  false,
			expectID:  uuid.Nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, ok := GetUserID(tt.ctx)
			assert.Equal(t, tt.expectOK, ok)
			if tt.expectOK {
				assert.Equal(t, tt.expectID, id)
			}
		})
	}
}

func TestGetUserRole(t *testing.T) {
	tests := []struct {
		name       string
		ctx        context.Context
		expectOK   bool
		expectRole model.UserRole
	}{
		{
			name:       "admin role in context",
			ctx:        context.WithValue(context.Background(), UserRole, model.RoleAdmin),
			expectOK:   true,
			expectRole: model.RoleAdmin,
		},
		{
			name:       "user role in context",
			ctx:        context.WithValue(context.Background(), UserRole, model.RoleUser),
			expectOK:   true,
			expectRole: model.RoleUser,
		},
		{
			name:       "no role in context",
			ctx:        context.Background(),
			expectOK:   false,
			expectRole: "",
		},
		{
			name:       "wrong type in context",
			ctx:        context.WithValue(context.Background(), UserRole, "not-a-role"),
			expectOK:   false,
			expectRole: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role, ok := GetUserRole(tt.ctx)
			assert.Equal(t, tt.expectOK, ok)
			if tt.expectOK {
				assert.Equal(t, tt.expectRole, role)
			}
		})
	}
}
