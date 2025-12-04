package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/order/domain/entity" // 导入订单领域的实体定义。
)

// OrderRepository 是订单模块的仓储接口。
// 它定义了对订单实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type OrderRepository interface {
	// Save 将订单实体保存到数据存储中。
	// 如果订单已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// order: 待保存的订单实体。
	Save(ctx context.Context, order *entity.Order) error
	// GetByID 根据ID获取订单实体。
	GetByID(ctx context.Context, id uint64) (*entity.Order, error)
	// GetByOrderNo 根据订单编号获取订单实体。
	GetByOrderNo(ctx context.Context, orderNo string) (*entity.Order, error)
	// List 列出指定用户ID的所有订单实体，支持通过状态过滤和分页。
	List(ctx context.Context, userID uint64, status *entity.OrderStatus, offset, limit int) ([]*entity.Order, int64, error)
}
