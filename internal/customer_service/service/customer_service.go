package service

import (
	"context"
	"errors"
	"time"

	"ecommerce/internal/customer_service/model"
	"ecommerce/internal/customer_service/repository"
	"github.com/google/uuid"
)

var (
	ErrTicketNotFound = errors.New("ticket not found")
)

// CustomerServiceService is the business logic for customer service.
type CustomerServiceService struct {
	repo repository.CustomerServiceRepo
}

// NewCustomerServiceService creates a new CustomerServiceService.
func NewCustomerServiceService(repo repository.CustomerServiceRepo) *CustomerServiceService {
	return &CustomerServiceService{repo: repo}
}

// CreateTicket creates a new support ticket.
func (s *CustomerServiceService) CreateTicket(ctx context.Context, userID uint64, subject, description string) (*model.Ticket, error) {
	ticketID := uuid.New().String()
	ticket := &model.Ticket{
		TicketID:    ticketID,
		UserID:      userID,
		Subject:     subject,
		Description: description,
		Status:      "OPEN", // Default status
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	return s.repo.CreateTicket(ctx, ticket)
}

// GetTicket retrieves a ticket by its ID, including all messages.
func (s *CustomerServiceService) GetTicket(ctx context.Context, ticketID string) (*model.Ticket, error) {
	ticket, err := s.repo.GetTicketByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if ticket == nil {
		return nil, ErrTicketNotFound
	}

	messages, err := s.repo.GetTicketMessages(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	ticket.Messages = messages
	return ticket, nil
}

// AddTicketMessage adds a new message to an existing ticket.
func (s *CustomerServiceService) AddTicketMessage(ctx context.Context, ticketID string, senderID uint64, senderType, content string) (*model.TicketMessage, error) {
	// Check if ticket exists
	ticket, err := s.repo.GetTicketByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if ticket == nil {
		return nil, ErrTicketNotFound
	}

	messageID := uuid.New().String()
	message := &model.TicketMessage{
		MessageID:  messageID,
		TicketID:   ticketID,
		SenderID:   senderID,
		SenderType: senderType,
		Content:    content,
		CreatedAt:  time.Now(),
	}
	return s.repo.AddTicketMessage(ctx, message)
}

// ListTickets lists tickets for a specific user.
func (s *CustomerServiceService) ListTickets(ctx context.Context, userID uint64, status string) ([]*model.Ticket, error) {
	return s.repo.ListTicketsByUserID(ctx, userID, status)
}
