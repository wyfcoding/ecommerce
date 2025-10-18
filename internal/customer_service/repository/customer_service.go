package repository

import (
	"context"

	"ecommerce/internal/customer_service/model"
)

// CustomerServiceRepo defines the interface for customer service data access.
type CustomerServiceRepo interface {
	CreateTicket(ctx context.Context, ticket *model.Ticket) (*model.Ticket, error)
	GetTicketByID(ctx context.Context, ticketID string) (*model.Ticket, error)
	AddTicketMessage(ctx context.Context, message *model.TicketMessage) (*model.TicketMessage, error)
	GetTicketMessages(ctx context.Context, ticketID string) ([]*model.TicketMessage, error)
	ListTicketsByUserID(ctx context.Context, userID uint64, status string) ([]*model.Ticket, error)
}