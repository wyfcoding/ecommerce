package domain

import "context"

// LogisticsReadRepository 定义物流读模型的高性能访问接口（Redis）。
type LogisticsReadRepository interface {
	Save(ctx context.Context, logistics *Logistics) error
	GetByID(ctx context.Context, id uint64) (*Logistics, error)
	GetByTrackingNo(ctx context.Context, trackingNo string) (*Logistics, error)
	GetByOrderID(ctx context.Context, orderID uint64) (*Logistics, error)
	Delete(ctx context.Context, id uint64, trackingNo string, orderID uint64) error
}
