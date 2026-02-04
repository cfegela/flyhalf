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
	Version     int       `json:"version"`
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
		SELECT id, user_id, name, description, version, created_at, updated_at
		FROM projects WHERE id = $1
	`
	project := &Project{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&project.ID, &project.UserID, &project.Name, &project.Description, &project.Version,
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
			SELECT id, user_id, name, description, version, created_at, updated_at
			FROM projects WHERE user_id = $1
			ORDER BY created_at DESC
		`
		args = append(args, *userID)
	} else {
		query = `
			SELECT id, user_id, name, description, version, created_at, updated_at
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
			&project.ID, &project.UserID, &project.Name, &project.Description, &project.Version,
			&project.CreatedAt, &project.UpdatedAt,
		); err != nil {
			return nil, err
		}

		projects = append(projects, project)
	}
	return projects, rows.Err()
}

func (r *ProjectRepository) ListPaginated(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]*Project, int, error) {
	// Get total count
	var countQuery string
	var countArgs []interface{}

	if userID != nil {
		countQuery = `SELECT COUNT(*) FROM projects WHERE user_id = $1`
		countArgs = append(countArgs, *userID)
	} else {
		countQuery = `SELECT COUNT(*) FROM projects`
	}

	var total int
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get paginated results
	var query string
	var args []interface{}

	if userID != nil {
		query = `
			SELECT id, user_id, name, description, version, created_at, updated_at
			FROM projects WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		args = append(args, *userID, limit, offset)
	} else {
		query = `
			SELECT id, user_id, name, description, version, created_at, updated_at
			FROM projects
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`
		args = append(args, limit, offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var projects []*Project
	for rows.Next() {
		project := &Project{}
		if err := rows.Scan(
			&project.ID, &project.UserID, &project.Name, &project.Description, &project.Version,
			&project.CreatedAt, &project.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		projects = append(projects, project)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

func (r *ProjectRepository) Update(ctx context.Context, project *Project) error {
	query := `
		UPDATE projects
		SET name = $1, description = $2, version = version + 1, updated_at = NOW()
		WHERE id = $3 AND version = $4
		RETURNING updated_at, version
	`
	err := r.db.QueryRow(ctx, query,
		project.Name, project.Description, project.ID, project.Version,
	).Scan(&project.UpdatedAt, &project.Version)

	if err != nil {
		if err.Error() == "no rows in result set" {
			return fmt.Errorf("project was modified by another user, please refresh and try again")
		}
		return err
	}

	return nil
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

func (r *ProjectRepository) DeleteAll(ctx context.Context) (int64, error) {
	result, err := r.db.Exec(ctx, "DELETE FROM projects")
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
