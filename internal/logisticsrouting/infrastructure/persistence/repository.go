package persistence

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/logisticsrouting/domain"

	"gorm.io/gorm"
)

type logisticsRoutingRepository struct {
	db *gorm.DB
}

// NewLogisticsRoutingRepository 创建并返回一个新的 logisticsRoutingRepository 实例。
func NewLogisticsRoutingRepository(db *gorm.DB) domain.LogisticsRoutingRepository {
	return &logisticsRoutingRepository{db: db}
}

// --- 配送商 (Carrier methods) ---

func (r *logisticsRoutingRepository) SaveCarrier(ctx context.Context, carrier *domain.Carrier) error {
	return r.db.WithContext(ctx).Save(carrier).Error
}

func (r *logisticsRoutingRepository) GetCarrier(ctx context.Context, id uint64) (*domain.Carrier, error) {
	var carrier domain.Carrier
	if err := r.db.WithContext(ctx).First(&carrier, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &carrier, nil
}

func (r *logisticsRoutingRepository) ListCarriers(ctx context.Context, activeOnly bool) ([]*domain.Carrier, error) {
	var carriers []*domain.Carrier
	db := r.db.WithContext(ctx)
	if activeOnly {
		db = db.Where("is_active = ?", true)
	}
	if err := db.Find(&carriers).Error; err != nil {
		return nil, err
	}
	return carriers, nil
}

// --- 路由 (OptimizedRoute methods) ---

func (r *logisticsRoutingRepository) SaveRoute(ctx context.Context, route *domain.OptimizedRoute) error {
	return r.db.WithContext(ctx).Save(route).Error
}

func (r *logisticsRoutingRepository) GetRoute(ctx context.Context, id uint64) (*domain.OptimizedRoute, error) {
	var route domain.OptimizedRoute
	if err := r.db.WithContext(ctx).First(&route, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &route, nil
}

// --- 统计 (Statistics methods) ---

func (r *logisticsRoutingRepository) SaveStatistics(ctx context.Context, stats *domain.RoutingStatistics) error {
	return r.db.WithContext(ctx).Save(stats).Error
}
