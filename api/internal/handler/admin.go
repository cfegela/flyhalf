package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cfegela/flyhalf/internal/auth"
	"github.com/cfegela/flyhalf/internal/model"
	"github.com/cfegela/flyhalf/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AdminHandler struct {
	userRepo    *model.UserRepository
	ticketRepo  *model.TicketRepository
	sprintRepo  *model.SprintRepository
	projectRepo *model.ProjectRepository
}

func NewAdminHandler(userRepo *model.UserRepository, ticketRepo *model.TicketRepository, sprintRepo *model.SprintRepository, projectRepo *model.ProjectRepository) *AdminHandler {
	return &AdminHandler{
		userRepo:    userRepo,
		ticketRepo:  ticketRepo,
		sprintRepo:  sprintRepo,
		projectRepo: projectRepo,
	}
}

type CreateUserRequest struct {
	Email     string          `json:"email"`
	Password  string          `json:"password"`
	Role      model.UserRole  `json:"role"`
	FirstName string          `json:"first_name"`
	LastName  string          `json:"last_name"`
	TeamID    *uuid.UUID      `json:"team_id,omitempty"`
}

type UpdateUserRequest struct {
	Email     string          `json:"email"`
	Password  string          `json:"password,omitempty"`
	Role      model.UserRole  `json:"role"`
	FirstName string          `json:"first_name"`
	LastName  string          `json:"last_name"`
	IsActive  bool            `json:"is_active"`
	TeamID    *uuid.UUID      `json:"team_id,omitempty"`
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userRepo.List(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list users"}`, http.StatusInternalServerError)
		return
	}

	for _, user := range users {
		user.PasswordHash = ""
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// ListUsersForAssignment returns a simplified list of users for ticket assignment
// Available to all authenticated users
func (h *AdminHandler) ListUsersForAssignment(w http.ResponseWriter, r *http.Request) {
	users, err := h.userRepo.List(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list users"}`, http.StatusInternalServerError)
		return
	}

	// Return simplified user info (no password hash, no sensitive fields)
	type UserForAssignment struct {
		ID        uuid.UUID `json:"id"`
		FirstName string    `json:"first_name"`
		LastName  string    `json:"last_name"`
		Email     string    `json:"email"`
	}

	simplified := make([]UserForAssignment, len(users))
	for i, user := range users {
		simplified[i] = UserForAssignment{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Email:     user.Email,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(simplified)
}

func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid user ID"}`, http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}

	user.PasswordHash = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Sanitize input
	req.Email = util.SanitizeString(req.Email)
	req.FirstName = util.SanitizeString(req.FirstName)
	req.LastName = util.SanitizeString(req.LastName)

	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		http.Error(w, `{"error":"missing required fields"}`, http.StatusBadRequest)
		return
	}

	// Validate email
	if err := util.ValidateEmail(req.Email); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	// Validate password strength
	if err := auth.ValidatePassword(req.Password); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	if req.Role != model.RoleAdmin && req.Role != model.RoleUser {
		req.Role = model.RoleUser
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, `{"error":"failed to hash password"}`, http.StatusInternalServerError)
		return
	}

	user := &model.User{
		Email:              req.Email,
		PasswordHash:       passwordHash,
		Role:               req.Role,
		FirstName:          req.FirstName,
		LastName:           req.LastName,
		IsActive:           true,
		MustChangePassword: true,
		TeamID:             req.TeamID,
	}

	if err := h.userRepo.Create(r.Context(), user); err != nil {
		http.Error(w, `{"error":"failed to create user"}`, http.StatusInternalServerError)
		return
	}

	user.PasswordHash = ""

	// Log user creation
	adminID, _ := auth.GetUserID(r.Context())
	util.LogSecurityEvent(util.EventUserCreated, &adminID, user.Email, util.GetIPFromRequest(r),
		fmt.Sprintf("created user %s with role %s", user.Email, user.Role))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid user ID"}`, http.StatusBadRequest)
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}

	// Sanitize input
	req.Email = util.SanitizeString(req.Email)
	req.FirstName = util.SanitizeString(req.FirstName)
	req.LastName = util.SanitizeString(req.LastName)

	if req.Email != "" {
		// Validate email
		if err := util.ValidateEmail(req.Email); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
			return
		}
		user.Email = req.Email
	}
	if req.Role == model.RoleAdmin || req.Role == model.RoleUser {
		user.Role = req.Role
	}
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	user.IsActive = req.IsActive
	user.TeamID = req.TeamID

	// Handle password reset
	if req.Password != "" {
		// Validate password strength
		if err := auth.ValidatePassword(req.Password); err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
			return
		}

		passwordHash, err := auth.HashPassword(req.Password)
		if err != nil {
			http.Error(w, `{"error":"failed to hash password"}`, http.StatusInternalServerError)
			return
		}
		user.PasswordHash = passwordHash
		user.MustChangePassword = true
	}

	if err := h.userRepo.Update(r.Context(), user); err != nil {
		http.Error(w, `{"error":"failed to update user"}`, http.StatusInternalServerError)
		return
	}

	user.PasswordHash = ""

	// Log user update
	adminID, _ := auth.GetUserID(r.Context())
	util.LogSecurityEvent(util.EventUserUpdated, &adminID, user.Email, util.GetIPFromRequest(r),
		fmt.Sprintf("updated user %s", user.Email))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid user ID"}`, http.StatusBadRequest)
		return
	}

	// Get user info before deletion for logging
	userToDelete, _ := h.userRepo.GetByID(r.Context(), id)

	if err := h.userRepo.Delete(r.Context(), id); err != nil {
		http.Error(w, `{"error":"failed to delete user"}`, http.StatusInternalServerError)
		return
	}

	// Log user deletion
	adminID, _ := auth.GetUserID(r.Context())
	email := "unknown"
	if userToDelete != nil {
		email = userToDelete.Email
	}
	util.LogSecurityEvent(util.EventUserDeleted, &adminID, email, util.GetIPFromRequest(r),
		fmt.Sprintf("deleted user %s", email))

	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) ResetDemo(w http.ResponseWriter, r *http.Request) {
	// Delete tickets first due to foreign key constraints
	ticketsDeleted, err := h.ticketRepo.DeleteAll(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to delete tickets"}`, http.StatusInternalServerError)
		return
	}

	// Delete sprints
	sprintsDeleted, err := h.sprintRepo.DeleteAll(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to delete sprints"}`, http.StatusInternalServerError)
		return
	}

	// Delete projects
	projectsDeleted, err := h.projectRepo.DeleteAll(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to delete projects"}`, http.StatusInternalServerError)
		return
	}

	// Log demo reset
	adminID, _ := auth.GetUserID(r.Context())
	util.LogSecurityEvent(util.EventDemoReset, &adminID, "", util.GetIPFromRequest(r),
		fmt.Sprintf("reset demo: %d tickets, %d sprints, %d projects deleted", ticketsDeleted, sprintsDeleted, projectsDeleted))

	response := map[string]interface{}{
		"message":          "Demo environment reset successfully",
		"tickets_deleted":  ticketsDeleted,
		"sprints_deleted":  sprintsDeleted,
		"projects_deleted": projectsDeleted,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AdminHandler) ReseedDemo(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user ID
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Create a demo project
	project := &model.Project{
		UserID:      userID,
		Name:        "Demo Project",
		Description: "Sample project for demonstration purposes",
	}
	if err := h.projectRepo.Create(r.Context(), project); err != nil {
		http.Error(w, `{"error":"failed to create demo project"}`, http.StatusInternalServerError)
		return
	}

	// Create a demo sprint (2 week sprint starting today)
	now := time.Now().UTC()
	sprint := &model.Sprint{
		UserID:    userID,
		Name:      "Demo Sprint",
		StartDate: now,
		EndDate:   now.AddDate(0, 0, 13), // 14 day sprint (start date + 13 days)
	}
	if err := h.sprintRepo.Create(r.Context(), sprint); err != nil {
		http.Error(w, `{"error":"failed to create demo sprint"}`, http.StatusInternalServerError)
		return
	}

	// Create 5 demo tickets with different valid statuses
	demoTickets := []struct {
		title       string
		description string
		status      string
		size        *int
	}{
		{"Implement user authentication", "Add JWT-based authentication to the API", "closed", intPtr(5)},
		{"Create dashboard UI", "Design and implement the main dashboard interface", "in-progress", intPtr(8)},
		{"Write API documentation", "Document all API endpoints with examples", "needs-review", intPtr(3)},
		{"Fix database connection pooling", "Investigate and resolve connection pool issues", "blocked", intPtr(5)},
		{"Add email notifications", "Implement email notifications for important events", "open", intPtr(3)},
	}

	ticketsCreated := 0
	for i, demo := range demoTickets {
		ticket := &model.Ticket{
			UserID:          userID,
			Title:           demo.title,
			Description:     demo.description,
			Status:          demo.status,
			ProjectID:       &project.ID,
			SprintID:        &sprint.ID,
			Size:            demo.size,
			Priority:        float64(5 - i), // Descending priority
			SprintOrder:     float64(5 - i), // Descending order
			AddedToSprintAt: &now,
		}
		if err := h.ticketRepo.Create(r.Context(), ticket); err != nil {
			http.Error(w, `{"error":"failed to create demo tickets"}`, http.StatusInternalServerError)
			return
		}
		ticketsCreated++
	}

	// Log demo reseed
	util.LogSecurityEvent(util.EventDemoReseed, &userID, "", util.GetIPFromRequest(r),
		fmt.Sprintf("reseeded demo: %d tickets, 1 sprint, 1 project created", ticketsCreated))

	response := map[string]interface{}{
		"message":          "Demo environment reseeded successfully",
		"tickets_created":  ticketsCreated,
		"sprints_created":  1,
		"projects_created": 1,
		"project_id":       project.ID,
		"sprint_id":        sprint.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper function to create int pointer
func intPtr(i int) *int {
	return &i
}
