package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AcceptanceCriteria struct {
	ID        uuid.UUID `json:"id"`
	TicketID  uuid.UUID `json:"ticket_id"`
	Content   string    `json:"content"`
	SortOrder int       `json:"sort_order"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AcceptanceCriteriaRepository struct {
	db *pgxpool.Pool
}

func NewAcceptanceCriteriaRepository(db *pgxpool.Pool) *AcceptanceCriteriaRepository {
	return &AcceptanceCriteriaRepository{db: db}
}

func (r *AcceptanceCriteriaRepository) Create(ctx context.Context, criteria *AcceptanceCriteria) error {
	query := `
		INSERT INTO acceptance_criteria (ticket_id, content, sort_order, completed)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query,
		criteria.TicketID, criteria.Content, criteria.SortOrder, criteria.Completed,
	).Scan(&criteria.ID, &criteria.CreatedAt, &criteria.UpdatedAt)
	return err
}

func (r *AcceptanceCriteriaRepository) ListByTicketID(ctx context.Context, ticketID uuid.UUID) ([]*AcceptanceCriteria, error) {
	query := `
		SELECT id, ticket_id, content, sort_order, completed, created_at, updated_at
		FROM acceptance_criteria
		WHERE ticket_id = $1
		ORDER BY sort_order ASC
	`
	rows, err := r.db.Query(ctx, query, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var criteriaList []*AcceptanceCriteria
	for rows.Next() {
		criteria := &AcceptanceCriteria{}
		if err := rows.Scan(
			&criteria.ID, &criteria.TicketID, &criteria.Content, &criteria.SortOrder, &criteria.Completed,
			&criteria.CreatedAt, &criteria.UpdatedAt,
		); err != nil {
			return nil, err
		}
		criteriaList = append(criteriaList, criteria)
	}
	return criteriaList, rows.Err()
}

func (r *AcceptanceCriteriaRepository) DeleteByTicketID(ctx context.Context, ticketID uuid.UUID) error {
	query := `DELETE FROM acceptance_criteria WHERE ticket_id = $1`
	_, err := r.db.Exec(ctx, query, ticketID)
	return err
}

func (r *AcceptanceCriteriaRepository) GetByID(ctx context.Context, id uuid.UUID) (*AcceptanceCriteria, error) {
	query := `
		SELECT id, ticket_id, content, sort_order, completed, created_at, updated_at
		FROM acceptance_criteria
		WHERE id = $1
	`
	criteria := &AcceptanceCriteria{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&criteria.ID, &criteria.TicketID, &criteria.Content, &criteria.SortOrder, &criteria.Completed,
		&criteria.CreatedAt, &criteria.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return criteria, nil
}

func (r *AcceptanceCriteriaRepository) UpdateCompleted(ctx context.Context, id uuid.UUID, completed bool) error {
	query := `
		UPDATE acceptance_criteria
		SET completed = $1, updated_at = NOW()
		WHERE id = $2
	`
	result, err := r.db.Exec(ctx, query, completed, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("acceptance criteria not found")
	}
	return nil
}

func (r *AcceptanceCriteriaRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM acceptance_criteria WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("acceptance criteria not found")
	}
	return nil
}
