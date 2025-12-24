package persistence

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/financialsettlement/domain"

	"gorm.io/gorm"
)

type settlementRepository struct {
	db *gorm.DB
}

// NewSettlementRepository 创建并返回一个新的 settlementRepository 实例.
func NewSettlementRepository(db *gorm.DB) domain.SettlementRepository {
	return &settlementRepository{db: db}
}

// --- 结算管理 (Settlement methods) ---

func (r *settlementRepository) SaveSettlement(ctx context.Context, settlement *domain.Settlement) error {
	return r.db.WithContext(ctx).Save(settlement).Error
}

func (r *settlementRepository) GetSettlement(ctx context.Context, settlementID uint64) (*domain.Settlement, error) {
	var settlement domain.Settlement
	if err := r.db.WithContext(ctx).First(&settlement, settlementID).Error; err != nil {
		return nil, err
	}
	return &settlement, nil
}

func (r *settlementRepository) ListSellerSettlements(ctx context.Context, sellerID uint64, offset, limit int) ([]*domain.Settlement, int64, error) {
	var list []*domain.Settlement
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.Settlement{})
	if sellerID != 0 {
		db = db.Where("seller_id = ?", sellerID)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *settlementRepository) GetSettlementsByPeriod(ctx context.Context, startDate, endDate time.Time) ([]*domain.Settlement, error) {
	var list []*domain.Settlement
	if err := r.db.WithContext(ctx).Where("start_date >= ? AND end_date <= ?", startDate, endDate).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// --- 订单管理 (SettlementOrder methods) ---

func (r *settlementRepository) SaveSettlementOrder(ctx context.Context, order *domain.SettlementOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *settlementRepository) GetSettlementOrders(ctx context.Context, settlementID uint64) ([]*domain.SettlementOrder, error) {
	var list []*domain.SettlementOrder
	if err := r.db.WithContext(ctx).Where("settlement_id = ?", settlementID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// --- 支付管理 (SettlementPayment methods) ---

func (r *settlementRepository) SaveSettlementPayment(ctx context.Context, payment *domain.SettlementPayment) error {
	return r.db.WithContext(ctx).Save(payment).Error
}

func (r *settlementRepository) GetSettlementPayment(ctx context.Context, paymentID uint64) (*domain.SettlementPayment, error) {
	var payment domain.SettlementPayment
	if err := r.db.WithContext(ctx).First(&payment, paymentID).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *settlementRepository) GetSettlementPaymentBySettlementID(ctx context.Context, settlementID uint64) (*domain.SettlementPayment, error) {
	var payment domain.SettlementPayment
	if err := r.db.WithContext(ctx).Where("settlement_id = ?", settlementID).First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

// --- 统计 (Statistics methods) ---

func (r *settlementRepository) GetSettlementStatistics(ctx context.Context, startDate, endDate time.Time) (*domain.SettlementStatistics, error) {
	var stats domain.SettlementStatistics
	stats.StartDate = startDate
	stats.EndDate = endDate

	db := r.db.WithContext(ctx).Model(&domain.Settlement{}).Where("created_at BETWEEN ? AND ?", startDate, endDate)

	if err := db.Count(&stats.TotalSettlements).Error; err != nil {
		return nil, err
	}

	// 结果 结构体定义。
	type Result struct {
		TotalAmount int64
	}
	var result Result
	if err := db.Select("sum(final_amount) as total_amount").Scan(&result).Error; err != nil {
		return nil, err
	}
	stats.TotalAmount = result.TotalAmount

	if stats.TotalSettlements > 0 {
		stats.AverageAmount = float64(stats.TotalAmount) / float64(stats.TotalSettlements)
	}

	if err := db.Where("status = ?", domain.SettlementStatusCompleted).Count(&stats.CompletedCount).Error; err != nil {
		return nil, err
	}
	if err := db.Where("status = ?", domain.SettlementStatusPending).Count(&stats.PendingCount).Error; err != nil {
		return nil, err
	}
	if err := db.Where("status = ?", domain.SettlementStatusRejected).Count(&stats.RejectedCount).Error; err != nil {
		return nil, err
	}

	return &stats, nil
}
