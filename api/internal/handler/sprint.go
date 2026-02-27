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

type SprintHandler struct {
	sprintRepo *model.SprintRepository
	ticketRepo *model.TicketRepository
	pool       *pgxpool.Pool
}

func NewSprintHandler(sprintRepo *model.SprintRepository, ticketRepo *model.TicketRepository, pool *pgxpool.Pool) *SprintHandler {
	return &SprintHandler{
		sprintRepo: sprintRepo,
		ticketRepo: ticketRepo,
		pool:       pool,
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

	// Check if pagination is requested
	if r.URL.Query().Get("page") != "" || r.URL.Query().Get("limit") != "" {
		params := util.GetPaginationParams(r)
		sprints, total, err := h.sprintRepo.ListPaginated(r.Context(), nil, params.Limit, params.CalculateOffset())
		if err != nil {
			http.Error(w, `{"error":"failed to list sprints"}`, http.StatusInternalServerError)
			return
		}

		if sprints == nil {
			sprints = []*model.Sprint{}
		}

		response := util.CreatePaginatedResponse(sprints, params.Page, params.Limit, total)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// All users can see all sprints (non-paginated for backward compatibility)
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

func (h *SprintHandler) GetSprintTickets(w http.ResponseWriter, r *http.Request) {
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

	// If sprint is closed, return tickets from snapshot
	if sprint.IsClosed {
		var ticketsJSON json.RawMessage
		query := `SELECT tickets FROM sprint_snapshots WHERE sprint_id = $1`
		err = h.pool.QueryRow(r.Context(), query, id).Scan(&ticketsJSON)
		if err != nil {
			http.Error(w, `{"error":"failed to fetch sprint snapshot"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(ticketsJSON)
		return
	}

	// For open sprints, return live tickets
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

	if sprintTickets == nil {
		sprintTickets = []*model.Ticket{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sprintTickets)
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

func (h *SprintHandler) CloseSprint(w http.ResponseWriter, r *http.Request) {
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

	// Validate sprint can be closed
	if sprint.IsClosed {
		http.Error(w, `{"error":"sprint is already closed"}`, http.StatusBadRequest)
		return
	}

	if sprint.Status == "upcoming" {
		http.Error(w, `{"error":"cannot close an upcoming sprint"}`, http.StatusBadRequest)
		return
	}

	// Begin transaction
	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to start transaction"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	// Query sprint tickets within transaction
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

	// Compute report metrics (reuse existing logic)
	totalPoints := 0
	committedPoints := 0
	committedTickets := 0
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

		// Calculate committed points and tickets: added at or before sprint start
		if ticket.AddedToSprintAt != nil {
			addedDate := ticket.AddedToSprintAt.Truncate(24 * time.Hour)
			startDate := sprint.StartDate.Truncate(24 * time.Hour)
			if !addedDate.After(startDate) {
				committedPoints += points
				committedTickets++
			}
		}

		if ticket.Status == "closed" {
			completedPoints += points
		}
	}

	remainingPoints := totalPoints - completedPoints
	adoptedPoints := totalPoints - committedPoints
	adoptedTickets := len(sprintTickets) - committedTickets

	// Generate ideal and actual burndown
	idealBurndown := generateIdealBurndown(sprint.StartDate, sprint.EndDate, totalPoints)
	actualBurndown := generateActualBurndown(sprint.StartDate, sprint.EndDate, totalPoints, sprintTickets)

	// Create ticket snapshots with relevant fields
	type TicketSnapshot struct {
		ID     uuid.UUID `json:"id"`
		Title  string    `json:"title"`
		Status string    `json:"status"`
		Size   *int      `json:"size"`
	}

	ticketSnapshots := make([]TicketSnapshot, 0, len(sprintTickets))
	for _, ticket := range sprintTickets {
		ticketSnapshots = append(ticketSnapshots, TicketSnapshot{
			ID:     ticket.ID,
			Title:  ticket.Title,
			Status: ticket.Status,
			Size:   ticket.Size,
		})
	}

	// Marshal to JSON
	idealBurndownJSON, _ := json.Marshal(idealBurndown)
	actualBurndownJSON, _ := json.Marshal(actualBurndown)
	ticketsByStatusJSON, _ := json.Marshal(ticketsByStatus)
	pointsByStatusJSON, _ := json.Marshal(pointsByStatus)
	ticketsJSON, _ := json.Marshal(ticketSnapshots)

	// Insert snapshot into sprint_snapshots
	_, err = tx.Exec(r.Context(), `
		INSERT INTO sprint_snapshots (
			sprint_id, total_points, committed_points, adopted_points, completed_points,
			remaining_points, total_tickets, committed_tickets, adopted_tickets,
			completed_tickets, ideal_burndown, actual_burndown, tickets_by_status, points_by_status, tickets
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`, id, totalPoints, committedPoints, adoptedPoints, completedPoints, remainingPoints,
		len(sprintTickets), committedTickets, adoptedTickets, ticketsByStatus["closed"],
		idealBurndownJSON, actualBurndownJSON, ticketsByStatusJSON, pointsByStatusJSON, ticketsJSON)

	if err != nil {
		http.Error(w, `{"error":"failed to create sprint snapshot"}`, http.StatusInternalServerError)
		return
	}

	// Remove non-closed tickets from sprint and reset their status to open
	_, err = tx.Exec(r.Context(), `
		UPDATE tickets
		SET sprint_id = NULL, added_to_sprint_at = NULL, sprint_order = 0, status = 'open'
		WHERE sprint_id = $1 AND status != 'closed'
	`, id)

	if err != nil {
		http.Error(w, `{"error":"failed to remove open tickets from sprint"}`, http.StatusInternalServerError)
		return
	}

	// Mark sprint as closed
	_, err = tx.Exec(r.Context(), `
		UPDATE sprints
		SET is_closed = true, updated_at = NOW()
		WHERE id = $1
	`, id)

	if err != nil {
		http.Error(w, `{"error":"failed to close sprint"}`, http.StatusInternalServerError)
		return
	}

	// Commit transaction
	if err := tx.Commit(r.Context()); err != nil {
		http.Error(w, `{"error":"failed to commit transaction"}`, http.StatusInternalServerError)
		return
	}

	// Return updated sprint
	sprint, err = h.sprintRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch updated sprint"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sprint)
}

type SprintReportResponse struct {
	Sprint              *model.Sprint       `json:"sprint"`
	TotalPoints         int                 `json:"total_points"`
	CommittedPoints     int                 `json:"committed_points"`
	AdoptedPoints       int                 `json:"adopted_points"`
	CompletedPoints     int                 `json:"completed_points"`
	RemainingPoints     int                 `json:"remaining_points"`
	TotalTickets        int                 `json:"total_tickets"`
	CommittedTickets    int                 `json:"committed_tickets"`
	AdoptedTickets      int                 `json:"adopted_tickets"`
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

	// If sprint is closed, return snapshot data
	if sprint.IsClosed {
		var snapshot struct {
			TotalPoints      int             `json:"total_points"`
			CommittedPoints  int             `json:"committed_points"`
			AdoptedPoints    int             `json:"adopted_points"`
			CompletedPoints  int             `json:"completed_points"`
			RemainingPoints  int             `json:"remaining_points"`
			TotalTickets     int             `json:"total_tickets"`
			CommittedTickets int             `json:"committed_tickets"`
			AdoptedTickets   int             `json:"adopted_tickets"`
			CompletedTickets int             `json:"completed_tickets"`
			IdealBurndown    json.RawMessage `json:"ideal_burndown"`
			ActualBurndown   json.RawMessage `json:"actual_burndown"`
			TicketsByStatus  json.RawMessage `json:"tickets_by_status"`
			PointsByStatus   json.RawMessage `json:"points_by_status"`
		}

		query := `
			SELECT total_points, committed_points, adopted_points, completed_points, remaining_points,
			       total_tickets, committed_tickets, adopted_tickets, completed_tickets,
			       ideal_burndown, actual_burndown, tickets_by_status, points_by_status
			FROM sprint_snapshots WHERE sprint_id = $1
		`
		err = h.pool.QueryRow(r.Context(), query, id).Scan(
			&snapshot.TotalPoints, &snapshot.CommittedPoints, &snapshot.AdoptedPoints,
			&snapshot.CompletedPoints, &snapshot.RemainingPoints, &snapshot.TotalTickets,
			&snapshot.CommittedTickets, &snapshot.AdoptedTickets, &snapshot.CompletedTickets,
			&snapshot.IdealBurndown, &snapshot.ActualBurndown, &snapshot.TicketsByStatus,
			&snapshot.PointsByStatus,
		)
		if err != nil {
			http.Error(w, `{"error":"failed to fetch sprint snapshot"}`, http.StatusInternalServerError)
			return
		}

		// Unmarshal JSON fields
		var idealBurndown []BurndownPoint
		var actualBurndown []BurndownPoint
		var ticketsByStatus map[string]int
		var pointsByStatus map[string]int

		json.Unmarshal(snapshot.IdealBurndown, &idealBurndown)
		json.Unmarshal(snapshot.ActualBurndown, &actualBurndown)
		json.Unmarshal(snapshot.TicketsByStatus, &ticketsByStatus)
		json.Unmarshal(snapshot.PointsByStatus, &pointsByStatus)

		response := SprintReportResponse{
			Sprint:           sprint,
			TotalPoints:      snapshot.TotalPoints,
			CommittedPoints:  snapshot.CommittedPoints,
			AdoptedPoints:    snapshot.AdoptedPoints,
			CompletedPoints:  snapshot.CompletedPoints,
			RemainingPoints:  snapshot.RemainingPoints,
			TotalTickets:     snapshot.TotalTickets,
			CommittedTickets: snapshot.CommittedTickets,
			AdoptedTickets:   snapshot.AdoptedTickets,
			CompletedTickets: snapshot.CompletedTickets,
			IdealBurndown:    idealBurndown,
			ActualBurndown:   actualBurndown,
			TicketsByStatus:  ticketsByStatus,
			PointsByStatus:   pointsByStatus,
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		json.NewEncoder(w).Encode(response)
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
	committedPoints := 0
	committedTickets := 0
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

		// Calculate committed points and tickets: added at or before sprint start
		if ticket.AddedToSprintAt != nil {
			addedDate := ticket.AddedToSprintAt.Truncate(24 * time.Hour)
			startDate := sprint.StartDate.Truncate(24 * time.Hour)
			if !addedDate.After(startDate) {
				committedPoints += points
				committedTickets++
			}
		}

		if ticket.Status == "closed" {
			completedPoints += points
		}
	}

	remainingPoints := totalPoints - completedPoints
	adoptedPoints := totalPoints - committedPoints
	adoptedTickets := len(sprintTickets) - committedTickets

	// Generate ideal burndown line
	idealBurndown := generateIdealBurndown(sprint.StartDate, sprint.EndDate, totalPoints)

	// Generate actual burndown based on ticket completion dates
	actualBurndown := generateActualBurndown(sprint.StartDate, sprint.EndDate, totalPoints, sprintTickets)

	response := SprintReportResponse{
		Sprint:           sprint,
		TotalPoints:      totalPoints,
		CommittedPoints:  committedPoints,
		AdoptedPoints:    adoptedPoints,
		CompletedPoints:  completedPoints,
		RemainingPoints:  remainingPoints,
		TotalTickets:     len(sprintTickets),
		CommittedTickets: committedTickets,
		AdoptedTickets:   adoptedTickets,
		CompletedTickets: ticketsByStatus["closed"],
		IdealBurndown:    idealBurndown,
		ActualBurndown:   actualBurndown,
		TicketsByStatus:  ticketsByStatus,
		PointsByStatus:   pointsByStatus,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	json.NewEncoder(w).Encode(response)
}

// isWeekend returns true if the date falls on Saturday or Sunday
func isWeekend(date time.Time) bool {
	weekday := date.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// countWorkingDays returns the number of working days (excluding weekends) between start and end dates (inclusive)
func countWorkingDays(startDate, endDate time.Time) int {
	workingDays := 0
	for date := startDate; !date.After(endDate); date = date.AddDate(0, 0, 1) {
		if !isWeekend(date) {
			workingDays++
		}
	}
	return workingDays
}

func generateIdealBurndown(startDate, endDate time.Time, totalPoints int) []BurndownPoint {
	var points []BurndownPoint

	// Start one day before the sprint to show initial capacity
	dayBeforeSprint := startDate.AddDate(0, 0, -1)
	points = append(points, BurndownPoint{
		Date:   dayBeforeSprint.Format("2006-01-02"),
		Points: totalPoints,
	})

	// Count working days in the sprint (excluding weekends)
	workingDays := countWorkingDays(startDate, endDate)

	// Calculate points to burn per working day
	pointsPerDay := float64(totalPoints) / float64(workingDays)

	// Generate ideal burndown for each working day
	workingDayCount := 0
	currentDate := startDate
	for !currentDate.After(endDate) {
		if !isWeekend(currentDate) {
			remainingPoints := totalPoints - int(float64(workingDayCount)*pointsPerDay)
			if remainingPoints < 0 {
				remainingPoints = 0
			}

			points = append(points, BurndownPoint{
				Date:   currentDate.Format("2006-01-02"),
				Points: remainingPoints,
			})
			workingDayCount++
		}
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return points
}

func generateActualBurndown(startDate, endDate time.Time, totalPoints int, tickets []*model.Ticket) []BurndownPoint {
	var points []BurndownPoint

	// Start one day before the sprint to show initial capacity
	dayBeforeSprint := startDate.AddDate(0, 0, -1)
	points = append(points, BurndownPoint{
		Date:   dayBeforeSprint.Format("2006-01-02"),
		Points: totalPoints,
	})

	// Generate actual burndown for each working day (excluding weekends)
	currentDate := startDate
	for !currentDate.After(endDate) {
		if !isWeekend(currentDate) {
			// Calculate end of day (23:59:59)
			endOfDay := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), 23, 59, 59, 0, currentDate.Location())

			// Calculate points completed by end of this day
			completedPoints := 0
			for _, ticket := range tickets {
				// If ticket is closed and was closed by end of this day, count its points as completed
				if ticket.Status == "closed" {
					// Use updated_at as proxy for when ticket was closed
					closedDate := ticket.UpdatedAt.UTC()
					if closedDate.Before(endOfDay) || closedDate.Equal(endOfDay) {
						ticketPoints := 0
						if ticket.Size != nil {
							ticketPoints = *ticket.Size
						}
						completedPoints += ticketPoints
					}
				}
			}

			remainingPoints := totalPoints - completedPoints
			if remainingPoints < 0 {
				remainingPoints = 0
			}

			points = append(points, BurndownPoint{
				Date:   currentDate.Format("2006-01-02"),
				Points: remainingPoints,
			})
		}
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return points
}
