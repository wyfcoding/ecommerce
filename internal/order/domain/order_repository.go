package domain

import (
	"context"
)

// OrderRepository 是订单模块的仓储接口。
// 它定义了对订单实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type OrderRepository interface {
	// Save 将订单实体保存到数据存储中。
	// 如果订单已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// order: 待保存的订单实体。
	Save(ctx context.Context, order *Order) error
	WithTx(tx any) OrderRepository
	Transaction(ctx context.Context, userID uint64, fn func(tx any) error) error
	FindByID(ctx context.Context, id uint) (*Order, error)
	FindByOrderNo(ctx context.Context, orderNo string) (*Order, error)
	Update(ctx context.Context, order *Order) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, offset, limit int) ([]*Order, int64, error)
	ListByUserID(ctx context.Context, userID uint, offset, limit int) ([]*Order, int64, error)
}
