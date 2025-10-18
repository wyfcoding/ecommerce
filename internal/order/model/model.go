package model

import (
	"time"

	"gorm.io/gorm"
)

// OrderStatus 定义了订单的各种状态
// 使用 iota 可以方便地给常量赋值
type OrderStatus int

const (
	StatusPendingPayment OrderStatus = iota + 1 // 1: 待支付
	StatusProcessing                          // 2: 处理中 (已支付，待发货)
	StatusShipped                             // 3: 已发货
	StatusCompleted                           // 4: 已完成 (已收货)
	StatusCancelled                           // 5: 已取消
	StatusRefunded                            // 6: 已退款
)

// Order 订单主模型
// 记录了订单的总体信息
type Order struct {
	ID              uint        `gorm:"primarykey" json:"id"`
	OrderSN         string      `gorm:"type:varchar(100);uniqueIndex;not null" json:"order_sn"` // 订单号，业务唯一标识
	UserID          uint        `gorm:"not null;index" json:"user_id"`                      // 用户ID
	TotalPrice      float64     `gorm:"type:decimal(10,2);not null" json:"total_price"`       // 订单总金额
	ShippingAddress string      `gorm:"type:varchar(255);not null" json:"shipping_address"`  // 收货地址
	ContactPhone    string      `gorm:"type:varchar(20);not null" json:"contact_phone"`     // 联系电话
	Remarks         string      `gorm:"type:varchar(255)" json:"remarks"`                   // 订单备注
	Status          OrderStatus `gorm:"not null;default:1" json:"status"`                   // 订单状态
	PaymentMethod   string      `gorm:"type:varchar(50)" json:"payment_method"`            // 支付方式
	PaymentSN       string      `gorm:"type:varchar(100);index" json:"payment_sn"`          // 支付流水号
	PaidAt          *time.Time  `json:"paid_at"`                                            // 支付时间
	ShippedAt       *time.Time  `json:"shipped_at"`                                         // 发货时间
	CompletedAt     *time.Time  `json:"completed_at"`                                       // 完成时间
	CancelledAt     *time.Time  `json:"cancelled_at"`                                       // 取消时间
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	Items []OrderItem `gorm:"foreignKey:OrderID" json:"items"` // 订单包含的商品项
}

// OrderItem 订单商品项模型
// 记录了订单中每一个商品的快照信息
type OrderItem struct {
	ID          uint    `gorm:"primarykey" json:"id"`
	OrderID     uint    `gorm:"not null;index" json:"order_id"`         // 所属订单ID
	ProductID   uint    `gorm:"not null;index" json:"product_id"`       // 商品ID
	ProductName string  `gorm:"type:varchar(255);not null" json:"product_name"` // 商品名称 (快照)
	ProductSKU  string  `gorm:"type:varchar(100);not null" json:"product_sku"`  // 商品SKU (快照)
	Price       float64 `gorm:"type:decimal(10,2);not null" json:"price"`        // 商品单价 (快照)
	Quantity    int     `gorm:"not null" json:"quantity"`                 // 商品数量
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 自定义表名
func (Order) TableName() string {
	return "orders"
}

func (OrderItem) TableName() string {
	return "order_items"
}
