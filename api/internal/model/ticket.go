package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Ticket struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	AssignedTo  *uuid.UUID `json:"assigned_to,omitempty"`
	ProjectID   *uuid.UUID `json:"project_id,omitempty"`
	SprintID    *uuid.UUID `json:"sprint_id,omitempty"`
	Size        *int       `json:"size,omitempty"`
	Priority    float64    `json:"priority"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type TicketRepository struct {
	db *pgxpool.Pool
}

func NewTicketRepository(db *pgxpool.Pool) *TicketRepository {
	return &TicketRepository{db: db}
}

func (r *TicketRepository) Create(ctx context.Context, ticket *Ticket) error {
	query := `
		INSERT INTO tickets (user_id, title, description, status, assigned_to, project_id, sprint_id, size, priority)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		ticket.UserID, ticket.Title, ticket.Description, ticket.Status, ticket.AssignedTo, ticket.ProjectID, ticket.SprintID, ticket.Size, ticket.Priority,
	).Scan(&ticket.ID, &ticket.CreatedAt, &ticket.UpdatedAt)
}

func (r *TicketRepository) GetByID(ctx context.Context, id uuid.UUID) (*Ticket, error) {
	query := `
		SELECT id, user_id, title, description, status, assigned_to, project_id, sprint_id, size, priority, created_at, updated_at
		FROM tickets WHERE id = $1
	`
	ticket := &Ticket{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&ticket.ID, &ticket.UserID, &ticket.Title, &ticket.Description,
		&ticket.Status, &ticket.AssignedTo, &ticket.ProjectID, &ticket.SprintID, &ticket.Size, &ticket.Priority, &ticket.CreatedAt, &ticket.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return ticket, nil
}

func (r *TicketRepository) List(ctx context.Context, userID *uuid.UUID) ([]*Ticket, error) {
	var query string
	var args []interface{}

	if userID != nil {
		query = `
			SELECT id, user_id, title, description, status, assigned_to, project_id, sprint_id, size, priority, created_at, updated_at
			FROM tickets WHERE user_id = $1
			ORDER BY priority DESC, created_at ASC
		`
		args = append(args, *userID)
	} else {
		query = `
			SELECT id, user_id, title, description, status, assigned_to, project_id, sprint_id, size, priority, created_at, updated_at
			FROM tickets
			ORDER BY priority DESC, created_at ASC
		`
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []*Ticket
	for rows.Next() {
		ticket := &Ticket{}
		if err := rows.Scan(
			&ticket.ID, &ticket.UserID, &ticket.Title, &ticket.Description,
			&ticket.Status, &ticket.AssignedTo, &ticket.ProjectID, &ticket.SprintID, &ticket.Size, &ticket.Priority, &ticket.CreatedAt, &ticket.UpdatedAt,
		); err != nil {
			return nil, err
		}

		tickets = append(tickets, ticket)
	}
	return tickets, rows.Err()
}

func (r *TicketRepository) Update(ctx context.Context, ticket *Ticket) error {
	query := `
		UPDATE tickets
		SET title = $1, description = $2, status = $3, assigned_to = $4, project_id = $5, sprint_id = $6, size = $7, priority = $8, updated_at = NOW()
		WHERE id = $9
		RETURNING updated_at
	`
	return r.db.QueryRow(ctx, query,
		ticket.Title, ticket.Description, ticket.Status, ticket.AssignedTo, ticket.ProjectID, ticket.SprintID, ticket.Size, ticket.Priority, ticket.ID,
	).Scan(&ticket.UpdatedAt)
}

func (r *TicketRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tickets WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("ticket not found")
	}
	return nil
}

func (r *TicketRepository) GetMaxPriority(ctx context.Context) (float64, error) {
	query := `SELECT COALESCE(MAX(priority), 0) FROM tickets`
	var maxPriority float64
	err := r.db.QueryRow(ctx, query).Scan(&maxPriority)
	return maxPriority, err
}

func (r *TicketRepository) GetMinPriority(ctx context.Context) (float64, error) {
	query := `SELECT COALESCE(MIN(priority), 0) FROM tickets`
	var minPriority float64
	err := r.db.QueryRow(ctx, query).Scan(&minPriority)
	return minPriority, err
}

func (r *TicketRepository) UpdatePriority(ctx context.Context, id uuid.UUID, priority float64) error {
	query := `UPDATE tickets SET priority = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.db.Exec(ctx, query, priority, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("ticket not found")
	}
	return nil
}


