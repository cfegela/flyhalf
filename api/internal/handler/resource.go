package handler

import (
	"encoding/json"
	"net/http"

	"github.com/cfegela/flyhalf/internal/auth"
	"github.com/cfegela/flyhalf/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ResourceHandler struct {
	resourceRepo *model.ResourceRepository
}

func NewResourceHandler(resourceRepo *model.ResourceRepository) *ResourceHandler {
	return &ResourceHandler{resourceRepo: resourceRepo}
}

type CreateResourceRequest struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type UpdateResourceRequest struct {
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Status      string                 `json:"status"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

func (h *ResourceHandler) ListResources(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	userRole, _ := auth.GetUserRole(r.Context())

	var resources []*model.Resource
	var err error

	if userRole == model.RoleAdmin {
		resources, err = h.resourceRepo.List(r.Context(), nil)
	} else {
		resources, err = h.resourceRepo.List(r.Context(), &userID)
	}

	if err != nil {
		http.Error(w, `{"error":"failed to list resources"}`, http.StatusInternalServerError)
		return
	}

	if resources == nil {
		resources = []*model.Resource{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resources)
}

func (h *ResourceHandler) GetResource(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid resource ID"}`, http.StatusBadRequest)
		return
	}

	resource, err := h.resourceRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"resource not found"}`, http.StatusNotFound)
		return
	}

	userID, _ := auth.GetUserID(r.Context())
	userRole, _ := auth.GetUserRole(r.Context())

	if userRole != model.RoleAdmin && resource.UserID != userID {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resource)
}

func (h *ResourceHandler) CreateResource(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req CreateResourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, `{"error":"title is required"}`, http.StatusBadRequest)
		return
	}

	if req.Status == "" {
		req.Status = "active"
	}

	resource := &model.Resource{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Metadata:    req.Metadata,
	}

	if err := h.resourceRepo.Create(r.Context(), resource); err != nil {
		http.Error(w, `{"error":"failed to create resource"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resource)
}

func (h *ResourceHandler) UpdateResource(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid resource ID"}`, http.StatusBadRequest)
		return
	}

	resource, err := h.resourceRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"resource not found"}`, http.StatusNotFound)
		return
	}

	userID, _ := auth.GetUserID(r.Context())
	userRole, _ := auth.GetUserRole(r.Context())

	if userRole != model.RoleAdmin && resource.UserID != userID {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return
	}

	var req UpdateResourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Title != "" {
		resource.Title = req.Title
	}
	if req.Description != "" {
		resource.Description = req.Description
	}
	if req.Status != "" {
		resource.Status = req.Status
	}
	if req.Metadata != nil {
		resource.Metadata = req.Metadata
	}

	if err := h.resourceRepo.Update(r.Context(), resource); err != nil {
		http.Error(w, `{"error":"failed to update resource"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resource)
}

func (h *ResourceHandler) DeleteResource(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid resource ID"}`, http.StatusBadRequest)
		return
	}

	resource, err := h.resourceRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"resource not found"}`, http.StatusNotFound)
		return
	}

	userID, _ := auth.GetUserID(r.Context())
	userRole, _ := auth.GetUserRole(r.Context())

	if userRole != model.RoleAdmin && resource.UserID != userID {
		http.Error(w, `{"error":"forbidden"}`, http.StatusForbidden)
		return
	}

	if err := h.resourceRepo.Delete(r.Context(), id); err != nil {
		http.Error(w, `{"error":"failed to delete resource"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
