// 变更说明：
// 高性能促销限额缓存接口，利用 Redis Lua 脚本保障分布式锁和原子扣减防超脱。
package domain

import "context"

type PromotionCache interface {
	// --- Query 侧：高速读取缓存 ---
	GetActiveByProduct(ctx context.Context, productID uint64) ([]*Promotion, error)
	SetProductPromotions(ctx context.Context, productID uint64, promos []*Promotion) error

	// --- Command 侧：高并发计数与防超发 ---
	// DeductQuota 执行 Lua 原子扣减（如前1000名优惠名额拦截）。返回 (-1: 全局名额耗尽, -2: 用户超限, 1: 成功)
	DeductQuota(ctx context.Context, promotionID, userID uint64, globalLimit int64, userLimit int32) (int, error)
	// RollbackQuota 订单取消/超时未支付的回退
	RollbackQuota(ctx context.Context, promotionID, userID uint64) error
}
