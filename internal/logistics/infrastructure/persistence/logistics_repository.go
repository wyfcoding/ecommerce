package persistence

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/logistics/domain" // 导入物流模块的领域层。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

type logisticsRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewLogisticsRepository 创建并返回一个新的 logisticsRepository 实例。
// db: GORM数据库连接实例。
func NewLogisticsRepository(db *gorm.DB) domain.LogisticsRepository {
	return &logisticsRepository{db: db}
}

// Save 将物流实体保存到数据库。
// 如果物流单已存在（通过ID），则更新其信息；如果不存在，则创建。
// 此方法在一个事务中保存物流主实体及其关联的新增轨迹和配送路线。
func (r *logisticsRepository) Save(ctx context.Context, logistics *domain.Logistics) error {
	// 使用事务确保物流主实体和轨迹的更新操作的原子性。
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 保存或更新物流主实体。
		if err := tx.Save(logistics).Error; err != nil {
			return err
		}
		// 遍历所有轨迹，只保存新增的轨迹（ID为0的轨迹）。
		for _, trace := range logistics.Traces {
			if trace.ID == 0 { // 检查是否是新轨迹。
				trace.LogisticsID = uint64(logistics.ID) // 关联物流ID。
				if err := tx.Save(trace).Error; err != nil {
					return err
				}
			}
		}

		// 保存关联的 DeliveryRoute 实体
		if logistics.Route != nil {
			logistics.Route.LogisticsID = uint64(logistics.ID)
			if err := tx.Save(logistics.Route).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GetByID 根据ID从数据库获取物流记录，并预加载其关联的轨迹和路线。
// 如果记录未找到，则返回 domain.ErrLogisticsNotFound 错误。
func (r *logisticsRepository) GetByID(ctx context.Context, id uint64) (*domain.Logistics, error) {
	var logistics domain.Logistics
	// Preload "Traces" and "Route" 确保在获取物流单时，同时加载所有关联的轨迹记录和路线。
	if err := r.db.WithContext(ctx).Preload("Traces").Preload("Route").First(&logistics, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrLogisticsNotFound // 返回自定义的“未找到”错误。
		}
		return nil, err
	}
	return &logistics, nil
}

// GetByTrackingNo 根据运单号从数据库获取物流记录，并预加载其关联的轨迹和路线。
// 如果记录未找到，则返回 domain.ErrLogisticsNotFound 错误。
func (r *logisticsRepository) GetByTrackingNo(ctx context.Context, trackingNo string) (*domain.Logistics, error) {
	var logistics domain.Logistics
	if err := r.db.WithContext(ctx).Preload("Traces").Preload("Route").Where("tracking_no = ?", trackingNo).First(&logistics).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrLogisticsNotFound // 返回自定义的“未找到”错误。
		}
		return nil, err
	}
	return &logistics, nil
}

// GetByOrderID 根据订单ID从数据库获取物流记录，并预加载其关联的轨迹和路线。
// 如果记录未找到，则返回 domain.ErrLogisticsNotFound 错误。
func (r *logisticsRepository) GetByOrderID(ctx context.Context, orderID uint64) (*domain.Logistics, error) {
	var logistics domain.Logistics
	if err := r.db.WithContext(ctx).Preload("Traces").Preload("Route").Where("order_id = ?", orderID).First(&logistics).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrLogisticsNotFound // 返回自定义的“未找到”错误。
		}
		return nil, err
	}
	return &logistics, nil
}

// List 从数据库列出所有物流记录，支持分页。
func (r *logisticsRepository) List(ctx context.Context, offset, limit int) ([]*domain.Logistics, int64, error) {
	var list []*domain.Logistics
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.Logistics{})

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序，并预加载关联数据
	if err := db.Preload("Traces").Preload("Route").Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
