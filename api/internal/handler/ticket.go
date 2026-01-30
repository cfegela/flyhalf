package handler

import (
	"encoding/json"
	"net/http"

	"github.com/cfegela/flyhalf/internal/auth"
	"github.com/cfegela/flyhalf/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type TicketHandler struct {
	ticketRepo *model.TicketRepository
}

func NewTicketHandler(ticketRepo *model.TicketRepository) *TicketHandler {
	return &TicketHandler{ticketRepo: ticketRepo}
}

type CreateTicketRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	AssignedTo  *uuid.UUID `json:"assigned_to,omitempty"`
	ProjectID   *uuid.UUID `json:"project_id,omitempty"`
	SprintID    *uuid.UUID `json:"sprint_id,omitempty"`
	Size        *int       `json:"size,omitempty"`
}

type UpdateTicketRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	AssignedTo  *uuid.UUID `json:"assigned_to,omitempty"`
	ProjectID   *uuid.UUID `json:"project_id,omitempty"`
	SprintID    *uuid.UUID `json:"sprint_id,omitempty"`
	Size        *int       `json:"size,omitempty"`
}

func (h *TicketHandler) ListTickets(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// All users can see all tickets
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

	// All authenticated users can view any ticket
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ticket)
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

	if req.Title == "" {
		http.Error(w, `{"error":"title is required"}`, http.StatusBadRequest)
		return
	}

	if req.Description == "" {
		http.Error(w, `{"error":"description is required"}`, http.StatusBadRequest)
		return
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

	if err := h.ticketRepo.Create(r.Context(), ticket); err != nil {
		http.Error(w, `{"error":"failed to create ticket"}`, http.StatusInternalServerError)
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

	// All authenticated users can update any ticket
	var req UpdateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

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
	if req.Status != "" {
		ticket.Status = req.Status
	}
	ticket.AssignedTo = req.AssignedTo
	ticket.ProjectID = req.ProjectID
	ticket.SprintID = req.SprintID
	ticket.Size = req.Size

	if err := h.ticketRepo.Update(r.Context(), ticket); err != nil {
		http.Error(w, `{"error":"failed to update ticket"}`, http.StatusInternalServerError)
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
