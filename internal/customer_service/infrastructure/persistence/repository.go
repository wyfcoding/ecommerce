package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/customer_service/domain/entity"     // 导入客户服务模块的领域实体定义。
	"github.com/wyfcoding/ecommerce/internal/customer_service/domain/repository" // 导入客户服务模块的领域仓储接口。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

// customerServiceRepository 是 CustomerServiceRepository 接口的GORM实现。
// 它负责将客户服务模块的领域实体映射到数据库，并执行持久化操作。
type customerServiceRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewCustomerServiceRepository 创建并返回一个新的 customerServiceRepository 实例。
// db: GORM数据库连接实例。
func NewCustomerServiceRepository(db *gorm.DB) repository.CustomerServiceRepository {
	return &customerServiceRepository{db: db}
}

// --- Ticket methods ---

// SaveTicket 将工单实体保存到数据库。
// 如果工单已存在（通过ID），则更新其信息；如果不存在，则创建。
func (r *customerServiceRepository) SaveTicket(ctx context.Context, ticket *entity.Ticket) error {
	return r.db.WithContext(ctx).Save(ticket).Error
}

// GetTicket 根据ID从数据库获取工单记录。
func (r *customerServiceRepository) GetTicket(ctx context.Context, id uint64) (*entity.Ticket, error) {
	var ticket entity.Ticket
	if err := r.db.WithContext(ctx).First(&ticket, id).Error; err != nil {
		return nil, err
	}
	return &ticket, nil
}

// GetTicketByNo 根据工单编号从数据库获取工单记录。
func (r *customerServiceRepository) GetTicketByNo(ctx context.Context, ticketNo string) (*entity.Ticket, error) {
	var ticket entity.Ticket
	if err := r.db.WithContext(ctx).Where("ticket_no = ?", ticketNo).First(&ticket).Error; err != nil {
		return nil, err
	}
	return &ticket, nil
}

// UpdateTicket 更新数据库中的工单记录。
func (r *customerServiceRepository) UpdateTicket(ctx context.Context, ticket *entity.Ticket) error {
	return r.db.WithContext(ctx).Save(ticket).Error
}

// ListTickets 从数据库列出所有工单记录，支持通过用户ID和状态过滤，并支持分页。
func (r *customerServiceRepository) ListTickets(ctx context.Context, userID uint64, status entity.TicketStatus, offset, limit int) ([]*entity.Ticket, int64, error) {
	var list []*entity.Ticket
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Ticket{})
	if userID != 0 { // 如果提供了用户ID，则按用户ID过滤。
		db = db.Where("user_id = ?", userID)
	}
	if status != 0 { // 如果提供了状态，则按状态过滤。
		db = db.Where("status = ?", status)
	}

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// --- Message methods ---

// SaveMessage 将消息实体保存到数据库。
func (r *customerServiceRepository) SaveMessage(ctx context.Context, message *entity.Message) error {
	return r.db.WithContext(ctx).Save(message).Error
}

// ListMessages 从数据库列出指定工单的所有消息记录，支持分页。
func (r *customerServiceRepository) ListMessages(ctx context.Context, ticketID uint64, offset, limit int) ([]*entity.Message, int64, error) {
	var list []*entity.Message
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Message{}).Where("ticket_id = ?", ticketID)

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	if err := db.Offset(offset).Limit(limit).Order("created_at asc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
