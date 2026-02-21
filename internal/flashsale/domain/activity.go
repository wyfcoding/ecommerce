// 变更说明：
// 从 seckill 服务合并多商品秒杀活动模型到 flashsale 统一域。
// seckill 与 flashsale 本质相同（限时限量抢购），合并后消除跨服务冗余调用。
// 新增 FlashsaleActivity 聚合根，支持一个活动包含多个秒杀商品项。
package domain

import (
	"context"
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

// 活动级别错误定义。
var (
	ErrActivityNotFound     = errors.New("flashsale activity not found")
	ErrActivityNotStarted   = errors.New("flashsale activity has not started")
	ErrActivityEnded        = errors.New("flashsale activity has ended")
	ErrActivityItemNotFound = errors.New("flashsale activity item not found")
	ErrDuplicatePurchase    = errors.New("user has already purchased this item")
	ErrPurchaseLimitExceed  = errors.New("purchase limit exceeded")
	ErrStockInsufficient    = errors.New("insufficient stock for flashsale")
)

// ActivityStatus 定义了秒杀活动的生命周期状态。
type ActivityStatus string

const (
	ActivityStatusDraft     ActivityStatus = "DRAFT"     // 草稿：活动已创建但未发布。
	ActivityStatusUpcoming  ActivityStatus = "UPCOMING"  // 即将开始：活动已发布，等待开始。
	ActivityStatusRunning   ActivityStatus = "RUNNING"   // 进行中：活动正在进行。
	ActivityStatusEnded     ActivityStatus = "ENDED"     // 已结束：活动已过结束时间。
	ActivityStatusCancelled ActivityStatus = "CANCELLED" // 已取消：活动被取消。
)

// FlashsaleActivity 多商品秒杀活动聚合根。
// 一个活动可以包含多个秒杀商品项，支持不同商品不同价格和库存。
// 并发控制策略：乐观锁 (Version 字段) + Redis Lua 原子预扣。
type FlashsaleActivity struct {
	ID          uint64          `json:"id"`
	Name        string          `json:"name"`        // 活动名称。
	Description string          `json:"description"` // 活动描述。
	StartTime   time.Time       `json:"start_time"`  // 活动开始时间。
	EndTime     time.Time       `json:"end_time"`    // 活动结束时间。
	Status      ActivityStatus  `json:"status"`      // 活动状态。
	Items       []*ActivityItem `json:"items"`       // 活动商品列表。
	WarmupTime  time.Time       `json:"warmup_time"` // 预热开始时间（允许浏览但不可购买）。
	MaxQPS      int32           `json:"max_qps"`     // 活动级别最大 QPS 限制。
	Version     int64           `json:"version"`     // 乐观锁版本号。
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// ActivityItem 秒杀活动中的单个商品项。
type ActivityItem struct {
	ID             uint64          `json:"id"`
	ActivityID     uint64          `json:"activity_id"`     // 所属活动 ID。
	ProductID      uint64          `json:"product_id"`      // 商品 ID。
	SkuID          uint64          `json:"sku_id"`          // SKU ID。
	OriginalPrice  decimal.Decimal `json:"original_price"`  // 原价。
	FlashsalePrice decimal.Decimal `json:"flashsale_price"` // 秒杀价。
	InitialStock   int32           `json:"initial_stock"`   // 初始库存。
	AvailableStock int32           `json:"available_stock"` // 可用库存。
	SoldCount      int32           `json:"sold_count"`      // 已售数量。
	PurchaseLimit  int32           `json:"purchase_limit"`  // 单用户限购数量（0 表示不限）。
	SortOrder      int32           `json:"sort_order"`      // 排序权重。
	Version        int64           `json:"version"`         // 乐观锁版本号。
}

// NewFlashsaleActivity 创建一个新的多商品秒杀活动。
func NewFlashsaleActivity(name, description string, startTime, endTime time.Time) *FlashsaleActivity {
	return &FlashsaleActivity{
		Name:        name,
		Description: description,
		StartTime:   startTime,
		EndTime:     endTime,
		Status:      ActivityStatusDraft,
		Items:       make([]*ActivityItem, 0),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// AddItem 向活动中添加一个秒杀商品项。
func (a *FlashsaleActivity) AddItem(productID, skuID uint64, originalPrice, flashsalePrice decimal.Decimal, stock, limit int32) *ActivityItem {
	item := &ActivityItem{
		ActivityID:     a.ID,
		ProductID:      productID,
		SkuID:          skuID,
		OriginalPrice:  originalPrice,
		FlashsalePrice: flashsalePrice,
		InitialStock:   stock,
		AvailableStock: stock,
		SoldCount:      0,
		PurchaseLimit:  limit,
		SortOrder:      int32(len(a.Items)),
	}
	a.Items = append(a.Items, item)
	return item
}

// Publish 发布活动，将状态从草稿变更为即将开始。
func (a *FlashsaleActivity) Publish() error {
	if a.Status != ActivityStatusDraft {
		return errors.New("only draft activity can be published")
	}
	if len(a.Items) == 0 {
		return errors.New("activity must have at least one item")
	}
	a.Status = ActivityStatusUpcoming
	a.UpdatedAt = time.Now()
	return nil
}

// Start 启动活动。
func (a *FlashsaleActivity) Start() error {
	if a.Status != ActivityStatusUpcoming {
		return errors.New("only upcoming activity can be started")
	}
	a.Status = ActivityStatusRunning
	a.UpdatedAt = time.Now()
	return nil
}

// End 结束活动。
func (a *FlashsaleActivity) End() {
	a.Status = ActivityStatusEnded
	a.UpdatedAt = time.Now()
}

// Cancel 取消活动。
func (a *FlashsaleActivity) Cancel() {
	a.Status = ActivityStatusCancelled
	a.UpdatedAt = time.Now()
}

// IsInWarmup 判断活动是否处于预热期。
func (a *FlashsaleActivity) IsInWarmup(now time.Time) bool {
	return !a.WarmupTime.IsZero() && now.After(a.WarmupTime) && now.Before(a.StartTime)
}

// ValidatePurchase 验证某次秒杀请求是否合法，返回匹配的商品项。
func (a *FlashsaleActivity) ValidatePurchase(now time.Time, itemID uint64, qty int32) (*ActivityItem, error) {
	if now.Before(a.StartTime) {
		return nil, ErrActivityNotStarted
	}
	if now.After(a.EndTime) {
		return nil, ErrActivityEnded
	}
	if a.Status != ActivityStatusRunning {
		return nil, errors.New("activity is not in running state")
	}

	for _, item := range a.Items {
		if item.ID == itemID {
			if item.PurchaseLimit > 0 && qty > item.PurchaseLimit {
				return nil, ErrPurchaseLimitExceed
			}
			if item.AvailableStock < qty {
				return nil, ErrStockInsufficient
			}
			return item, nil
		}
	}
	return nil, ErrActivityItemNotFound
}

// DeductStock 扣减商品项库存（数据库层面，需配合乐观锁）。
func (item *ActivityItem) DeductStock(qty int32) error {
	if item.AvailableStock < qty {
		return ErrStockInsufficient
	}
	item.AvailableStock -= qty
	item.SoldCount += qty
	return nil
}

// RollbackStock 回滚商品项库存。
func (item *ActivityItem) RollbackStock(qty int32) {
	item.AvailableStock += qty
	item.SoldCount -= qty
	if item.SoldCount < 0 {
		item.SoldCount = 0
	}
}

// RemainingStock 计算剩余库存。
func (item *ActivityItem) RemainingStock() int32 {
	return item.AvailableStock
}

// DiscountRate 计算折扣率。
func (item *ActivityItem) DiscountRate() decimal.Decimal {
	if item.OriginalPrice.IsZero() {
		return decimal.Zero
	}
	return item.FlashsalePrice.Div(item.OriginalPrice).Round(2)
}

// ActivityRepository 多商品秒杀活动仓储接口。
type ActivityRepository interface {
	// SaveActivity 保存活动。
	SaveActivity(ctx context.Context, activity *FlashsaleActivity) error
	// GetActivityByID 根据 ID 获取活动（含商品项）。
	GetActivityByID(ctx context.Context, id uint64) (*FlashsaleActivity, error)
	// ListActiveActivities 获取当前进行中的活动列表。
	ListActiveActivities(ctx context.Context) ([]*FlashsaleActivity, error)
	// ListUpcomingActivities 获取即将开始的活动列表。
	ListUpcomingActivities(ctx context.Context) ([]*FlashsaleActivity, error)
	// UpdateActivityStatus 更新活动状态。
	UpdateActivityStatus(ctx context.Context, id uint64, status ActivityStatus, version int64) error
}

// ActivityStockCache 活动级别的缓存库存管理器。
type ActivityStockCache interface {
	// PreheatActivityStock 预热活动所有商品项的库存到 Redis。
	PreheatActivityStock(ctx context.Context, activity *FlashsaleActivity) error
	// PreDeductStock 原子预扣库存 + 用户限购检查。
	PreDeductStock(ctx context.Context, activityID, itemID, userID uint64, qty, limit int32) error
	// RollbackStock 回滚预扣库存。
	RollbackStock(ctx context.Context, activityID, itemID, userID uint64, qty int32) error
	// GetRemainingStock 获取剩余库存。
	GetRemainingStock(ctx context.Context, activityID, itemID uint64) (int32, error)
}
