package handler

import (
	"encoding/json"
	"net/http"

	"github.com/cfegela/flyhalf/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type LeagueHandler struct {
	leagueRepo *model.LeagueRepository
}

func NewLeagueHandler(leagueRepo *model.LeagueRepository) *LeagueHandler {
	return &LeagueHandler{leagueRepo: leagueRepo}
}

type CreateLeagueRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateLeagueRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *LeagueHandler) ListLeagues(w http.ResponseWriter, r *http.Request) {
	leagues, err := h.leagueRepo.List(r.Context())
	if err != nil {
		http.Error(w, `{"error":"failed to list leagues"}`, http.StatusInternalServerError)
		return
	}

	if leagues == nil {
		leagues = []*model.League{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leagues)
}

func (h *LeagueHandler) GetLeague(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid league ID"}`, http.StatusBadRequest)
		return
	}

	league, err := h.leagueRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"league not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(league)
}

func (h *LeagueHandler) CreateLeague(w http.ResponseWriter, r *http.Request) {
	var req CreateLeagueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}

	league := &model.League{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.leagueRepo.Create(r.Context(), league); err != nil {
		http.Error(w, `{"error":"failed to create league"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(league)
}

func (h *LeagueHandler) UpdateLeague(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid league ID"}`, http.StatusBadRequest)
		return
	}

	league, err := h.leagueRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"league not found"}`, http.StatusNotFound)
		return
	}

	var req UpdateLeagueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, `{"error":"name is required"}`, http.StatusBadRequest)
		return
	}

	league.Name = req.Name
	league.Description = req.Description

	if err := h.leagueRepo.Update(r.Context(), league); err != nil {
		http.Error(w, `{"error":"failed to update league"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(league)
}

func (h *LeagueHandler) DeleteLeague(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, `{"error":"invalid league ID"}`, http.StatusBadRequest)
		return
	}

	if err := h.leagueRepo.Delete(r.Context(), id); err != nil {
		http.Error(w, `{"error":"failed to delete league"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
