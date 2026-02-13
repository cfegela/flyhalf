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

		// Convert priority from INTEGER to DOUBLE PRECISION with fractional indexing
		`ALTER TABLE tickets ADD COLUMN IF NOT EXISTS priority_new DOUBLE PRECISION`,

		`DO $$
		BEGIN
			IF EXISTS (SELECT 1 FROM information_schema.columns
					   WHERE table_name = 'tickets'
					   AND column_name = 'priority_new'
					   AND data_type = 'double precision') THEN
				-- Reindex existing tickets with unique values (assigns N, N-1, N-2... based on current order)
				WITH ranked AS (
					SELECT id, ROW_NUMBER() OVER (ORDER BY priority DESC, created_at ASC) as rn,
						   COUNT(*) OVER () as total
					FROM tickets
				)
				UPDATE tickets SET priority_new = (SELECT total - rn + 1 FROM ranked WHERE ranked.id = tickets.id);

				-- Set default for any NULLs (shouldn't happen, but safety)
				UPDATE tickets SET priority_new = 0 WHERE priority_new IS NULL;

				-- Drop old column and rename
				ALTER TABLE tickets DROP COLUMN IF EXISTS priority;
				ALTER TABLE tickets RENAME COLUMN priority_new TO priority;
				ALTER TABLE tickets ALTER COLUMN priority SET NOT NULL;
				ALTER TABLE tickets ALTER COLUMN priority SET DEFAULT 0;
			END IF;
		END $$`,

		`CREATE TABLE IF NOT EXISTS projects (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,

		`CREATE INDEX IF NOT EXISTS idx_projects_user_id ON projects(user_id)`,

		`ALTER TABLE tickets ADD COLUMN IF NOT EXISTS project_id UUID REFERENCES projects(id) ON DELETE SET NULL`,
		`CREATE INDEX IF NOT EXISTS idx_tickets_project_id ON tickets(project_id)`,

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

		`CREATE TABLE IF NOT EXISTS retro_items (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			sprint_id UUID NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			content TEXT NOT NULL,
			category VARCHAR(10) NOT NULL CHECK (category IN ('good', 'bad')),
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,

		`CREATE INDEX IF NOT EXISTS idx_retro_items_sprint_id ON retro_items(sprint_id)`,
		`CREATE INDEX IF NOT EXISTS idx_retro_items_user_id ON retro_items(user_id)`,

		`ALTER TABLE retro_items ADD COLUMN IF NOT EXISTS vote_count INTEGER NOT NULL DEFAULT 0`,

		`CREATE TABLE IF NOT EXISTS teams (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(255) NOT NULL,
			description TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,

		`CREATE INDEX IF NOT EXISTS idx_teams_name ON teams(name)`,

		`ALTER TABLE users ADD COLUMN IF NOT EXISTS team_id UUID REFERENCES teams(id) ON DELETE SET NULL`,
		`CREATE INDEX IF NOT EXISTS idx_users_team_id ON users(team_id)`,

		// Add sprint_order column for independent sprint board ordering
		`ALTER TABLE tickets ADD COLUMN IF NOT EXISTS sprint_order DOUBLE PRECISION NOT NULL DEFAULT 0`,
		`CREATE INDEX IF NOT EXISTS idx_tickets_sprint_order ON tickets(sprint_order)`,

		// Initialize sprint_order values based on existing priority
		`UPDATE tickets SET sprint_order = priority WHERE sprint_order = 0`,

		// Add added_to_sprint_at timestamp for tracking when tickets were committed to sprints
		`ALTER TABLE tickets ADD COLUMN IF NOT EXISTS added_to_sprint_at TIMESTAMP`,

		// Backfill existing tickets: assume they were added at sprint start date
		`UPDATE tickets t SET added_to_sprint_at = s.start_date
		FROM sprints s WHERE t.sprint_id = s.id AND t.added_to_sprint_at IS NULL`,

		// Create default admin user if it doesn't exist
		// Default password: admin123 (bcrypt hash with cost 12)
		// IMPORTANT: Admin must change password on first login!
		`INSERT INTO users (email, password_hash, role, first_name, last_name, is_active, must_change_password)
VALUES ('admin@flyhalf.app', '$2a$12$R2iQS4ZXc0z1h7Oq2wAOKeqslDynZTXBkt9chHBIVIRUuUVO.nbPi', 'admin', 'System', 'Administrator', true, true)
ON CONFLICT (email) DO NOTHING`,

		// Acceptance criteria table for tickets
		`CREATE TABLE IF NOT EXISTS acceptance_criteria (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			ticket_id UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
			content VARCHAR(256) NOT NULL,
			sort_order INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,

		`CREATE INDEX IF NOT EXISTS idx_acceptance_criteria_ticket_id ON acceptance_criteria(ticket_id)`,

		// Add completed field to acceptance criteria
		`ALTER TABLE acceptance_criteria ADD COLUMN IF NOT EXISTS completed BOOLEAN NOT NULL DEFAULT false`,

		// Add version column for optimistic locking
		`ALTER TABLE tickets ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE projects ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE sprints ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1`,

		// Ticket updates table for tracking updates with timestamps
		`CREATE TABLE IF NOT EXISTS ticket_updates (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			ticket_id UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
			content VARCHAR(500) NOT NULL,
			sort_order INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,

		`CREATE INDEX IF NOT EXISTS idx_ticket_updates_ticket_id ON ticket_updates(ticket_id)`,

		// Leagues table
		`CREATE TABLE IF NOT EXISTS leagues (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name VARCHAR(255) NOT NULL,
			description TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,

		`CREATE INDEX IF NOT EXISTS idx_leagues_name ON leagues(name)`,

		// Add league_id to teams
		`ALTER TABLE teams ADD COLUMN IF NOT EXISTS league_id UUID REFERENCES leagues(id) ON DELETE SET NULL`,
		`CREATE INDEX IF NOT EXISTS idx_teams_league_id ON teams(league_id)`,

		// Add is_closed column to sprints
		`ALTER TABLE sprints ADD COLUMN IF NOT EXISTS is_closed BOOLEAN NOT NULL DEFAULT false`,

		// Create sprint_snapshots table for closed sprint reports
		`CREATE TABLE IF NOT EXISTS sprint_snapshots (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			sprint_id UUID UNIQUE NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
			total_points INTEGER NOT NULL DEFAULT 0,
			committed_points INTEGER NOT NULL DEFAULT 0,
			adopted_points INTEGER NOT NULL DEFAULT 0,
			completed_points INTEGER NOT NULL DEFAULT 0,
			remaining_points INTEGER NOT NULL DEFAULT 0,
			total_tickets INTEGER NOT NULL DEFAULT 0,
			committed_tickets INTEGER NOT NULL DEFAULT 0,
			adopted_tickets INTEGER NOT NULL DEFAULT 0,
			completed_tickets INTEGER NOT NULL DEFAULT 0,
			ideal_burndown JSONB NOT NULL DEFAULT '[]',
			actual_burndown JSONB NOT NULL DEFAULT '[]',
			tickets_by_status JSONB NOT NULL DEFAULT '{}',
			points_by_status JSONB NOT NULL DEFAULT '{}',
			closed_at TIMESTAMP NOT NULL DEFAULT NOW(),
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,

		`CREATE INDEX IF NOT EXISTS idx_sprint_snapshots_sprint_id ON sprint_snapshots(sprint_id)`,
	}

	for _, migration := range migrations {
		if _, err := pool.Exec(ctx, migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}
