package persistence

import (
	"context"
	"errors"
	"github.com/wyfcoding/ecommerce/internal/logistics_routing/domain/entity"     // 导入物流路由模块的领域实体定义。
	"github.com/wyfcoding/ecommerce/internal/logistics_routing/domain/repository" // 导入物流路由模块的领域仓储接口。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

type logisticsRoutingRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewLogisticsRoutingRepository 创建并返回一个新的 logisticsRoutingRepository 实例。
// db: GORM数据库连接实例。
func NewLogisticsRoutingRepository(db *gorm.DB) repository.LogisticsRoutingRepository {
	return &logisticsRoutingRepository{db: db}
}

// --- 配送商 (Carrier methods) ---

// SaveCarrier 将配送商实体保存到数据库。
// 如果配送商已存在（通过ID），则更新其信息；如果不存在，则创建。
func (r *logisticsRoutingRepository) SaveCarrier(ctx context.Context, carrier *entity.Carrier) error {
	return r.db.WithContext(ctx).Save(carrier).Error
}

// GetCarrier 根据ID从数据库获取配送商记录。
// 如果记录未找到，则返回nil。
func (r *logisticsRoutingRepository) GetCarrier(ctx context.Context, id uint64) (*entity.Carrier, error) {
	var carrier entity.Carrier
	if err := r.db.WithContext(ctx).First(&carrier, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &carrier, nil
}

// ListCarriers 从数据库列出所有配送商记录。
// activeOnly: 布尔值，如果为true，则只列出活跃的配送商。
func (r *logisticsRoutingRepository) ListCarriers(ctx context.Context, activeOnly bool) ([]*entity.Carrier, error) {
	var carriers []*entity.Carrier
	db := r.db.WithContext(ctx)
	if activeOnly { // 根据activeOnly参数过滤。
		db = db.Where("is_active = ?", true)
	}
	if err := db.Find(&carriers).Error; err != nil {
		return nil, err
	}
	return carriers, nil
}

// --- 路由 (OptimizedRoute methods) ---

// SaveRoute 将优化路由实体保存到数据库。
// 如果路由已存在（通过ID），则更新其信息；如果不存在，则创建。
func (r *logisticsRoutingRepository) SaveRoute(ctx context.Context, route *entity.OptimizedRoute) error {
	return r.db.WithContext(ctx).Save(route).Error
}

// GetRoute 根据ID从数据库获取优化路由记录。
// 如果记录未找到，则返回nil。
func (r *logisticsRoutingRepository) GetRoute(ctx context.Context, id uint64) (*entity.OptimizedRoute, error) {
	var route entity.OptimizedRoute
	if err := r.db.WithContext(ctx).First(&route, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &route, nil
}

// --- 统计 (Statistics methods) ---

// SaveStatistics 将路由统计实体保存到数据库。
func (r *logisticsRoutingRepository) SaveStatistics(ctx context.Context, stats *entity.RoutingStatistics) error {
	return r.db.WithContext(ctx).Save(stats).Error
}
