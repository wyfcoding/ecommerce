package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/customer_service/domain"
)

// CustomerService 作为客户服务操作的门面。
type CustomerService struct {
	manager *CustomerServiceManager
	query   *CustomerServiceQuery
}

// NewCustomerService creates a new CustomerService facade.
func NewCustomerService(manager *CustomerServiceManager, query *CustomerServiceQuery) *CustomerService {
	return &CustomerService{
		manager: manager,
		query:   query,
	}
}

// --- 写操作（委托给 Manager）---

// CreateTicket 创建一个新的客服工单。
func (s *CustomerService) CreateTicket(ctx context.Context, userID uint64, subject, description, category string, priority domain.TicketPriority) (*domain.Ticket, error) {
	return s.manager.CreateTicket(ctx, userID, subject, description, category, priority)
}

// ReplyTicket 为指定工单添加一条新回复。
func (s *CustomerService) ReplyTicket(ctx context.Context, ticketID, senderID uint64, senderType, content string, msgType domain.MessageType) (*domain.Message, error) {
	return s.manager.ReplyTicket(ctx, ticketID, senderID, senderType, content, msgType)
}

// CloseTicket 关闭指定的客服工单。
func (s *CustomerService) CloseTicket(ctx context.Context, id uint64) error {
	return s.manager.CloseTicket(ctx, id)
}

// ResolveTicket 将工单状态标记为已解决。
func (s *CustomerService) ResolveTicket(ctx context.Context, id uint64) error {
	return s.manager.ResolveTicket(ctx, id)
}

// --- 读操作（委托给 Query）---

// GetTicket 获取指定ID的工单详情。
func (s *CustomerService) GetTicket(ctx context.Context, id uint64) (*domain.Ticket, error) {
	return s.query.GetTicket(ctx, id)
}

// ListTickets 获取用户的工单列表。
func (s *CustomerService) ListTickets(ctx context.Context, userID uint64, status domain.TicketStatus, page, pageSize int) ([]*domain.Ticket, int64, error) {
	return s.query.ListTickets(ctx, userID, status, page, pageSize)
}

// ListMessages 获取指定工单下的所有聊天消息。
func (s *CustomerService) ListMessages(ctx context.Context, ticketID uint64, page, pageSize int) ([]*domain.Message, int64, error) {
	return s.query.ListMessages(ctx, ticketID, page, pageSize)
}
