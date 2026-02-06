package mysql

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/aftersales/domain"
	"gorm.io/gorm"
)

type afterSalesRepository struct {
	db *gorm.DB
}

// NewAfterSalesRepository 创建并返回一个新的 AfterSalesRepository 实例。
func NewAfterSalesRepository(db *gorm.DB) domain.AfterSalesRepository {
	return &afterSalesRepository{db: db}
}

func (r *afterSalesRepository) BeginTx(ctx context.Context) any {
	return r.db.WithContext(ctx).Begin()
}

func (r *afterSalesRepository) CommitTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Commit().Error
}

func (r *afterSalesRepository) RollbackTx(tx any) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return gormTx.Rollback().Error
}

func (r *afterSalesRepository) WithTx(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// --- AfterSales methods ---

func (r *afterSalesRepository) Create(ctx context.Context, afterSales *domain.AfterSales) error {
	return r.createAfterSalesWithTx(ctx, r.db, afterSales)
}

func (r *afterSalesRepository) CreateInTx(ctx context.Context, tx any, afterSales *domain.AfterSales) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.createAfterSalesWithTx(ctx, gormTx, afterSales)
}

func (r *afterSalesRepository) GetByID(ctx context.Context, id uint64) (*domain.AfterSales, error) {
	var model AfterSalesModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	item := toAfterSales(&model)
	items, err := r.listItems(ctx, model.ID)
	if err != nil {
		return nil, err
	}
	logs, err := r.listLogs(ctx, model.ID)
	if err != nil {
		return nil, err
	}
	item.Items = items
	item.Logs = logs
	return item, nil
}

func (r *afterSalesRepository) GetByNo(ctx context.Context, no string) (*domain.AfterSales, error) {
	var model AfterSalesModel
	if err := r.db.WithContext(ctx).Where("after_sales_no = ?", no).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	item := toAfterSales(&model)
	items, err := r.listItems(ctx, model.ID)
	if err != nil {
		return nil, err
	}
	logs, err := r.listLogs(ctx, model.ID)
	if err != nil {
		return nil, err
	}
	item.Items = items
	item.Logs = logs
	return item, nil
}

func (r *afterSalesRepository) Update(ctx context.Context, afterSales *domain.AfterSales) error {
	return r.updateAfterSalesWithTx(ctx, r.db, afterSales)
}

func (r *afterSalesRepository) UpdateInTx(ctx context.Context, tx any, afterSales *domain.AfterSales) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.updateAfterSalesWithTx(ctx, gormTx, afterSales)
}

func (r *afterSalesRepository) List(ctx context.Context, query *domain.AfterSalesQuery) ([]*domain.AfterSales, int64, error) {
	var list []*AfterSalesModel
	var total int64

	db := r.db.WithContext(ctx).Model(&AfterSalesModel{})

	if query.UserID > 0 {
		db = db.Where("user_id = ?", query.UserID)
	}
	if query.OrderID > 0 {
		db = db.Where("order_id = ?", query.OrderID)
	}
	if query.Type > 0 {
		db = db.Where("type = ?", query.Type)
	}
	if query.Status > 0 {
		db = db.Where("status = ?", query.Status)
	}
	if query.AfterSalesNo != "" {
		db = db.Where("after_sales_no = ?", query.AfterSalesNo)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (query.Page - 1) * query.PageSize
	if err := db.Offset(offset).Limit(query.PageSize).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*domain.AfterSales, len(list))
	for i, model := range list {
		item := toAfterSales(model)
		if item != nil {
			item.Items, _ = r.listItems(ctx, model.ID)
		}
		items[i] = item
	}

	return items, total, nil
}

// --- Log methods ---

func (r *afterSalesRepository) CreateLog(ctx context.Context, log *domain.AfterSalesLog) error {
	return r.createLogWithTx(ctx, r.db, log)
}

func (r *afterSalesRepository) CreateLogInTx(ctx context.Context, tx any, log *domain.AfterSalesLog) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.createLogWithTx(ctx, gormTx, log)
}

func (r *afterSalesRepository) ListLogs(ctx context.Context, afterSalesID uint64) ([]*domain.AfterSalesLog, error) {
	return r.listLogs(ctx, uint(afterSalesID))
}

// --- Support Ticket methods ---

func (r *afterSalesRepository) CreateSupportTicket(ctx context.Context, ticket *domain.SupportTicket) error {
	return r.createSupportTicketWithTx(ctx, r.db, ticket)
}

func (r *afterSalesRepository) CreateSupportTicketInTx(ctx context.Context, tx any, ticket *domain.SupportTicket) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.createSupportTicketWithTx(ctx, gormTx, ticket)
}

func (r *afterSalesRepository) GetSupportTicket(ctx context.Context, id uint64) (*domain.SupportTicket, error) {
	var model SupportTicketModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	item := toSupportTicket(&model)
	msgs, err := r.listSupportTicketMessages(ctx, model.ID)
	if err != nil {
		return nil, err
	}
	item.Messages = msgs
	return item, nil
}

func (r *afterSalesRepository) UpdateSupportTicket(ctx context.Context, ticket *domain.SupportTicket) error {
	return r.updateSupportTicketWithTx(ctx, r.db, ticket)
}

func (r *afterSalesRepository) UpdateSupportTicketInTx(ctx context.Context, tx any, ticket *domain.SupportTicket) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.updateSupportTicketWithTx(ctx, gormTx, ticket)
}

func (r *afterSalesRepository) ListSupportTickets(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*domain.SupportTicket, int64, error) {
	var list []*SupportTicketModel
	var total int64

	db := r.db.WithContext(ctx).Model(&SupportTicketModel{})
	if userID > 0 {
		db = db.Where("user_id = ?", userID)
	}
	if status != nil {
		db = db.Where("status = ?", *status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	items := make([]*domain.SupportTicket, len(list))
	for i, model := range list {
		items[i] = toSupportTicket(model)
	}
	return items, total, nil
}

func (r *afterSalesRepository) CreateSupportTicketMessage(ctx context.Context, msg *domain.SupportTicketMessage) error {
	return r.createSupportTicketMessageWithTx(ctx, r.db, msg)
}

func (r *afterSalesRepository) CreateSupportTicketMessageInTx(ctx context.Context, tx any, msg *domain.SupportTicketMessage) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.createSupportTicketMessageWithTx(ctx, gormTx, msg)
}

func (r *afterSalesRepository) GetSupportTicketMessage(ctx context.Context, id uint64) (*domain.SupportTicketMessage, error) {
	var model SupportTicketMessageModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toSupportTicketMessage(&model), nil
}

func (r *afterSalesRepository) ListSupportTicketMessages(ctx context.Context, ticketID uint64) ([]*domain.SupportTicketMessage, error) {
	return r.listSupportTicketMessages(ctx, uint(ticketID))
}

// --- Config methods ---

func (r *afterSalesRepository) GetConfig(ctx context.Context, key string) (*domain.AfterSalesConfig, error) {
	var model AfterSalesConfigModel
	if err := r.db.WithContext(ctx).Where("`key` = ?", key).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toConfig(&model), nil
}

func (r *afterSalesRepository) SetConfig(ctx context.Context, config *domain.AfterSalesConfig) error {
	return r.setConfigWithTx(ctx, r.db, config)
}

func (r *afterSalesRepository) SetConfigInTx(ctx context.Context, tx any, config *domain.AfterSalesConfig) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok || gormTx == nil {
		return errors.New("invalid transaction")
	}
	return r.setConfigWithTx(ctx, gormTx, config)
}

func (r *afterSalesRepository) createAfterSalesWithTx(ctx context.Context, tx *gorm.DB, afterSales *domain.AfterSales) error {
	if afterSales == nil {
		return nil
	}
	model := toAfterSalesModel(afterSales)
	if err := tx.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	afterSales.ID = uint64(model.ID)
	afterSales.CreatedAt = model.CreatedAt
	afterSales.UpdatedAt = model.UpdatedAt
	if len(afterSales.Items) > 0 {
		for _, item := range afterSales.Items {
			if item == nil {
				continue
			}
			item.AfterSalesID = afterSales.ID
			itemModel := toAfterSalesItemModel(item)
			if err := tx.WithContext(ctx).Create(itemModel).Error; err != nil {
				return err
			}
			item.ID = uint64(itemModel.ID)
			item.CreatedAt = itemModel.CreatedAt
			item.UpdatedAt = itemModel.UpdatedAt
		}
	}
	return nil
}

func (r *afterSalesRepository) updateAfterSalesWithTx(ctx context.Context, tx *gorm.DB, afterSales *domain.AfterSales) error {
	if afterSales == nil {
		return nil
	}
	model := toAfterSalesModel(afterSales)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	afterSales.ID = uint64(model.ID)
	afterSales.CreatedAt = model.CreatedAt
	afterSales.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *afterSalesRepository) createLogWithTx(ctx context.Context, tx *gorm.DB, log *domain.AfterSalesLog) error {
	if log == nil {
		return nil
	}
	model := toAfterSalesLogModel(log)
	if err := tx.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	log.ID = uint64(model.ID)
	log.CreatedAt = model.CreatedAt
	log.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *afterSalesRepository) listItems(ctx context.Context, afterSalesID uint) ([]*domain.AfterSalesItem, error) {
	var models []*AfterSalesItemModel
	if err := r.db.WithContext(ctx).Where("after_sales_id = ?", afterSalesID).Order("id asc").Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]*domain.AfterSalesItem, len(models))
	for i, model := range models {
		items[i] = toAfterSalesItem(model)
	}
	return items, nil
}

func (r *afterSalesRepository) listLogs(ctx context.Context, afterSalesID uint) ([]*domain.AfterSalesLog, error) {
	var models []*AfterSalesLogModel
	if err := r.db.WithContext(ctx).Where("after_sales_id = ?", afterSalesID).Order("created_at asc").Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]*domain.AfterSalesLog, len(models))
	for i, model := range models {
		items[i] = toAfterSalesLog(model)
	}
	return items, nil
}

func (r *afterSalesRepository) createSupportTicketWithTx(ctx context.Context, tx *gorm.DB, ticket *domain.SupportTicket) error {
	if ticket == nil {
		return nil
	}
	model := toSupportTicketModel(ticket)
	if err := tx.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	ticket.ID = uint64(model.ID)
	ticket.CreatedAt = model.CreatedAt
	ticket.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *afterSalesRepository) updateSupportTicketWithTx(ctx context.Context, tx *gorm.DB, ticket *domain.SupportTicket) error {
	if ticket == nil {
		return nil
	}
	model := toSupportTicketModel(ticket)
	if err := tx.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}
	ticket.ID = uint64(model.ID)
	ticket.CreatedAt = model.CreatedAt
	ticket.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *afterSalesRepository) createSupportTicketMessageWithTx(ctx context.Context, tx *gorm.DB, msg *domain.SupportTicketMessage) error {
	if msg == nil {
		return nil
	}
	model := toSupportTicketMessageModel(msg)
	if err := tx.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	msg.ID = uint64(model.ID)
	msg.CreatedAt = model.CreatedAt
	msg.UpdatedAt = model.UpdatedAt
	return nil
}

func (r *afterSalesRepository) listSupportTicketMessages(ctx context.Context, ticketID uint) ([]*domain.SupportTicketMessage, error) {
	var models []*SupportTicketMessageModel
	if err := r.db.WithContext(ctx).Where("ticket_id = ?", ticketID).Order("created_at asc").Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]*domain.SupportTicketMessage, len(models))
	for i, model := range models {
		items[i] = toSupportTicketMessage(model)
	}
	return items, nil
}

func (r *afterSalesRepository) setConfigWithTx(ctx context.Context, tx *gorm.DB, config *domain.AfterSalesConfig) error {
	if config == nil {
		return nil
	}
	var existing AfterSalesConfigModel
	err := tx.WithContext(ctx).Where("`key` = ?", config.Key).First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	model := toConfigModel(config)
	if existing.ID != 0 {
		model.ID = existing.ID
		model.CreatedAt = existing.CreatedAt
		if err := tx.WithContext(ctx).Save(model).Error; err != nil {
			return err
		}
	} else {
		if err := tx.WithContext(ctx).Create(model).Error; err != nil {
			return err
		}
	}
	config.ID = uint64(model.ID)
	config.CreatedAt = model.CreatedAt
	config.UpdatedAt = model.UpdatedAt
	return nil
}
