package database

import (
	"context"
	"fmt"
	"time"

	"github.com/cfegela/flyhalf/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(cfg *config.DatabaseConfig) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	poolConfig, err := pgxpool.ParseConfig(cfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	poolConfig.MaxConns = 25
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrations := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,

		`DO $$ BEGIN
			CREATE TYPE user_role AS ENUM ('admin', 'user');
		EXCEPTION
			WHEN duplicate_object THEN null;
		END $$`,

		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			role user_role NOT NULL DEFAULT 'user',
			first_name VARCHAR(100) NOT NULL,
			last_name VARCHAR(100) NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,

		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,

		`ALTER TABLE users ADD COLUMN IF NOT EXISTS must_change_password BOOLEAN NOT NULL DEFAULT false`,

		`CREATE TABLE IF NOT EXISTS refresh_tokens (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token_hash VARCHAR(255) NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			revoked_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,

		`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash)`,

		`CREATE TABLE IF NOT EXISTS tickets (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			title VARCHAR(255) NOT NULL,
			description TEXT,
			status VARCHAR(50) NOT NULL DEFAULT 'open',
			assigned_to UUID REFERENCES users(id) ON DELETE SET NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,

		`CREATE INDEX IF NOT EXISTS idx_tickets_user_id ON tickets(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tickets_status ON tickets(status)`,

		`ALTER TABLE tickets DROP COLUMN IF EXISTS metadata`,
		`CREATE INDEX IF NOT EXISTS idx_tickets_assigned_to ON tickets(assigned_to)`,

		`ALTER TABLE tickets ADD COLUMN IF NOT EXISTS priority INTEGER NOT NULL DEFAULT 0`,
		`CREATE INDEX IF NOT EXISTS idx_tickets_priority ON tickets(priority)`,

		`CREATE TABLE IF NOT EXISTS epics (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,

		`CREATE INDEX IF NOT EXISTS idx_epics_user_id ON epics(user_id)`,

		`ALTER TABLE tickets ADD COLUMN IF NOT EXISTS epic_id UUID REFERENCES epics(id) ON DELETE SET NULL`,
		`CREATE INDEX IF NOT EXISTS idx_tickets_epic_id ON tickets(epic_id)`,

		`CREATE TABLE IF NOT EXISTS sprints (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			start_date DATE NOT NULL,
			end_date DATE NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,

		`CREATE INDEX IF NOT EXISTS idx_sprints_user_id ON sprints(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sprints_start_date ON sprints(start_date)`,

		`ALTER TABLE tickets ADD COLUMN IF NOT EXISTS sprint_id UUID REFERENCES sprints(id) ON DELETE SET NULL`,
		`CREATE INDEX IF NOT EXISTS idx_tickets_sprint_id ON tickets(sprint_id)`,

		`ALTER TABLE tickets ADD COLUMN IF NOT EXISTS size INTEGER`,
		`ALTER TABLE tickets ALTER COLUMN size DROP NOT NULL`,
		`ALTER TABLE tickets ALTER COLUMN size DROP DEFAULT`,

		`UPDATE tickets SET status = 'open' WHERE status = 'new'`,
		`ALTER TABLE tickets ALTER COLUMN status SET DEFAULT 'open'`,

		// Create default admin user if it doesn't exist
		// Default password: admin123 (bcrypt hash with cost 12)
		// IMPORTANT: Change this password immediately after first login!
		`INSERT INTO users (email, password_hash, role, first_name, last_name, is_active)
VALUES ('admin@flyhalf.local', '$2a$12$R2iQS4ZXc0z1h7Oq2wAOKeqslDynZTXBkt9chHBIVIRUuUVO.nbPi', 'admin', 'System', 'Administrator', true)
ON CONFLICT (email) DO NOTHING`,
	}

	for _, migration := range migrations {
		if _, err := pool.Exec(ctx, migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}
