package domain

import (
	"errors" // 导入标准错误处理库。
	"time"   // 导入时间库。

	"gorm.io/gorm" // 导入GORM库。
)

// 定义Flashsale模块的业务错误。
var (
	ErrFlashsaleNotFound   = errors.New("秒杀活动不存在") // 秒杀活动记录未找到。
	ErrFlashsaleNotStarted = errors.New("秒杀活动未开始") // 秒杀活动尚未开始。
	ErrFlashsaleEnded      = errors.New("秒杀活动已结束") // 秒杀活动已结束。
	ErrFlashsaleSoldOut    = errors.New("秒杀商品已售罄") // 秒杀商品库存已售罄。
	ErrFlashsaleLimit      = errors.New("超过购买限制")  // 用户购买数量超过限制。
)

// FlashsaleStatus 定义了秒杀活动的生命周期状态。
type FlashsaleStatus int8

const (
	FlashsaleStatusPending  FlashsaleStatus = 0 // 未开始：秒杀活动已创建但尚未到开始时间。
	FlashsaleStatusOngoing  FlashsaleStatus = 1 // 进行中：秒杀活动正在进行。
	FlashsaleStatusEnded    FlashsaleStatus = 2 // 已结束：秒杀活动已过结束时间。
	FlashsaleStatusCanceled FlashsaleStatus = 3 // 已取消：秒杀活动被取消。
)

// Flashsale 实体是秒杀模块的聚合根。
// 它代表一个秒杀活动，包含了秒杀商品的详细信息、价格、库存、时间范围和状态。
type Flashsale struct {
	gorm.Model                    // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	Name          string          `gorm:"type:varchar(255);not null;comment:活动名称" json:"name"`     // 活动名称。
	ProductID     uint64          `gorm:"not null;index;comment:商品ID" json:"product_id"`           // 关联的商品ID，索引字段。
	SkuID         uint64          `gorm:"not null;index;comment:SKU ID" json:"sku_id"`             // 关联的SKU ID，索引字段。
	OriginalPrice int64           `gorm:"not null;comment:原价" json:"original_price"`               // 商品原价（单位：分）。
	FlashPrice    int64           `gorm:"not null;comment:秒杀价" json:"flash_price"`                 // 秒杀价格（单位：分）。
	TotalStock    int32           `gorm:"not null;comment:总库存" json:"total_stock"`                 // 秒杀活动的总库存。
	SoldCount     int32           `gorm:"not null;default:0;comment:已售数量" json:"sold_count"`       // 已售出的数量。
	LimitPerUser  int32           `gorm:"not null;default:1;comment:每人限购数量" json:"limit_per_user"` // 每位用户限购数量。
	StartTime     time.Time       `gorm:"not null;comment:开始时间" json:"start_time"`                 // 秒杀活动开始时间。
	EndTime       time.Time       `gorm:"not null;comment:结束时间" json:"end_time"`                   // 秒杀活动结束时间。
	Status        FlashsaleStatus `gorm:"default:0;comment:状态" json:"status"`                      // 秒杀活动状态，默认为未开始。
	Description   string          `gorm:"type:text;comment:描述" json:"description"`                 // 活动描述。
}

// NewFlashsale 创建并返回一个新的 Flashsale 实体实例。
// name: 活动名称。
// productID, skuID: 商品和SKU ID。
// originalPrice, flashPrice: 原价和秒杀价。
// totalStock, limitPerUser: 总库存和每人限购。
// startTime, endTime: 活动时间。
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
		Status:        FlashsaleStatusPending, // 默认状态为未开始。
	}
}

// RemainingStock 计算秒杀活动的剩余库存。
func (f *Flashsale) RemainingStock() int32 {
	return f.TotalStock - f.SoldCount
}

// IsAvailable 检查秒杀活动当前是否可用（正在进行中且有库存）。
func (f *Flashsale) IsAvailable() bool {
	now := time.Now()
	return f.Status == FlashsaleStatusOngoing && // 状态为进行中。
		now.After(f.StartTime) && // 当前时间在开始时间之后。
		now.Before(f.EndTime) && // 当前时间在结束时间之前。
		f.SoldCount < f.TotalStock // 仍有剩余库存。
}

// CanBuy 检查用户是否可以购买指定数量的秒杀商品。
// userBoughtCount: 用户在此次秒杀活动中已购买的数量。
// quantity: 用户本次尝试购买的数量。
// 返回一个错误，如果不能购买则说明具体原因。
func (f *Flashsale) CanBuy(userBoughtCount int32, quantity int32) error {
	now := time.Now()

	// 检查秒杀活动状态和时间。
	if f.Status != FlashsaleStatusOngoing {
		if now.Before(f.StartTime) {
			return ErrFlashsaleNotStarted
		}
		return ErrFlashsaleEnded
	}

	// 再次检查时间，确保在活动进行中。
	if now.Before(f.StartTime) {
		return ErrFlashsaleNotStarted
	}
	if now.After(f.EndTime) {
		return ErrFlashsaleEnded
	}

	// 检查总库存是否充足。
	if f.SoldCount >= f.TotalStock {
		return ErrFlashsaleSoldOut
	}

	// 检查用户是否超过限购数量。
	if f.LimitPerUser > 0 && userBoughtCount+quantity > f.LimitPerUser {
		return ErrFlashsaleLimit
	}

	// 检查本次购买后，是否会超出总库存。
	if f.SoldCount+quantity > f.TotalStock {
		return ErrFlashsaleSoldOut
	}

	return nil // 可以购买。
}

// Buy 模拟购买指定数量的秒杀商品，减少库存。
func (f *Flashsale) Buy(quantity int32) error {
	if f.SoldCount+quantity > f.TotalStock {
		return ErrFlashsaleSoldOut // 再次检查库存，防止超卖。
	}

	f.SoldCount += quantity // 增加已售数量。
	return nil
}

// Start 启动秒杀活动，将其状态设置为“进行中”。
func (f *Flashsale) Start() {
	f.Status = FlashsaleStatusOngoing
}

// End 结束秒杀活动，将其状态设置为“已结束”。
func (f *Flashsale) End() {
	f.Status = FlashsaleStatusEnded
}

// Cancel 取消秒杀活动，将其状态设置为“已取消”。
func (f *Flashsale) Cancel() {
	f.Status = FlashsaleStatusCanceled
}

// FlashsaleOrderStatus 定义了秒杀订单的生命周期状态。
type FlashsaleOrderStatus int8

const (
	FlashsaleOrderStatusPending   FlashsaleOrderStatus = 0 // 待支付：订单已创建，等待用户支付。
	FlashsaleOrderStatusPaid      FlashsaleOrderStatus = 1 // 已支付：订单已支付成功。
	FlashsaleOrderStatusCancelled FlashsaleOrderStatus = 2 // 已取消：订单已取消。
)

// FlashsaleOrder 实体代表一个秒杀订单。
// 它记录了秒杀活动、用户、商品、数量和价格等信息。
type FlashsaleOrder struct {
	gorm.Model                       // 嵌入gorm.Model。
	FlashsaleID uint64               `gorm:"not null;index;comment:秒杀活动ID" json:"flashsale_id"` // 关联的秒杀活动ID，索引字段。
	UserID      uint64               `gorm:"not null;index;comment:用户ID" json:"user_id"`        // 下单用户ID，索引字段。
	ProductID   uint64               `gorm:"not null;comment:商品ID" json:"product_id"`           // 商品ID。
	SkuID       uint64               `gorm:"not null;comment:SKU ID" json:"sku_id"`             // SKU ID。
	Quantity    int32                `gorm:"not null;comment:数量" json:"quantity"`               // 购买数量。
	Price       int64                `gorm:"not null;comment:单价" json:"price"`                  // 商品单价（秒杀价格）。
	TotalAmount int64                `gorm:"not null;comment:总金额" json:"total_amount"`          // 订单总金额。
	Status      FlashsaleOrderStatus `gorm:"default:0;comment:状态" json:"status"`                // 订单状态，默认为待支付。
}

// NewFlashsaleOrder 创建并返回一个新的 FlashsaleOrder 实体实例。
// flashsaleID, userID, productID, skuID: 关联的活动、用户和商品信息。
// quantity: 购买数量。
// price: 商品单价。
func NewFlashsaleOrder(flashsaleID, userID, productID, skuID uint64, quantity int32, price int64) *FlashsaleOrder {
	return &FlashsaleOrder{
		FlashsaleID: flashsaleID,
		UserID:      userID,
		ProductID:   productID,
		SkuID:       skuID,
		Quantity:    quantity,
		Price:       price,
		TotalAmount: price * int64(quantity),     // 计算总金额。
		Status:      FlashsaleOrderStatusPending, // 默认状态为待支付。
	}
}

// Pay 支付订单，将其状态设置为“已支付”。
func (o *FlashsaleOrder) Pay() {
	o.Status = FlashsaleOrderStatusPaid
}

// Cancel 取消订单，将其状态设置为“已取消”。
func (o *FlashsaleOrder) Cancel() {
	o.Status = FlashsaleOrderStatusCancelled
}
