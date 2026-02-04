package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cfegela/flyhalf/internal/auth"
	"github.com/cfegela/flyhalf/internal/model"
	"github.com/cfegela/flyhalf/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TicketHandler struct {
	ticketRepo   *model.TicketRepository
	criteriaRepo *model.AcceptanceCriteriaRepository
	pool         *pgxpool.Pool
}

func NewTicketHandler(ticketRepo *model.TicketRepository, criteriaRepo *model.AcceptanceCriteriaRepository, pool *pgxpool.Pool) *TicketHandler {
	return &TicketHandler{
		ticketRepo:   ticketRepo,
		criteriaRepo: criteriaRepo,
		pool:         pool,
	}
}

type AcceptanceCriteriaInput struct {
	ID        string `json:"id,omitempty"`
	Content   string `json:"content"`
	Completed bool   `json:"completed"`
}

type CreateTicketRequest struct {
	Title              string                     `json:"title"`
	Description        string                     `json:"description"`
	Status             string                     `json:"status"`
	AssignedTo         *uuid.UUID                 `json:"assigned_to,omitempty"`
	ProjectID          *uuid.UUID                 `json:"project_id,omitempty"`
	SprintID           *uuid.UUID                 `json:"sprint_id,omitempty"`
	Size               *int                       `json:"size,omitempty"`
	AcceptanceCriteria []AcceptanceCriteriaInput  `json:"acceptance_criteria"`
}

type UpdateTicketRequest struct {
	Title              string                     `json:"title"`
	Description        string                     `json:"description"`
	Status             string                     `json:"status"`
	AssignedTo         *uuid.UUID                 `json:"assigned_to,omitempty"`
	ProjectID          *uuid.UUID                 `json:"project_id,omitempty"`
	SprintID           *uuid.UUID                 `json:"sprint_id,omitempty"`
	Size               *int                       `json:"size,omitempty"`
	AcceptanceCriteria []AcceptanceCriteriaInput  `json:"acceptance_criteria"`
}

func (h *TicketHandler) ListTickets(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Check if pagination is requested
	if r.URL.Query().Get("page") != "" || r.URL.Query().Get("limit") != "" {
		params := util.GetPaginationParams(r)
		tickets, total, err := h.ticketRepo.ListPaginated(r.Context(), nil, params.Limit, params.CalculateOffset())
		if err != nil {
			http.Error(w, `{"error":"failed to list tickets"}`, http.StatusInternalServerError)
			return
		}

		if tickets == nil {
			tickets = []*model.Ticket{}
		}

		response := util.CreatePaginatedResponse(tickets, params.Page, params.Limit, total)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// All users can see all tickets (non-paginated for backward compatibility)
	tickets, err := h.ticketRepo.List(r.Context(), nil)
	if err != nil {
		http.Error(w, `{"error":"failed to list tickets"}`, http.StatusInternalServerError)
		return
	}

	if tickets == nil {
		tickets = []*model.Ticket{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tickets)
}

type TicketWithCriteria struct {
	*model.Ticket
	AcceptanceCriteria []*model.AcceptanceCriteria `json:"acceptance_criteria"`
}

func (h *TicketHandler) GetTicket(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid ticket ID"}`, http.StatusBadRequest)
		return
	}

	ticket, err := h.ticketRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"ticket not found"}`, http.StatusNotFound)
		return
	}

	// Get acceptance criteria
	criteriaList, err := h.criteriaRepo.ListByTicketID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"failed to get acceptance criteria"}`, http.StatusInternalServerError)
		return
	}

	response := TicketWithCriteria{
		Ticket:             ticket,
		AcceptanceCriteria: criteriaList,
	}

	// All authenticated users can view any ticket
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *TicketHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Sanitize input
	req.Title = util.SanitizeString(req.Title)
	req.Description = util.SanitizeString(req.Description)

	if req.Title == "" {
		http.Error(w, `{"error":"title is required"}`, http.StatusBadRequest)
		return
	}

	if req.Description == "" {
		http.Error(w, `{"error":"description is required"}`, http.StatusBadRequest)
		return
	}

	// Validate and sanitize acceptance criteria
	if len(req.AcceptanceCriteria) < 1 {
		http.Error(w, `{"error":"at least 1 acceptance criterion is required"}`, http.StatusBadRequest)
		return
	}
	if len(req.AcceptanceCriteria) > 6 {
		http.Error(w, `{"error":"maximum 6 acceptance criteria allowed"}`, http.StatusBadRequest)
		return
	}
	for i := range req.AcceptanceCriteria {
		req.AcceptanceCriteria[i].Content = util.SanitizeString(req.AcceptanceCriteria[i].Content)
		if len(req.AcceptanceCriteria[i].Content) < 1 {
			http.Error(w, `{"error":"acceptance criteria cannot be empty"}`, http.StatusBadRequest)
			return
		}
		if len(req.AcceptanceCriteria[i].Content) > 256 {
			http.Error(w, `{"error":"acceptance criteria must be 256 characters or less"}`, http.StatusBadRequest)
			return
		}
	}

	if req.Status == "" {
		req.Status = "open"
	}

	// Get min priority to place new tickets at the bottom
	minPriority, err := h.ticketRepo.GetMinPriority(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to get min priority"}`, http.StatusInternalServerError)
		return
	}

	ticket := &model.Ticket{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		AssignedTo:  req.AssignedTo,
		ProjectID:   req.ProjectID,
		SprintID:    req.SprintID,
		Size:        req.Size,
		Priority:    minPriority - 1.0,
	}

	// Set added_to_sprint_at timestamp if ticket is created with a sprint
	if req.SprintID != nil {
		now := time.Now()
		ticket.AddedToSprintAt = &now
	}

	// Use transaction to ensure atomicity
	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to start transaction"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	// Create ticket within transaction
	query := `
		INSERT INTO tickets (user_id, title, description, status, assigned_to, project_id, sprint_id, size, priority, sprint_order, added_to_sprint_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`
	err = tx.QueryRow(r.Context(), query,
		ticket.UserID, ticket.Title, ticket.Description, ticket.Status, ticket.AssignedTo,
		ticket.ProjectID, ticket.SprintID, ticket.Size, ticket.Priority, ticket.SprintOrder, ticket.AddedToSprintAt,
	).Scan(&ticket.ID, &ticket.CreatedAt, &ticket.UpdatedAt)
	if err != nil {
		http.Error(w, `{"error":"failed to create ticket"}`, http.StatusInternalServerError)
		return
	}

	// Create acceptance criteria within transaction
	for i, criterionInput := range req.AcceptanceCriteria {
		criteriaQuery := `
			INSERT INTO acceptance_criteria (ticket_id, content, sort_order, completed)
			VALUES ($1, $2, $3, $4)
		`
		_, err := tx.Exec(r.Context(), criteriaQuery, ticket.ID, criterionInput.Content, i, criterionInput.Completed)
		if err != nil {
			http.Error(w, `{"error":"failed to create acceptance criteria"}`, http.StatusInternalServerError)
			return
		}
	}

	// Commit transaction
	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, `{"error":"failed to commit transaction"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ticket)
}

func (h *TicketHandler) UpdateTicket(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid ticket ID"}`, http.StatusBadRequest)
		return
	}

	ticket, err := h.ticketRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"ticket not found"}`, http.StatusNotFound)
		return
	}

	// Capture old sprint ID before updates
	oldSprintID := ticket.SprintID

	// All authenticated users can update any ticket
	var req UpdateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Sanitize input
	req.Title = util.SanitizeString(req.Title)
	req.Description = util.SanitizeString(req.Description)

	if req.Title == "" {
		http.Error(w, `{"error":"title is required"}`, http.StatusBadRequest)
		return
	}
	ticket.Title = req.Title

	if req.Description == "" {
		http.Error(w, `{"error":"description is required"}`, http.StatusBadRequest)
		return
	}
	ticket.Description = req.Description

	// Validate and sanitize acceptance criteria
	if len(req.AcceptanceCriteria) < 1 {
		http.Error(w, `{"error":"at least 1 acceptance criterion is required"}`, http.StatusBadRequest)
		return
	}
	if len(req.AcceptanceCriteria) > 6 {
		http.Error(w, `{"error":"maximum 6 acceptance criteria allowed"}`, http.StatusBadRequest)
		return
	}
	for i := range req.AcceptanceCriteria {
		req.AcceptanceCriteria[i].Content = util.SanitizeString(req.AcceptanceCriteria[i].Content)
		if len(req.AcceptanceCriteria[i].Content) < 1 {
			http.Error(w, `{"error":"acceptance criteria cannot be empty"}`, http.StatusBadRequest)
			return
		}
		if len(req.AcceptanceCriteria[i].Content) > 256 {
			http.Error(w, `{"error":"acceptance criteria must be 256 characters or less"}`, http.StatusBadRequest)
			return
		}
	}

	if req.Status != "" {
		ticket.Status = req.Status
	}
	ticket.AssignedTo = req.AssignedTo
	ticket.ProjectID = req.ProjectID
	ticket.SprintID = req.SprintID
	ticket.Size = req.Size

	// Handle sprint assignment timestamp
	if ticket.SprintID == nil {
		// Removed from sprint
		ticket.AddedToSprintAt = nil
	} else if oldSprintID == nil || *ticket.SprintID != *oldSprintID {
		// Added to new sprint or changed sprint
		now := time.Now()
		ticket.AddedToSprintAt = &now
	}
	// If sprint unchanged, AddedToSprintAt stays as-is

	// Use transaction to ensure atomicity
	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to start transaction"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	// Update ticket within transaction with optimistic locking
	updateQuery := `
		UPDATE tickets
		SET title = $1, description = $2, status = $3, assigned_to = $4, project_id = $5, sprint_id = $6,
		    size = $7, priority = $8, sprint_order = $9, added_to_sprint_at = $10, version = version + 1, updated_at = NOW()
		WHERE id = $11 AND version = $12
		RETURNING updated_at, version
	`
	err = tx.QueryRow(r.Context(), updateQuery,
		ticket.Title, ticket.Description, ticket.Status, ticket.AssignedTo, ticket.ProjectID, ticket.SprintID,
		ticket.Size, ticket.Priority, ticket.SprintOrder, ticket.AddedToSprintAt, ticket.ID, ticket.Version,
	).Scan(&ticket.UpdatedAt, &ticket.Version)
	if err != nil {
		if err.Error() == "no rows in result set" {
			http.Error(w, `{"error":"ticket was modified by another user, please refresh and try again"}`, http.StatusConflict)
		} else {
			http.Error(w, `{"error":"failed to update ticket"}`, http.StatusInternalServerError)
		}
		return
	}

	// Delete old acceptance criteria within transaction
	deleteQuery := `DELETE FROM acceptance_criteria WHERE ticket_id = $1`
	_, err = tx.Exec(r.Context(), deleteQuery, ticket.ID)
	if err != nil {
		http.Error(w, `{"error":"failed to update acceptance criteria"}`, http.StatusInternalServerError)
		return
	}

	// Insert new acceptance criteria within transaction
	for i, criterionInput := range req.AcceptanceCriteria {
		criteriaQuery := `
			INSERT INTO acceptance_criteria (ticket_id, content, sort_order, completed)
			VALUES ($1, $2, $3, $4)
		`
		_, err := tx.Exec(r.Context(), criteriaQuery, ticket.ID, criterionInput.Content, i, criterionInput.Completed)
		if err != nil {
			http.Error(w, `{"error":"failed to update acceptance criteria"}`, http.StatusInternalServerError)
			return
		}
	}

	// Commit transaction
	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, `{"error":"failed to commit transaction"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticket)
}

func (h *TicketHandler) DeleteTicket(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid ticket ID"}`, http.StatusBadRequest)
		return
	}

	ticket, err := h.ticketRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"ticket not found"}`, http.StatusNotFound)
		return
	}

	userID, _ := auth.GetUserID(r.Context())
	userRole, _ := auth.GetUserRole(r.Context())

	// Allow deletion if user is admin OR if user created the ticket
	if userRole != model.RoleAdmin && ticket.UserID != userID {
		http.Error(w, `{"error":"you can only delete tickets you created"}`, http.StatusForbidden)
		return
	}

	if err := h.ticketRepo.Delete(r.Context(), id); err != nil {
		http.Error(w, `{"error":"failed to delete ticket"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TicketHandler) PromoteTicket(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid ticket ID"}`, http.StatusBadRequest)
		return
	}

	// Get current max priority
	maxPriority, err := h.ticketRepo.GetMaxPriority(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to get max priority"}`, http.StatusInternalServerError)
		return
	}

	// Set ticket priority to max + 1.0
	if err := h.ticketRepo.UpdatePriority(r.Context(), id, maxPriority+1.0); err != nil {
		http.Error(w, `{"error":"failed to promote ticket"}`, http.StatusInternalServerError)
		return
	}

	// Get updated ticket and return it
	ticket, err := h.ticketRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"ticket not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticket)
}

type UpdatePriorityRequest struct {
	Priority float64 `json:"priority"`
}

func (h *TicketHandler) UpdateTicketPriority(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid ticket ID"}`, http.StatusBadRequest)
		return
	}

	var req UpdatePriorityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Update ticket priority
	if err := h.ticketRepo.UpdatePriority(r.Context(), id, req.Priority); err != nil {
		http.Error(w, `{"error":"failed to update ticket priority"}`, http.StatusInternalServerError)
		return
	}

	// Get updated ticket and return it
	ticket, err := h.ticketRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"ticket not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticket)
}

type UpdateSprintOrderRequest struct {
	SprintOrder float64 `json:"sprint_order"`
}

func (h *TicketHandler) UpdateTicketSprintOrder(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid ticket ID"}`, http.StatusBadRequest)
		return
	}

	var req UpdateSprintOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Update ticket sprint order
	if err := h.ticketRepo.UpdateSprintOrder(r.Context(), id, req.SprintOrder); err != nil {
		http.Error(w, `{"error":"failed to update ticket sprint order"}`, http.StatusInternalServerError)
		return
	}

	// Get updated ticket and return it
	ticket, err := h.ticketRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"ticket not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticket)
}

type UpdateAcceptanceCriteriaCompletedRequest struct {
	Completed bool `json:"completed"`
}

func (h *TicketHandler) UpdateAcceptanceCriteriaCompleted(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	criteriaIDParam := chi.URLParam(r, "criteriaId")
	criteriaID, err := uuid.Parse(criteriaIDParam)
	if err != nil {
		http.Error(w, `{"error":"invalid acceptance criteria ID"}`, http.StatusBadRequest)
		return
	}

	var req UpdateAcceptanceCriteriaCompletedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Update completion status
	if err := h.criteriaRepo.UpdateCompleted(r.Context(), criteriaID, req.Completed); err != nil {
		http.Error(w, `{"error":"failed to update acceptance criteria"}`, http.StatusInternalServerError)
		return
	}

	// Get updated criteria and return it
	criteria, err := h.criteriaRepo.GetByID(r.Context(), criteriaID)
	if err != nil {
		http.Error(w, `{"error":"acceptance criteria not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(criteria)
}
