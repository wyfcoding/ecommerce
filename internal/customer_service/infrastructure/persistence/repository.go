package persistence

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/customer_service/domain"

	"gorm.io/gorm"
)

type customerServiceRepository struct {
	db *gorm.DB
}

// NewCustomerServiceRepository 创建并返回一个新的 customerServiceRepository 实例。
func NewCustomerServiceRepository(db *gorm.DB) domain.CustomerServiceRepository {
	return &customerServiceRepository{db: db}
}

// --- Ticket methods ---

func (r *customerServiceRepository) SaveTicket(ctx context.Context, ticket *domain.Ticket) error {
	return r.db.WithContext(ctx).Save(ticket).Error
}

func (r *customerServiceRepository) GetTicket(ctx context.Context, id uint64) (*domain.Ticket, error) {
	var ticket domain.Ticket
	if err := r.db.WithContext(ctx).First(&ticket, id).Error; err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *customerServiceRepository) GetTicketByNo(ctx context.Context, ticketNo string) (*domain.Ticket, error) {
	var ticket domain.Ticket
	if err := r.db.WithContext(ctx).Where("ticket_no = ?", ticketNo).First(&ticket).Error; err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *customerServiceRepository) UpdateTicket(ctx context.Context, ticket *domain.Ticket) error {
	return r.db.WithContext(ctx).Save(ticket).Error
}

func (r *customerServiceRepository) ListTickets(ctx context.Context, userID uint64, status domain.TicketStatus, offset, limit int) ([]*domain.Ticket, int64, error) {
	var list []*domain.Ticket
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.Ticket{})
	if userID != 0 {
		db = db.Where("user_id = ?", userID)
	}
	if status != 0 {
		db = db.Where("status = ?", status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// --- Message methods ---

func (r *customerServiceRepository) SaveMessage(ctx context.Context, message *domain.Message) error {
	return r.db.WithContext(ctx).Save(message).Error
}

func (r *customerServiceRepository) ListMessages(ctx context.Context, ticketID uint64, offset, limit int) ([]*domain.Message, int64, error) {
	var list []*domain.Message
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.Message{}).Where("ticket_id = ?", ticketID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at asc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
