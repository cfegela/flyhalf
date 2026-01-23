package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Sprint struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Name        string    `json:"name"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SprintRepository struct {
	db *pgxpool.Pool
}

func NewSprintRepository(db *pgxpool.Pool) *SprintRepository {
	return &SprintRepository{db: db}
}

func (r *SprintRepository) Create(ctx context.Context, sprint *Sprint) error {
	query := `
		INSERT INTO sprints (user_id, name, start_date, end_date)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		sprint.UserID, sprint.Name, sprint.StartDate, sprint.EndDate,
	).Scan(&sprint.ID, &sprint.CreatedAt, &sprint.UpdatedAt)
}

func (r *SprintRepository) GetByID(ctx context.Context, id uuid.UUID) (*Sprint, error) {
	query := `
		SELECT id, user_id, name, start_date, end_date, created_at, updated_at
		FROM sprints WHERE id = $1
	`
	sprint := &Sprint{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&sprint.ID, &sprint.UserID, &sprint.Name, &sprint.StartDate, &sprint.EndDate,
		&sprint.CreatedAt, &sprint.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return sprint, nil
}

func (r *SprintRepository) List(ctx context.Context, userID *uuid.UUID) ([]*Sprint, error) {
	var query string
	var args []interface{}

	if userID != nil {
		query = `
			SELECT id, user_id, name, start_date, end_date, created_at, updated_at
			FROM sprints WHERE user_id = $1
			ORDER BY start_date DESC
		`
		args = append(args, *userID)
	} else {
		query = `
			SELECT id, user_id, name, start_date, end_date, created_at, updated_at
			FROM sprints
			ORDER BY start_date DESC
		`
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sprints []*Sprint
	for rows.Next() {
		sprint := &Sprint{}
		if err := rows.Scan(
			&sprint.ID, &sprint.UserID, &sprint.Name, &sprint.StartDate, &sprint.EndDate,
			&sprint.CreatedAt, &sprint.UpdatedAt,
		); err != nil {
			return nil, err
		}

		sprints = append(sprints, sprint)
	}
	return sprints, rows.Err()
}

func (r *SprintRepository) Update(ctx context.Context, sprint *Sprint) error {
	query := `
		UPDATE sprints
		SET name = $1, start_date = $2, end_date = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at
	`
	return r.db.QueryRow(ctx, query,
		sprint.Name, sprint.StartDate, sprint.EndDate, sprint.ID,
	).Scan(&sprint.UpdatedAt)
}

func (r *SprintRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM sprints WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("sprint not found")
	}
	return nil
}
