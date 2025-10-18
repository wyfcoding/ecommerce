package model

import (
	"time"
)

// OrderStatus 定义订单状态的枚举。
type OrderStatus int32

const (
	OrderStatusUnspecified OrderStatus = 0 // 未指定
	PendingPayment         OrderStatus = 1 // 待支付
	Paid                   OrderStatus = 2 // 已支付
	Shipped                OrderStatus = 3 // 已发货
	Delivered              OrderStatus = 4 // 已送达
	Completed              OrderStatus = 5 // 已完成 (用户确认收货)
	Cancelled              OrderStatus = 6 // 已取消 (用户或系统取消)
	RefundRequested        OrderStatus = 7 // 退款申请中
	Refunded               OrderStatus = 8 // 已退款
	Closed                 OrderStatus = 9 // 已关闭 (交易失败或超时)
)

// PaymentStatus 定义支付状态的枚举。
type PaymentStatus int32

const (
	PaymentStatusUnspecified PaymentStatus = 0 // 未指定
	Unpaid                   PaymentStatus = 1 // 未支付
	Processing               PaymentStatus = 2 // 支付处理中
	Success                  PaymentStatus = 3 // 支付成功
	Failed                   PaymentStatus = 4 // 支付失败
	Refunding                PaymentStatus = 5 // 退款中
	RefundSuccess            PaymentStatus = 6 // 退款成功
	RefundFailed             PaymentStatus = 7 // 退款失败
)

// ShippingStatus 定义配送状态的枚举。
type ShippingStatus int32

const (
	ShippingStatusUnspecified ShippingStatus = 0 // 未指定
	PendingShipment           ShippingStatus = 1 // 待发货
	Shipped                   ShippingStatus = 2 // 已发货
	InTransit                 ShippingStatus = 3 // 运输中
	Delivered                 ShippingStatus = 4 // 已送达
	Exception                 ShippingStatus = 5 // 配送异常
)

// Order 代表一个订单的完整信息。
type Order struct {
	ID             uint64        `gorm:"primarykey" json:"id"`                               // 订单ID
	OrderNo        string        `gorm:"type:varchar(64);uniqueIndex;not null" json:"order_no"` // 订单编号 (唯一)
	UserID         uint64        `gorm:"index;not null" json:"user_id"`                      // 用户ID
	Status         OrderStatus   `gorm:"type:tinyint;not null" json:"status"`                // 订单当前状态
	PaymentStatus  PaymentStatus `gorm:"type:tinyint;not null" json:"payment_status"`        // 支付状态
	ShippingStatus ShippingStatus `gorm:"type:tinyint;not null" json:"shipping_status"`       // 配送状态
	TotalAmount    int64         `gorm:"type:bigint;not null" json:"total_amount"`           // 订单总金额 (单位: 分)
	ActualAmount   int64         `gorm:"type:bigint;not null" json:"actual_amount"`          // 实际支付金额 (单位: 分)
	ShippingFee    int64         `gorm:"type:bigint;not null" json:"shipping_fee"`           // 运费 (单位: 分)
	DiscountAmount int64         `gorm:"type:bigint;not null" json:"discount_amount"`        // 优惠金额 (单位: 分)
	PaymentMethod  string        `gorm:"type:varchar(50)" json:"payment_method"`             // 支付方式
	Remark         string        `gorm:"type:varchar(500)" json:"remark"`                    // 订单备注
	CreatedAt      time.Time     `gorm:"autoCreateTime" json:"created_at"`                   // 订单创建时间
	UpdatedAt      time.Time     `gorm:"autoUpdateTime" json:"updated_at"`                   // 最后更新时间
	PaidAt         *time.Time    `json:"paid_at,omitempty"`                                  // 支付时间
	ShippedAt      *time.Time    `json:"shipped_at,omitempty"`                               // 发货时间
	DeliveredAt    *time.Time    `json:"delivered_at,omitempty"`                             // 送达时间
	CompletedAt    *time.Time    `json:"completed_at,omitempty"`                             // 完成时间
	CancelledAt    *time.Time    `json:"cancelled_at,omitempty"`                             // 取消时间
	DeletedAt      *time.Time    `gorm:"index" json:"deleted_at,omitempty"`                  // 软删除时间

	ShippingAddress ShippingAddress `gorm:"foreignKey:OrderID" json:"shipping_address"` // 收货地址信息
	Items           []OrderItem     `gorm:"foreignKey:OrderID" json:"items"`            // 订单包含的商品列表
	Logs            []OrderLog      `gorm:"foreignKey:OrderID" json:"logs"`             // 订单操作日志
}

// OrderItem 代表订单中的一个商品项。
type OrderItem struct {
	ID              uint64     `gorm:"primarykey" json:"id"`                               // 订单项ID
	OrderID         uint64     `gorm:"index;not null" json:"order_id"`                     // 所属订单ID
	ProductID       uint64     `gorm:"index;not null" json:"product_id"`                   // 商品SPU ID
	SKUID           uint64     `gorm:"index;not null" json:"sku_id"`                       // 商品SKU ID
	ProductName     string     `gorm:"type:varchar(255);not null" json:"product_name"`     // 商品名称
	SKUName         string     `gorm:"type:varchar(255);not null" json:"sku_name"`         // SKU名称
	ProductImageURL string     `gorm:"type:varchar(255)" json:"product_image_url"`         // 商品图片URL
	Price           int64      `gorm:"type:bigint;not null" json:"price"`                  // 购买时单价 (单位: 分)
	Quantity        int32      `gorm:"type:int;not null" json:"quantity"`                  // 购买数量
	TotalPrice      int64      `gorm:"type:bigint;not null" json:"total_price"`            // 该订单项总价 (单位: 分)
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`                   // 创建时间
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updated_at"`                   // 最后更新时间
	DeletedAt       *time.Time `gorm:"index" json:"deleted_at,omitempty"`                  // 软删除时间
}

// ShippingAddress 代表订单的收货地址信息。
type ShippingAddress struct {
	ID              uint64     `gorm:"primarykey" json:"id"`
	OrderID         uint64     `gorm:"uniqueIndex;not null" json:"order_id"`               // 所属订单ID
	RecipientName   string     `gorm:"type:varchar(100);not null" json:"recipient_name"`   // 收货人姓名
	PhoneNumber     string     `gorm:"type:varchar(20);not null" json:"phone_number"`      // 手机号
	Province        string     `gorm:"type:varchar(50);not null" json:"province"`          // 省份
	City            string     `gorm:"type:varchar(50);not null" json:"city"`              // 城市
	District        string     `gorm:"type:varchar(50);not null" json:"district"`          // 区县
	DetailedAddress string     `gorm:"type:varchar(255);not null" json:"detailed_address"` // 详细地址
	PostalCode      string     `gorm:"type:varchar(10)" json:"postal_code"`                // 邮政编码
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"created_at"`                   // 创建时间
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updated_at"`                   // 最后更新时间
	DeletedAt       *time.Time `gorm:"index" json:"deleted_at,omitempty"`                  // 软删除时间
}

// OrderLog 记录订单的关键操作和状态变更。
type OrderLog struct {
	ID        uint64     `gorm:"primarykey" json:"id"`
	OrderID   uint64     `gorm:"index;not null" json:"order_id"`                     // 所属订单ID
	Operator  string     `gorm:"type:varchar(100);not null" json:"operator"`         // 操作人 (用户ID或系统)
	Action    string     `gorm:"type:varchar(255);not null" json:"action"`           // 操作描述
	OldStatus string     `gorm:"type:varchar(50)" json:"old_status"`                 // 旧状态
	NewStatus string     `gorm:"type:varchar(50);not null" json:"new_status"`        // 新状态
	Remark    string     `gorm:"type:varchar(500)" json:"remark"`                    // 备注
	CreatedAt time.Time  `gorm:"autoCreateTime" json:"created_at"`                   // 操作时间
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`                  // 软删除时间
}
