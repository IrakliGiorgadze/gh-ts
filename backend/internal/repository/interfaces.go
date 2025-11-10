package repository

import "context"
import "gh-ts/internal/models"

type TicketRepository interface {
	List(ctx context.Context, q string, status string, limit, offset int) ([]models.Ticket, error)
	Get(ctx context.Context, id string) (*models.Ticket, error)
	Create(ctx context.Context, t *models.Ticket) error
	Update(ctx context.Context, t *models.Ticket) error
	AddComment(ctx context.Context, ticketID string, text string) (*models.Comment, error)
}

type UserRepository interface {
	Create(ctx context.Context, email, name, role, passwordHash string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, string, error)
	GetByID(ctx context.Context, id string) (*models.User, error)

	List(ctx context.Context, q, role string, active *bool, limit, offset int) ([]models.User, int, error)
	UpdateBasic(ctx context.Context, id, name string) (*models.User, error)
	UpdateRole(ctx context.Context, id, role string) (*models.User, error)
	SetActive(ctx context.Context, id string, active bool) (*models.User, error)
	UpdatePasswordHash(ctx context.Context, id, hash string) error
}
