package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cfegela/flyhalf/internal/auth"
	"github.com/cfegela/flyhalf/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type SprintHandler struct {
	sprintRepo *model.SprintRepository
}

func NewSprintHandler(sprintRepo *model.SprintRepository) *SprintHandler {
	return &SprintHandler{sprintRepo: sprintRepo}
}

type CreateSprintRequest struct {
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
}

type UpdateSprintRequest struct {
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
}

func (h *SprintHandler) ListSprints(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// All users can see all sprints
	sprints, err := h.sprintRepo.List(r.Context(), nil)

	if err != nil {
		http.Error(w, `{"error":"failed to list sprints"}`, http.StatusInternalServerError)
		return
	}

	if sprints == nil {
		sprints = []*model.Sprint{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sprints)
}

func (h *SprintHandler) GetSprint(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid sprint ID"}`, http.StatusBadRequest)
		return
	}

	sprint, err := h.sprintRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"sprint not found"}`, http.StatusNotFound)
		return
	}

	// All authenticated users can view any sprint
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sprint)
}

func (h *SprintHandler) CreateSprint(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req CreateSprintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}

	if req.StartDate == "" {
		http.Error(w, `{"error":"start_date is required"}`, http.StatusBadRequest)
		return
	}

	// Parse start date
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		http.Error(w, `{"error":"invalid start_date format, use YYYY-MM-DD"}`, http.StatusBadRequest)
		return
	}

	// Calculate end date (2 weeks after start date)
	endDate := startDate.AddDate(0, 0, 14)

	sprint := &model.Sprint{
		UserID:    userID,
		Name:      req.Name,
		StartDate: startDate,
		EndDate:   endDate,
	}

	if err := h.sprintRepo.Create(r.Context(), sprint); err != nil {
		http.Error(w, `{"error":"failed to create sprint"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sprint)
}

func (h *SprintHandler) UpdateSprint(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid sprint ID"}`, http.StatusBadRequest)
		return
	}

	sprint, err := h.sprintRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"sprint not found"}`, http.StatusNotFound)
		return
	}

	// All authenticated users can update any sprint
	var req UpdateSprintRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Name != "" {
		sprint.Name = req.Name
	}
	if req.StartDate != "" {
		startDate, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			http.Error(w, `{"error":"invalid start_date format, use YYYY-MM-DD"}`, http.StatusBadRequest)
			return
		}
		sprint.StartDate = startDate
		// Recalculate end date when start date changes
		sprint.EndDate = startDate.AddDate(0, 0, 14)
	}

	if err := h.sprintRepo.Update(r.Context(), sprint); err != nil {
		http.Error(w, `{"error":"failed to update sprint"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sprint)
}

func (h *SprintHandler) DeleteSprint(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid sprint ID"}`, http.StatusBadRequest)
		return
	}

	sprint, err := h.sprintRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"sprint not found"}`, http.StatusNotFound)
		return
	}

	userID, _ := auth.GetUserID(r.Context())
	userRole, _ := auth.GetUserRole(r.Context())

	// Allow deletion if user is admin OR if user created the sprint
	if userRole != model.RoleAdmin && sprint.UserID != userID {
		http.Error(w, `{"error":"you can only delete sprints you created"}`, http.StatusForbidden)
		return
	}

	if err := h.sprintRepo.Delete(r.Context(), id); err != nil {
		http.Error(w, `{"error":"failed to delete sprint"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
