package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadDefaults(t *testing.T) {
	// Clear all env vars to test defaults
	t.Setenv("SERVER_PORT", "")
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_PORT", "")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("DB_NAME", "")
	t.Setenv("DB_SSLMODE", "")
	t.Setenv("JWT_ACCESS_SECRET", "")
	t.Setenv("JWT_REFRESH_SECRET", "")
	t.Setenv("JWT_ACCESS_EXPIRY_MIN", "")
	t.Setenv("JWT_REFRESH_EXPIRY_DAY", "")
	t.Setenv("BCRYPT_COST", "")
	t.Setenv("ALLOWED_ORIGIN", "")
	t.Setenv("ENVIRONMENT", "")

	cfg, err := Load()
	assert.NoError(t, err)

	// Server defaults
	assert.Equal(t, "8080", cfg.Server.Port)
	assert.Equal(t, []string{"http://localhost:3000"}, cfg.Server.AllowedOrigins)
	assert.Equal(t, "development", cfg.Server.Environment)

	// Database defaults
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 5432, cfg.Database.Port)
	assert.Equal(t, "flyhalf", cfg.Database.User)
	assert.Equal(t, "", cfg.Database.Password)
	assert.Equal(t, "flyhalf", cfg.Database.DBName)
	assert.Equal(t, "disable", cfg.Database.SSLMode)

	// JWT defaults
	assert.Equal(t, "", cfg.JWT.AccessSecret)
	assert.Equal(t, "", cfg.JWT.RefreshSecret)
	assert.Equal(t, 15, cfg.JWT.AccessExpiryMin)
	assert.Equal(t, 7, cfg.JWT.RefreshExpiryDay)

	// Security defaults
	assert.Equal(t, 12, cfg.Security.BcryptCost)
}

func TestLoadEnvOverrides(t *testing.T) {
	t.Setenv("SERVER_PORT", "9000")
	t.Setenv("DB_HOST", "db.example.com")
	t.Setenv("DB_PORT", "5433")
	t.Setenv("DB_USER", "customuser")
	t.Setenv("DB_PASSWORD", "secret123")
	t.Setenv("DB_NAME", "customdb")
	t.Setenv("DB_SSLMODE", "require")
	t.Setenv("JWT_ACCESS_SECRET", "access-secret-key")
	t.Setenv("JWT_REFRESH_SECRET", "refresh-secret-key")
	t.Setenv("JWT_ACCESS_EXPIRY_MIN", "30")
	t.Setenv("JWT_REFRESH_EXPIRY_DAY", "14")
	t.Setenv("BCRYPT_COST", "10")
	t.Setenv("ALLOWED_ORIGIN", "https://example.com")
	t.Setenv("ENVIRONMENT", "production")

	cfg, err := Load()
	assert.NoError(t, err)

	// Server overrides
	assert.Equal(t, "9000", cfg.Server.Port)
	assert.Equal(t, []string{"https://example.com"}, cfg.Server.AllowedOrigins)
	assert.Equal(t, "production", cfg.Server.Environment)

	// Database overrides
	assert.Equal(t, "db.example.com", cfg.Database.Host)
	assert.Equal(t, 5433, cfg.Database.Port)
	assert.Equal(t, "customuser", cfg.Database.User)
	assert.Equal(t, "secret123", cfg.Database.Password)
	assert.Equal(t, "customdb", cfg.Database.DBName)
	assert.Equal(t, "require", cfg.Database.SSLMode)

	// JWT overrides
	assert.Equal(t, "access-secret-key", cfg.JWT.AccessSecret)
	assert.Equal(t, "refresh-secret-key", cfg.JWT.RefreshSecret)
	assert.Equal(t, 30, cfg.JWT.AccessExpiryMin)
	assert.Equal(t, 14, cfg.JWT.RefreshExpiryDay)

	// Security overrides
	assert.Equal(t, 10, cfg.Security.BcryptCost)
}

func TestLoadInvalidDBPort(t *testing.T) {
	t.Setenv("DB_PORT", "not-a-number")

	_, err := Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid DB_PORT")
}

func TestLoadInvalidAccessExpiry(t *testing.T) {
	t.Setenv("JWT_ACCESS_EXPIRY_MIN", "not-a-number")

	_, err := Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JWT_ACCESS_EXPIRY_MIN")
}

func TestLoadInvalidRefreshExpiry(t *testing.T) {
	t.Setenv("JWT_REFRESH_EXPIRY_DAY", "not-a-number")

	_, err := Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JWT_REFRESH_EXPIRY_DAY")
}

func TestLoadInvalidBcryptCost(t *testing.T) {
	t.Setenv("BCRYPT_COST", "not-a-number")

	_, err := Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid BCRYPT_COST")
}

func TestLoadBcryptCostTooLow(t *testing.T) {
	t.Setenv("BCRYPT_COST", "3")

	_, err := Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "BCRYPT_COST must be between 4 and 31")
}

func TestLoadBcryptCostTooHigh(t *testing.T) {
	t.Setenv("BCRYPT_COST", "32")

	_, err := Load()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "BCRYPT_COST must be between 4 and 31")
}

func TestLoadBcryptCostMinBoundary(t *testing.T) {
	t.Setenv("BCRYPT_COST", "4")

	cfg, err := Load()
	assert.NoError(t, err)
	assert.Equal(t, 4, cfg.Security.BcryptCost)
}

func TestLoadBcryptCostMaxBoundary(t *testing.T) {
	t.Setenv("BCRYPT_COST", "31")

	cfg, err := Load()
	assert.NoError(t, err)
	assert.Equal(t, 31, cfg.Security.BcryptCost)
}

func TestConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		config   DatabaseConfig
		expected string
	}{
		{
			name: "basic connection string",
			config: DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "password",
				DBName:   "testdb",
				SSLMode:  "disable",
			},
			expected: "host=localhost port=5432 user=postgres password=password dbname=testdb sslmode=disable",
		},
		{
			name: "empty password",
			config: DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "",
				DBName:   "testdb",
				SSLMode:  "disable",
			},
			expected: "host=localhost port=5432 user=postgres password= dbname=testdb sslmode=disable",
		},
		{
			name: "ssl enabled",
			config: DatabaseConfig{
				Host:     "db.example.com",
				Port:     5432,
				User:     "user",
				Password: "secret",
				DBName:   "proddb",
				SSLMode:  "require",
			},
			expected: "host=db.example.com port=5432 user=user password=secret dbname=proddb sslmode=require",
		},
		{
			name: "custom port",
			config: DatabaseConfig{
				Host:     "localhost",
				Port:     5433,
				User:     "postgres",
				Password: "password",
				DBName:   "testdb",
				SSLMode:  "disable",
			},
			expected: "host=localhost port=5433 user=postgres password=password dbname=testdb sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.ConnectionString()
			assert.Equal(t, tt.expected, result)
		})
	}
}
