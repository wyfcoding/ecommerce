package data

import (
	"context"
	"ecommerce/internal/customer_service/biz"
	"ecommerce/internal/customer_service/data/model"

	"gorm.io/gorm"
)

type customerServiceRepo struct {
	data *Data
}

// NewCustomerServiceRepo creates a new CustomerServiceRepo.
func NewCustomerServiceRepo(data *Data) biz.CustomerServiceRepo {
	return &customerServiceRepo{data: data}
}

// CreateTicket creates a new support ticket.
func (r *customerServiceRepo) CreateTicket(ctx context.Context, ticket *biz.Ticket) (*biz.Ticket, error) {
	po := &model.Ticket{
		TicketID:    ticket.TicketID,
		UserID:      ticket.UserID,
		Subject:     ticket.Subject,
		Description: ticket.Description,
		Status:      ticket.Status,
	}
	if err := r.data.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	ticket.ID = po.ID
	return ticket, nil
}

// GetTicketByID retrieves a ticket by its ID.
func (r *customerServiceRepo) GetTicketByID(ctx context.Context, ticketID string) (*biz.Ticket, error) {
	var po model.Ticket
	if err := r.data.db.WithContext(ctx).Where("ticket_id = ?", ticketID).First(&po).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Ticket not found
		}
		return nil, err
	}
	return &biz.Ticket{
		ID:          po.ID,
		TicketID:    po.TicketID,
		UserID:      po.UserID,
		Subject:     po.Subject,
		Description: po.Description,
		Status:      po.Status,
		CreatedAt:   po.CreatedAt,
		UpdatedAt:   po.UpdatedAt,
	}, nil
}

// AddTicketMessage adds a new message to a ticket.
func (r *customerServiceRepo) AddTicketMessage(ctx context.Context, message *biz.TicketMessage) (*biz.TicketMessage, error) {
	po := &model.TicketMessage{
		MessageID:  message.MessageID,
		TicketID:   message.TicketID,
		SenderID:   message.SenderID,
		SenderType: message.SenderType,
		Content:    message.Content,
	}
	if err := r.data.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, err
	}
	message.ID = po.ID
	return message, nil
}

// GetTicketMessages retrieves all messages for a given ticket.
func (r *customerServiceRepo) GetTicketMessages(ctx context.Context, ticketID string) ([]*biz.TicketMessage, error) {
	var messages []*model.TicketMessage
	if err := r.data.db.WithContext(ctx).Where("ticket_id = ?", ticketID).Order("created_at ASC").Find(&messages).Error; err != nil {
		return nil, err
	}
	bizMessages := make([]*biz.TicketMessage, len(messages))
	for i, msg := range messages {
		bizMessages[i] = &biz.TicketMessage{
			ID:         msg.ID,
			MessageID:  msg.MessageID,
			TicketID:   msg.TicketID,
			SenderID:   msg.SenderID,
			SenderType: msg.SenderType,
			Content:    msg.Content,
			CreatedAt:  msg.CreatedAt,
		}
	}
	return bizMessages, nil
}

// ListTicketsByUserID lists tickets for a specific user.
func (r *customerServiceRepo) ListTicketsByUserID(ctx context.Context, userID uint64, status string) ([]*biz.Ticket, error) {
	var tickets []*model.Ticket
	query := r.data.db.WithContext(ctx).Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Find(&tickets).Error; err != nil {
		return nil, err
	}
	bizTickets := make([]*biz.Ticket, len(tickets))
	for i, t := range tickets {
		bizTickets[i] = &biz.Ticket{
			ID:          t.ID,
			TicketID:    t.TicketID,
			UserID:      t.UserID,
			Subject:     t.Subject,
			Description: t.Description,
			Status:      t.Status,
			CreatedAt:   t.CreatedAt,
			UpdatedAt:   t.UpdatedAt,
		}
	}
	return bizTickets, nil
}
