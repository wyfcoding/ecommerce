package persistence

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/gateway/domain"
	"gorm.io/gorm"
)

type gatewayRepository struct {
	db *gorm.DB
}

// NewGatewayRepository 创建并返回一个新的 gatewayRepository 实例。
func NewGatewayRepository(db *gorm.DB) domain.GatewayRepository {
	return &gatewayRepository{db: db}
}

// --- Route methods ---

func (r *gatewayRepository) SaveRoute(ctx context.Context, route *domain.Route) error {
	return r.db.WithContext(ctx).Save(route).Error
}

func (r *gatewayRepository) GetRoute(ctx context.Context, id uint64) (*domain.Route, error) {
	var route domain.Route
	if err := r.db.WithContext(ctx).First(&route, id).Error; err != nil {
		return nil, err
	}
	return &route, nil
}

func (r *gatewayRepository) GetRouteByPath(ctx context.Context, path, method string) (*domain.Route, error) {
	var route domain.Route
	if err := r.db.WithContext(ctx).Where("path = ? AND method = ?", path, method).First(&route).Error; err != nil {
		return nil, err
	}
	return &route, nil
}

func (r *gatewayRepository) ListRoutes(ctx context.Context, offset, limit int) ([]*domain.Route, int64, error) {
	var list []*domain.Route
	var total int64
	db := r.db.WithContext(ctx).Model(&domain.Route{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *gatewayRepository) DeleteRoute(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.Route{}, id).Error
}

func (r *gatewayRepository) GetRouteByExternalID(ctx context.Context, externalID string) (*domain.Route, error) {
	var route domain.Route
	err := r.db.WithContext(ctx).Where("external_id = ?", externalID).First(&route).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &route, nil
}

// --- RateLimitRule methods ---

func (r *gatewayRepository) SaveRateLimitRule(ctx context.Context, rule *domain.RateLimitRule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

func (r *gatewayRepository) GetRateLimitRule(ctx context.Context, id uint64) (*domain.RateLimitRule, error) {
	var rule domain.RateLimitRule
	if err := r.db.WithContext(ctx).First(&rule, id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *gatewayRepository) ListRateLimitRules(ctx context.Context, offset, limit int) ([]*domain.RateLimitRule, int64, error) {
	var list []*domain.RateLimitRule
	var total int64
	db := r.db.WithContext(ctx).Model(&domain.RateLimitRule{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *gatewayRepository) DeleteRateLimitRule(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.RateLimitRule{}, id).Error
}

// --- APILog methods ---

func (r *gatewayRepository) SaveAPILog(ctx context.Context, log *domain.APILog) error {
	return r.db.WithContext(ctx).Create(log).Error
}
