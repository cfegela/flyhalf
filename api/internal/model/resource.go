package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Resource struct {
	ID          uuid.UUID              `json:"id"`
	UserID      uuid.UUID              `json:"user_id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type ResourceRepository struct {
	db *pgxpool.Pool
}

func NewResourceRepository(db *pgxpool.Pool) *ResourceRepository {
	return &ResourceRepository{db: db}
}

func (r *ResourceRepository) Create(ctx context.Context, resource *Resource) error {
	var metadata []byte
	var err error
	if resource.Metadata != nil {
		metadata, err = json.Marshal(resource.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	query := `
		INSERT INTO resources (user_id, title, description, status, metadata)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		resource.UserID, resource.Title, resource.Description, resource.Status, metadata,
	).Scan(&resource.ID, &resource.CreatedAt, &resource.UpdatedAt)
}

func (r *ResourceRepository) GetByID(ctx context.Context, id uuid.UUID) (*Resource, error) {
	query := `
		SELECT id, user_id, title, description, status, metadata, created_at, updated_at
		FROM resources WHERE id = $1
	`
	resource := &Resource{}
	var metadata []byte
	err := r.db.QueryRow(ctx, query, id).Scan(
		&resource.ID, &resource.UserID, &resource.Title, &resource.Description,
		&resource.Status, &metadata, &resource.CreatedAt, &resource.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(metadata) > 0 {
		if err := json.Unmarshal(metadata, &resource.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return resource, nil
}

func (r *ResourceRepository) List(ctx context.Context, userID *uuid.UUID) ([]*Resource, error) {
	var query string
	var args []interface{}

	if userID != nil {
		query = `
			SELECT id, user_id, title, description, status, metadata, created_at, updated_at
			FROM resources WHERE user_id = $1 ORDER BY created_at DESC
		`
		args = append(args, *userID)
	} else {
		query = `
			SELECT id, user_id, title, description, status, metadata, created_at, updated_at
			FROM resources ORDER BY created_at DESC
		`
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resources []*Resource
	for rows.Next() {
		resource := &Resource{}
		var metadata []byte
		if err := rows.Scan(
			&resource.ID, &resource.UserID, &resource.Title, &resource.Description,
			&resource.Status, &metadata, &resource.CreatedAt, &resource.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if len(metadata) > 0 {
			if err := json.Unmarshal(metadata, &resource.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		resources = append(resources, resource)
	}
	return resources, rows.Err()
}

func (r *ResourceRepository) Update(ctx context.Context, resource *Resource) error {
	var metadata []byte
	var err error
	if resource.Metadata != nil {
		metadata, err = json.Marshal(resource.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	query := `
		UPDATE resources
		SET title = $1, description = $2, status = $3, metadata = $4, updated_at = NOW()
		WHERE id = $5
		RETURNING updated_at
	`
	return r.db.QueryRow(ctx, query,
		resource.Title, resource.Description, resource.Status, metadata, resource.ID,
	).Scan(&resource.UpdatedAt)
}

func (r *ResourceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM resources WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("resource not found")
	}
	return nil
}
