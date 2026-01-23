package handler

import (
	"encoding/json"
	"net/http"

	"github.com/cfegela/flyhalf/internal/auth"
	"github.com/cfegela/flyhalf/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AdminHandler struct {
	userRepo *model.UserRepository
}

func NewAdminHandler(userRepo *model.UserRepository) *AdminHandler {
	return &AdminHandler{userRepo: userRepo}
}

type CreateUserRequest struct {
	Email     string          `json:"email"`
	Password  string          `json:"password"`
	Role      model.UserRole  `json:"role"`
	FirstName string          `json:"first_name"`
	LastName  string          `json:"last_name"`
}

type UpdateUserRequest struct {
	Email     string          `json:"email"`
	Role      model.UserRole  `json:"role"`
	FirstName string          `json:"first_name"`
	LastName  string          `json:"last_name"`
	IsActive  bool            `json:"is_active"`
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

	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		http.Error(w, `{"error":"missing required fields"}`, http.StatusBadRequest)
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
	}

	if err := h.userRepo.Create(r.Context(), user); err != nil {
		http.Error(w, `{"error":"failed to create user"}`, http.StatusInternalServerError)
		return
	}

	user.PasswordHash = ""

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

	if req.Email != "" {
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

	if err := h.userRepo.Update(r.Context(), user); err != nil {
		http.Error(w, `{"error":"failed to update user"}`, http.StatusInternalServerError)
		return
	}

	user.PasswordHash = ""

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

	if err := h.userRepo.Delete(r.Context(), id); err != nil {
		http.Error(w, `{"error":"failed to delete user"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
