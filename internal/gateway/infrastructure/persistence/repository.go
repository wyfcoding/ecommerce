package persistence

import (
	"context"
	"errors"
	"github.com/wyfcoding/ecommerce/internal/gateway/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/gateway/domain/repository"

	"gorm.io/gorm"
)

type gatewayRepository struct {
	db *gorm.DB
}

func NewGatewayRepository(db *gorm.DB) repository.GatewayRepository {
	return &gatewayRepository{db: db}
}

// 路由管理
func (r *gatewayRepository) SaveRoute(ctx context.Context, route *entity.Route) error {
	return r.db.WithContext(ctx).Save(route).Error
}

func (r *gatewayRepository) GetRoute(ctx context.Context, id uint64) (*entity.Route, error) {
	var route entity.Route
	if err := r.db.WithContext(ctx).First(&route, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrRouteNotFound
		}
		return nil, err
	}
	return &route, nil
}

func (r *gatewayRepository) GetRouteByPath(ctx context.Context, path, method string) (*entity.Route, error) {
	var route entity.Route
	if err := r.db.WithContext(ctx).Where("path = ? AND method = ?", path, method).First(&route).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrRouteNotFound
		}
		return nil, err
	}
	return &route, nil
}

func (r *gatewayRepository) ListRoutes(ctx context.Context, offset, limit int) ([]*entity.Route, int64, error) {
	var list []*entity.Route
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Route{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *gatewayRepository) DeleteRoute(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.Route{}, id).Error
}

// 限流规则管理
func (r *gatewayRepository) SaveRateLimitRule(ctx context.Context, rule *entity.RateLimitRule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

func (r *gatewayRepository) GetRateLimitRule(ctx context.Context, id uint64) (*entity.RateLimitRule, error) {
	var rule entity.RateLimitRule
	if err := r.db.WithContext(ctx).First(&rule, id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *gatewayRepository) ListRateLimitRules(ctx context.Context, offset, limit int) ([]*entity.RateLimitRule, int64, error) {
	var list []*entity.RateLimitRule
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.RateLimitRule{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *gatewayRepository) DeleteRateLimitRule(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.RateLimitRule{}, id).Error
}

// 日志管理
func (r *gatewayRepository) SaveAPILog(ctx context.Context, log *entity.APILog) error {
	return r.db.WithContext(ctx).Save(log).Error
}
