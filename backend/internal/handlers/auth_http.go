package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"gh-ts/internal/middleware"
	"gh-ts/internal/repository"
	"gh-ts/internal/service"
	"gh-ts/internal/utils"
)

type AuthHTTP struct {
	svc   *service.AuthService
	users repository.UserRepository
}

func NewAuthHTTP(s *service.AuthService, users repository.UserRepository) *AuthHTTP {
	return &AuthHTTP{svc: s, users: users}
}

func (h *AuthHTTP) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Email    string `json:"email"`
			Name     string `json:"name"`
			Password string `json:"password"`
			Role     string `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			utils.Error(w, http.StatusBadRequest, "invalid json")
			return
		}
		u, err := h.svc.Register(r.Context(), in.Email, in.Name, in.Password, in.Role)
		if err != nil {
			utils.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		utils.JSON(w, http.StatusCreated, u)
	}
}

func (h *AuthHTTP) Login(secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			utils.Error(w, http.StatusBadRequest, "invalid json")
			return
		}

		token, u, err := h.svc.Login(r.Context(), in.Email, in.Password)
		if err != nil {
			utils.Error(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		// Issue httpOnly session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			// Lax works for same-origin (frontend via Nginx proxy)
			SameSite: http.SameSiteLaxMode,
			// Set true behind HTTPS in prod
			Secure:  false,
			Expires: time.Now().Add(24 * time.Hour),
		})

		// Return the public profile as body
		utils.JSON(w, http.StatusOK, map[string]any{
			"id":        u.ID,
			"name":      u.Name,
			"email":     u.Email,
			"role":      u.Role,
			"createdAt": u.CreatedAt,
			"updatedAt": u.UpdatedAt,
		})
	}
}

func (h *AuthHTTP) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   -1,              // expire immediately
			Expires:  time.Unix(0, 0), // for older browsers
		})
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *AuthHTTP) Me() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uid, ok := utils.GetString(r.Context(), middleware.CtxUserID)
		if !ok || uid == "" {
			utils.Error(w, http.StatusUnauthorized, "not authenticated")
			return
		}

		// Load full user profile
		u, err := h.users.GetByID(r.Context(), uid)
		if err != nil || u == nil {
			utils.Error(w, http.StatusNotFound, "user not found")
			return
		}

		utils.JSON(w, http.StatusOK, map[string]any{
			"id":        u.ID,
			"name":      u.Name,
			"email":     u.Email,
			"role":      u.Role,
			"createdAt": u.CreatedAt,
			"updatedAt": u.UpdatedAt,
		})
	}
}
