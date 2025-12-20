package domain

import (
	"context"
)

// CustomerServiceRepository 是客服模块的仓储接口。
type CustomerServiceRepository interface {
	// --- Ticket methods ---
	SaveTicket(ctx context.Context, ticket *Ticket) error
	GetTicket(ctx context.Context, id uint64) (*Ticket, error)
	GetTicketByNo(ctx context.Context, ticketNo string) (*Ticket, error)
	UpdateTicket(ctx context.Context, ticket *Ticket) error
	ListTickets(ctx context.Context, userID uint64, status TicketStatus, offset, limit int) ([]*Ticket, int64, error)

	// --- Message methods ---
	SaveMessage(ctx context.Context, message *Message) error
	ListMessages(ctx context.Context, ticketID uint64, offset, limit int) ([]*Message, int64, error)
}
