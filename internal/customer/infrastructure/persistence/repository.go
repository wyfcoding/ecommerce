package persistence

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/customer/domain"

	"gorm.io/gorm"
)

type customerRepository struct {
	db *gorm.DB
}

// NewCustomerRepository 创建并返回一个新的 customerRepository 实例。
func NewCustomerRepository(db *gorm.DB) domain.CustomerRepository {
	return &customerRepository{db: db}
}

// --- Ticket methods ---

func (r *customerRepository) SaveTicket(ctx context.Context, ticket *domain.Ticket) error {
	return r.db.WithContext(ctx).Save(ticket).Error
}

func (r *customerRepository) GetTicket(ctx context.Context, id uint64) (*domain.Ticket, error) {
	var ticket domain.Ticket
	if err := r.db.WithContext(ctx).First(&ticket, id).Error; err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *customerRepository) GetTicketByNo(ctx context.Context, ticketNo string) (*domain.Ticket, error) {
	var ticket domain.Ticket
	if err := r.db.WithContext(ctx).Where("ticket_no = ?", ticketNo).First(&ticket).Error; err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *customerRepository) UpdateTicket(ctx context.Context, ticket *domain.Ticket) error {
	return r.db.WithContext(ctx).Save(ticket).Error
}

func (r *customerRepository) ListTickets(ctx context.Context, userID uint64, status domain.TicketStatus, offset, limit int) ([]*domain.Ticket, int64, error) {
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

func (r *customerRepository) SaveMessage(ctx context.Context, message *domain.Message) error {
	return r.db.WithContext(ctx).Save(message).Error
}

func (r *customerRepository) ListMessages(ctx context.Context, ticketID uint64, offset, limit int) ([]*domain.Message, int64, error) {
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

// GetCustomerSegmentationStats 通过聚合工单数据获取用户分群统计。
func (r *customerRepository) GetCustomerSegmentationStats(ctx context.Context) ([]struct {
	UserID      uint64
	TicketCount float64
	AvgPriority float64
}, error) {
	var results []struct {
		UserID      uint64
		TicketCount float64
		AvgPriority float64
	}

	// 聚合查询：按用户分组，统计工单数和平均优先级
	err := r.db.WithContext(ctx).Table("tickets").
		Select("user_id, COUNT(*) as ticket_count, AVG(priority) as avg_priority").
		Group("user_id").
		Scan(&results).Error

	return results, err
}
