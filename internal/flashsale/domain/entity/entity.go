package entity

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrFlashsaleNotFound   = errors.New("秒杀活动不存在")
	ErrFlashsaleNotStarted = errors.New("秒杀活动未开始")
	ErrFlashsaleEnded      = errors.New("秒杀活动已结束")
	ErrFlashsaleSoldOut    = errors.New("秒杀商品已售罄")
	ErrFlashsaleLimit      = errors.New("超过购买限制")
)

// FlashsaleStatus 秒杀状态
type FlashsaleStatus int8

const (
	FlashsaleStatusPending  FlashsaleStatus = 0 // 未开始
	FlashsaleStatusOngoing  FlashsaleStatus = 1 // 进行中
	FlashsaleStatusEnded    FlashsaleStatus = 2 // 已结束
	FlashsaleStatusCanceled FlashsaleStatus = 3 // 已取消
)

// Flashsale 秒杀活动实体
type Flashsale struct {
	gorm.Model
	Name          string          `gorm:"type:varchar(255);not null;comment:活动名称" json:"name"`
	ProductID     uint64          `gorm:"not null;index;comment:商品ID" json:"product_id"`
	SkuID         uint64          `gorm:"not null;index;comment:SKU ID" json:"sku_id"`
	OriginalPrice int64           `gorm:"not null;comment:原价" json:"original_price"`
	FlashPrice    int64           `gorm:"not null;comment:秒杀价" json:"flash_price"`
	TotalStock    int32           `gorm:"not null;comment:总库存" json:"total_stock"`
	SoldCount     int32           `gorm:"not null;default:0;comment:已售数量" json:"sold_count"`
	LimitPerUser  int32           `gorm:"not null;default:1;comment:每人限购数量" json:"limit_per_user"`
	StartTime     time.Time       `gorm:"not null;comment:开始时间" json:"start_time"`
	EndTime       time.Time       `gorm:"not null;comment:结束时间" json:"end_time"`
	Status        FlashsaleStatus `gorm:"default:0;comment:状态" json:"status"`
	Description   string          `gorm:"type:text;comment:描述" json:"description"`
}

// NewFlashsale 创建秒杀活动
func NewFlashsale(name string, productID, skuID uint64, originalPrice, flashPrice int64, totalStock, limitPerUser int32, startTime, endTime time.Time) *Flashsale {
	return &Flashsale{
		Name:          name,
		ProductID:     productID,
		SkuID:         skuID,
		OriginalPrice: originalPrice,
		FlashPrice:    flashPrice,
		TotalStock:    totalStock,
		SoldCount:     0,
		LimitPerUser:  limitPerUser,
		StartTime:     startTime,
		EndTime:       endTime,
		Status:        FlashsaleStatusPending,
	}
}

// RemainingStock 剩余库存
func (f *Flashsale) RemainingStock() int32 {
	return f.TotalStock - f.SoldCount
}

// IsAvailable 是否可用
func (f *Flashsale) IsAvailable() bool {
	now := time.Now()
	return f.Status == FlashsaleStatusOngoing &&
		now.After(f.StartTime) &&
		now.Before(f.EndTime) &&
		f.SoldCount < f.TotalStock
}

// CanBuy 是否可以购买
func (f *Flashsale) CanBuy(userBoughtCount int32, quantity int32) error {
	now := time.Now()

	if f.Status != FlashsaleStatusOngoing {
		if now.Before(f.StartTime) {
			return ErrFlashsaleNotStarted
		}
		return ErrFlashsaleEnded
	}

	if now.Before(f.StartTime) {
		return ErrFlashsaleNotStarted
	}

	if now.After(f.EndTime) {
		return ErrFlashsaleEnded
	}

	if f.SoldCount >= f.TotalStock {
		return ErrFlashsaleSoldOut
	}

	if userBoughtCount+quantity > f.LimitPerUser {
		return ErrFlashsaleLimit
	}

	if f.SoldCount+quantity > f.TotalStock {
		return ErrFlashsaleSoldOut
	}

	return nil
}

// Buy 购买
func (f *Flashsale) Buy(quantity int32) error {
	if f.SoldCount+quantity > f.TotalStock {
		return ErrFlashsaleSoldOut
	}

	f.SoldCount += quantity
	return nil
}

// Start 开始秒杀
func (f *Flashsale) Start() {
	f.Status = FlashsaleStatusOngoing
}

// End 结束秒杀
func (f *Flashsale) End() {
	f.Status = FlashsaleStatusEnded
}

// Cancel 取消秒杀
func (f *Flashsale) Cancel() {
	f.Status = FlashsaleStatusCanceled
}

// FlashsaleOrderStatus 秒杀订单状态
type FlashsaleOrderStatus int8

const (
	FlashsaleOrderStatusPending   FlashsaleOrderStatus = 0 // 待支付
	FlashsaleOrderStatusPaid      FlashsaleOrderStatus = 1 // 已支付
	FlashsaleOrderStatusCancelled FlashsaleOrderStatus = 2 // 已取消
)

// FlashsaleOrder 秒杀订单实体
type FlashsaleOrder struct {
	gorm.Model
	FlashsaleID uint64               `gorm:"not null;index;comment:秒杀活动ID" json:"flashsale_id"`
	UserID      uint64               `gorm:"not null;index;comment:用户ID" json:"user_id"`
	ProductID   uint64               `gorm:"not null;comment:商品ID" json:"product_id"`
	SkuID       uint64               `gorm:"not null;comment:SKU ID" json:"sku_id"`
	Quantity    int32                `gorm:"not null;comment:数量" json:"quantity"`
	Price       int64                `gorm:"not null;comment:单价" json:"price"`
	TotalAmount int64                `gorm:"not null;comment:总金额" json:"total_amount"`
	Status      FlashsaleOrderStatus `gorm:"default:0;comment:状态" json:"status"`
}

// NewFlashsaleOrder 创建秒杀订单
func NewFlashsaleOrder(flashsaleID, userID, productID, skuID uint64, quantity int32, price int64) *FlashsaleOrder {
	return &FlashsaleOrder{
		FlashsaleID: flashsaleID,
		UserID:      userID,
		ProductID:   productID,
		SkuID:       skuID,
		Quantity:    quantity,
		Price:       price,
		TotalAmount: price * int64(quantity),
		Status:      FlashsaleOrderStatusPending,
	}
}

// Pay 支付
func (o *FlashsaleOrder) Pay() {
	o.Status = FlashsaleOrderStatusPaid
}

// Cancel 取消
func (o *FlashsaleOrder) Cancel() {
	o.Status = FlashsaleOrderStatusCancelled
}
