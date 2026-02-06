// 生成摘要：定义购物车读模型仓储接口（Redis），用于高频查询。
// 假设：读模型以 user_id 为主键索引。
package domain

import "context"

// CartReadRepository 定义购物车读模型的高性能访问接口。
type CartReadRepository interface {
	// Save 保存或更新购物车读模型。
	Save(ctx context.Context, cart *Cart) error
	// GetByUserID 根据用户ID获取购物车读模型。
	GetByUserID(ctx context.Context, userID uint64) (*Cart, error)
	// Delete 删除购物车读模型。
	Delete(ctx context.Context, userID uint64) error
}
