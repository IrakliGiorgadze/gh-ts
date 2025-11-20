package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"gh-ts/internal/middleware"
	"gh-ts/internal/repository"
	"gh-ts/internal/utils"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type UserHTTP struct {
	repo repository.UserRepository
}

func NewUserHTTP(r repository.UserRepository) *UserHTTP {
	return &UserHTTP{repo: r}
}

// POST /api/users (admin-only)
func (h *UserHTTP) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email    string `json:"email"`
			Name     string `json:"name"`
			Password string `json:"password"`
			Role     string `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.Error(w, http.StatusBadRequest, "invalid json")
			return
		}

		// Validate required fields
		email := strings.TrimSpace(req.Email)
		name := strings.TrimSpace(req.Name)
		password := req.Password
		role := strings.ToLower(strings.TrimSpace(req.Role))

		if email == "" || name == "" || password == "" || role == "" {
			utils.Error(w, http.StatusBadRequest, "email, name, password, and role are required")
			return
		}

		// Validate name length
		if !utils.ValidateLength(name, utils.MaxNameLength) {
			utils.Error(w, http.StatusBadRequest, "name too long (max 100 characters)")
			return
		}

		// Validate email format
		if !utils.ValidateEmail(email) {
			utils.Error(w, http.StatusBadRequest, "invalid email format")
			return
		}

		// Validate password strength
		if err := utils.ValidatePasswordStrength(password); err != nil {
			// Password validation errors are safe to expose
			utils.Error(w, http.StatusBadRequest, err.Error())
			return
		}

		// Validate role
		allowedRoles := map[string]bool{
			"admin":    true,
			"agent":    true,
			"end_user": true,
		}
		if !allowedRoles[role] {
			utils.Error(w, http.StatusBadRequest, "invalid role. allowed: admin, agent, end_user")
			return
		}

		// Hash password
		hash, err := utils.HashPassword(password)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, "failed to hash password")
			return
		}

		// Create user
		u, err := h.repo.Create(r.Context(), email, name, role, hash)
		if err != nil {
			env := utils.GetEnv(r.Context())
			// Check for duplicate email error
			errStr := strings.ToLower(err.Error())
			if strings.Contains(errStr, "duplicate") || strings.Contains(errStr, "unique") {
				utils.Error(w, http.StatusBadRequest, "email already exists")
				return
			}
			utils.Error(w, http.StatusInternalServerError, utils.SafeError(env, err))
			return
		}

		utils.JSON(w, http.StatusCreated, u)
	}
}

// GET /api/users?q=&role=&active=&limit=&offset=
func (h *UserHTTP) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		qv := r.URL.Query()
		q := qv.Get("q")
		role := qv.Get("role")
		var active *bool
		if s := qv.Get("active"); s != "" {
			v, _ := strconv.ParseBool(s)
			active = &v
		}
		limit := utils.QueryInt(qv, "limit", 20)
		offset := utils.QueryInt(qv, "offset", 0)

		users, total, err := h.repo.List(r.Context(), q, role, active, limit, offset)
		if err != nil {
			env := utils.GetEnv(r.Context())
			utils.Error(w, http.StatusInternalServerError, utils.SafeError(env, err))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"items": users, "total": total})
	}
}

// PATCH /api/users/{id}/role
func (h *UserHTTP) UpdateRole() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var req struct {
			Role string `json:"role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Role == "" {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		u, err := h.repo.UpdateRole(r.Context(), id, req.Role)
		if err != nil {
			env := utils.GetEnv(r.Context())
			utils.Error(w, http.StatusInternalServerError, utils.SafeError(env, err))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(u)
	}
}

// PATCH /api/users/{id}/active
func (h *UserHTTP) SetActive() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var req struct {
			Active bool `json:"active"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		u, err := h.repo.SetActive(r.Context(), id, req.Active)
		if err != nil {
			env := utils.GetEnv(r.Context())
			utils.Error(w, http.StatusInternalServerError, utils.SafeError(env, err))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(u)
	}
}

// PATCH /api/users/{id}/basic
func (h *UserHTTP) UpdateBasic() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		name := strings.TrimSpace(req.Name)
		if !utils.ValidateLength(name, utils.MaxNameLength) {
			utils.Error(w, http.StatusBadRequest, "name too long (max 100 characters)")
			return
		}
		u, err := h.repo.UpdateBasic(r.Context(), id, name)
		if err != nil {
			env := utils.GetEnv(r.Context())
			utils.Error(w, http.StatusInternalServerError, utils.SafeError(env, err))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(u)
	}
}

// PATCH /api/users/{id}/password
func (h *UserHTTP) UpdatePassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		
		// Validate UUID format
		if _, err := uuid.Parse(id); err != nil {
			utils.Error(w, http.StatusBadRequest, "invalid user id")
			return
		}
		
		// Verify user can only update their own password (unless admin)
		uid, _ := utils.GetString(r.Context(), middleware.CtxUserID)
		role, _ := utils.GetString(r.Context(), middleware.CtxRole)
		
		if role != "admin" && uid != id {
			utils.Error(w, http.StatusForbidden, "can only update your own password")
			return
		}
		
		var req struct {
			Password        string `json:"password"`
			CurrentPassword string `json:"currentPassword"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.Error(w, http.StatusBadRequest, "invalid json")
			return
		}
		
		// Validate new password strength (use same validation as registration)
		if err := utils.ValidatePasswordStrength(req.Password); err != nil {
			utils.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		
		// For non-admin users, verify current password
		if role != "admin" {
			if req.CurrentPassword == "" {
				utils.Error(w, http.StatusBadRequest, "current password is required")
				return
			}
			
			// Get user with password hash
			u, hash, err := h.repo.GetByIDWithHash(r.Context(), id)
			if err != nil || u == nil {
				utils.Error(w, http.StatusNotFound, "user not found")
				return
			}
			
			// Verify current password
			if !utils.CheckPassword(hash, req.CurrentPassword) {
				utils.Error(w, http.StatusUnauthorized, "current password is incorrect")
				return
			}
		}
		
		// Hash the new password
		hash, err := utils.HashPassword(req.Password)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, "failed to hash password")
			return
		}
		
		// Update password
		if err := h.repo.UpdatePasswordHash(r.Context(), id, hash); err != nil {
			utils.Error(w, http.StatusInternalServerError, "failed to update password")
			return
		}
		
		utils.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
