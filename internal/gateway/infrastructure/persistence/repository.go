package persistence

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/gateway/domain/entity"     // 导入网关模块的领域实体定义。
	"github.com/wyfcoding/ecommerce/internal/gateway/domain/repository" // 导入网关模块的领域仓储接口。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

type gatewayRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewGatewayRepository 创建并返回一个新的 gatewayRepository 实例。
// db: GORM数据库连接实例。
func NewGatewayRepository(db *gorm.DB) repository.GatewayRepository {
	return &gatewayRepository{db: db}
}

// --- 路由管理 (Route methods) ---

// SaveRoute 将路由实体保存到数据库。
// 如果路由已存在（通过ID），则更新其信息；如果不存在，则创建。
func (r *gatewayRepository) SaveRoute(ctx context.Context, route *entity.Route) error {
	return r.db.WithContext(ctx).Save(route).Error
}

// GetRoute 根据ID从数据库获取路由记录。
// 如果记录未找到，则返回 entity.ErrRouteNotFound 错误。
func (r *gatewayRepository) GetRoute(ctx context.Context, id uint64) (*entity.Route, error) {
	var route entity.Route
	if err := r.db.WithContext(ctx).First(&route, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrRouteNotFound // 返回自定义的“未找到”错误。
		}
		return nil, err
	}
	return &route, nil
}

// GetRouteByPath 根据路径和HTTP方法从数据库获取路由记录。
// 如果记录未找到，则返回 entity.ErrRouteNotFound 错误。
func (r *gatewayRepository) GetRouteByPath(ctx context.Context, path, method string) (*entity.Route, error) {
	var route entity.Route
	if err := r.db.WithContext(ctx).Where("path = ? AND method = ?", path, method).First(&route).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrRouteNotFound // 返回自定义的“未找到”错误。
		}
		return nil, err
	}
	return &route, nil
}

// ListRoutes 从数据库列出所有路由记录，支持分页。
func (r *gatewayRepository) ListRoutes(ctx context.Context, offset, limit int) ([]*entity.Route, int64, error) {
	var list []*entity.Route
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Route{})

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

// DeleteRoute 根据ID从数据库删除路由记录。
func (r *gatewayRepository) DeleteRoute(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.Route{}, id).Error
}

// --- 限流规则管理 (RateLimitRule methods) ---

// SaveRateLimitRule 将限流规则实体保存到数据库。
func (r *gatewayRepository) SaveRateLimitRule(ctx context.Context, rule *entity.RateLimitRule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

// GetRateLimitRule 根据ID从数据库获取限流规则记录。
func (r *gatewayRepository) GetRateLimitRule(ctx context.Context, id uint64) (*entity.RateLimitRule, error) {
	var rule entity.RateLimitRule
	if err := r.db.WithContext(ctx).First(&rule, id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

// ListRateLimitRules 从数据库列出所有限流规则记录，支持分页。
func (r *gatewayRepository) ListRateLimitRules(ctx context.Context, offset, limit int) ([]*entity.RateLimitRule, int64, error) {
	var list []*entity.RateLimitRule
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.RateLimitRule{})

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

// DeleteRateLimitRule 根据ID从数据库删除限流规则记录。
func (r *gatewayRepository) DeleteRateLimitRule(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&entity.RateLimitRule{}, id).Error
}

// --- 日志管理 (APILog methods) ---

// SaveAPILog 将API日志实体保存到数据库。
func (r *gatewayRepository) SaveAPILog(ctx context.Context, log *entity.APILog) error {
	return r.db.WithContext(ctx).Save(log).Error
}
