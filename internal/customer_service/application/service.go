package application

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/customer_service/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/customer_service/domain/repository"
	"fmt"
	"time"

	"log/slog"
)

type CustomerService struct {
	repo   repository.CustomerServiceRepository
	logger *slog.Logger
}

func NewCustomerService(repo repository.CustomerServiceRepository, logger *slog.Logger) *CustomerService {
	return &CustomerService{
		repo:   repo,
		logger: logger,
	}
}

// CreateTicket 创建工单
func (s *CustomerService) CreateTicket(ctx context.Context, userID uint64, subject, description, category string, priority entity.TicketPriority) (*entity.Ticket, error) {
	ticketNo := fmt.Sprintf("TKT%d", time.Now().UnixNano())
	ticket := entity.NewTicket(ticketNo, userID, subject, description, category, priority)

	if err := s.repo.SaveTicket(ctx, ticket); err != nil {
		s.logger.Error("failed to create ticket", "error", err)
		return nil, err
	}
	return ticket, nil
}

// ReplyTicket 回复工单
func (s *CustomerService) ReplyTicket(ctx context.Context, ticketID, senderID uint64, senderType, content string, msgType entity.MessageType) (*entity.Message, error) {
	ticket, err := s.repo.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	// Update ticket status if replied by admin/support
	if senderType != "user" && ticket.Status == entity.TicketStatusOpen {
		ticket.Status = entity.TicketStatusInProgress
		if err := s.repo.UpdateTicket(ctx, ticket); err != nil {
			return nil, err
		}
	}

	message := entity.NewMessage(ticketID, senderID, senderType, content, msgType, false)
	if err := s.repo.SaveMessage(ctx, message); err != nil {
		s.logger.Error("failed to save message", "error", err)
		return nil, err
	}

	return message, nil
}

// GetTicket 获取工单详情
func (s *CustomerService) GetTicket(ctx context.Context, id uint64) (*entity.Ticket, error) {
	return s.repo.GetTicket(ctx, id)
}

// ListTickets 获取工单列表
func (s *CustomerService) ListTickets(ctx context.Context, userID uint64, status entity.TicketStatus, page, pageSize int) ([]*entity.Ticket, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListTickets(ctx, userID, status, offset, pageSize)
}

// ListMessages 获取工单消息
func (s *CustomerService) ListMessages(ctx context.Context, ticketID uint64, page, pageSize int) ([]*entity.Message, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListMessages(ctx, ticketID, offset, pageSize)
}

// CloseTicket 关闭工单
func (s *CustomerService) CloseTicket(ctx context.Context, id uint64) error {
	ticket, err := s.repo.GetTicket(ctx, id)
	if err != nil {
		return err
	}

	ticket.Close()
	return s.repo.UpdateTicket(ctx, ticket)
}

// ResolveTicket 解决工单
func (s *CustomerService) ResolveTicket(ctx context.Context, id uint64) error {
	ticket, err := s.repo.GetTicket(ctx, id)
	if err != nil {
		return err
	}

	ticket.Resolve()
	return s.repo.UpdateTicket(ctx, ticket)
}
