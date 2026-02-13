package handler

import (
	"encoding/json"
	"net/http"

	"github.com/cfegela/flyhalf/internal/auth"
	"github.com/cfegela/flyhalf/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type RetroItemHandler struct {
	retroItemRepo *model.RetroItemRepository
	userRepo      *model.UserRepository
	sprintRepo    *model.SprintRepository
}

func NewRetroItemHandler(retroItemRepo *model.RetroItemRepository, userRepo *model.UserRepository, sprintRepo *model.SprintRepository) *RetroItemHandler {
	return &RetroItemHandler{
		retroItemRepo: retroItemRepo,
		userRepo:      userRepo,
		sprintRepo:    sprintRepo,
	}
}

type CreateRetroItemRequest struct {
	Content  string `json:"content"`
	Category string `json:"category"`
}

type UpdateRetroItemRequest struct {
	Content  string `json:"content"`
	Category string `json:"category"`
}

type RetroItemWithAuthor struct {
	ID         uuid.UUID `json:"id"`
	SprintID   uuid.UUID `json:"sprint_id"`
	UserID     uuid.UUID `json:"user_id"`
	Content    string    `json:"content"`
	Category   string    `json:"category"`
	VoteCount  int       `json:"vote_count"`
	AuthorName string    `json:"author_name"`
	CreatedAt  string    `json:"created_at"`
	UpdatedAt  string    `json:"updated_at"`
}

func (h *RetroItemHandler) ListRetroItems(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	sprintIDParam := chi.URLParam(r, "sprintId")
	sprintID, err := uuid.Parse(sprintIDParam)
	if err != nil {
		http.Error(w, `{"error":"invalid sprint ID"}`, http.StatusBadRequest)
		return
	}

	items, err := h.retroItemRepo.ListBySprintID(r.Context(), sprintID)
	if err != nil {
		http.Error(w, `{"error":"failed to list retro items"}`, http.StatusInternalServerError)
		return
	}

	if items == nil {
		items = []*model.RetroItem{}
	}

	// Enrich items with author names
	itemsWithAuthors := make([]RetroItemWithAuthor, 0, len(items))
	for _, item := range items {
		user, err := h.userRepo.GetByID(r.Context(), item.UserID)
		authorName := "Unknown"
		if err == nil {
			authorName = user.FirstName + " " + user.LastName
		}

		itemsWithAuthors = append(itemsWithAuthors, RetroItemWithAuthor{
			ID:         item.ID,
			SprintID:   item.SprintID,
			UserID:     item.UserID,
			Content:    item.Content,
			Category:   item.Category,
			VoteCount:  item.VoteCount,
			AuthorName: authorName,
			CreatedAt:  item.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:  item.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(itemsWithAuthors)
}

func (h *RetroItemHandler) CreateRetroItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	sprintIDParam := chi.URLParam(r, "sprintId")
	sprintID, err := uuid.Parse(sprintIDParam)
	if err != nil {
		http.Error(w, `{"error":"invalid sprint ID"}`, http.StatusBadRequest)
		return
	}

	// Check if sprint is closed
	sprint, err := h.sprintRepo.GetByID(r.Context(), sprintID)
	if err != nil {
		http.Error(w, `{"error":"sprint not found"}`, http.StatusNotFound)
		return
	}
	if sprint.IsClosed {
		http.Error(w, `{"error":"cannot modify retro items for closed sprints"}`, http.StatusBadRequest)
		return
	}

	var req CreateRetroItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Content == "" {
		http.Error(w, `{"error":"content is required"}`, http.StatusBadRequest)
		return
	}

	if len(req.Content) > 500 {
		http.Error(w, `{"error":"content must be 500 characters or less"}`, http.StatusBadRequest)
		return
	}

	if req.Category != "good" && req.Category != "bad" {
		http.Error(w, `{"error":"category must be 'good' or 'bad'"}`, http.StatusBadRequest)
		return
	}

	item := &model.RetroItem{
		SprintID: sprintID,
		UserID:   userID,
		Content:  req.Content,
		Category: req.Category,
	}

	if err := h.retroItemRepo.Create(r.Context(), item); err != nil {
		http.Error(w, `{"error":"failed to create retro item"}`, http.StatusInternalServerError)
		return
	}

	// Get author name
	user, err := h.userRepo.GetByID(r.Context(), item.UserID)
	authorName := "Unknown"
	if err == nil {
		authorName = user.FirstName + " " + user.LastName
	}

	response := RetroItemWithAuthor{
		ID:         item.ID,
		SprintID:   item.SprintID,
		UserID:     item.UserID,
		Content:    item.Content,
		Category:   item.Category,
		VoteCount:  item.VoteCount,
		AuthorName: authorName,
		CreatedAt:  item.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  item.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *RetroItemHandler) UpdateRetroItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid retro item ID"}`, http.StatusBadRequest)
		return
	}

	item, err := h.retroItemRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"retro item not found"}`, http.StatusNotFound)
		return
	}

	// Check if sprint is closed
	sprint, err := h.sprintRepo.GetByID(r.Context(), item.SprintID)
	if err != nil {
		http.Error(w, `{"error":"sprint not found"}`, http.StatusNotFound)
		return
	}
	if sprint.IsClosed {
		http.Error(w, `{"error":"cannot modify retro items for closed sprints"}`, http.StatusBadRequest)
		return
	}

	// Check authorization: only creator or admin can edit
	userRole, _ := auth.GetUserRole(r.Context())
	if userRole != model.RoleAdmin && item.UserID != userID {
		http.Error(w, `{"error":"you can only edit retro items you created"}`, http.StatusForbidden)
		return
	}

	var req UpdateRetroItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Content != "" {
		if len(req.Content) > 500 {
			http.Error(w, `{"error":"content must be 500 characters or less"}`, http.StatusBadRequest)
			return
		}
		item.Content = req.Content
	}

	if req.Category != "" {
		if req.Category != "good" && req.Category != "bad" {
			http.Error(w, `{"error":"category must be 'good' or 'bad'"}`, http.StatusBadRequest)
			return
		}
		item.Category = req.Category
	}

	if err := h.retroItemRepo.Update(r.Context(), item); err != nil {
		http.Error(w, `{"error":"failed to update retro item"}`, http.StatusInternalServerError)
		return
	}

	// Get author name
	user, err := h.userRepo.GetByID(r.Context(), item.UserID)
	authorName := "Unknown"
	if err == nil {
		authorName = user.FirstName + " " + user.LastName
	}

	response := RetroItemWithAuthor{
		ID:         item.ID,
		SprintID:   item.SprintID,
		UserID:     item.UserID,
		Content:    item.Content,
		Category:   item.Category,
		VoteCount:  item.VoteCount,
		AuthorName: authorName,
		CreatedAt:  item.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  item.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *RetroItemHandler) DeleteRetroItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid retro item ID"}`, http.StatusBadRequest)
		return
	}

	item, err := h.retroItemRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"retro item not found"}`, http.StatusNotFound)
		return
	}

	// Check if sprint is closed
	sprint, err := h.sprintRepo.GetByID(r.Context(), item.SprintID)
	if err != nil {
		http.Error(w, `{"error":"sprint not found"}`, http.StatusNotFound)
		return
	}
	if sprint.IsClosed {
		http.Error(w, `{"error":"cannot modify retro items for closed sprints"}`, http.StatusBadRequest)
		return
	}

	// Check authorization: only creator or admin can delete
	userRole, _ := auth.GetUserRole(r.Context())
	if userRole != model.RoleAdmin && item.UserID != userID {
		http.Error(w, `{"error":"you can only delete retro items you created"}`, http.StatusForbidden)
		return
	}

	if err := h.retroItemRepo.Delete(r.Context(), id); err != nil {
		http.Error(w, `{"error":"failed to delete retro item"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *RetroItemHandler) VoteRetroItem(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid retro item ID"}`, http.StatusBadRequest)
		return
	}

	// Check if sprint is closed
	item, err := h.retroItemRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"retro item not found"}`, http.StatusNotFound)
		return
	}
	sprint, err := h.sprintRepo.GetByID(r.Context(), item.SprintID)
	if err != nil {
		http.Error(w, `{"error":"sprint not found"}`, http.StatusNotFound)
		return
	}
	if sprint.IsClosed {
		http.Error(w, `{"error":"cannot modify retro items for closed sprints"}`, http.StatusBadRequest)
		return
	}

	if err := h.retroItemRepo.Vote(r.Context(), id); err != nil {
		http.Error(w, `{"error":"failed to vote on retro item"}`, http.StatusInternalServerError)
		return
	}

	// Get updated item to return new vote count
	item, err = h.retroItemRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"retro item not found"}`, http.StatusNotFound)
		return
	}

	// Get author name
	user, err := h.userRepo.GetByID(r.Context(), item.UserID)
	authorName := "Unknown"
	if err == nil {
		authorName = user.FirstName + " " + user.LastName
	}

	response := RetroItemWithAuthor{
		ID:         item.ID,
		SprintID:   item.SprintID,
		UserID:     item.UserID,
		Content:    item.Content,
		Category:   item.Category,
		VoteCount:  item.VoteCount,
		AuthorName: authorName,
		CreatedAt:  item.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  item.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *RetroItemHandler) UnvoteRetroItem(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid retro item ID"}`, http.StatusBadRequest)
		return
	}

	// Check if sprint is closed
	item, err := h.retroItemRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"retro item not found"}`, http.StatusNotFound)
		return
	}
	sprint, err := h.sprintRepo.GetByID(r.Context(), item.SprintID)
	if err != nil {
		http.Error(w, `{"error":"sprint not found"}`, http.StatusNotFound)
		return
	}
	if sprint.IsClosed {
		http.Error(w, `{"error":"cannot modify retro items for closed sprints"}`, http.StatusBadRequest)
		return
	}

	if err := h.retroItemRepo.Unvote(r.Context(), id); err != nil {
		http.Error(w, `{"error":"failed to unvote retro item"}`, http.StatusInternalServerError)
		return
	}

	// Get updated item to return new vote count
	item, err = h.retroItemRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"retro item not found"}`, http.StatusNotFound)
		return
	}

	// Get author name
	user, err := h.userRepo.GetByID(r.Context(), item.UserID)
	authorName := "Unknown"
	if err == nil {
		authorName = user.FirstName + " " + user.LastName
	}

	response := RetroItemWithAuthor{
		ID:         item.ID,
		SprintID:   item.SprintID,
		UserID:     item.UserID,
		Content:    item.Content,
		Category:   item.Category,
		VoteCount:  item.VoteCount,
		AuthorName: authorName,
		CreatedAt:  item.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  item.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
