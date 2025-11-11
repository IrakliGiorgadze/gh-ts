package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"gh-ts/internal/middleware"
	"gh-ts/internal/models"
	"gh-ts/internal/repository"
	"gh-ts/internal/utils"
)

// TicketHTTP wires HTTP endpoints to repositories.
type TicketHTTP struct {
	tickets repository.TicketRepository
	users   repository.UserRepository
}

func NewTicketHTTP(tickets repository.TicketRepository, users repository.UserRepository) *TicketHTTP {
	return &TicketHTTP{tickets: tickets, users: users}
}

// -----------------------------------------------------------------------------
// Optional repo capability for DB-based default admin lookup (kept for future).
// -----------------------------------------------------------------------------
type defaultAdminFinder interface {
	FirstActiveAdminID(ctx context.Context) (string, error)
}

func (h *TicketHTTP) getDefaultAdminID(ctx context.Context) (string, error) {
	// Prefer the users repo (which we have)
	if h.users != nil {
		if id, err := h.users.FirstActiveAdminID(ctx); err == nil && strings.TrimSpace(id) != "" {
			return id, nil
		}
	}
	// As a fallback, if ticket repo happened to implement it (unlikely)
	if af, ok := any(h.tickets).(defaultAdminFinder); ok {
		if id, err := af.FirstActiveAdminID(ctx); err == nil && strings.TrimSpace(id) != "" {
			return id, nil
		}
	}
	return "", context.DeadlineExceeded // sentinel "not found"
}

// -----------------------------------------------------------------------------
// GET /api/tickets ... (unchanged list logic, just swapped h.repo -> h.tickets)
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

		role, _ := utils.GetString(r.Context(), middleware.CtxRole)
		uid, _ := utils.GetString(r.Context(), middleware.CtxUserID)

		type adv interface {
			ListAdv(ctx context.Context, q, status, priority, category, assignee, sort, order string, limit, offset int) ([]models.Ticket, error)
			CountAdv(ctx context.Context, q, status, priority, category, assignee string) (int, error)
		}
		if ar, ok := h.tickets.(adv); ok {
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
			if role == "end_user" && uid != "" {
				filtered := make([]models.Ticket, 0, len(items))
				for _, t := range items {
					if t.CreatedBy == uid {
						filtered = append(filtered, t)
					}
				}
				items = filtered
				total = len(filtered)
			}
			w.Header().Set("X-Total-Count", strconv.Itoa(total))
			utils.JSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
			return
		}

		// legacy
		items, err := h.tickets.List(r.Context(), q, status, limit, offset)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		if role == "end_user" && uid != "" {
			filtered := make([]models.Ticket, 0, len(items))
			for _, t := range items {
				if t.CreatedBy == uid {
					filtered = append(filtered, t)
				}
			}
			items = filtered
		}
		w.Header().Set("X-Total-Count", strconv.Itoa(len(items)))
		utils.JSON(w, http.StatusOK, map[string]any{"items": items, "total": len(items)})
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
		t, err := h.tickets.Get(r.Context(), id)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		if t == nil {
			utils.Error(w, http.StatusNotFound, "not found")
			return
		}
		role, _ := utils.GetString(r.Context(), middleware.CtxRole)
		uid, _ := utils.GetString(r.Context(), middleware.CtxUserID)
		if role == "end_user" && uid != "" && t.CreatedBy != uid {
			utils.Error(w, http.StatusForbidden, "forbidden")
			return
		}
		utils.JSON(w, http.StatusOK, t)
	}
}

// -----------------------------------------------------------------------------
// POST /api/tickets
// If creator is end_user → ALWAYS auto-assign to first active admin from DB.
// (If none available → 500 with a clear message.)
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

		uid, _ := utils.GetString(r.Context(), middleware.CtxUserID)
		if uid == "" {
			utils.Error(w, http.StatusUnauthorized, "not authenticated")
			return
		}
		role, _ := utils.GetString(r.Context(), middleware.CtxRole)

		assignee := strings.TrimSpace(in.Assignee)

		// Hard rule: end users' tickets must be assigned to an admin
		if role == "end_user" {
			adminID, err := h.getDefaultAdminID(r.Context())
			if err != nil || strings.TrimSpace(adminID) == "" {
				utils.Error(w, http.StatusInternalServerError, "no active admin available for assignment")
				return
			}
			assignee = adminID
		} else if role == "admin" && assignee == "" {
			assignee = uid
		}

		t := &models.Ticket{
			Title:       in.Title,
			Description: strings.TrimSpace(in.Description),
			Category:    strings.TrimSpace(in.Category),
			Priority:    strings.TrimSpace(in.Priority),
			Status:      "New",
			Assignee:    assignee,
			Department:  strings.TrimSpace(in.Department),
			CreatedBy:   uid,
		}

		if err := h.tickets.Create(r.Context(), t); err != nil {
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		utils.JSON(w, http.StatusCreated, t)
	}
}

// -----------------------------------------------------------------------------
// PATCH /api/tickets/{id}
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
		role, _ := utils.GetString(r.Context(), middleware.CtxRole)
		switch role {
		case "admin", "agent", "supervisor":
		default:
			utils.Error(w, http.StatusForbidden, "forbidden")
			return
		}

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

		t, err := h.tickets.Get(r.Context(), id)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		if t == nil {
			utils.Error(w, http.StatusNotFound, "not found")
			return
		}

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

		if err := h.tickets.Update(r.Context(), t); err != nil {
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}

		// Fetch the updated ticket with assignee name/email populated via JOIN
		updated, err := h.tickets.Get(r.Context(), t.ID)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		if updated == nil {
			utils.Error(w, http.StatusInternalServerError, "ticket not found after update")
			return
		}
		utils.JSON(w, http.StatusOK, updated)
	}
}

// -----------------------------------------------------------------------------
// POST /api/tickets/{id}/comments
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

		if _, err := h.tickets.AddComment(r.Context(), id, in.Text); err != nil {
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}
		t, err := h.tickets.Get(r.Context(), id)
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
