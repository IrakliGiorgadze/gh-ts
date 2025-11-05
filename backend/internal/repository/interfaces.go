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
