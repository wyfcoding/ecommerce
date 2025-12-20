package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/customer_service/domain"
)

// CustomerService acts as a facade for customer service operations.
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

// --- Write Operations (Delegated to Manager) ---

func (s *CustomerService) CreateTicket(ctx context.Context, userID uint64, subject, description, category string, priority domain.TicketPriority) (*domain.Ticket, error) {
	return s.manager.CreateTicket(ctx, userID, subject, description, category, priority)
}

func (s *CustomerService) ReplyTicket(ctx context.Context, ticketID, senderID uint64, senderType, content string, msgType domain.MessageType) (*domain.Message, error) {
	return s.manager.ReplyTicket(ctx, ticketID, senderID, senderType, content, msgType)
}

func (s *CustomerService) CloseTicket(ctx context.Context, id uint64) error {
	return s.manager.CloseTicket(ctx, id)
}

func (s *CustomerService) ResolveTicket(ctx context.Context, id uint64) error {
	return s.manager.ResolveTicket(ctx, id)
}

// --- Read Operations (Delegated to Query) ---

func (s *CustomerService) GetTicket(ctx context.Context, id uint64) (*domain.Ticket, error) {
	return s.query.GetTicket(ctx, id)
}

func (s *CustomerService) ListTickets(ctx context.Context, userID uint64, status domain.TicketStatus, page, pageSize int) ([]*domain.Ticket, int64, error) {
	return s.query.ListTickets(ctx, userID, status, page, pageSize)
}

func (s *CustomerService) ListMessages(ctx context.Context, ticketID uint64, page, pageSize int) ([]*domain.Message, int64, error) {
	return s.query.ListMessages(ctx, ticketID, page, pageSize)
}
