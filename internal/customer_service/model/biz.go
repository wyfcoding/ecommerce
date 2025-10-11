package biz

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTicketNotFound = errors.New("ticket not found")
)

// Ticket represents a customer service ticket in the business logic layer.
type Ticket struct {
	ID          uint
	TicketID    string
	UserID      uint64
	Subject     string
	Description string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Messages    []*TicketMessage // Populated when getting ticket details
}

// TicketMessage represents a message within a customer service ticket in the business logic layer.
type TicketMessage struct {
	ID         uint
	MessageID  string
	TicketID   string
	SenderID   uint64
	SenderType string
	Content    string
	CreatedAt  time.Time
}

// CustomerServiceRepo defines the interface for customer service data access.
type CustomerServiceRepo interface {
	CreateTicket(ctx context.Context, ticket *Ticket) (*Ticket, error)
	GetTicketByID(ctx context.Context, ticketID string) (*Ticket, error)
	AddTicketMessage(ctx context.Context, message *TicketMessage) (*TicketMessage, error)
	GetTicketMessages(ctx context.Context, ticketID string) ([]*TicketMessage, error)
	ListTicketsByUserID(ctx context.Context, userID uint64, status string) ([]*Ticket, error)
}

// CustomerServiceUsecase is the business logic for customer service.
type CustomerServiceUsecase struct {
	repo CustomerServiceRepo
}

// NewCustomerServiceUsecase creates a new CustomerServiceUsecase.
func NewCustomerServiceUsecase(repo CustomerServiceRepo) *CustomerServiceUsecase {
	return &CustomerServiceUsecase{repo: repo}
}

// CreateTicket creates a new support ticket.
func (uc *CustomerServiceUsecase) CreateTicket(ctx context.Context, userID uint64, subject, description string) (*Ticket, error) {
	ticketID := uuid.New().String()
	ticket := &Ticket{
		TicketID:    ticketID,
		UserID:      userID,
		Subject:     subject,
		Description: description,
		Status:      "OPEN", // Default status
	}
	return uc.repo.CreateTicket(ctx, ticket)
}

// GetTicket retrieves a ticket by its ID, including all messages.
func (uc *CustomerServiceUsecase) GetTicket(ctx context.Context, ticketID string) (*Ticket, error) {
	ticket, err := uc.repo.GetTicketByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if ticket == nil {
		return nil, ErrTicketNotFound
	}

	messages, err := uc.repo.GetTicketMessages(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	ticket.Messages = messages
	return ticket, nil
}

// AddTicketMessage adds a new message to an existing ticket.
func (uc *CustomerServiceUsecase) AddTicketMessage(ctx context.Context, ticketID string, senderID uint64, senderType, content string) (*TicketMessage, error) {
	// Check if ticket exists
	ticket, err := uc.repo.GetTicketByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if ticket == nil {
		return nil, ErrTicketNotFound
	}

	messageID := uuid.New().String()
	message := &TicketMessage{
		MessageID:  messageID,
		TicketID:   ticketID,
		SenderID:   senderID,
		SenderType: senderType,
		Content:    content,
	}
	return uc.repo.AddTicketMessage(ctx, message)
}

// ListTickets lists tickets for a specific user.
func (uc *CustomerServiceUsecase) ListTickets(ctx context.Context, userID uint64, status string) ([]*Ticket, error) {
	return uc.repo.ListTicketsByUserID(ctx, userID, status)
}
