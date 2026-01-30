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
	ticketRepo *model.TicketRepository
}

func NewSprintHandler(sprintRepo *model.SprintRepository, ticketRepo *model.TicketRepository) *SprintHandler {
	return &SprintHandler{
		sprintRepo: sprintRepo,
		ticketRepo: ticketRepo,
	}
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

	// Calculate end date (14 days total: start date + 13 days)
	endDate := startDate.AddDate(0, 0, 13)

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
		// Recalculate end date when start date changes (14 days total: start date + 13 days)
		sprint.EndDate = startDate.AddDate(0, 0, 13)
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

type SprintReportResponse struct {
	Sprint              *model.Sprint       `json:"sprint"`
	TotalPoints         int                 `json:"total_points"`
	CompletedPoints     int                 `json:"completed_points"`
	RemainingPoints     int                 `json:"remaining_points"`
	TotalTickets        int                 `json:"total_tickets"`
	CompletedTickets    int                 `json:"completed_tickets"`
	IdealBurndown       []BurndownPoint     `json:"ideal_burndown"`
	ActualBurndown      []BurndownPoint     `json:"actual_burndown"`
	TicketsByStatus     map[string]int      `json:"tickets_by_status"`
	PointsByStatus      map[string]int      `json:"points_by_status"`
}

type BurndownPoint struct {
	Date   string `json:"date"`
	Points int    `json:"points"`
}

func (h *SprintHandler) GetSprintReport(w http.ResponseWriter, r *http.Request) {
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

	// Get all tickets
	allTickets, err := h.ticketRepo.List(r.Context(), nil)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch tickets"}`, http.StatusInternalServerError)
		return
	}

	// Filter tickets for this sprint
	var sprintTickets []*model.Ticket
	for _, ticket := range allTickets {
		if ticket.SprintID != nil && *ticket.SprintID == id {
			sprintTickets = append(sprintTickets, ticket)
		}
	}

	// Calculate statistics
	totalPoints := 0
	completedPoints := 0
	ticketsByStatus := make(map[string]int)
	pointsByStatus := make(map[string]int)

	for _, ticket := range sprintTickets {
		points := 0
		if ticket.Size != nil {
			points = *ticket.Size
		}

		totalPoints += points
		ticketsByStatus[ticket.Status]++
		pointsByStatus[ticket.Status] += points

		if ticket.Status == "closed" {
			completedPoints += points
		}
	}

	remainingPoints := totalPoints - completedPoints

	// Generate ideal burndown line
	idealBurndown := generateIdealBurndown(sprint.StartDate, sprint.EndDate, totalPoints)

	// Generate actual burndown based on ticket completion dates
	actualBurndown := generateActualBurndown(sprint.StartDate, sprint.EndDate, totalPoints, sprintTickets)

	response := SprintReportResponse{
		Sprint:           sprint,
		TotalPoints:      totalPoints,
		CompletedPoints:  completedPoints,
		RemainingPoints:  remainingPoints,
		TotalTickets:     len(sprintTickets),
		CompletedTickets: ticketsByStatus["closed"],
		IdealBurndown:    idealBurndown,
		ActualBurndown:   actualBurndown,
		TicketsByStatus:  ticketsByStatus,
		PointsByStatus:   pointsByStatus,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func generateIdealBurndown(startDate, endDate time.Time, totalPoints int) []BurndownPoint {
	var points []BurndownPoint

	// Calculate number of days in the sprint
	duration := endDate.Sub(startDate).Hours() / 24
	days := int(duration) + 1 // Include both start and end day

	// Calculate points to burn per day
	pointsPerDay := float64(totalPoints) / float64(days-1)

	// Generate ideal burndown for each day
	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i)
		remainingPoints := totalPoints - int(float64(i)*pointsPerDay)
		if remainingPoints < 0 {
			remainingPoints = 0
		}

		points = append(points, BurndownPoint{
			Date:   date.Format("2006-01-02"),
			Points: remainingPoints,
		})
	}

	return points
}

func generateActualBurndown(startDate, endDate time.Time, totalPoints int, tickets []*model.Ticket) []BurndownPoint {
	var points []BurndownPoint

	// Calculate number of days in the sprint
	duration := endDate.Sub(startDate).Hours() / 24
	days := int(duration) + 1 // Include both start and end day

	// Generate actual burndown for each day
	for i := 0; i < days; i++ {
		date := startDate.AddDate(0, 0, i)
		// Calculate end of day (23:59:59)
		endOfDay := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, date.Location())

		// Calculate points completed by end of this day
		completedPoints := 0
		for _, ticket := range tickets {
			// If ticket is closed and was closed by end of this day, count its points as completed
			if ticket.Status == "closed" {
				// Use updated_at as proxy for when ticket was closed
				closedDate := ticket.UpdatedAt.UTC()
				if closedDate.Before(endOfDay) || closedDate.Equal(endOfDay) {
					points := 0
					if ticket.Size != nil {
						points = *ticket.Size
					}
					completedPoints += points
				}
			}
		}

		remainingPoints := totalPoints - completedPoints
		if remainingPoints < 0 {
			remainingPoints = 0
		}

		points = append(points, BurndownPoint{
			Date:   date.Format("2006-01-02"),
			Points: remainingPoints,
		})
	}

	return points
}
