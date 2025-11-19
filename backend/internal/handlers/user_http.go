package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"gh-ts/internal/repository"
	"gh-ts/internal/utils"

	"github.com/go-chi/chi/v5"
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

		if len(password) < 6 {
			utils.Error(w, http.StatusBadRequest, "password must be at least 6 characters")
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
			// Check for duplicate email error
			if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
				utils.Error(w, http.StatusBadRequest, "email already exists")
				return
			}
			utils.Error(w, http.StatusInternalServerError, err.Error())
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
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
		u, err := h.repo.UpdateBasic(r.Context(), id, req.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
		var req struct {
			Hash string `json:"hash"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Hash == "" {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		if err := h.repo.UpdatePasswordHash(r.Context(), id, req.Hash); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}
