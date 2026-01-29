package domain

import (
	"context"
)

// LogisticsRepository 是物流模块的仓储接口。
type LogisticsRepository interface {
	// Save 将物流实体保存到数据存储中。
	Save(ctx context.Context, logistics *Logistics) error

	// SaveInTx 在事务中保存（由 OutboxPublisher 需求）
	SaveInTx(ctx context.Context, tx any, logistics *Logistics) error

	// GetByID 根据ID获取物流实体。
	GetByID(ctx context.Context, id uint64) (*Logistics, error)

	// GetByTrackingNo 根据运单号获取物流实体。
	GetByTrackingNo(ctx context.Context, trackingNo string) (*Logistics, error)

	// GetByOrderID 根据订单ID获取物流实体。
	GetByOrderID(ctx context.Context, orderID uint64) (*Logistics, error)

	// List 列出所有物流实体，支持分页。
	List(ctx context.Context, offset, limit int) ([]*Logistics, int64, error)
}
