package persistence

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/logistics/domain"
	"gorm.io/gorm"
)

type logisticsRepository struct {
	db *gorm.DB
}

// NewLogisticsRepository 创建并返回一个新的 logisticsRepository 实例。
func NewLogisticsRepository(db *gorm.DB) domain.LogisticsRepository {
	return &logisticsRepository{db: db}
}

func (r *logisticsRepository) save(ctx context.Context, db *gorm.DB, logistics *domain.Logistics) error {
	// 保存或更新物流主实体。
	if err := db.Save(logistics).Error; err != nil {
		return err
	}
	// 遍历所有轨迹，只保存新增的轨迹。
	for _, trace := range logistics.Traces {
		if trace.ID == 0 {
			trace.LogisticsID = uint64(logistics.ID)
			if err := db.Save(trace).Error; err != nil {
				return err
			}
		}
	}
	// 保存关联的 DeliveryRoute 实体
	if logistics.Route != nil {
		logistics.Route.LogisticsID = uint64(logistics.ID)
		if err := db.Save(logistics.Route).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *logisticsRepository) Save(ctx context.Context, logistics *domain.Logistics) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return r.save(ctx, tx, logistics)
	})
}

func (r *logisticsRepository) SaveInTx(ctx context.Context, tx any, logistics *domain.Logistics) error {
	gormTx, ok := tx.(*gorm.DB)
	if !ok {
		return r.Save(ctx, logistics)
	}
	return r.save(ctx, gormTx.WithContext(ctx), logistics)
}

func (r *logisticsRepository) GetByID(ctx context.Context, id uint64) (*domain.Logistics, error) {
	var logistics domain.Logistics
	if err := r.db.WithContext(ctx).Preload("Traces").Preload("Route").First(&logistics, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrLogisticsNotFound
		}
		return nil, err
	}
	return &logistics, nil
}

func (r *logisticsRepository) GetByTrackingNo(ctx context.Context, trackingNo string) (*domain.Logistics, error) {
	var logistics domain.Logistics
	if err := r.db.WithContext(ctx).Preload("Traces").Preload("Route").Where("tracking_no = ?", trackingNo).First(&logistics).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrLogisticsNotFound
		}
		return nil, err
	}
	return &logistics, nil
}

func (r *logisticsRepository) GetByOrderID(ctx context.Context, orderID uint64) (*domain.Logistics, error) {
	var logistics domain.Logistics
	if err := r.db.WithContext(ctx).Preload("Traces").Preload("Route").Where("order_id = ?", orderID).First(&logistics).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrLogisticsNotFound
		}
		return nil, err
	}
	return &logistics, nil
}

func (r *logisticsRepository) List(ctx context.Context, offset, limit int) ([]*domain.Logistics, int64, error) {
	var list []*domain.Logistics
	var total int64
	db := r.db.WithContext(ctx).Model(&domain.Logistics{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Preload("Traces").Preload("Route").Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
