package postgres

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gh-ts/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TicketRepo struct{ db *pgxpool.Pool }

func NewTicketRepo(db *pgxpool.Pool) *TicketRepo { return &TicketRepo{db: db} }

// -----------------------------------------------------------------------------
// Simple list (backward compatible) + assignee name/email join
// -----------------------------------------------------------------------------
func (r *TicketRepo) List(ctx context.Context, q, status string, limit, offset int) ([]models.Ticket, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	args := []any{}
	conds := []string{"1=1"}

	if q = strings.TrimSpace(q); q != "" {
		p := "%" + q + "%"
		args = append(args, p, p)
		// Case-insensitive match on title or description
		conds = append(conds, "(t.title ILIKE $"+itoa(len(args)-1)+" OR t.description ILIKE $"+itoa(len(args))+")")
	}
	if status != "" {
		args = append(args, status)
		conds = append(conds, "t.status = $"+itoa(len(args)))
	}
	args = append(args, limit, offset)

	sql := `
		SELECT
			t.id, t.alias, t.title, t.description, t.category, t.priority, t.status,
			COALESCE(t.assignee, ''), t.department, t.created_by, t.created_at, t.updated_at,
			COALESCE(u.name, ''), COALESCE(u.email, '')
		FROM tickets t
		LEFT JOIN users u ON u.id = NULLIF(t.assignee, '')::uuid
		WHERE ` + strings.Join(conds, " AND ") + `
		ORDER BY t.updated_at DESC
		LIMIT $` + itoa(len(args)-1) + ` OFFSET $` + itoa(len(args))

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Ticket
	for rows.Next() {
		var t models.Ticket
		if err := rows.Scan(
			&t.ID, &t.Alias, &t.Title, &t.Description, &t.Category, &t.Priority,
			&t.Status, &t.Assignee, &t.Department, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
			&t.AssigneeName, &t.AssigneeEmail,
		); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// -----------------------------------------------------------------------------
// Advanced listing with filters + pagination + sort + assignee name/email join
// -----------------------------------------------------------------------------

// ListAdv returns a page of tickets filtered by multiple fields and sorted.
// - q:        free-text search (title/description, ILIKE)
// - status:   exact
// - priority: exact
// - category: exact
// - assignee: exact
// - sort:     created_at|updated_at|priority (default updated_at)
// - order:    asc|desc (default desc)
// - limit/offset: pagination
func (r *TicketRepo) ListAdv(
	ctx context.Context,
	q, status, priority, category, assignee, sort, order string,
	limit, offset int,
) ([]models.Ticket, error) {

	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	whereSQL, args := buildTicketWhere(q, status, priority, category, assignee)

	sortCol := sanitizeSort(sort, "updated_at")
	sortOrd := sanitizeOrder(order, "desc")

	sql := fmt.Sprintf(`
		SELECT
			t.id, t.alias, t.title, t.description, t.category, t.priority, t.status,
			COALESCE(t.assignee, ''), t.department, t.created_by, t.created_at, t.updated_at,
			COALESCE(u.name, ''), COALESCE(u.email, '')
		FROM tickets t
		LEFT JOIN users u ON u.id = NULLIF(t.assignee, '')::uuid
		%s
		ORDER BY t.%s %s
		LIMIT $%d OFFSET $%d
	`, whereSQL, sortCol, sortOrd, len(args)+1, len(args)+2)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Ticket
	for rows.Next() {
		var t models.Ticket
		if err := rows.Scan(
			&t.ID, &t.Alias, &t.Title, &t.Description, &t.Category, &t.Priority,
			&t.Status, &t.Assignee, &t.Department, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
			&t.AssigneeName, &t.AssigneeEmail,
		); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// CountAdv returns the total number of tickets for the same filter set (for pagination).
func (r *TicketRepo) CountAdv(ctx context.Context, q, status, priority, category, assignee string) (int, error) {
	whereSQL, args := buildTicketWhere(q, status, priority, category, assignee)
	sql := `SELECT COUNT(*) FROM tickets t ` + whereSQL

	var n int
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

// -----------------------------------------------------------------------------
// Single ticket + create/update + comments (Get joined with assignee name/email)
// -----------------------------------------------------------------------------
func (r *TicketRepo) Get(ctx context.Context, id string) (*models.Ticket, error) {
	var t models.Ticket
	err := r.db.QueryRow(ctx, `
		SELECT
			t.id, t.alias, t.title, t.description, t.category, t.priority, t.status,
			t.assignee, t.department, t.created_by, t.created_at, t.updated_at,
			COALESCE(u.name, ''), COALESCE(u.email, '')
		FROM tickets t
		LEFT JOIN users u ON u.id = NULLIF(t.assignee, '')::uuid
		WHERE t.id = $1
	`, id).Scan(
		&t.ID, &t.Alias, &t.Title, &t.Description, &t.Category, &t.Priority,
		&t.Status, &t.Assignee, &t.Department, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
		&t.AssigneeName, &t.AssigneeEmail,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// load comments
	rows, err := r.db.Query(ctx, `
		SELECT id, ticket_id, text, created_at
		FROM comments
		WHERE ticket_id = $1
		ORDER BY created_at ASC
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var c models.Comment
		if err := rows.Scan(&c.ID, &c.TicketID, &c.Text, &c.CreatedAt); err != nil {
			return nil, err
		}
		t.Comments = append(t.Comments, c)
	}
	return &t, nil
}

func (r *TicketRepo) Create(ctx context.Context, t *models.Ticket) error {
	now := time.Now()
	err := r.db.QueryRow(ctx, `
		INSERT INTO tickets (title, description, category, priority, status, assignee, department, created_by, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING id, alias, created_at, updated_at
	`,
		t.Title, t.Description, t.Category, t.Priority, "New", nullIfEmpty(t.Assignee), t.Department, t.CreatedBy, now, now,
	).Scan(&t.ID, &t.Alias, &t.CreatedAt, &t.UpdatedAt)
	return err
}

func (r *TicketRepo) Update(ctx context.Context, t *models.Ticket) error {
	t.UpdatedAt = time.Now()
	ct, err := r.db.Exec(ctx, `
		UPDATE tickets SET
			title=$1, description=$2, category=$3, priority=$4, status=$5, assignee=$6, department=$7, updated_at=$8
		WHERE id=$9
	`,
		t.Title, t.Description, t.Category, t.Priority, t.Status, nullIfEmpty(t.Assignee), t.Department, t.UpdatedAt, t.ID,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *TicketRepo) AddComment(ctx context.Context, ticketID string, text string) (*models.Comment, error) {
	var c models.Comment
	err := r.db.QueryRow(ctx, `
		INSERT INTO comments (ticket_id, text)
		VALUES ($1,$2)
		RETURNING id, ticket_id, text, created_at
	`, ticketID, text).Scan(&c.ID, &c.TicketID, &c.Text, &c.CreatedAt)
	return &c, err
}

// -----------------------------------------------------------------------------
// Reporting helpers (optional, used by /api/reports)
// -----------------------------------------------------------------------------

// CountByStatus counts tickets IN or NOT IN the given statuses.
// If inclusive == true → count IN (statuses); otherwise NOT IN (statuses).
func (r *TicketRepo) CountByStatus(ctx context.Context, statuses []string, inclusive bool) (int, error) {
	op := "NOT IN"
	if inclusive {
		op = "IN"
	}
	sql := `SELECT COUNT(*) FROM tickets WHERE status ` + op + ` (SELECT UNNEST($1::text[]))`
	var n int
	if err := r.db.QueryRow(ctx, sql, statuses).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

// CountResolvedSince counts tickets resolved/closed since the provided time.
func (r *TicketRepo) CountResolvedSince(ctx context.Context, since time.Time) (int, error) {
	sql := `SELECT COUNT(*) FROM tickets WHERE status IN ('Resolved','Closed') AND updated_at >= $1`
	var n int
	if err := r.db.QueryRow(ctx, sql, since).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

// CountOpenByPriorities counts open tickets (not Resolved/Closed) with given priorities.
func (r *TicketRepo) CountOpenByPriorities(ctx context.Context, prios []string) (int, error) {
	sql := `SELECT COUNT(*) FROM tickets WHERE status NOT IN ('Resolved','Closed') AND priority = ANY($1)`
	var n int
	if err := r.db.QueryRow(ctx, sql, prios).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

// buildTicketWhere composes WHERE clause and args for advanced filters (with aliases).
func buildTicketWhere(q, status, priority, category, assignee string) (string, []any) {
	clauses := []string{"1=1"}
	args := []any{}

	// free-text search (ILIKE)
	if s := strings.TrimSpace(q); s != "" {
		p := "%" + s + "%"
		args = append(args, p, p)
		clauses = append(clauses, "(t.title ILIKE $"+itoa(len(args)-1)+" OR t.description ILIKE $"+itoa(len(args))+")")
	}

	// exact filters
	if s := strings.TrimSpace(status); s != "" {
		args = append(args, s)
		clauses = append(clauses, "t.status = $"+itoa(len(args)))
	}
	if p := strings.TrimSpace(priority); p != "" {
		args = append(args, p)
		clauses = append(clauses, "t.priority = $"+itoa(len(args)))
	}
	if c := strings.TrimSpace(category); c != "" {
		args = append(args, c)
		clauses = append(clauses, "t.category = $"+itoa(len(args)))
	}
	if a := strings.TrimSpace(assignee); a != "" {
		args = append(args, a)
		// Cast assignee to UUID for comparison since it's stored as TEXT but represents a UUID
		// Handle NULL assignee values by checking both NULL and UUID cast
		clauses = append(clauses, "(NULLIF(t.assignee,'')::uuid = $"+itoa(len(args))+"::uuid)")
	}

	return "WHERE " + strings.Join(clauses, " AND "), args
}

func sanitizeSort(s, def string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "created_at", "updated_at", "priority":
		return s
	default:
		return def
	}
}

func sanitizeOrder(o, def string) string {
	switch strings.ToLower(strings.TrimSpace(o)) {
	case "asc", "desc":
		return o
	default:
		return def
	}
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

// small helper to avoid fmt for performance-sensitive path.
func itoa(i int) string { return strconv.Itoa(i) }
