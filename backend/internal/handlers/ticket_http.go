package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"gh-ts/internal/models"
	"gh-ts/internal/repository"
	"gh-ts/internal/utils"
)

// TicketHTTP wires HTTP endpoints to a TicketRepository.
type TicketHTTP struct {
	repo repository.TicketRepository
}

func NewTicketHTTP(r repository.TicketRepository) *TicketHTTP { return &TicketHTTP{repo: r} }

// -----------------------------------------------------------------------------
// GET /api/tickets?q=&status=&priority=&category=&assignee=&limit=&offset=&sort=&order=
// Uses advanced repo when available, falls back to legacy List() otherwise.
// Returns: { items: [...], total: N }
// -----------------------------------------------------------------------------
func (h *TicketHTTP) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		qv := r.URL.Query()
		q := strings.TrimSpace(qv.Get("q"))

		status := strings.TrimSpace(qv.Get("status"))
		priority := strings.TrimSpace(qv.Get("priority"))
		category := strings.TrimSpace(qv.Get("category"))
		assignee := strings.TrimSpace(qv.Get("assignee"))

		limit := utils.QueryInt(qv, "limit", 10)
		offset := utils.QueryInt(qv, "offset", 0)
		sort := qv.Get("sort")
		order := qv.Get("order")

		// Prefer advanced methods if the concrete repo provides them.
		type adv interface {
			ListAdv(ctx_ interface{ Done() <-chan struct{} }, q, status, priority, category, assignee, sort, order string, limit, offset int) ([]models.Ticket, error)
			CountAdv(ctx_ interface{ Done() <-chan struct{} }, q, status, priority, category, assignee string) (int, error)
		}
		if ar, ok := h.repo.(adv); ok {
			items, err := ar.ListAdv(r.Context(), q, status, priority, category, assignee, sort, order, limit, offset)
			if err != nil {
				utils.Error(w, http.StatusInternalServerError, err.Error())
				return
			}
			total, err := ar.CountAdv(r.Context(), q, status, priority, category, assignee)
			if err != nil {
				utils.Error(w, http.StatusInternalServerError, err.Error())
				return
			}
			w.Header().Set("X-Total-Count", strconv.Itoa(total))
			utils.JSON(w, http.StatusOK, map[string]any{
				"items": items,
				"total": total,
			})
			return
		}

		// Fallback to legacy List (q + status + paging only)
		items, err := h.repo.List(r.Context(), q, status, limit, offset)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Header().Set("X-Total-Count", strconv.Itoa(len(items)))
		utils.JSON(w, http.StatusOK, map[string]any{
			"items": items,
			"total": len(items),
		})
	}
}

// -----------------------------------------------------------------------------
// GET /api/tickets/{id}
// -----------------------------------------------------------------------------
func (h *TicketHTTP) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			utils.Error(w, http.StatusBadRequest, "missing id")
			return
		}
		t, err := h.repo.Get(r.Context(), id)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		if t == nil {
			utils.Error(w, http.StatusNotFound, "not found")
			return
		}
		utils.JSON(w, http.StatusOK, t)
	}
}

// -----------------------------------------------------------------------------
// POST /api/tickets
// Body: { title, description, category, priority, department, assignee? }
// -----------------------------------------------------------------------------
func (h *TicketHTTP) Create() http.HandlerFunc {
	type inDTO struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Category    string `json:"category"`
		Priority    string `json:"priority"`
		Department  string `json:"department"`
		Assignee    string `json:"assignee"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var in inDTO
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			utils.Error(w, http.StatusBadRequest, "invalid json")
			return
		}
		in.Title = strings.TrimSpace(in.Title)
		if in.Title == "" {
			utils.Error(w, http.StatusBadRequest, "title is required")
			return
		}

		t := &models.Ticket{
			Title:       in.Title,
			Description: strings.TrimSpace(in.Description),
			Category:    strings.TrimSpace(in.Category),
			Priority:    strings.TrimSpace(in.Priority),
			Status:      "New",
			Assignee:    strings.TrimSpace(in.Assignee),
			Department:  strings.TrimSpace(in.Department),
		}

		if err := h.repo.Create(r.Context(), t); err != nil {
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		utils.JSON(w, http.StatusCreated, t)
	}
}

// -----------------------------------------------------------------------------
// PATCH /api/tickets/{id}
// Body: any of { title, description, category, priority, status, assignee, department }
// -----------------------------------------------------------------------------
func (h *TicketHTTP) Update() http.HandlerFunc {
	type inDTO struct {
		Title       *string `json:"title"`
		Description *string `json:"description"`
		Category    *string `json:"category"`
		Priority    *string `json:"priority"`
		Status      *string `json:"status"`
		Assignee    *string `json:"assignee"`
		Department  *string `json:"department"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			utils.Error(w, http.StatusBadRequest, "missing id")
			return
		}

		var in inDTO
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			utils.Error(w, http.StatusBadRequest, "invalid json")
			return
		}

		// load existing
		t, err := h.repo.Get(r.Context(), id)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		if t == nil {
			utils.Error(w, http.StatusNotFound, "not found")
			return
		}

		// apply partial updates
		if in.Title != nil {
			t.Title = strings.TrimSpace(*in.Title)
		}
		if in.Description != nil {
			t.Description = strings.TrimSpace(*in.Description)
		}
		if in.Category != nil {
			t.Category = strings.TrimSpace(*in.Category)
		}
		if in.Priority != nil {
			t.Priority = strings.TrimSpace(*in.Priority)
		}
		if in.Status != nil {
			t.Status = strings.TrimSpace(*in.Status)
		}
		if in.Assignee != nil {
			t.Assignee = strings.TrimSpace(*in.Assignee)
		}
		if in.Department != nil {
			t.Department = strings.TrimSpace(*in.Department)
		}

		if err := h.repo.Update(r.Context(), t); err != nil {
			if err == nil {
				// noop
			}
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		utils.JSON(w, http.StatusOK, t)
	}
}

// -----------------------------------------------------------------------------
// POST /api/tickets/{id}/comments
// Body: { text }
// Returns: the updated ticket (so FE can re-render comments easily)
// -----------------------------------------------------------------------------
func (h *TicketHTTP) AddComment() http.HandlerFunc {
	type inDTO struct {
		Text string `json:"text"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			utils.Error(w, http.StatusBadRequest, "missing id")
			return
		}
		var in inDTO
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			utils.Error(w, http.StatusBadRequest, "invalid json")
			return
		}
		in.Text = strings.TrimSpace(in.Text)
		if in.Text == "" {
			utils.Error(w, http.StatusBadRequest, "text is required")
			return
		}

		if _, err := h.repo.AddComment(r.Context(), id, in.Text); err != nil {
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		// Return the refreshed ticket with comments (matches your FE expectation)
		t, err := h.repo.Get(r.Context(), id)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		if t == nil {
			utils.Error(w, http.StatusNotFound, "not found")
			return
		}
		utils.JSON(w, http.StatusOK, t)
	}
}
