package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cfegela/flyhalf/internal/auth"
	"github.com/cfegela/flyhalf/internal/model"
	"github.com/cfegela/flyhalf/internal/util"
)

type AuthHandler struct {
	userRepo   *model.UserRepository
	jwtService *auth.JWTService
	isProduction bool
}

func NewAuthHandler(userRepo *model.UserRepository, jwtService *auth.JWTService, isProduction bool) *AuthHandler {
	return &AuthHandler{
		userRepo:   userRepo,
		jwtService: jwtService,
		isProduction: isProduction,
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

	// Validate email format
	if err := util.ValidateEmail(req.Email); err != nil {
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	user, err := h.userRepo.GetByEmail(r.Context(), req.Email)

	// Always check password even if user not found to prevent timing attacks
	// Use a dummy hash with the same cost as real passwords
	// nosemgrep: detected-bcrypt-hash
	dummyHash := "$2a$12$R2iQS4ZXc0z1h7Oq2wAOKeqslDynZTXBkt9chHBIVIRUuUVO.nbPi"
	passwordHash := dummyHash

	if err == nil && user != nil {
		passwordHash = user.PasswordHash
	}

	// Check password (always runs, preventing timing attacks)
	passwordValid := auth.CheckPassword(req.Password, passwordHash)

	// Now verify all conditions
	if err != nil || user == nil || !user.IsActive || !passwordValid {
		// Log failed login attempt
		util.LogSecurityEvent(util.EventLoginFailure, nil, req.Email, util.GetIPFromRequest(r), "invalid credentials")
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	// Log successful login
	util.LogSecurityEvent(util.EventLoginSuccess, &user.ID, user.Email, util.GetIPFromRequest(r), "login successful")

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
		Secure:   true, // Always require HTTPS for secure cookies
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
		Secure:   true, // Always require HTTPS for secure cookies
		SameSite: http.SameSiteStrictMode,
	})

	user.PasswordHash = ""

	// Log token refresh
	util.LogSecurityEvent(util.EventTokenRefresh, &user.ID, user.Email, util.GetIPFromRequest(r), "token refreshed")

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

	// Log logout
	util.LogSecurityEvent(util.EventLogout, &userID, "", util.GetIPFromRequest(r), "user logged out")

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true, // Always require HTTPS for secure cookies
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

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		http.Error(w, `{"error":"current password and new password are required"}`, http.StatusBadRequest)
		return
	}

	// Validate password strength
	if err := auth.ValidatePassword(req.NewPassword); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}

	if !auth.CheckPassword(req.CurrentPassword, user.PasswordHash) {
		http.Error(w, `{"error":"current password is incorrect"}`, http.StatusUnauthorized)
		return
	}

	newPasswordHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		http.Error(w, `{"error":"failed to hash password"}`, http.StatusInternalServerError)
		return
	}

	if err := h.userRepo.UpdatePassword(r.Context(), userID, newPasswordHash); err != nil {
		http.Error(w, `{"error":"failed to update password"}`, http.StatusInternalServerError)
		return
	}

	// Log password change
	util.LogSecurityEvent(util.EventPasswordChange, &userID, user.Email, util.GetIPFromRequest(r), "password changed successfully")

	w.WriteHeader(http.StatusNoContent)
}
