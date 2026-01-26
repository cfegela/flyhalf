package model

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RetroItem struct {
	ID        uuid.UUID `json:"id"`
	SprintID  uuid.UUID `json:"sprint_id"`
	UserID    uuid.UUID `json:"user_id"`
	Content   string    `json:"content"`
	Category  string    `json:"category"` // "good" or "bad"
	VoteCount int       `json:"vote_count"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RetroItemRepository struct {
	db *pgxpool.Pool
}

func NewRetroItemRepository(db *pgxpool.Pool) *RetroItemRepository {
	return &RetroItemRepository{db: db}
}

func (r *RetroItemRepository) Create(ctx context.Context, item *RetroItem) error {
	query := `
		INSERT INTO retro_items (sprint_id, user_id, content, category)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query,
		item.SprintID, item.UserID, item.Content, item.Category,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
	return err
}

func (r *RetroItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*RetroItem, error) {
	query := `
		SELECT id, sprint_id, user_id, content, category, vote_count, created_at, updated_at
		FROM retro_items WHERE id = $1
	`
	item := &RetroItem{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&item.ID, &item.SprintID, &item.UserID, &item.Content, &item.Category, &item.VoteCount,
		&item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (r *RetroItemRepository) ListBySprintID(ctx context.Context, sprintID uuid.UUID) ([]*RetroItem, error) {
	query := `
		SELECT id, sprint_id, user_id, content, category, vote_count, created_at, updated_at
		FROM retro_items
		WHERE sprint_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(ctx, query, sprintID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*RetroItem
	for rows.Next() {
		item := &RetroItem{}
		if err := rows.Scan(
			&item.ID, &item.SprintID, &item.UserID, &item.Content, &item.Category, &item.VoteCount,
			&item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *RetroItemRepository) Update(ctx context.Context, item *RetroItem) error {
	query := `
		UPDATE retro_items
		SET content = $1, category = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING updated_at
	`
	err := r.db.QueryRow(ctx, query,
		item.Content, item.Category, item.ID,
	).Scan(&item.UpdatedAt)
	return err
}

func (r *RetroItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM retro_items WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("retro item not found")
	}
	return nil
}

func (r *RetroItemRepository) Vote(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE retro_items
		SET vote_count = vote_count + 1
		WHERE id = $1
	`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("retro item not found")
	}
	return nil
}

func (r *RetroItemRepository) Unvote(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE retro_items
		SET vote_count = GREATEST(vote_count - 1, 0)
		WHERE id = $1
	`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("retro item not found")
	}
	return nil
}
