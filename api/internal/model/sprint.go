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
	Status      string    `json:"status"`
	IsClosed    bool      `json:"is_closed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s *Sprint) CalculateStatus() {
	// If sprint is closed, set status to "closed" and return early
	if s.IsClosed {
		s.Status = "closed"
		return
	}

	now := time.Now().UTC()
	startDate := s.StartDate.UTC()
	endDate := s.EndDate.UTC()

	// Normalize to start of day for date-only comparisons
	nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	startDateOnly := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	endDateOnly := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, time.UTC)

	if nowDate.Before(startDateOnly) {
		s.Status = "upcoming"
	} else if nowDate.After(endDateOnly) {
		s.Status = "completed"
	} else {
		s.Status = "active"
	}
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
		RETURNING id, is_closed, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query,
		sprint.UserID, sprint.Name, sprint.StartDate, sprint.EndDate,
	).Scan(&sprint.ID, &sprint.IsClosed, &sprint.CreatedAt, &sprint.UpdatedAt)
	if err != nil {
		return err
	}

	sprint.CalculateStatus()
	return nil
}

func (r *SprintRepository) GetByID(ctx context.Context, id uuid.UUID) (*Sprint, error) {
	query := `
		SELECT id, user_id, name, start_date, end_date, is_closed, created_at, updated_at
		FROM sprints WHERE id = $1
	`
	sprint := &Sprint{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&sprint.ID, &sprint.UserID, &sprint.Name, &sprint.StartDate, &sprint.EndDate,
		&sprint.IsClosed, &sprint.CreatedAt, &sprint.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	sprint.CalculateStatus()
	return sprint, nil
}

func (r *SprintRepository) List(ctx context.Context, userID *uuid.UUID) ([]*Sprint, error) {
	var query string
	var args []interface{}

	if userID != nil {
		query = `
			SELECT id, user_id, name, start_date, end_date, is_closed, created_at, updated_at
			FROM sprints WHERE user_id = $1
			ORDER BY start_date DESC
		`
		args = append(args, *userID)
	} else {
		query = `
			SELECT id, user_id, name, start_date, end_date, is_closed, created_at, updated_at
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
			&sprint.IsClosed, &sprint.CreatedAt, &sprint.UpdatedAt,
		); err != nil {
			return nil, err
		}

		sprint.CalculateStatus()
		sprints = append(sprints, sprint)
	}
	return sprints, rows.Err()
}

func (r *SprintRepository) ListPaginated(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]*Sprint, int, error) {
	// Get total count
	var countQuery string
	var countArgs []interface{}

	if userID != nil {
		countQuery = `SELECT COUNT(*) FROM sprints WHERE user_id = $1`
		countArgs = append(countArgs, *userID)
	} else {
		countQuery = `SELECT COUNT(*) FROM sprints`
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
			SELECT id, user_id, name, start_date, end_date, is_closed, created_at, updated_at
			FROM sprints WHERE user_id = $1
			ORDER BY start_date DESC
			LIMIT $2 OFFSET $3
		`
		args = append(args, *userID, limit, offset)
	} else {
		query = `
			SELECT id, user_id, name, start_date, end_date, is_closed, created_at, updated_at
			FROM sprints
			ORDER BY start_date DESC
			LIMIT $1 OFFSET $2
		`
		args = append(args, limit, offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var sprints []*Sprint
	for rows.Next() {
		sprint := &Sprint{}
		if err := rows.Scan(
			&sprint.ID, &sprint.UserID, &sprint.Name, &sprint.StartDate, &sprint.EndDate,
			&sprint.IsClosed, &sprint.CreatedAt, &sprint.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		sprint.CalculateStatus()
		sprints = append(sprints, sprint)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return sprints, total, nil
}

func (r *SprintRepository) Update(ctx context.Context, sprint *Sprint) error {
	query := `
		UPDATE sprints
		SET name = $1, start_date = $2, end_date = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING updated_at
	`
	err := r.db.QueryRow(ctx, query,
		sprint.Name, sprint.StartDate, sprint.EndDate, sprint.ID,
	).Scan(&sprint.UpdatedAt)
	if err != nil {
		return err
	}

	sprint.CalculateStatus()
	return nil
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

func (r *SprintRepository) DeleteAll(ctx context.Context) (int64, error) {
	result, err := r.db.Exec(ctx, "DELETE FROM sprints")
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}
