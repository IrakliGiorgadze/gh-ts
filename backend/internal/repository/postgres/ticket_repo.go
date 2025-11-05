package postgres

import (
	"context"
	"strconv"
	"strings"
	"time"

	"gh-ts/internal/models"
	"gh-ts/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TicketRepo struct{ db *pgxpool.Pool }

func NewTicketRepo(db *pgxpool.Pool) repository.TicketRepository { return &TicketRepo{db: db} }

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
		args = append(args, "%"+strings.ToLower(q)+"%")
		conds = append(conds, "(lower(title) LIKE $"+itoa(len(args))+" OR lower(description) LIKE $"+itoa(len(args))+" )")
	}
	if status != "" {
		args = append(args, status)
		conds = append(conds, "status = $"+itoa(len(args)))
	}
	args = append(args, limit, offset)

	sql := `SELECT id, title, description, category, priority, status, assignee, department, created_at, updated_at
	        FROM tickets
	        WHERE ` + strings.Join(conds, " AND ") + `
	        ORDER BY updated_at DESC
	        LIMIT $` + itoa(len(args)-1) + ` OFFSET $` + itoa(len(args))

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Ticket
	for rows.Next() {
		var t models.Ticket
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Category, &t.Priority, &t.Status, &t.Assignee, &t.Department, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *TicketRepo) Get(ctx context.Context, id string) (*models.Ticket, error) {
	var t models.Ticket
	err := r.db.QueryRow(ctx, `SELECT id, title, description, category, priority, status, assignee, department, created_at, updated_at
	                           FROM tickets WHERE id = $1`, id).
		Scan(&t.ID, &t.Title, &t.Description, &t.Category, &t.Priority, &t.Status, &t.Assignee, &t.Department, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	// load comments
	rows, err := r.db.Query(ctx, `SELECT id, ticket_id, text, created_at FROM comments WHERE ticket_id=$1 ORDER BY created_at ASC`, id)
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
	err := r.db.QueryRow(ctx, `INSERT INTO tickets (title, description, category, priority, status, assignee, department, created_at, updated_at)
	                           VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id, created_at, updated_at`,
		t.Title, t.Description, t.Category, t.Priority, "New", t.Assignee, t.Department, now, now).
		Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	return err
}

func (r *TicketRepo) Update(ctx context.Context, t *models.Ticket) error {
	t.UpdatedAt = time.Now()
	ct, err := r.db.Exec(ctx, `UPDATE tickets SET 
		title=$1, description=$2, category=$3, priority=$4, status=$5, assignee=$6, department=$7, updated_at=$8
		WHERE id=$9`,
		t.Title, t.Description, t.Category, t.Priority, t.Status, t.Assignee, t.Department, t.UpdatedAt, t.ID)
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
	err := r.db.QueryRow(ctx, `INSERT INTO comments (ticket_id, text) VALUES ($1,$2) RETURNING id, ticket_id, text, created_at`,
		ticketID, text).Scan(&c.ID, &c.TicketID, &c.Text, &c.CreatedAt)
	return &c, err
}

// small helper to avoid fmt for performance-sensitive path; fine here.
func itoa(i int) string { return strconv.Itoa(i) }
