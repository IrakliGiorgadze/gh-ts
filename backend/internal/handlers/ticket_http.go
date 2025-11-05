package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"gh-ts/internal/models"
	"gh-ts/internal/repository"
	"gh-ts/internal/utils"
)

type TicketHTTP struct {
	repo repository.TicketRepository
}

func NewTicketHTTP(repo repository.TicketRepository) *TicketHTTP {
	return &TicketHTTP{repo: repo}
}

func (h *TicketHTTP) List() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		status := r.URL.Query().Get("status")
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
		list, err := h.repo.List(r.Context(), q, status, limit, offset)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, "query failed")
			return
		}
		utils.JSON(w, http.StatusOK, list)
	}
}

func (h *TicketHTTP) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Title, Description, Category, Priority, Department, Assignee string
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			utils.Error(w, http.StatusBadRequest, "invalid json")
			return
		}
		t := &models.Ticket{
			Title:       strings.TrimSpace(in.Title),
			Description: strings.TrimSpace(in.Description),
			Category:    choose(in.Category, "Software"),
			Priority:    choose(in.Priority, "Low"),
			Department:  strings.TrimSpace(in.Department),
			Assignee:    strings.TrimSpace(in.Assignee),
			Status:      "New",
		}
		if t.Title == "" {
			utils.Error(w, http.StatusBadRequest, "title is required")
			return
		}
		if err := h.repo.Create(r.Context(), t); err != nil {
			utils.Error(w, http.StatusInternalServerError, "create failed")
			return
		}
		utils.JSON(w, http.StatusCreated, t)
	}
}

func (h *TicketHTTP) Get() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/tickets/")
		t, err := h.repo.Get(r.Context(), id)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, "get failed")
			return
		}
		if t == nil {
			utils.Error(w, http.StatusNotFound, "not found")
			return
		}
		utils.JSON(w, http.StatusOK, t)
	}
}

func (h *TicketHTTP) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/api/tickets/")
		cur, err := h.repo.Get(r.Context(), id)
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, "get failed")
			return
		}
		if cur == nil {
			utils.Error(w, http.StatusNotFound, "not found")
			return
		}

		var patch map[string]any
		if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
			utils.Error(w, http.StatusBadRequest, "invalid json")
			return
		}
		if v, ok := patch["title"].(string); ok {
			cur.Title = v
		}
		if v, ok := patch["description"].(string); ok {
			cur.Description = v
		}
		if v, ok := patch["category"].(string); ok {
			cur.Category = v
		}
		if v, ok := patch["priority"].(string); ok {
			cur.Priority = v
		}
		if v, ok := patch["status"].(string); ok {
			cur.Status = v
		}
		if v, ok := patch["assignee"].(string); ok {
			cur.Assignee = v
		}
		if v, ok := patch["department"].(string); ok {
			cur.Department = v
		}

		if err := h.repo.Update(r.Context(), cur); err != nil {
			utils.Error(w, http.StatusInternalServerError, "update failed")
			return
		}
		utils.JSON(w, http.StatusOK, cur)
	}
}

func (h *TicketHTTP) AddComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/tickets/"), "/comments")
		var in struct{ Text string }
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil || strings.TrimSpace(in.Text) == "" {
			utils.Error(w, http.StatusBadRequest, "text required")
			return
		}
		if _, err := h.repo.AddComment(r.Context(), id, in.Text); err != nil {
			utils.Error(w, http.StatusInternalServerError, "comment failed")
			return
		}
		// return refreshed ticket
		cur, _ := h.repo.Get(r.Context(), id)
		if cur == nil {
			utils.Error(w, http.StatusNotFound, "not found")
			return
		}
		utils.JSON(w, http.StatusOK, cur)
	}
}

func choose(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}
