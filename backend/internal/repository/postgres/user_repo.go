package postgres

import (
	"context"

	"gh-ts/internal/models"
	"gh-ts/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct{ db *pgxpool.Pool }

func NewUserRepo(db *pgxpool.Pool) repository.UserRepository { return &UserRepo{db: db} }

func (r *UserRepo) Create(ctx context.Context, email, name, role, passwordHash string) (*models.User, error) {
	var u models.User
	err := r.db.QueryRow(ctx, `
		INSERT INTO users (email, name, role, password_h)
		VALUES ($1,$2,$3,$4)
		RETURNING id, email, name, role, created_at, updated_at`,
		email, name, role, passwordHash).
		Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	return &u, err
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, string, error) {
	var u models.User
	var ph string
	err := r.db.QueryRow(ctx, `
		SELECT id, email, name, role, password_h, created_at, updated_at
		FROM users WHERE email=$1`, email).
		Scan(&u.ID, &u.Email, &u.Name, &u.Role, &ph, &u.CreatedAt, &u.UpdatedAt)
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
		SELECT id, email, name, role, created_at, updated_at
		FROM users WHERE id=$1`, id).
		Scan(&u.ID, &u.Email, &u.Name, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}
