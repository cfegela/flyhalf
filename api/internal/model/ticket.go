package model

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Ticket struct {
	ID          uuid.UUID              `json:"id"`
	UserID      uuid.UUID              `json:"user_id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"`
	Priority    string                 `json:"priority"`
	AssignedTo  *uuid.UUID             `json:"assigned_to,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type TicketRepository struct {
	db *pgxpool.Pool
}

func NewTicketRepository(db *pgxpool.Pool) *TicketRepository {
	return &TicketRepository{db: db}
}

func (r *TicketRepository) Create(ctx context.Context, ticket *Ticket) error {
	var metadata []byte
	var err error
	if ticket.Metadata != nil {
		metadata, err = json.Marshal(ticket.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	query := `
		INSERT INTO tickets (user_id, title, description, status, priority, assigned_to, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query,
		ticket.UserID, ticket.Title, ticket.Description, ticket.Status, ticket.Priority, ticket.AssignedTo, metadata,
	).Scan(&ticket.ID, &ticket.CreatedAt, &ticket.UpdatedAt)
}

func (r *TicketRepository) GetByID(ctx context.Context, id uuid.UUID) (*Ticket, error) {
	query := `
		SELECT id, user_id, title, description, status, priority, assigned_to, metadata, created_at, updated_at
		FROM tickets WHERE id = $1
	`
	ticket := &Ticket{}
	var metadata []byte
	err := r.db.QueryRow(ctx, query, id).Scan(
		&ticket.ID, &ticket.UserID, &ticket.Title, &ticket.Description,
		&ticket.Status, &ticket.Priority, &ticket.AssignedTo, &metadata, &ticket.CreatedAt, &ticket.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if len(metadata) > 0 {
		if err := json.Unmarshal(metadata, &ticket.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return ticket, nil
}

func (r *TicketRepository) List(ctx context.Context, userID *uuid.UUID) ([]*Ticket, error) {
	var query string
	var args []interface{}

	if userID != nil {
		query = `
			SELECT id, user_id, title, description, status, priority, assigned_to, metadata, created_at, updated_at
			FROM tickets WHERE user_id = $1 ORDER BY created_at DESC
		`
		args = append(args, *userID)
	} else {
		query = `
			SELECT id, user_id, title, description, status, priority, assigned_to, metadata, created_at, updated_at
			FROM tickets ORDER BY created_at DESC
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
		var metadata []byte
		if err := rows.Scan(
			&ticket.ID, &ticket.UserID, &ticket.Title, &ticket.Description,
			&ticket.Status, &ticket.Priority, &ticket.AssignedTo, &metadata, &ticket.CreatedAt, &ticket.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if len(metadata) > 0 {
			if err := json.Unmarshal(metadata, &ticket.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		tickets = append(tickets, ticket)
	}
	return tickets, rows.Err()
}

func (r *TicketRepository) Update(ctx context.Context, ticket *Ticket) error {
	var metadata []byte
	var err error
	if ticket.Metadata != nil {
		metadata, err = json.Marshal(ticket.Metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
	}

	query := `
		UPDATE tickets
		SET title = $1, description = $2, status = $3, priority = $4, assigned_to = $5, metadata = $6, updated_at = NOW()
		WHERE id = $7
		RETURNING updated_at
	`
	return r.db.QueryRow(ctx, query,
		ticket.Title, ticket.Description, ticket.Status, ticket.Priority, ticket.AssignedTo, metadata, ticket.ID,
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
