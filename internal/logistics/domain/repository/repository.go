package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/logistics/domain/entity"
)

// LogisticsRepository 物流仓储接口
type LogisticsRepository interface {
	Save(ctx context.Context, logistics *entity.Logistics) error
	GetByID(ctx context.Context, id uint64) (*entity.Logistics, error)
	GetByTrackingNo(ctx context.Context, trackingNo string) (*entity.Logistics, error)
	GetByOrderID(ctx context.Context, orderID uint64) (*entity.Logistics, error)
	List(ctx context.Context, offset, limit int) ([]*entity.Logistics, int64, error)
}
