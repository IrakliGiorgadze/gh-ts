package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"gh-ts/internal/middleware"
	"gh-ts/internal/service"
	"gh-ts/internal/utils"
)

type AuthHTTP struct {
	svc *service.AuthService
}

func NewAuthHTTP(s *service.AuthService) *AuthHTTP { return &AuthHTTP{svc: s} }

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

		// httpOnly cookie session
		c := &http.Cookie{
			Name:     "session",
			Value:    token,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   false, // set true behind HTTPS/production
			Expires:  time.Now().Add(24 * time.Hour),
		}
		http.SetCookie(w, c)

		utils.JSON(w, http.StatusOK, u)
	}
}

func (h *AuthHTTP) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c := &http.Cookie{
			Name: "session", Value: "",
			Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode,
			MaxAge: -1,
		}
		http.SetCookie(w, c)
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
		role, _ := utils.GetString(r.Context(), middleware.CtxRole)
		utils.JSON(w, http.StatusOK, map[string]string{"id": uid, "role": role})
	}
}
