package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/customer/domain"
)

// CustomerManager 处理客户服务的写操作。
type CustomerManager struct {
	repo   domain.CustomerRepository
	logger *slog.Logger
}

// NewCustomerManager 创建并返回一个新的 CustomerManager 实例。
func NewCustomerManager(repo domain.CustomerRepository, logger *slog.Logger) *CustomerManager {
	return &CustomerManager{
		repo:   repo,
		logger: logger,
	}
}

// CreateTicket 创建一个新的客户服务工单。
func (m *CustomerManager) CreateTicket(ctx context.Context, userID uint64, subject, description, category string, priority domain.TicketPriority) (*domain.Ticket, error) {
	ticketNo := fmt.Sprintf("TKT%d", time.Now().UnixNano())
	ticket := domain.NewTicket(ticketNo, userID, subject, description, category, priority)

	if err := m.repo.SaveTicket(ctx, ticket); err != nil {
		m.logger.ErrorContext(ctx, "failed to create ticket", "user_id", userID, "subject", subject, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "ticket created successfully", "ticket_id", ticket.ID, "ticket_no", ticketNo)
	return ticket, nil
}

// ReplyTicket 回复一个工单。
func (m *CustomerManager) ReplyTicket(ctx context.Context, ticketID, senderID uint64, senderType, content string, msgType domain.MessageType) (*domain.Message, error) {
	ticket, err := m.repo.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	if senderType != "user" && ticket.Status == domain.TicketStatusOpen {
		ticket.Status = domain.TicketStatusInProgress
		if err := m.repo.UpdateTicket(ctx, ticket); err != nil {
			return nil, err
		}
	}

	message := domain.NewMessage(ticketID, senderID, senderType, content, msgType, false)
	if err := m.repo.SaveMessage(ctx, message); err != nil {
		m.logger.ErrorContext(ctx, "failed to save message", "ticket_id", ticketID, "sender_id", senderID, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "message saved successfully", "message_id", message.ID, "ticket_id", ticketID)

	return message, nil
}

// CloseTicket 关闭一个工单。
func (m *CustomerManager) CloseTicket(ctx context.Context, id uint64) error {
	ticket, err := m.repo.GetTicket(ctx, id)
	if err != nil {
		return err
	}

	ticket.Close()
	return m.repo.UpdateTicket(ctx, ticket)
}

// ResolveTicket 解决一个工单。
func (m *CustomerManager) ResolveTicket(ctx context.Context, id uint64) error {
	ticket, err := m.repo.GetTicket(ctx, id)
	if err != nil {
		return err
	}

	ticket.Resolve()
	return m.repo.UpdateTicket(ctx, ticket)
}
