package handler

import (
	"encoding/json"
	"net/http"

	"github.com/cfegela/flyhalf/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type TeamHandler struct {
	teamRepo *model.TeamRepository
}

func NewTeamHandler(teamRepo *model.TeamRepository) *TeamHandler {
	return &TeamHandler{teamRepo: teamRepo}
}

type CreateTeamRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateTeamRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *TeamHandler) ListTeams(w http.ResponseWriter, r *http.Request) {
	teams, err := h.teamRepo.List(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list teams"}`, http.StatusInternalServerError)
		return
	}

	if teams == nil {
		teams = []*model.Team{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teams)
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid team ID"}`, http.StatusBadRequest)
		return
	}

	team, err := h.teamRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"team not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(team)
}

func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var req CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}

	team := &model.Team{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.teamRepo.Create(r.Context(), team); err != nil {
		http.Error(w, `{"error":"failed to create team"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(team)
}

func (h *TeamHandler) UpdateTeam(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid team ID"}`, http.StatusBadRequest)
		return
	}

	team, err := h.teamRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"team not found"}`, http.StatusNotFound)
		return
	}

	var req UpdateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}

	team.Name = req.Name
	team.Description = req.Description

	if err := h.teamRepo.Update(r.Context(), team); err != nil {
		http.Error(w, `{"error":"failed to update team"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(team)
}

func (h *TeamHandler) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid team ID"}`, http.StatusBadRequest)
		return
	}

	if err := h.teamRepo.Delete(r.Context(), id); err != nil {
		http.Error(w, `{"error":"failed to delete team"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
