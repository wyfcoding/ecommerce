package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/logistics/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/logistics/domain/repository"
	"errors"

	"gorm.io/gorm"
)

type logisticsRepository struct {
	db *gorm.DB
}

func NewLogisticsRepository(db *gorm.DB) repository.LogisticsRepository {
	return &logisticsRepository{db: db}
}

func (r *logisticsRepository) Save(ctx context.Context, logistics *entity.Logistics) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(logistics).Error; err != nil {
			return err
		}
		for _, trace := range logistics.Traces {
			if trace.ID == 0 { // Only save new traces
				trace.LogisticsID = uint64(logistics.ID)
				if err := tx.Save(trace).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (r *logisticsRepository) GetByID(ctx context.Context, id uint64) (*entity.Logistics, error) {
	var logistics entity.Logistics
	if err := r.db.WithContext(ctx).Preload("Traces").First(&logistics, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrLogisticsNotFound
		}
		return nil, err
	}
	return &logistics, nil
}

func (r *logisticsRepository) GetByTrackingNo(ctx context.Context, trackingNo string) (*entity.Logistics, error) {
	var logistics entity.Logistics
	if err := r.db.WithContext(ctx).Preload("Traces").Where("tracking_no = ?", trackingNo).First(&logistics).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrLogisticsNotFound
		}
		return nil, err
	}
	return &logistics, nil
}

func (r *logisticsRepository) GetByOrderID(ctx context.Context, orderID uint64) (*entity.Logistics, error) {
	var logistics entity.Logistics
	if err := r.db.WithContext(ctx).Preload("Traces").Where("order_id = ?", orderID).First(&logistics).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrLogisticsNotFound
		}
		return nil, err
	}
	return &logistics, nil
}

func (r *logisticsRepository) List(ctx context.Context, offset, limit int) ([]*entity.Logistics, int64, error) {
	var list []*entity.Logistics
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.Logistics{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
