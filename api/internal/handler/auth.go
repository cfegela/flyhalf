package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cfegela/flyhalf/internal/auth"
	"github.com/cfegela/flyhalf/internal/model"
)

type AuthHandler struct {
	userRepo   *model.UserRepository
	jwtService *auth.JWTService
}

func NewAuthHandler(userRepo *model.UserRepository, jwtService *auth.JWTService) *AuthHandler {
	return &AuthHandler{
		userRepo:   userRepo,
		jwtService: jwtService,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string      `json:"access_token"`
	User        *model.User `json:"user"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, `{"error":"email and password are required"}`, http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetByEmail(r.Context(), req.Email)
	if err != nil {
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	if !user.IsActive {
		http.Error(w, `{"error":"account is inactive"}`, http.StatusUnauthorized)
		return
	}

	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	tokenPair, refreshTokenHash, err := h.jwtService.GenerateTokenPair(user)
	if err != nil {
		http.Error(w, `{"error":"failed to generate tokens"}`, http.StatusInternalServerError)
		return
	}

	refreshToken := &model.RefreshToken{
		UserID:    user.ID,
		TokenHash: refreshTokenHash,
		ExpiresAt: h.jwtService.RefreshTokenExpiry(),
	}

	if err := h.userRepo.CreateRefreshToken(r.Context(), refreshToken); err != nil {
		http.Error(w, `{"error":"failed to store refresh token"}`, http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    tokenPair.RefreshToken,
		Path:     "/",
		Expires:  refreshToken.ExpiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	user.PasswordHash = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{
		AccessToken: tokenPair.AccessToken,
		User:        user,
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, `{"error":"refresh token not found"}`, http.StatusUnauthorized)
		return
	}

	refreshTokenHash := h.jwtService.HashRefreshToken(cookie.Value)

	storedToken, err := h.userRepo.GetRefreshToken(r.Context(), refreshTokenHash)
	if err != nil {
		http.Error(w, `{"error":"invalid or expired refresh token"}`, http.StatusUnauthorized)
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), storedToken.UserID)
	if err != nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusUnauthorized)
		return
	}

	if !user.IsActive {
		http.Error(w, `{"error":"account is inactive"}`, http.StatusUnauthorized)
		return
	}

	if err := h.userRepo.RevokeRefreshToken(r.Context(), refreshTokenHash); err != nil {
		http.Error(w, `{"error":"failed to revoke old token"}`, http.StatusInternalServerError)
		return
	}

	tokenPair, newRefreshTokenHash, err := h.jwtService.GenerateTokenPair(user)
	if err != nil {
		http.Error(w, `{"error":"failed to generate tokens"}`, http.StatusInternalServerError)
		return
	}

	newRefreshToken := &model.RefreshToken{
		UserID:    user.ID,
		TokenHash: newRefreshTokenHash,
		ExpiresAt: h.jwtService.RefreshTokenExpiry(),
	}

	if err := h.userRepo.CreateRefreshToken(r.Context(), newRefreshToken); err != nil {
		http.Error(w, `{"error":"failed to store refresh token"}`, http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    tokenPair.RefreshToken,
		Path:     "/",
		Expires:  newRefreshToken.ExpiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	user.PasswordHash = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{
		AccessToken: tokenPair.AccessToken,
		User:        user,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	if err := h.userRepo.RevokeAllUserTokens(r.Context(), userID); err != nil {
		http.Error(w, `{"error":"failed to logout"}`, http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}

	user.PasswordHash = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
