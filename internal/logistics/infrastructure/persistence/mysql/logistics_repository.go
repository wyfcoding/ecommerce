package mysql

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/logistics/domain"
	"gorm.io/gorm"
)

type logisticsRepository struct {
	db *gorm.DB
}

// NewLogisticsRepository 创建并返回一个新的物流仓储实例。
func NewLogisticsRepository(db *gorm.DB) domain.LogisticsRepository {
	return &logisticsRepository{db: db}
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
	var model LogisticsModel
	if err := r.db.WithContext(ctx).Preload("Traces").Preload("Route").First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrLogisticsNotFound
		}
		return nil, err
	}
	return toDomainLogistics(&model), nil
}

func (r *logisticsRepository) GetByTrackingNo(ctx context.Context, trackingNo string) (*domain.Logistics, error) {
	var model LogisticsModel
	if err := r.db.WithContext(ctx).Preload("Traces").Preload("Route").Where("tracking_no = ?", trackingNo).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrLogisticsNotFound
		}
		return nil, err
	}
	return toDomainLogistics(&model), nil
}

func (r *logisticsRepository) GetByOrderID(ctx context.Context, orderID uint64) (*domain.Logistics, error) {
	var model LogisticsModel
	if err := r.db.WithContext(ctx).Preload("Traces").Preload("Route").Where("order_id = ?", orderID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrLogisticsNotFound
		}
		return nil, err
	}
	return toDomainLogistics(&model), nil
}

func (r *logisticsRepository) List(ctx context.Context, offset, limit int) ([]*domain.Logistics, int64, error) {
	var list []*LogisticsModel
	var total int64
	db := r.db.WithContext(ctx).Model(&LogisticsModel{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Preload("Traces").Preload("Route").Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}
	result := make([]*domain.Logistics, 0, len(list))
	for _, m := range list {
		result = append(result, toDomainLogistics(m))
	}
	return result, total, nil
}

func (r *logisticsRepository) WithTx(ctx context.Context, fn func(tx any) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

func (r *logisticsRepository) save(ctx context.Context, tx *gorm.DB, logistics *domain.Logistics) error {
	if logistics == nil {
		return nil
	}

	model := toLogisticsModel(logistics)
	if model.ID == 0 {
		if err := tx.WithContext(ctx).Create(model).Error; err != nil {
			return err
		}
	} else {
		if err := tx.WithContext(ctx).Save(model).Error; err != nil {
			return err
		}
	}

	// 保存新增轨迹
	for _, trace := range logistics.Traces {
		if trace == nil || trace.ID != 0 {
			continue
		}
		trace.LogisticsID = uint64(model.ID)
		trace.TrackingNo = model.TrackingNo
		traceModel := toLogisticsTraceModel(trace)
		if err := tx.WithContext(ctx).Create(traceModel).Error; err != nil {
			return err
		}
		if synced := toDomainLogisticsTrace(traceModel); synced != nil {
			*trace = *synced
		}
	}

	// 保存配送路线
	if logistics.Route != nil {
		logistics.Route.LogisticsID = uint64(model.ID)
		routeModel := toDeliveryRouteModel(logistics.Route)
		if routeModel.ID == 0 {
			if err := tx.WithContext(ctx).Create(routeModel).Error; err != nil {
				return err
			}
		} else {
			if err := tx.WithContext(ctx).Save(routeModel).Error; err != nil {
				return err
			}
		}
		if synced := toDomainDeliveryRoute(routeModel); synced != nil {
			*logistics.Route = *synced
		}
	}

	if synced := toDomainLogistics(model); synced != nil {
		logistics.ID = synced.ID
		logistics.CreatedAt = synced.CreatedAt
		logistics.UpdatedAt = synced.UpdatedAt
	}

	return nil
}
