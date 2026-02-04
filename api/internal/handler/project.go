package handler

import (
	"encoding/json"
	"net/http"

	"github.com/cfegela/flyhalf/internal/auth"
	"github.com/cfegela/flyhalf/internal/model"
	"github.com/cfegela/flyhalf/internal/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ProjectHandler struct {
	projectRepo *model.ProjectRepository
}

func NewProjectHandler(projectRepo *model.ProjectRepository) *ProjectHandler {
	return &ProjectHandler{projectRepo: projectRepo}
}

type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	_, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Check if pagination is requested
	if r.URL.Query().Get("page") != "" || r.URL.Query().Get("limit") != "" {
		params := util.GetPaginationParams(r)
		projects, total, err := h.projectRepo.ListPaginated(r.Context(), nil, params.Limit, params.CalculateOffset())
		if err != nil {
			http.Error(w, `{"error":"failed to list projects"}`, http.StatusInternalServerError)
			return
		}

		if projects == nil {
			projects = []*model.Project{}
		}

		response := util.CreatePaginatedResponse(projects, params.Page, params.Limit, total)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// All users can see all projects (non-paginated for backward compatibility)
	projects, err := h.projectRepo.List(r.Context(), nil)
	if err != nil {
		http.Error(w, `{"error":"failed to list projects"}`, http.StatusInternalServerError)
		return
	}

	if projects == nil {
		projects = []*model.Project{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid project ID"}`, http.StatusBadRequest)
		return
	}

	project, err := h.projectRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"project not found"}`, http.StatusNotFound)
		return
	}

	// All authenticated users can view any project
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Sanitize input
	req.Name = util.SanitizeString(req.Name)
	req.Description = util.SanitizeString(req.Description)

	if req.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}

	if req.Description == "" {
		http.Error(w, `{"error":"description is required"}`, http.StatusBadRequest)
		return
	}

	project := &model.Project{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.projectRepo.Create(r.Context(), project); err != nil {
		http.Error(w, `{"error":"failed to create project"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(project)
}

func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid project ID"}`, http.StatusBadRequest)
		return
	}

	project, err := h.projectRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"project not found"}`, http.StatusNotFound)
		return
	}

	// All authenticated users can update any project
	var req UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Sanitize input
	req.Name = util.SanitizeString(req.Name)
	req.Description = util.SanitizeString(req.Description)

	if req.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}
	project.Name = req.Name

	if req.Description == "" {
		http.Error(w, `{"error":"description is required"}`, http.StatusBadRequest)
		return
	}
	project.Description = req.Description

	if err := h.projectRepo.Update(r.Context(), project); err != nil {
		http.Error(w, `{"error":"failed to update project"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid project ID"}`, http.StatusBadRequest)
		return
	}

	project, err := h.projectRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"project not found"}`, http.StatusNotFound)
		return
	}

	userID, _ := auth.GetUserID(r.Context())
	userRole, _ := auth.GetUserRole(r.Context())

	// Allow deletion if user is admin OR if user created the project
	if userRole != model.RoleAdmin && project.UserID != userID {
		http.Error(w, `{"error":"you can only delete projects you created"}`, http.StatusForbidden)
		return
	}

	if err := h.projectRepo.Delete(r.Context(), id); err != nil {
		http.Error(w, `{"error":"failed to delete project"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
