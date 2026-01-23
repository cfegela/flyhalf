package handler

import (
	"encoding/json"
	"net/http"

	"github.com/cfegela/flyhalf/internal/auth"
	"github.com/cfegela/flyhalf/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type EpicHandler struct {
	epicRepo *model.EpicRepository
}

func NewEpicHandler(epicRepo *model.EpicRepository) *EpicHandler {
	return &EpicHandler{epicRepo: epicRepo}
}

type CreateEpicRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateEpicRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *EpicHandler) ListEpics(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// All users can see all epics
	epics, err := h.epicRepo.List(r.Context(), nil)

	if err != nil {
		http.Error(w, `{"error":"failed to list epics"}`, http.StatusInternalServerError)
		return
	}

	if epics == nil {
		epics = []*model.Epic{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(epics)
}

func (h *EpicHandler) GetEpic(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid epic ID"}`, http.StatusBadRequest)
		return
	}

	epic, err := h.epicRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"epic not found"}`, http.StatusNotFound)
		return
	}

	// All authenticated users can view any epic
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(epic)
}

func (h *EpicHandler) CreateEpic(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req CreateEpicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}

	epic := &model.Epic{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.epicRepo.Create(r.Context(), epic); err != nil {
		http.Error(w, `{"error":"failed to create epic"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(epic)
}

func (h *EpicHandler) UpdateEpic(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid epic ID"}`, http.StatusBadRequest)
		return
	}

	epic, err := h.epicRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"epic not found"}`, http.StatusNotFound)
		return
	}

	// All authenticated users can update any epic
	var req UpdateEpicRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Name != "" {
		epic.Name = req.Name
	}
	if req.Description != "" {
		epic.Description = req.Description
	}

	if err := h.epicRepo.Update(r.Context(), epic); err != nil {
		http.Error(w, `{"error":"failed to update epic"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(epic)
}

func (h *EpicHandler) DeleteEpic(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid epic ID"}`, http.StatusBadRequest)
		return
	}

	epic, err := h.epicRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"epic not found"}`, http.StatusNotFound)
		return
	}

	userID, _ := auth.GetUserID(r.Context())
	userRole, _ := auth.GetUserRole(r.Context())

	// Allow deletion if user is admin OR if user created the epic
	if userRole != model.RoleAdmin && epic.UserID != userID {
		http.Error(w, `{"error":"you can only delete epics you created"}`, http.StatusForbidden)
		return
	}

	if err := h.epicRepo.Delete(r.Context(), id); err != nil {
		http.Error(w, `{"error":"failed to delete epic"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
