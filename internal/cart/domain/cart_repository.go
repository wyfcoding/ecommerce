package domain

import (
	"context"
)

// CartRepository 是购物车模块的仓储接口。
// 它定义了对购物车实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type CartRepository interface {
	// 事务支持
	BeginTx(ctx context.Context) any
	CommitTx(tx any) error
	RollbackTx(tx any) error
	WithTx(ctx context.Context, fn func(tx any) error) error

	// Save 将购物车实体保存到数据存储中。
	Save(ctx context.Context, cart *Cart) error
	SaveInTx(ctx context.Context, tx any, cart *Cart) error

	// GetByUserID 根据用户ID从数据存储中获取购物车实体。
	GetByUserID(ctx context.Context, userID uint64) (*Cart, error)

	// Delete 根据购物车ID从数据存储中删除购物车实体。
	Delete(ctx context.Context, id uint64) error
	DeleteInTx(ctx context.Context, tx any, id uint64) error

	// Clear 清空指定购物车ID的所有商品项。
	Clear(ctx context.Context, cartID uint64) error
	ClearInTx(ctx context.Context, tx any, cartID uint64) error
}
