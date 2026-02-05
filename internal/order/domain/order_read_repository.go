// 生成摘要：定义订单读模型仓储接口（Redis），用于高频查询。
// 假设：读模型以订单ID与订单号为主键索引。
package domain

import "context"

// OrderReadRepository 定义订单读模型的高性能访问接口。
type OrderReadRepository interface {
	// Save 保存或更新订单读模型。
	Save(ctx context.Context, order *Order) error
	// GetByID 根据订单ID获取读模型。
	GetByID(ctx context.Context, userID uint64, orderID uint64) (*Order, error)
	// GetByOrderNo 根据订单号获取读模型。
	GetByOrderNo(ctx context.Context, userID uint64, orderNo string) (*Order, error)
	// Delete 删除读模型数据（用于清理）。
	Delete(ctx context.Context, userID uint64, orderID uint64, orderNo string) error
}
