package postgres

import (
	"context"
	"fmt"
	"strings"

	"gh-ts/internal/models"
	"gh-ts/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct{ db *pgxpool.Pool }

func NewUserRepo(db *pgxpool.Pool) repository.UserRepository { return &UserRepo{db: db} }

// Create user (stores bcrypt hash in password_h)
func (r *UserRepo) Create(ctx context.Context, email, name, role, passwordHash string) (*models.User, error) {
	var u models.User
	err := r.db.QueryRow(ctx, `
		INSERT INTO users (email, name, role, password_h)
		VALUES ($1,$2,$3,$4)
		RETURNING id, email, name, role, active, created_at, updated_at`,
		email, name, role, passwordHash).
		Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.Active, &u.CreatedAt, &u.UpdatedAt)
	return &u, err
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, string, error) {
	var u models.User
	var ph string
	err := r.db.QueryRow(ctx, `
		SELECT id, email, name, role, active, password_h, created_at, updated_at
		FROM users WHERE email=$1`, email).
		Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.Active, &ph, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, "", nil
		}
		return nil, "", err
	}
	return &u, ph, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*models.User, error) {
	var u models.User
	err := r.db.QueryRow(ctx, `
		SELECT id, email, name, role, active, created_at, updated_at
		FROM users WHERE id=$1`, id).
		Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.Active, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

// -----------------------------------------------------------------------------
// Admin/list/update operations
// -----------------------------------------------------------------------------

// List returns a filtered, paginated list of users and total count.
// Filters: q (matches email or name, ILIKE), role (exact), active (*bool).
func (r *UserRepo) List(ctx context.Context, q, role string, active *bool, limit, offset int) ([]models.User, int, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	clauses := []string{"1=1"}
	args := []any{}

	if s := strings.TrimSpace(q); s != "" {
		p := "%" + s + "%"
		args = append(args, p, p)
		clauses = append(clauses, "(email ILIKE $"+itoa(len(args)-1)+" OR name ILIKE $"+itoa(len(args))+")")
	}
	if s := strings.TrimSpace(role); s != "" {
		args = append(args, s)
		clauses = append(clauses, "role = $"+itoa(len(args)))
	}
	if active != nil {
		args = append(args, *active)
		clauses = append(clauses, "active = $"+itoa(len(args)))
	}

	// Count
	countSQL := `SELECT COUNT(*) FROM users WHERE ` + strings.Join(clauses, " AND ")
	var total int
	if err := r.db.QueryRow(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Page
	args = append(args, limit, offset)
	listSQL := fmt.Sprintf(`
		SELECT id, email, name, role, active, created_at, updated_at
		FROM users
		WHERE %s
		ORDER BY updated_at DESC
		LIMIT $%d OFFSET $%d
	`, strings.Join(clauses, " AND "), len(args)-1, len(args))
	rows, err := r.db.Query(ctx, listSQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.Active, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, u)
	}
	return out, total, rows.Err()
}

func (r *UserRepo) UpdateRole(ctx context.Context, id, role string) (*models.User, error) {
	var u models.User
	err := r.db.QueryRow(ctx, `
		UPDATE users
		SET role=$1, updated_at=now()
		WHERE id=$2
		RETURNING id, email, name, role, active, created_at, updated_at
	`, role, id).Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.Active, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) SetActive(ctx context.Context, id string, active bool) (*models.User, error) {
	var u models.User
	err := r.db.QueryRow(ctx, `
		UPDATE users
		SET active=$1, updated_at=now()
		WHERE id=$2
		RETURNING id, email, name, role, active, created_at, updated_at
	`, active, id).Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.Active, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) UpdateBasic(ctx context.Context, id, name string) (*models.User, error) {
	var u models.User
	err := r.db.QueryRow(ctx, `
		UPDATE users
		SET name=$1, updated_at=now()
		WHERE id=$2
		RETURNING id, email, name, role, active, created_at, updated_at
	`, name, id).Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.Active, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) UpdatePasswordHash(ctx context.Context, id, passwordHash string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users
		SET password_h=$1, updated_at=now()
		WHERE id=$2
	`, passwordHash, id)
	return err
}

// small helper
//func itoa(i int) string { return fmt.Sprint(i) }
