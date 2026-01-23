package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Epic struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type EpicRepository struct {
	db *pgxpool.Pool
}

func NewEpicRepository(db *pgxpool.Pool) *EpicRepository {
	return &EpicRepository{db: db}
}

func (r *EpicRepository) Create(ctx context.Context, epic *Epic) error {
	query := `
		INSERT INTO epics (user_id, name, description)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		epic.UserID, epic.Name, epic.Description,
	).Scan(&epic.ID, &epic.CreatedAt, &epic.UpdatedAt)
}

func (r *EpicRepository) GetByID(ctx context.Context, id uuid.UUID) (*Epic, error) {
	query := `
		SELECT id, user_id, name, description, created_at, updated_at
		FROM epics WHERE id = $1
	`
	epic := &Epic{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&epic.ID, &epic.UserID, &epic.Name, &epic.Description,
		&epic.CreatedAt, &epic.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return epic, nil
}

func (r *EpicRepository) List(ctx context.Context, userID *uuid.UUID) ([]*Epic, error) {
	var query string
	var args []interface{}

	if userID != nil {
		query = `
			SELECT id, user_id, name, description, created_at, updated_at
			FROM epics WHERE user_id = $1
			ORDER BY created_at DESC
		`
		args = append(args, *userID)
	} else {
		query = `
			SELECT id, user_id, name, description, created_at, updated_at
			FROM epics
			ORDER BY created_at DESC
		`
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var epics []*Epic
	for rows.Next() {
		epic := &Epic{}
		if err := rows.Scan(
			&epic.ID, &epic.UserID, &epic.Name, &epic.Description,
			&epic.CreatedAt, &epic.UpdatedAt,
		); err != nil {
			return nil, err
		}

		epics = append(epics, epic)
	}
	return epics, rows.Err()
}

func (r *EpicRepository) Update(ctx context.Context, epic *Epic) error {
	query := `
		UPDATE epics
		SET name = $1, description = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING updated_at
	`
	return r.db.QueryRow(ctx, query,
		epic.Name, epic.Description, epic.ID,
	).Scan(&epic.UpdatedAt)
}

func (r *EpicRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM epics WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("epic not found")
	}
	return nil
}
