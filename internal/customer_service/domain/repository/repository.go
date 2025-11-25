package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/customer_service/domain/entity"
)

// CustomerServiceRepository 客服仓储接口
type CustomerServiceRepository interface {
	// Ticket methods
	SaveTicket(ctx context.Context, ticket *entity.Ticket) error
	GetTicket(ctx context.Context, id uint64) (*entity.Ticket, error)
	GetTicketByNo(ctx context.Context, ticketNo string) (*entity.Ticket, error)
	UpdateTicket(ctx context.Context, ticket *entity.Ticket) error
	ListTickets(ctx context.Context, userID uint64, status entity.TicketStatus, offset, limit int) ([]*entity.Ticket, int64, error)

	// Message methods
	SaveMessage(ctx context.Context, message *entity.Message) error
	ListMessages(ctx context.Context, ticketID uint64, offset, limit int) ([]*entity.Message, int64, error)
}
