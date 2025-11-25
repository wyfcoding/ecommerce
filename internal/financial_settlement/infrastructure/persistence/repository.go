package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/financial_settlement/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/financial_settlement/domain/repository"
	"time"

	"gorm.io/gorm"
)

type settlementRepository struct {
	db *gorm.DB
}

func NewSettlementRepository(db *gorm.DB) repository.SettlementRepository {
	return &settlementRepository{db: db}
}

// 结算管理
func (r *settlementRepository) SaveSettlement(ctx context.Context, settlement *entity.Settlement) error {
	return r.db.WithContext(ctx).Save(settlement).Error
}

func (r *settlementRepository) GetSettlement(ctx context.Context, settlementID uint64) (*entity.Settlement, error) {
	var settlement entity.Settlement
	if err := r.db.WithContext(ctx).First(&settlement, settlementID).Error; err != nil {
		return nil, err
	}
	return &settlement, nil
}

func (r *settlementRepository) ListSellerSettlements(ctx context.Context, sellerID uint64, offset, limit int) ([]*entity.Settlement, int64, error) {
	var list []*entity.Settlement
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Settlement{})
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

func (r *settlementRepository) GetSettlementsByPeriod(ctx context.Context, startDate, endDate time.Time) ([]*entity.Settlement, error) {
	var list []*entity.Settlement
	if err := r.db.WithContext(ctx).Where("start_date >= ? AND end_date <= ?", startDate, endDate).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// 订单管理
func (r *settlementRepository) SaveSettlementOrder(ctx context.Context, order *entity.SettlementOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *settlementRepository) GetSettlementOrders(ctx context.Context, settlementID uint64) ([]*entity.SettlementOrder, error) {
	var list []*entity.SettlementOrder
	if err := r.db.WithContext(ctx).Where("settlement_id = ?", settlementID).Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// 支付管理
func (r *settlementRepository) SaveSettlementPayment(ctx context.Context, payment *entity.SettlementPayment) error {
	return r.db.WithContext(ctx).Save(payment).Error
}

func (r *settlementRepository) GetSettlementPayment(ctx context.Context, paymentID uint64) (*entity.SettlementPayment, error) {
	var payment entity.SettlementPayment
	if err := r.db.WithContext(ctx).First(&payment, paymentID).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *settlementRepository) GetSettlementPaymentBySettlementID(ctx context.Context, settlementID uint64) (*entity.SettlementPayment, error) {
	var payment entity.SettlementPayment
	if err := r.db.WithContext(ctx).Where("settlement_id = ?", settlementID).First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

// 统计
func (r *settlementRepository) GetSettlementStatistics(ctx context.Context, startDate, endDate time.Time) (*entity.SettlementStatistics, error) {
	var stats entity.SettlementStatistics
	stats.StartDate = startDate
	stats.EndDate = endDate

	db := r.db.WithContext(ctx).Model(&entity.Settlement{}).Where("created_at BETWEEN ? AND ?", startDate, endDate)

	if err := db.Count(&stats.TotalSettlements).Error; err != nil {
		return nil, err
	}

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
