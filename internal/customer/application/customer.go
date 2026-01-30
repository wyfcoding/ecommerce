package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/customer/domain"
)

// Customer 作为客户服务操作的门面。
type Customer struct {
	Command *CustomerCommandService
	Query   *CustomerQueryService
}

// NewCustomer 创建并返回一个新的 Customer 门面实例。
func NewCustomer(command *CustomerCommandService, query *CustomerQueryService) *Customer {
	return &Customer{
		Command: command,
		Query:   query,
	}
}

// --- 写操作（委托给 Manager）---

// CreateTicket 创建一个新的客服工单。
func (s *Customer) CreateTicket(ctx context.Context, userID uint64, subject, description, category string, priority domain.TicketPriority) (*domain.Ticket, error) {
	return s.Command.CreateTicket(ctx, userID, subject, description, category, priority)
}

// ReplyTicket 为指定工单添加一条新回复。
func (s *Customer) ReplyTicket(ctx context.Context, ticketID, senderID uint64, senderType, content string, msgType domain.MessageType) (*domain.Message, error) {
	return s.Command.ReplyTicket(ctx, ticketID, senderID, senderType, content, msgType)
}

// CloseTicket 关闭指定的客服工单。
func (s *Customer) CloseTicket(ctx context.Context, id uint64) error {
	return s.Command.CloseTicket(ctx, id)
}

// ResolveTicket 将工单状态标记为已解决。
func (s *Customer) ResolveTicket(ctx context.Context, id uint64) error {
	return s.Command.ResolveTicket(ctx, id)
}

// --- 读操作（委托给 Query）---

// GetTicket 获取指定ID的工单详情。
func (s *Customer) GetTicket(ctx context.Context, id uint64) (*domain.Ticket, error) {
	return s.Query.GetTicket(ctx, id)
}

// ListTickets 获取用户的工单列表。
func (s *Customer) ListTickets(ctx context.Context, userID uint64, status domain.TicketStatus, page, pageSize int) ([]*domain.Ticket, int64, error) {
	return s.Query.ListTickets(ctx, userID, status, page, pageSize)
}

// ListMessages 获取指定工单下的所有聊天消息。
func (s *Customer) ListMessages(ctx context.Context, ticketID uint64, page, pageSize int) ([]*domain.Message, int64, error) {
	return s.Query.ListMessages(ctx, ticketID, page, pageSize)
}
