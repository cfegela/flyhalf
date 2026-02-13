package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type League struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type LeagueRepository struct {
	db *pgxpool.Pool
}

func NewLeagueRepository(db *pgxpool.Pool) *LeagueRepository {
	return &LeagueRepository{db: db}
}

func (r *LeagueRepository) Create(ctx context.Context, league *League) error {
	query := `
		INSERT INTO leagues (name, description)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		league.Name, league.Description,
	).Scan(&league.ID, &league.CreatedAt, &league.UpdatedAt)
}

func (r *LeagueRepository) GetByID(ctx context.Context, id uuid.UUID) (*League, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM leagues WHERE id = $1
	`
	league := &League{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&league.ID, &league.Name, &league.Description,
		&league.CreatedAt, &league.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return league, nil
}

func (r *LeagueRepository) List(ctx context.Context) ([]*League, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM leagues
		ORDER BY name ASC
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var leagues []*League
	for rows.Next() {
		league := &League{}
		if err := rows.Scan(
			&league.ID, &league.Name, &league.Description,
			&league.CreatedAt, &league.UpdatedAt,
		); err != nil {
			return nil, err
		}

		leagues = append(leagues, league)
	}
	return leagues, rows.Err()
}

func (r *LeagueRepository) Update(ctx context.Context, league *League) error {
	query := `
		UPDATE leagues
		SET name = $1, description = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING updated_at
	`
	return r.db.QueryRow(ctx, query,
		league.Name, league.Description, league.ID,
	).Scan(&league.UpdatedAt)
}

func (r *LeagueRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM leagues WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("league not found")
	}
	return nil
}
