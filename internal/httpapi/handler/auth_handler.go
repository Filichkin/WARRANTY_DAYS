package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"warranty_days/internal/service"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type authTokensResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

type userResponse struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	user, err := h.authSvc.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidEmail):
			writeJSONError(w, http.StatusBadRequest, "invalid email")
			return
		case errors.Is(err, service.ErrWeakPassword):
			writeJSONError(w, http.StatusBadRequest, "weak password")
			return
		case errors.Is(err, service.ErrEmailAlreadyExists):
			writeJSONError(w, http.StatusConflict, "email already exists")
			return
		default:
			writeJSONError(w, http.StatusInternalServerError, "internal error")
			return
		}
	}

	resp := userResponse{
		ID:       user.ID,
		Email:    user.Email,
		IsActive: user.IsActive,
	}

	writeJSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	tokens, err := h.authSvc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			writeJSONError(w, http.StatusUnauthorized, "invalid credentials")
			return
		case errors.Is(err, service.ErrUserInactive):
			writeJSONError(w, http.StatusForbidden, "user is inactive")
			return
		default:
			writeJSONError(w, http.StatusInternalServerError, "internal error")
			return
		}
	}

	resp := authTokensResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    "Bearer",
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	tokens, err := h.authSvc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			writeJSONError(w, http.StatusUnauthorized, "invalid refresh token")
			return
		default:
			writeJSONError(w, http.StatusInternalServerError, "internal error")
			return
		}
	}

	resp := authTokensResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    "Bearer",
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
