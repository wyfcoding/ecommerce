package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/logistics/domain/entity" // 导入物流领域的实体定义。
)

// LogisticsRepository 是物流模块的仓储接口。
// 它定义了对物流单实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type LogisticsRepository interface {
	// Save 将物流实体保存到数据存储中。
	// 如果物流单已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// logistics: 待保存的物流实体。
	Save(ctx context.Context, logistics *entity.Logistics) error
	// GetByID 根据ID获取物流实体。
	GetByID(ctx context.Context, id uint64) (*entity.Logistics, error)
	// GetByTrackingNo 根据运单号获取物流实体。
	GetByTrackingNo(ctx context.Context, trackingNo string) (*entity.Logistics, error)
	// GetByOrderID 根据订单ID获取物流实体。
	GetByOrderID(ctx context.Context, orderID uint64) (*entity.Logistics, error)
	// List 列出所有物流实体，支持分页。
	List(ctx context.Context, offset, limit int) ([]*entity.Logistics, int64, error)
}
