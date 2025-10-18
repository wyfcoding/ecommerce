package model

import (
	"time"
)

// Cart 代表一个用户的购物车。
// 购物车本身可以是一个逻辑概念，也可以有对应的数据库记录。
// 这里我们假设每个用户有一个唯一的购物车记录，其中包含多个购物车项。
type Cart struct {
	UserID         uint64     `gorm:"primarykey" json:"user_id"`                          // 用户ID，作为主键
	TotalQuantity  int32      `gorm:"type:int;not null;default:0" json:"total_quantity"` // 购物车中商品总数量
	TotalAmount    int64      `gorm:"type:bigint;not null;default:0" json:"total_amount"`   // 购物车中商品总金额 (原价，单位: 分)
	DiscountAmount int64      `gorm:"type:bigint;not null;default:0" json:"discount_amount"` // 优惠金额 (单位: 分)
	ActualAmount   int64      `gorm:"type:bigint;not null;default:0" json:"actual_amount"`  // 实际支付金额 (总金额 - 优惠金额，单位: 分)
	AppliedCouponCode string    `gorm:"type:varchar(50)" json:"applied_coupon_code"`        // 已应用的优惠券码
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`                   // 购物车创建时间
	UpdatedAt      time.Time  `gorm:"autoUpdateTime" json:"updated_at"`                   // 最后更新时间
	DeletedAt      *time.Time `gorm:"index" json:"deleted_at,omitempty"`                  // 软删除时间

	Items []CartItem `gorm:"foreignKey:UserID;references:UserID" json:"items"` // 购物车中的商品列表
}

// CartItem 代表购物车中的一个商品项。
type CartItem struct {
	ID              uint64     `gorm:"primarykey" json:"id"`                               // 购物车项ID
	UserID          uint64     `gorm:"index;not null" json:"user_id"`                      // 所属用户ID
	ProductID       uint64     `gorm:"index;not null" json:"product_id"`                   // 商品SPU ID
	SKUID           uint64     `gorm:"index;not null" json:"sku_id"`                       // 商品SKU ID
	ProductName     string     `gorm:"type:varchar(255);not null" json:"product_name"`     // 商品名称
	SKUName         string     `gorm:"type:varchar(255);not null" json:"sku_name"`         // SKU名称
	ProductImageURL string     `gorm:"type:varchar(255)" json:"product_image_url"`         // 商品图片URL
	Price           int64      `gorm:"type:bigint;not null" json:"price"`                  // 商品单价 (单位: 分)
	Quantity        int32      `gorm:"type:int;not null" json:"quantity"`                  // 购买数量
	TotalPrice      int64      `gorm:"type:bigint;not null" json:"total_price"`            // 该购物车项总价 (单位: 分)
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`                   // 添加到购物车时间
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updated_at"`                   // 最后更新时间
	DeletedAt       *time.Time `gorm:"index" json:"deleted_at,omitempty"`                  // 软删除时间
}
