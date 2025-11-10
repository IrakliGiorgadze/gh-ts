package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

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
