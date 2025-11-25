package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/customer_service/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/customer_service/domain/repository"

	"gorm.io/gorm"
)

type customerServiceRepository struct {
	db *gorm.DB
}

func NewCustomerServiceRepository(db *gorm.DB) repository.CustomerServiceRepository {
	return &customerServiceRepository{db: db}
}

// Ticket methods
func (r *customerServiceRepository) SaveTicket(ctx context.Context, ticket *entity.Ticket) error {
	return r.db.WithContext(ctx).Save(ticket).Error
}

func (r *customerServiceRepository) GetTicket(ctx context.Context, id uint64) (*entity.Ticket, error) {
	var ticket entity.Ticket
	if err := r.db.WithContext(ctx).First(&ticket, id).Error; err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *customerServiceRepository) GetTicketByNo(ctx context.Context, ticketNo string) (*entity.Ticket, error) {
	var ticket entity.Ticket
	if err := r.db.WithContext(ctx).Where("ticket_no = ?", ticketNo).First(&ticket).Error; err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *customerServiceRepository) UpdateTicket(ctx context.Context, ticket *entity.Ticket) error {
	return r.db.WithContext(ctx).Save(ticket).Error
}

func (r *customerServiceRepository) ListTickets(ctx context.Context, userID uint64, status entity.TicketStatus, offset, limit int) ([]*entity.Ticket, int64, error) {
	var list []*entity.Ticket
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Ticket{})
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

// Message methods
func (r *customerServiceRepository) SaveMessage(ctx context.Context, message *entity.Message) error {
	return r.db.WithContext(ctx).Save(message).Error
}

func (r *customerServiceRepository) ListMessages(ctx context.Context, ticketID uint64, offset, limit int) ([]*entity.Message, int64, error) {
	var list []*entity.Message
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Message{}).Where("ticket_id = ?", ticketID)

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at asc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
