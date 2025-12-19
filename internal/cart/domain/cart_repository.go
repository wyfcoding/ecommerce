package domain

import (
	"context"
)

// CartRepository 是购物车模块的仓储接口。
// 它定义了对购物车实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type CartRepository interface {
	// Save 将购物车实体保存到数据存储中。
	// 如果购物车已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// cart: 待保存的购物车实体。
	Save(ctx context.Context, cart *Cart) error
	// GetByUserID 根据用户ID从数据存储中获取购物车实体。
	// ctx: 上下文。
	// userID: 用户的唯一标识符。
	GetByUserID(ctx context.Context, userID uint64) (*Cart, error)
	// Delete 根据购物车ID从数据存储中删除购物车实体。
	// ctx: 上下文。
	// id: 购物车的唯一标识符。
	Delete(ctx context.Context, id uint64) error
	// Clear 清空指定购物车ID的所有商品项。
	// ctx: 上下文。
	// cartID: 购物车的唯一标识符。
	Clear(ctx context.Context, cartID uint64) error
}
