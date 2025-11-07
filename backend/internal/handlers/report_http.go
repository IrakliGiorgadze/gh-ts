package handlers

import (
	"net/http"
	"time"

	"gh-ts/internal/repository"
	"gh-ts/internal/utils"
)

type ReportsHTTP struct {
	repo repository.TicketRepository
}

func NewReportsHTTP(r repository.TicketRepository) *ReportsHTTP { return &ReportsHTTP{repo: r} }

// GET /api/reports/summary
// Returns: { open, resolved7d, highCriticalOpen }
func (h *ReportsHTTP) Summary() http.HandlerFunc {
	type adv interface {
		CountByStatus(ctx interface{ Done() <-chan struct{} }, statuses []string, inclusive bool) (int, error)
		CountResolvedSince(ctx interface{ Done() <-chan struct{} }, since time.Time) (int, error)
		CountOpenByPriorities(ctx interface{ Done() <-chan struct{} }, prios []string) (int, error)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// Fast path if the concrete repo supports counters
		if rr, ok := h.repo.(adv); ok {
			open, err := rr.CountByStatus(r.Context(), []string{"Resolved", "Closed"}, false)
			if err != nil {
				utils.Error(w, http.StatusInternalServerError, err.Error())
				return
			}

			resolved7d, err := rr.CountResolvedSince(r.Context(), time.Now().Add(-7*24*time.Hour))
			if err != nil {
				utils.Error(w, http.StatusInternalServerError, err.Error())
				return
			}

			highCritOpen, err := rr.CountOpenByPriorities(r.Context(), []string{"High", "Critical"})
			if err != nil {
				utils.Error(w, http.StatusInternalServerError, err.Error())
				return
			}

			utils.JSON(w, http.StatusOK, map[string]int{
				"open":             open,
				"resolved7d":       resolved7d,
				"highCriticalOpen": highCritOpen,
			})
			return
		}

		// Fallback (works with any repo): list & compute
		items, err := h.repo.List(r.Context(), "", "", 1000, 0) // cap to 1000 to avoid heavy scans
		if err != nil {
			utils.Error(w, http.StatusInternalServerError, err.Error())
			return
		}

		now := time.Now()
		open, resolved7d, highCritOpen := 0, 0, 0
		for _, t := range items {
			closed := t.Status == "Resolved" || t.Status == "Closed"
			if !closed {
				open++
			}
			updated := t.UpdatedAt
			if updated.IsZero() {
				updated = t.CreatedAt
			}
			if closed && now.Sub(updated) <= 7*24*time.Hour {
				resolved7d++
			}
			if (t.Priority == "High" || t.Priority == "Critical") && !closed {
				highCritOpen++
			}
		}
		utils.JSON(w, http.StatusOK, map[string]int{
			"open":             open,
			"resolved7d":       resolved7d,
			"highCriticalOpen": highCritOpen,
		})
	}
}
