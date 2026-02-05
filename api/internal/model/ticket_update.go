package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TicketUpdate struct {
	ID        uuid.UUID `json:"id"`
	TicketID  uuid.UUID `json:"ticket_id"`
	Content   string    `json:"content"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TicketUpdateRepository struct {
	db *pgxpool.Pool
}

func NewTicketUpdateRepository(db *pgxpool.Pool) *TicketUpdateRepository {
	return &TicketUpdateRepository{db: db}
}

func (r *TicketUpdateRepository) Create(ctx context.Context, update *TicketUpdate) error {
	query := `
		INSERT INTO ticket_updates (ticket_id, content, sort_order)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query,
		update.TicketID, update.Content, update.SortOrder,
	).Scan(&update.ID, &update.CreatedAt, &update.UpdatedAt)
	return err
}

func (r *TicketUpdateRepository) ListByTicketID(ctx context.Context, ticketID uuid.UUID) ([]*TicketUpdate, error) {
	query := `
		SELECT id, ticket_id, content, sort_order, created_at, updated_at
		FROM ticket_updates
		WHERE ticket_id = $1
		ORDER BY sort_order ASC
	`
	rows, err := r.db.Query(ctx, query, ticketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var updatesList []*TicketUpdate
	for rows.Next() {
		update := &TicketUpdate{}
		if err := rows.Scan(
			&update.ID, &update.TicketID, &update.Content, &update.SortOrder,
			&update.CreatedAt, &update.UpdatedAt,
		); err != nil {
			return nil, err
		}
		updatesList = append(updatesList, update)
	}
	return updatesList, rows.Err()
}

func (r *TicketUpdateRepository) Update(ctx context.Context, id uuid.UUID, content string) error {
	query := `
		UPDATE ticket_updates
		SET content = $1, updated_at = NOW()
		WHERE id = $2
	`
	result, err := r.db.Exec(ctx, query, content, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("ticket update not found")
	}
	return nil
}

func (r *TicketUpdateRepository) DeleteByTicketID(ctx context.Context, ticketID uuid.UUID) error {
	query := `DELETE FROM ticket_updates WHERE ticket_id = $1`
	_, err := r.db.Exec(ctx, query, ticketID)
	return err
}

func (r *TicketUpdateRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM ticket_updates WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("ticket update not found")
	}
	return nil
}
