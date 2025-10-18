package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"ecommerce/internal/logistics/model"
)

// LogisticsRepository 定义了物流数据仓库的接口
type LogisticsRepository interface {
	CreateShipment(ctx context.Context, shipment *model.Shipment) error
	GetShipmentByOrderSN(ctx context.Context, orderSN string) (*model.Shipment, error)
	GetShipmentByTrackingNumber(ctx context.Context, trackingNumber string) (*model.Shipment, error)
	UpdateShipment(ctx context.Context, shipment *model.Shipment) error
	// AddTrackingEvents 批量添加追踪事件
	AddTrackingEvents(ctx context.Context, events []model.TrackingEvent) error
}

// logisticsRepository 是接口的具体实现
type logisticsRepository struct {
	db *gorm.DB
}

// NewLogisticsRepository 创建一个新的 logisticsRepository 实例
func NewLogisticsRepository(db *gorm.DB) LogisticsRepository {
	return &logisticsRepository{db: db}
}

// CreateShipment 在数据库中创建一条新的货运记录
func (r *logisticsRepository) CreateShipment(ctx context.Context, shipment *model.Shipment) error {
	if err := r.db.WithContext(ctx).Create(shipment).Error; err != nil {
		return fmt.Errorf("数据库创建货运单失败: %w", err)
	}
	return nil
}

// GetShipmentByOrderSN 根据订单号获取货运信息
func (r *logisticsRepository) GetShipmentByOrderSN(ctx context.Context, orderSN string) (*model.Shipment, error) {
	var shipment model.Shipment
	if err := r.db.WithContext(ctx).Preload("TrackingEvents").Where("order_sn = ?", orderSN).First(&shipment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 记录不存在
		}
		return nil, fmt.Errorf("数据库查询货运单失败: %w", err)
	}
	return &shipment, nil
}

// GetShipmentByTrackingNumber 根据运单号获取货运信息
func (r *logisticsRepository) GetShipmentByTrackingNumber(ctx context.Context, trackingNumber string) (*model.Shipment, error) {
	var shipment model.Shipment
	if err := r.db.WithContext(ctx).Preload("TrackingEvents").Where("tracking_number = ?", trackingNumber).First(&shipment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("数据库查询货运单失败: %w", err)
	}
	return &shipment, nil
}

// UpdateShipment 更新货运单信息 (例如，状态、实际成本等)
func (r *logisticsRepository) UpdateShipment(ctx context.Context, shipment *model.Shipment) error {
	if err := r.db.WithContext(ctx).Save(shipment).Error; err != nil {
		return fmt.Errorf("数据库更新货运单失败: %w", err)
	}
	return nil
}

// AddTrackingEvents 批量为货运单添加新的追踪事件
// 使用事务确保所有事件要么全部添加，要么全部失败
func (r *logisticsRepository) AddTrackingEvents(ctx context.Context, events []model.TrackingEvent) error {
	if len(events) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&events).Error; err != nil {
			return fmt.Errorf("数据库批量创建追踪事件失败: %w", err)
		}
		return nil
	})
}
