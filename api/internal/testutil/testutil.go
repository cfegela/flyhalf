//go:build integration

package testutil

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/cfegela/flyhalf/internal/auth"
	"github.com/cfegela/flyhalf/internal/config"
	"github.com/cfegela/flyhalf/internal/database"
	"github.com/cfegela/flyhalf/internal/handler"
	"github.com/cfegela/flyhalf/internal/model"
	"github.com/cfegela/flyhalf/internal/router"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TestUser wraps model.User with the plain password for testing
type TestUser struct {
	*model.User
	PlainPassword string
}

// SetupTestDB connects to test PostgreSQL database and runs migrations
func SetupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()

	// Get test database configuration from environment
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

	// Verify connection
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		t.Fatalf("failed to ping test database: %v", err)
	}

	// Run migrations to set up schema
	if err := database.RunMigrations(context.Background(), pool); err != nil {
		pool.Close()
		t.Fatalf("failed to run migrations: %v", err)
	}

	// Clean database before tests
	CleanDB(pool)

	// Return cleanup function
	cleanup := func() {
		CleanDB(pool)
		pool.Close()
	}

	return pool, cleanup
}

// CleanDB truncates all tables in FK-safe order
func CleanDB(pool *pgxpool.Pool) {
	ctx := context.Background()

	// Disable triggers temporarily for faster truncate
	_, _ = pool.Exec(ctx, "SET session_replication_role = replica")

	// Truncate in reverse dependency order
	tables := []string{
		"retro_items",
		"sprint_snapshots",
		"ticket_updates",
		"acceptance_criteria",
		"tickets",
		"sprints",
		"projects",
		"teams",
		"leagues",
		"refresh_tokens",
		"users",
	}

	for _, table := range tables {
		_, _ = pool.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
	}

	// Re-enable triggers
	_, _ = pool.Exec(ctx, "SET session_replication_role = DEFAULT")
}

// CreateTestUser creates a test user with the given role
func CreateTestUser(pool *pgxpool.Pool, role model.UserRole) (*TestUser, error) {
	ctx := context.Background()

	// Set low bcrypt cost for faster tests
	auth.SetBcryptCost(4)

	email := fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8])
	password := "TestP@ssw0rd123!"

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return &TestUser{
		User:          user,
		PlainPassword: password,
	}, nil
}

// CreateTestAdmin creates a test admin user
func CreateTestAdmin(pool *pgxpool.Pool) (*TestUser, error) {
	return CreateTestUser(pool, model.RoleAdmin)
}

// NewTestJWTService returns a JWT service configured for testing
func NewTestJWTService() *auth.JWTService {
	cfg := &config.JWTConfig{
		AccessSecret:     "test-access-secret-key",
		RefreshSecret:    "test-refresh-secret-key",
		AccessExpiryMin:  15,
		RefreshExpiryDay: 7,
	}
	return auth.NewJWTService(cfg)
}

// AuthHeader generates an Authorization header value for the given user
func AuthHeader(jwtService *auth.JWTService, user *model.User) string {
	token, _ := jwtService.GenerateAccessToken(user)
	return "Bearer " + token
}

// SetupTestRouter creates a chi router with all handlers for integration tests
func SetupTestRouter(pool *pgxpool.Pool) http.Handler {
	// Create repositories
	userRepo := model.NewUserRepository(pool)
	projectRepo := model.NewProjectRepository(pool)
	ticketRepo := model.NewTicketRepository(pool)
	criteriaRepo := model.NewAcceptanceCriteriaRepository(pool)
	updateRepo := model.NewTicketUpdateRepository(pool)
	sprintRepo := model.NewSprintRepository(pool)
	retroItemRepo := model.NewRetroItemRepository(pool)
	leagueRepo := model.NewLeagueRepository(pool)
	teamRepo := model.NewTeamRepository(pool)

	// Create JWT service
	jwtService := NewTestJWTService()

	// Create handlers
	healthHandler := handler.NewHealthHandler(pool)
	metricsHandler := handler.NewMetricsHandler(pool)
	authHandler := handler.NewAuthHandler(userRepo, jwtService, false) // isProduction = false for tests
	projectHandler := handler.NewProjectHandler(projectRepo)
	ticketHandler := handler.NewTicketHandler(ticketRepo, criteriaRepo, updateRepo, pool)
	sprintHandler := handler.NewSprintHandler(sprintRepo, ticketRepo, pool)
	retroItemHandler := handler.NewRetroItemHandler(retroItemRepo, userRepo, sprintRepo)
	adminHandler := handler.NewAdminHandler(userRepo, ticketRepo, sprintRepo, projectRepo)
	leagueHandler := handler.NewLeagueHandler(leagueRepo)
	teamHandler := handler.NewTeamHandler(teamRepo)

	// Create middleware
	authMiddleware := auth.NewAuthMiddleware(jwtService)

	// Create test config
	cfg := &config.Config{
		Server: config.ServerConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
		},
	}

	// Create router
	r := router.New(
		healthHandler,
		metricsHandler,
		authHandler,
		adminHandler,
		teamHandler,
		leagueHandler,
		ticketHandler,
		projectHandler,
		sprintHandler,
		retroItemHandler,
		authMiddleware,
		cfg,
	)

	return r.Setup()
}

// ExecuteRequest executes an HTTP request on a test router and returns the response recorder
func ExecuteRequest(router http.Handler, req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
