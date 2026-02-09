// Package domain 提供了订单领域的先进业务逻辑。
// 变更说明：实现虚拟商品订单逻辑，支持电子卡密（Redeem Code）的自动化发放、有效期管理与核销状态追踪。
package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/wyfcoding/pkg/security"
)

// VirtualItemType 虚拟商品类型
type VirtualItemType string

const (
	VirtualTypeCard    VirtualItemType = "CARD"    // 礼品卡/充值卡
	VirtualTypeCoupon  VirtualItemType = "COUPON"  // 优惠券/代金券
	VirtualTypeDigital VirtualItemType = "DIGITAL" // 数字下载/激活码
)

// VirtualAsset 虚拟资产实体（如具体的卡密）
type VirtualAsset struct {
	AssetID   string          `json:"asset_id"`
	OrderID   uint64          `json:"order_id"`
	Type      VirtualItemType `json:"type"`
	Code      string          `json:"code"`   // 加密存储或加盐哈希的券码
	Secret    string          `json:"secret"` // 密钥/密码
	ValidFrom time.Time       `json:"valid_from"`
	ValidTo   time.Time       `json:"valid_to"`
	IsUsed    bool            `json:"is_used"`
	UsedAt    *time.Time      `json:"used_at"`
	Value     float64         `json:"value"` // 面值或额度
}

// VirtualDelivery 虚拟发货聚合
type VirtualDelivery struct {
	DeliveryID  string          `json:"delivery_id"`
	OrderID     uint64          `json:"order_id"`
	UserID      uint64          `json:"user_id"`
	Assets      []*VirtualAsset `json:"assets"`
	DeliveredAt time.Time       `json:"delivered_at"`
	Status      string          `json:"status"` // PENDING, SENT, FAILED
}

// VirtualOrderService 虚拟订单领域服务
type VirtualOrderService interface {
	// GenerateAssets 为订单生成虚拟资产
	GenerateAssets(ctx context.Context, orderID uint64, itemType VirtualItemType, count int) ([]*VirtualAsset, error)
	// Deliver 发送虚拟资产给用户（如通过邮件/站内信）
	Deliver(ctx context.Context, delivery *VirtualDelivery) error
	// Redeem 核销虚拟资产
	Redeem(ctx context.Context, code string) error
}

// VirtualAssetFactory 虚拟资产工厂
type VirtualAssetFactory struct{}

// CreateCard 创建一个新的虚拟卡密
func (f *VirtualAssetFactory) CreateCard(orderID uint64, val float64, days int) *VirtualAsset {
	// 简单生成高强度券码
	code := security.HashSHA256(fmt.Sprintf("ORD-%d-%d", orderID, time.Now().UnixNano()))[:16]

	return &VirtualAsset{
		AssetID:   fmt.Sprintf("VA-%d", time.Now().UnixNano()),
		OrderID:   orderID,
		Type:      VirtualTypeCard,
		Code:      code,
		ValidFrom: time.Now(),
		ValidTo:   time.Now().AddDate(0, 0, days),
		IsUsed:    false,
		Value:     val,
	}
}

// MarkAsUsed 标记为已使用（领域逻辑）
func (a *VirtualAsset) MarkAsUsed() error {
	if a.IsUsed {
		return errors.New("asset already used")
	}
	if time.Now().After(a.ValidTo) {
		return errors.New("asset expired")
	}

	a.IsUsed = true
	now := time.Now()
	a.UsedAt = &now
	return nil
}
