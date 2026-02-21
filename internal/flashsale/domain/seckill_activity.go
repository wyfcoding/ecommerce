// 生成摘要：
// - 从 seckill 服务合并到 flashsale 域。
// - 秒杀活动为限时抢购域子聚合，关注极高并发和防超卖。
// - 关键实体：SeckillActivity（秒杀活动聚合根）、SeckillItem（秒杀商品项）。
// - 并发控制策略：乐观锁 (Version) + Redis Lua 预扣库存。
package domain

import (
	"context"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

// 秒杀域业务错误。
var (
	ErrActivityNotStarted = errors.New("activity has not started")
	ErrActivityEnded      = errors.New("activity has ended")
	ErrStockNotEnough     = errors.New("insufficient stock for seckill")
	ErrDuplicatePurchase  = errors.New("user has already purchased this item")
	ErrInvalidLimit       = errors.New("purchase limit exceeded")
)

// SeckillActivityStatus 秒杀活动状态。
type SeckillActivityStatus string

const (
	// SeckillDraft 草稿：活动尚未发布。
	SeckillDraft SeckillActivityStatus = "DRAFT"
	// SeckillUpcoming 即将开始。
	SeckillUpcoming SeckillActivityStatus = "UPCOMING"
	// SeckillRunning 进行中。
	SeckillRunning SeckillActivityStatus = "RUNNING"
	// SeckillEnded 已结束。
	SeckillEnded SeckillActivityStatus = "ENDED"
	// SeckillCancelled 已取消。
	SeckillCancelled SeckillActivityStatus = "CANCELLED"
)

// SeckillActivity 秒杀活动聚合根。
type SeckillActivity struct {
	// ID 活动 ID。
	ID uint64 `json:"id"`
	// Name 活动名称。
	Name string `json:"name"`
	// StartTime 活动开始时间。
	StartTime time.Time `json:"start_time"`
	// EndTime 活动结束时间。
	EndTime time.Time `json:"end_time"`
	// Status 活动状态。
	Status SeckillActivityStatus `json:"status"`
	// Version 乐观锁版本号。
	Version int64 `json:"version"`
	// Items 秒杀商品列表。
	Items []*SeckillItem `json:"items"`
}

// SeckillItem 秒杀商品项。
type SeckillItem struct {
	// ID 秒杀商品 ID。
	ID uint64 `json:"id"`
	// ActivityID 所属活动 ID。
	ActivityID uint64 `json:"activity_id"`
	// ProductID 商品 ID。
	ProductID uint64 `json:"product_id"`
	// SkuID SKU ID。
	SkuID uint64 `json:"sku_id"`
	// SeckillPrice 秒杀价格。
	SeckillPrice decimal.Decimal `json:"seckill_price"`
	// InitialStock 初始库存。
	InitialStock int32 `json:"initial_stock"`
	// AvailableStock 剩余可售库存。
	AvailableStock int32 `json:"available_stock"`
	// PurchaseLimit 单用户限购数量。
	PurchaseLimit int32 `json:"purchase_limit"`
	// Version 乐观锁版本号。
	Version int64 `json:"version"`
}

// ValidatePurchase 验证某次秒杀请求是否合法。
func (a *SeckillActivity) ValidatePurchase(now time.Time, itemID uint64, qty int32) (*SeckillItem, error) {
	if now.Before(a.StartTime) {
		return nil, ErrActivityNotStarted
	}
	if now.After(a.EndTime) {
		return nil, ErrActivityEnded
	}
	if a.Status != SeckillRunning {
		return nil, errors.New("activity is not in running state")
	}
	for _, item := range a.Items {
		if item.ID == itemID {
			if qty > item.PurchaseLimit {
				return nil, ErrInvalidLimit
			}
			return item, nil
		}
	}
	return nil, errors.New("item not found in this activity")
}

// SeckillActivityRepository 秒杀活动仓储。
type SeckillActivityRepository interface {
	GetActiveActivities(ctx context.Context) ([]*SeckillActivity, error)
	GetActivityByID(ctx context.Context, id uint64) (*SeckillActivity, error)
}

// SeckillStockCache 秒杀极速库存管理器（基于 Redis Lua 脚本）。
type SeckillStockCache interface {
	// PreDeductStock 预扣库存，超出抛错，已购用户抛错。
	PreDeductStock(ctx context.Context, activityID, itemID, userID uint64, qty int32) error
	// RollbackStock 订单超时未支付，回滚秒杀库存。
	RollbackStock(ctx context.Context, activityID, itemID, userID uint64, qty int32) error
}
