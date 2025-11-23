package persistence

import (
	"context"
	"ecommerce/internal/logistics_routing/domain/entity"
	"ecommerce/internal/logistics_routing/domain/repository"
	"errors"

	"gorm.io/gorm"
)

type logisticsRoutingRepository struct {
	db *gorm.DB
}

func NewLogisticsRoutingRepository(db *gorm.DB) repository.LogisticsRoutingRepository {
	return &logisticsRoutingRepository{db: db}
}

// 配送商
func (r *logisticsRoutingRepository) SaveCarrier(ctx context.Context, carrier *entity.Carrier) error {
	return r.db.WithContext(ctx).Save(carrier).Error
}

func (r *logisticsRoutingRepository) GetCarrier(ctx context.Context, id uint64) (*entity.Carrier, error) {
	var carrier entity.Carrier
	if err := r.db.WithContext(ctx).First(&carrier, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &carrier, nil
}

func (r *logisticsRoutingRepository) ListCarriers(ctx context.Context, activeOnly bool) ([]*entity.Carrier, error) {
	var carriers []*entity.Carrier
	db := r.db.WithContext(ctx)
	if activeOnly {
		db = db.Where("is_active = ?", true)
	}
	if err := db.Find(&carriers).Error; err != nil {
		return nil, err
	}
	return carriers, nil
}

// 路由
func (r *logisticsRoutingRepository) SaveRoute(ctx context.Context, route *entity.OptimizedRoute) error {
	return r.db.WithContext(ctx).Save(route).Error
}

func (r *logisticsRoutingRepository) GetRoute(ctx context.Context, id uint64) (*entity.OptimizedRoute, error) {
	var route entity.OptimizedRoute
	if err := r.db.WithContext(ctx).First(&route, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &route, nil
}

// 统计
func (r *logisticsRoutingRepository) SaveStatistics(ctx context.Context, stats *entity.RoutingStatistics) error {
	return r.db.WithContext(ctx).Save(stats).Error
}
