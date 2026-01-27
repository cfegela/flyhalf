package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Project struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProjectRepository struct {
	db *pgxpool.Pool
}

func NewProjectRepository(db *pgxpool.Pool) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(ctx context.Context, project *Project) error {
	query := `
		INSERT INTO projects (user_id, name, description)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		project.UserID, project.Name, project.Description,
	).Scan(&project.ID, &project.CreatedAt, &project.UpdatedAt)
}

func (r *ProjectRepository) GetByID(ctx context.Context, id uuid.UUID) (*Project, error) {
	query := `
		SELECT id, user_id, name, description, created_at, updated_at
		FROM projects WHERE id = $1
	`
	project := &Project{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&project.ID, &project.UserID, &project.Name, &project.Description,
		&project.CreatedAt, &project.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (r *ProjectRepository) List(ctx context.Context, userID *uuid.UUID) ([]*Project, error) {
	var query string
	var args []interface{}

	if userID != nil {
		query = `
			SELECT id, user_id, name, description, created_at, updated_at
			FROM projects WHERE user_id = $1
			ORDER BY created_at DESC
		`
		args = append(args, *userID)
	} else {
		query = `
			SELECT id, user_id, name, description, created_at, updated_at
			FROM projects
			ORDER BY created_at DESC
		`
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		project := &Project{}
		if err := rows.Scan(
			&project.ID, &project.UserID, &project.Name, &project.Description,
			&project.CreatedAt, &project.UpdatedAt,
		); err != nil {
			return nil, err
		}

		projects = append(projects, project)
	}
	return projects, rows.Err()
}

func (r *ProjectRepository) Update(ctx context.Context, project *Project) error {
	query := `
		UPDATE projects
		SET name = $1, description = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING updated_at
	`
	return r.db.QueryRow(ctx, query,
		project.Name, project.Description, project.ID,
	).Scan(&project.UpdatedAt)
}

func (r *ProjectRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM projects WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("project not found")
	}
	return nil
}
