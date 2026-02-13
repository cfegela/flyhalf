package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Team struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	LeagueID    *uuid.UUID `json:"league_id,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type TeamRepository struct {
	db *pgxpool.Pool
}

func NewTeamRepository(db *pgxpool.Pool) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) Create(ctx context.Context, team *Team) error {
	query := `
		INSERT INTO teams (name, description, league_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		team.Name, team.Description, team.LeagueID,
	).Scan(&team.ID, &team.CreatedAt, &team.UpdatedAt)
}

func (r *TeamRepository) GetByID(ctx context.Context, id uuid.UUID) (*Team, error) {
	query := `
		SELECT id, name, description, league_id, created_at, updated_at
		FROM teams WHERE id = $1
	`
	team := &Team{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&team.ID, &team.Name, &team.Description, &team.LeagueID,
		&team.CreatedAt, &team.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return team, nil
}

func (r *TeamRepository) List(ctx context.Context) ([]*Team, error) {
	query := `
		SELECT id, name, description, league_id, created_at, updated_at
		FROM teams
		ORDER BY name ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*Team
	for rows.Next() {
		team := &Team{}
		if err := rows.Scan(
			&team.ID, &team.Name, &team.Description, &team.LeagueID,
			&team.CreatedAt, &team.UpdatedAt,
		); err != nil {
			return nil, err
		}

		teams = append(teams, team)
	}
	return teams, rows.Err()
}

func (r *TeamRepository) Update(ctx context.Context, team *Team) error {
	query := `
		UPDATE teams
		SET name = $1, description = $2, league_id = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at
	`
	return r.db.QueryRow(ctx, query,
		team.Name, team.Description, team.LeagueID, team.ID,
	).Scan(&team.UpdatedAt)
}

func (r *TeamRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM teams WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("team not found")
	}
	return nil
}
