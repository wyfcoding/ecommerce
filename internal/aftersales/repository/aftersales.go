package repository

import (
	"context"

	"gorm.io/gorm"

	"ecommerce/internal/aftersales/model"
)

// AfterSalesRepo 售后仓储接口
type AfterSalesRepo interface {
	// 退款订单
	CreateRefundOrder(ctx context.Context, refund *model.RefundOrder) error
	GetRefundOrderByID(ctx context.Context, id uint64) (*model.RefundOrder, error)
	GetRefundOrderByNo(ctx context.Context, refundNo string) (*model.RefundOrder, error)
	UpdateRefundOrder(ctx context.Context, refund *model.RefundOrder) error
	ListRefundOrders(ctx context.Context, userID uint64, pageSize, pageNum int32) ([]*model.RefundOrder, int64, error)
	
	// 换货订单
	CreateExchangeOrder(ctx context.Context, exchange *model.ExchangeOrder) error
	GetExchangeOrderByID(ctx context.Context, id uint64) (*model.ExchangeOrder, error)
	GetExchangeOrderByNo(ctx context.Context, exchangeNo string) (*model.ExchangeOrder, error)
	UpdateExchangeOrder(ctx context.Context, exchange *model.ExchangeOrder) error
	ListExchangeOrders(ctx context.Context, userID uint64, pageSize, pageNum int32) ([]*model.ExchangeOrder, int64, error)
	
	// 维修订单
	CreateRepairOrder(ctx context.Context, repair *model.RepairOrder) error
	GetRepairOrderByID(ctx context.Context, id uint64) (*model.RepairOrder, error)
	GetRepairOrderByNo(ctx context.Context, repairNo string) (*model.RepairOrder, error)
	UpdateRepairOrder(ctx context.Context, repair *model.RepairOrder) error
	ListRepairOrders(ctx context.Context, userID uint64, pageSize, pageNum int32) ([]*model.RepairOrder, int64, error)
	
	// 售后工单
	CreateTicket(ctx context.Context, ticket *model.AfterSalesTicket) error
	GetTicketByID(ctx context.Context, id uint64) (*model.AfterSalesTicket, error)
	GetTicketByNo(ctx context.Context, ticketNo string) (*model.AfterSalesTicket, error)
	UpdateTicket(ctx context.Context, ticket *model.AfterSalesTicket) error
	ListTickets(ctx context.Context, userID uint64, pageSize, pageNum int32) ([]*model.AfterSalesTicket, int64, error)
	ListTicketsByAgent(ctx context.Context, agentID uint64, status string, pageSize, pageNum int32) ([]*model.AfterSalesTicket, int64, error)
	
	// 事务支持
	InTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type afterSalesRepoImpl struct {
	db *gorm.DB
}

// NewAfterSalesRepo 创建售后仓储实例
func NewAfterSalesRepo(db *gorm.DB) AfterSalesRepo {
	return &afterSalesRepoImpl{db: db}
}

// CreateRefundOrder 创建退款订单
func (r *afterSalesRepoImpl) CreateRefundOrder(ctx context.Context, refund *model.RefundOrder) error {
	return r.db.WithContext(ctx).Create(refund).Error
}

// GetRefundOrderByID 根据ID获取退款订单
func (r *afterSalesRepoImpl) GetRefundOrderByID(ctx context.Context, id uint64) (*model.RefundOrder, error) {
	var refund model.RefundOrder
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&refund).Error
	if err != nil {
		return nil, err
	}
	return &refund, nil
}

// GetRefundOrderByNo 根据退款单号获取退款订单
func (r *afterSalesRepoImpl) GetRefundOrderByNo(ctx context.Context, refundNo string) (*model.RefundOrder, error) {
	var refund model.RefundOrder
	err := r.db.WithContext(ctx).Where("refund_no = ?", refundNo).First(&refund).Error
	if err != nil {
		return nil, err
	}
	return &refund, nil
}

// UpdateRefundOrder 更新退款订单
func (r *afterSalesRepoImpl) UpdateRefundOrder(ctx context.Context, refund *model.RefundOrder) error {
	return r.db.WithContext(ctx).Save(refund).Error
}

// ListRefundOrders 获取退款订单列表
func (r *afterSalesRepoImpl) ListRefundOrders(ctx context.Context, userID uint64, pageSize, pageNum int32) ([]*model.RefundOrder, int64, error) {
	var refunds []*model.RefundOrder
	var total int64

	db := r.db.WithContext(ctx).Model(&model.RefundOrder{})
	
	if userID > 0 {
		db = db.Where("user_id = ?", userID)
	}
	
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	err := db.Order("created_at DESC").
		Limit(int(pageSize)).
		Offset(int(offset)).
		Find(&refunds).Error
	
	if err != nil {
		return nil, 0, err
	}

	return refunds, total, nil
}

// CreateExchangeOrder 创建换货订单
func (r *afterSalesRepoImpl) CreateExchangeOrder(ctx context.Context, exchange *model.ExchangeOrder) error {
	return r.db.WithContext(ctx).Create(exchange).Error
}

// GetExchangeOrderByID 根据ID获取换货订单
func (r *afterSalesRepoImpl) GetExchangeOrderByID(ctx context.Context, id uint64) (*model.ExchangeOrder, error) {
	var exchange model.ExchangeOrder
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&exchange).Error
	if err != nil {
		return nil, err
	}
	return &exchange, nil
}

// GetExchangeOrderByNo 根据换货单号获取换货订单
func (r *afterSalesRepoImpl) GetExchangeOrderByNo(ctx context.Context, exchangeNo string) (*model.ExchangeOrder, error) {
	var exchange model.ExchangeOrder
	err := r.db.WithContext(ctx).Where("exchange_no = ?", exchangeNo).First(&exchange).Error
	if err != nil {
		return nil, err
	}
	return &exchange, nil
}

// UpdateExchangeOrder 更新换货订单
func (r *afterSalesRepoImpl) UpdateExchangeOrder(ctx context.Context, exchange *model.ExchangeOrder) error {
	return r.db.WithContext(ctx).Save(exchange).Error
}

// ListExchangeOrders 获取换货订单列表
func (r *afterSalesRepoImpl) ListExchangeOrders(ctx context.Context, userID uint64, pageSize, pageNum int32) ([]*model.ExchangeOrder, int64, error) {
	var exchanges []*model.ExchangeOrder
	var total int64

	db := r.db.WithContext(ctx).Model(&model.ExchangeOrder{})
	
	if userID > 0 {
		db = db.Where("user_id = ?", userID)
	}
	
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	err := db.Order("created_at DESC").
		Limit(int(pageSize)).
		Offset(int(offset)).
		Find(&exchanges).Error
	
	if err != nil {
		return nil, 0, err
	}

	return exchanges, total, nil
}

// CreateRepairOrder 创建维修订单
func (r *afterSalesRepoImpl) CreateRepairOrder(ctx context.Context, repair *model.RepairOrder) error {
	return r.db.WithContext(ctx).Create(repair).Error
}

// GetRepairOrderByID 根据ID获取维修订单
func (r *afterSalesRepoImpl) GetRepairOrderByID(ctx context.Context, id uint64) (*model.RepairOrder, error) {
	var repair model.RepairOrder
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&repair).Error
	if err != nil {
		return nil, err
	}
	return &repair, nil
}

// GetRepairOrderByNo 根据维修单号获取维修订单
func (r *afterSalesRepoImpl) GetRepairOrderByNo(ctx context.Context, repairNo string) (*model.RepairOrder, error) {
	var repair model.RepairOrder
	err := r.db.WithContext(ctx).Where("repair_no = ?", repairNo).First(&repair).Error
	if err != nil {
		return nil, err
	}
	return &repair, nil
}

// UpdateRepairOrder 更新维修订单
func (r *afterSalesRepoImpl) UpdateRepairOrder(ctx context.Context, repair *model.RepairOrder) error {
	return r.db.WithContext(ctx).Save(repair).Error
}

// ListRepairOrders 获取维修订单列表
func (r *afterSalesRepoImpl) ListRepairOrders(ctx context.Context, userID uint64, pageSize, pageNum int32) ([]*model.RepairOrder, int64, error) {
	var repairs []*model.RepairOrder
	var total int64

	db := r.db.WithContext(ctx).Model(&model.RepairOrder{})
	
	if userID > 0 {
		db = db.Where("user_id = ?", userID)
	}
	
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	err := db.Order("created_at DESC").
		Limit(int(pageSize)).
		Offset(int(offset)).
		Find(&repairs).Error
	
	if err != nil {
		return nil, 0, err
	}

	return repairs, total, nil
}

// CreateTicket 创建售后工单
func (r *afterSalesRepoImpl) CreateTicket(ctx context.Context, ticket *model.AfterSalesTicket) error {
	return r.db.WithContext(ctx).Create(ticket).Error
}

// GetTicketByID 根据ID获取工单
func (r *afterSalesRepoImpl) GetTicketByID(ctx context.Context, id uint64) (*model.AfterSalesTicket, error) {
	var ticket model.AfterSalesTicket
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&ticket).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

// GetTicketByNo 根据工单号获取工单
func (r *afterSalesRepoImpl) GetTicketByNo(ctx context.Context, ticketNo string) (*model.AfterSalesTicket, error) {
	var ticket model.AfterSalesTicket
	err := r.db.WithContext(ctx).Where("ticket_no = ?", ticketNo).First(&ticket).Error
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

// UpdateTicket 更新工单
func (r *afterSalesRepoImpl) UpdateTicket(ctx context.Context, ticket *model.AfterSalesTicket) error {
	return r.db.WithContext(ctx).Save(ticket).Error
}

// ListTickets 获取工单列表
func (r *afterSalesRepoImpl) ListTickets(ctx context.Context, userID uint64, pageSize, pageNum int32) ([]*model.AfterSalesTicket, int64, error) {
	var tickets []*model.AfterSalesTicket
	var total int64

	db := r.db.WithContext(ctx).Model(&model.AfterSalesTicket{})
	
	if userID > 0 {
		db = db.Where("user_id = ?", userID)
	}
	
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	err := db.Order("created_at DESC").
		Limit(int(pageSize)).
		Offset(int(offset)).
		Find(&tickets).Error
	
	if err != nil {
		return nil, 0, err
	}

	return tickets, total, nil
}

// ListTicketsByAgent 获取客服的工单列表
func (r *afterSalesRepoImpl) ListTicketsByAgent(ctx context.Context, agentID uint64, status string, pageSize, pageNum int32) ([]*model.AfterSalesTicket, int64, error) {
	var tickets []*model.AfterSalesTicket
	var total int64

	db := r.db.WithContext(ctx).Model(&model.AfterSalesTicket{})
	
	if agentID > 0 {
		db = db.Where("agent_id = ?", agentID)
	}
	
	if status != "" {
		db = db.Where("status = ?", status)
	}
	
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	err := db.Order("created_at DESC").
		Limit(int(pageSize)).
		Offset(int(offset)).
		Find(&tickets).Error
	
	if err != nil {
		return nil, 0, err
	}

	return tickets, total, nil
}

// InTx 在事务中执行操作
func (r *afterSalesRepoImpl) InTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, "tx", tx)
		return fn(txCtx)
	})
}
