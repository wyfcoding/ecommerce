package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/financial_settlement/domain/entity"     // 导入财务结算模块的领域实体定义。
	"github.com/wyfcoding/ecommerce/internal/financial_settlement/domain/repository" // 导入财务结算模块的领域仓储接口。
	"time"                                                                           // 导入时间包，用于查询条件。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

// settlementRepository 是 SettlementRepository 接口的GORM实现。
// 它负责将财务结算模块的领域实体映射到数据库，并执行持久化操作。
type settlementRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewSettlementRepository 创建并返回一个新的 settlementRepository 实例。
// db: GORM数据库连接实例。
func NewSettlementRepository(db *gorm.DB) repository.SettlementRepository {
	return &settlementRepository{db: db}
}

// --- 结算管理 (Settlement methods) ---

// SaveSettlement 将结算单实体保存到数据库。
// 如果结算单已存在（通过ID），则更新其信息；如果不存在，则创建。
func (r *settlementRepository) SaveSettlement(ctx context.Context, settlement *entity.Settlement) error {
	return r.db.WithContext(ctx).Save(settlement).Error
}

// GetSettlement 根据ID从数据库获取结算单记录。
func (r *settlementRepository) GetSettlement(ctx context.Context, settlementID uint64) (*entity.Settlement, error) {
	var settlement entity.Settlement
	if err := r.db.WithContext(ctx).First(&settlement, settlementID).Error; err != nil {
		return nil, err
	}
	return &settlement, nil
}

// ListSellerSettlements 从数据库列出指定卖家ID的所有结算单记录，支持分页。
func (r *settlementRepository) ListSellerSettlements(ctx context.Context, sellerID uint64, offset, limit int) ([]*entity.Settlement, int64, error) {
	var list []*entity.Settlement
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Settlement{})
	if sellerID != 0 { // 如果提供了卖家ID，则按卖家ID过滤。
		db = db.Where("seller_id = ?", sellerID)
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

// GetSettlementsByPeriod 获取指定时间范围内的所有结算单记录。
func (r *settlementRepository) GetSettlementsByPeriod(ctx context.Context, startDate, endDate time.Time) ([]*entity.Settlement, error) {
	var list []*entity.Settlement
	// 查询开始日期大于等于startDate且结束日期小于等于endDate的结算单。
	if err := r.db.WithContext(ctx).Where("start_date >= ? AND end_date <= ?", startDate, endDate).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// --- 订单管理 (SettlementOrder methods) ---

// SaveSettlementOrder 将结算订单实体保存到数据库。
func (r *settlementRepository) SaveSettlementOrder(ctx context.Context, order *entity.SettlementOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

// GetSettlementOrders 获取指定结算单ID的所有结算订单记录。
func (r *settlementRepository) GetSettlementOrders(ctx context.Context, settlementID uint64) ([]*entity.SettlementOrder, error) {
	var list []*entity.SettlementOrder
	if err := r.db.WithContext(ctx).Where("settlement_id = ?", settlementID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// --- 支付管理 (SettlementPayment methods) ---

// SaveSettlementPayment 将结算支付实体保存到数据库。
func (r *settlementRepository) SaveSettlementPayment(ctx context.Context, payment *entity.SettlementPayment) error {
	return r.db.WithContext(ctx).Save(payment).Error
}

// GetSettlementPayment 根据支付ID从数据库获取结算支付记录。
func (r *settlementRepository) GetSettlementPayment(ctx context.Context, paymentID uint64) (*entity.SettlementPayment, error) {
	var payment entity.SettlementPayment
	if err := r.db.WithContext(ctx).First(&payment, paymentID).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

// GetSettlementPaymentBySettlementID 根据结算单ID从数据库获取对应的结算支付记录。
func (r *settlementRepository) GetSettlementPaymentBySettlementID(ctx context.Context, settlementID uint64) (*entity.SettlementPayment, error) {
	var payment entity.SettlementPayment
	if err := r.db.WithContext(ctx).Where("settlement_id = ?", settlementID).First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

// --- 统计 (Statistics methods) ---

// GetSettlementStatistics 获取指定时间范围内的结算统计数据。
func (r *settlementRepository) GetSettlementStatistics(ctx context.Context, startDate, endDate time.Time) (*entity.SettlementStatistics, error) {
	var stats entity.SettlementStatistics
	stats.StartDate = startDate
	stats.EndDate = endDate

	// 查询指定时间范围内的结算单。
	db := r.db.WithContext(ctx).Model(&entity.Settlement{}).Where("created_at BETWEEN ? AND ?", startDate, endDate)

	// 统计总结算单数量。
	if err := db.Count(&stats.TotalSettlements).Error; err != nil {
		return nil, err
	}

	// 统计总结算金额。
	type Result struct {
		TotalAmount int64
	}
	var result Result
	if err := db.Select("sum(final_amount) as total_amount").Scan(&result).Error; err != nil {
		return nil, err
	}
	stats.TotalAmount = result.TotalAmount

	// 计算平均结算金额。
	if stats.TotalSettlements > 0 {
		stats.AverageAmount = float64(stats.TotalAmount) / float64(stats.TotalSettlements)
	}

	// 统计不同状态的结算单数量。
	if err := db.Where("status = ?", entity.SettlementStatusCompleted).Count(&stats.CompletedCount).Error; err != nil {
		return nil, err
	}
	if err := db.Where("status = ?", entity.SettlementStatusPending).Count(&stats.PendingCount).Error; err != nil {
		return nil, err
	}
	if err := db.Where("status = ?", entity.SettlementStatusRejected).Count(&stats.RejectedCount).Error; err != nil {
		return nil, err
	}

	return &stats, nil
}
