package entity

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// OrderStatus 订单状态
type OrderStatus int

const (
	PendingPayment  OrderStatus = 1
	Paid            OrderStatus = 2
	Shipped         OrderStatus = 3
	Delivered       OrderStatus = 4
	Completed       OrderStatus = 5
	Cancelled       OrderStatus = 6
	RefundRequested OrderStatus = 7
	Refunded        OrderStatus = 8
	Closed          OrderStatus = 9
)

func (s OrderStatus) String() string {
	names := map[OrderStatus]string{
		PendingPayment:  "PendingPayment",
		Paid:            "Paid",
		Shipped:         "Shipped",
		Delivered:       "Delivered",
		Completed:       "Completed",
		Cancelled:       "Cancelled",
		RefundRequested: "RefundRequested",
		Refunded:        "Refunded",
		Closed:          "Closed",
	}
	return names[s]
}

// Order 订单聚合根
type Order struct {
	gorm.Model
	OrderNo         string           `gorm:"type:varchar(64);uniqueIndex;not null;comment:订单编号" json:"order_no"`
	UserID          uint64           `gorm:"index;not null;comment:用户ID" json:"user_id"`
	Status          OrderStatus      `gorm:"type:tinyint;not null;default:1;comment:订单状态" json:"status"`
	TotalAmount     int64            `gorm:"not null;comment:订单总金额(分)" json:"total_amount"`
	ActualAmount    int64            `gorm:"not null;comment:实际支付金额(分)" json:"actual_amount"`
	ShippingFee     int64            `gorm:"not null;default:0;comment:运费(分)" json:"shipping_fee"`
	DiscountAmount  int64            `gorm:"not null;default:0;comment:优惠金额(分)" json:"discount_amount"`
	PaymentMethod   string           `gorm:"type:varchar(32);comment:支付方式" json:"payment_method"`
	Remark          string           `gorm:"type:varchar(255);comment:订单备注" json:"remark"`
	ShippingAddress *ShippingAddress `gorm:"embedded;embeddedPrefix:shipping_" json:"shipping_address"`
	Items           []*OrderItem     `gorm:"foreignKey:OrderID" json:"items"`
	Logs            []*OrderLog      `gorm:"foreignKey:OrderID" json:"logs"`
	PaidAt          *time.Time       `gorm:"comment:支付时间" json:"paid_at"`
	ShippedAt       *time.Time       `gorm:"comment:发货时间" json:"shipped_at"`
	DeliveredAt     *time.Time       `gorm:"comment:送达时间" json:"delivered_at"`
	CompletedAt     *time.Time       `gorm:"comment:完成时间" json:"completed_at"`
	CancelledAt     *time.Time       `gorm:"comment:取消时间" json:"cancelled_at"`
}

// OrderItem 订单项实体
type OrderItem struct {
	gorm.Model
	OrderID         uint64 `gorm:"index;not null;comment:订单ID" json:"order_id"`
	ProductID       uint64 `gorm:"not null;comment:商品ID" json:"product_id"`
	SkuID           uint64 `gorm:"not null;comment:SKU ID" json:"sku_id"`
	ProductName     string `gorm:"type:varchar(255);not null;comment:商品名称" json:"product_name"`
	SkuName         string `gorm:"type:varchar(255);not null;comment:SKU名称" json:"sku_name"`
	ProductImageURL string `gorm:"type:varchar(255);comment:商品图片URL" json:"product_image_url"`
	Price           int64  `gorm:"not null;comment:单价(分)" json:"price"`
	Quantity        int32  `gorm:"not null;comment:数量" json:"quantity"`
	TotalPrice      int64  `gorm:"not null;comment:总价(分)" json:"total_price"`
}

// ShippingAddress 收货地址值对象
type ShippingAddress struct {
	RecipientName   string `gorm:"type:varchar(64);comment:收货人姓名" json:"recipient_name"`
	PhoneNumber     string `gorm:"type:varchar(20);comment:手机号" json:"phone_number"`
	Province        string `gorm:"type:varchar(64);comment:省份" json:"province"`
	City            string `gorm:"type:varchar(64);comment:城市" json:"city"`
	District        string `gorm:"type:varchar(64);comment:区县" json:"district"`
	DetailedAddress string `gorm:"type:varchar(255);comment:详细地址" json:"detailed_address"`
	PostalCode      string `gorm:"type:varchar(20);comment:邮政编码" json:"postal_code"`
}

// OrderLog 订单日志值对象
type OrderLog struct {
	gorm.Model
	OrderID   uint64 `gorm:"index;not null;comment:订单ID" json:"order_id"`
	Operator  string `gorm:"type:varchar(64);not null;comment:操作人" json:"operator"`
	Action    string `gorm:"type:varchar(64);not null;comment:操作动作" json:"action"`
	OldStatus string `gorm:"type:varchar(32);comment:旧状态" json:"old_status"`
	NewStatus string `gorm:"type:varchar(32);comment:新状态" json:"new_status"`
	Remark    string `gorm:"type:varchar(255);comment:备注" json:"remark"`
}

// NewOrder 创建订单
func NewOrder(orderNo string, userID uint64, items []*OrderItem, shippingAddr *ShippingAddress) *Order {
	var totalAmount int64
	for _, item := range items {
		item.TotalPrice = item.Price * int64(item.Quantity)
		totalAmount += item.TotalPrice
	}

	order := &Order{
		OrderNo:         orderNo,
		UserID:          userID,
		Status:          PendingPayment,
		TotalAmount:     totalAmount,
		ActualAmount:    totalAmount,
		ShippingFee:     0,
		DiscountAmount:  0,
		ShippingAddress: shippingAddr,
		Items:           items,
		Logs:            []*OrderLog{},
	}

	order.AddLog("System", "Order Created", "", PendingPayment.String(), "Initial order creation")
	return order
}

// CanPay 是否可以支付
func (o *Order) CanPay() error {
	if o.Status != PendingPayment {
		return fmt.Errorf("order status must be PendingPayment, current: %s", o.Status.String())
	}
	return nil
}

// Pay 支付订单
func (o *Order) Pay(paymentMethod string, operator string) error {
	if err := o.CanPay(); err != nil {
		return err
	}

	oldStatus := o.Status
	o.Status = Paid
	o.PaymentMethod = paymentMethod
	now := time.Now()
	o.PaidAt = &now

	o.AddLog(operator, "Order Paid", oldStatus.String(), Paid.String(), fmt.Sprintf("Payment method: %s", paymentMethod))
	return nil
}

// CanShip 是否可以发货
func (o *Order) CanShip() error {
	if o.Status != Paid {
		return fmt.Errorf("order status must be Paid, current: %s", o.Status.String())
	}
	return nil
}

// Ship 发货
func (o *Order) Ship(operator string) error {
	if err := o.CanShip(); err != nil {
		return err
	}

	oldStatus := o.Status
	o.Status = Shipped
	now := time.Now()
	o.ShippedAt = &now

	o.AddLog(operator, "Order Shipped", oldStatus.String(), Shipped.String(), "Order has been shipped")
	return nil
}

// CanDeliver 是否可以送达
func (o *Order) CanDeliver() error {
	if o.Status != Shipped {
		return fmt.Errorf("order status must be Shipped, current: %s", o.Status.String())
	}
	return nil
}

// Deliver 送达
func (o *Order) Deliver(operator string) error {
	if err := o.CanDeliver(); err != nil {
		return err
	}

	oldStatus := o.Status
	o.Status = Delivered
	now := time.Now()
	o.DeliveredAt = &now

	o.AddLog(operator, "Order Delivered", oldStatus.String(), Delivered.String(), "Order has been delivered")
	return nil
}

// CanComplete 是否可以完成
func (o *Order) CanComplete() error {
	if o.Status != Delivered {
		return fmt.Errorf("order status must be Delivered, current: %s", o.Status.String())
	}
	return nil
}

// Complete 完成订单
func (o *Order) Complete(operator string) error {
	if err := o.CanComplete(); err != nil {
		return err
	}

	oldStatus := o.Status
	o.Status = Completed
	now := time.Now()
	o.CompletedAt = &now

	o.AddLog(operator, "Order Completed", oldStatus.String(), Completed.String(), "Order has been completed")
	return nil
}

// CanCancel 是否可以取消
func (o *Order) CanCancel() error {
	if o.Status != PendingPayment && o.Status != Paid {
		return fmt.Errorf("order cannot be cancelled in current status: %s", o.Status.String())
	}
	return nil
}

// Cancel 取消订单
func (o *Order) Cancel(operator, reason string) error {
	if err := o.CanCancel(); err != nil {
		return err
	}

	oldStatus := o.Status
	o.Status = Cancelled
	now := time.Now()
	o.CancelledAt = &now

	o.AddLog(operator, "Order Cancelled", oldStatus.String(), Cancelled.String(), reason)
	return nil
}

// CanRequestRefund 是否可以申请退款
func (o *Order) CanRequestRefund() error {
	if o.Status != Paid && o.Status != Shipped && o.Status != Delivered {
		return fmt.Errorf("order cannot request refund in current status: %s", o.Status.String())
	}
	return nil
}

// RequestRefund 申请退款
func (o *Order) RequestRefund(operator, reason string) error {
	if err := o.CanRequestRefund(); err != nil {
		return err
	}

	oldStatus := o.Status
	o.Status = RefundRequested

	o.AddLog(operator, "Refund Requested", oldStatus.String(), RefundRequested.String(), reason)
	return nil
}

// ApproveRefund 批准退款
func (o *Order) ApproveRefund(operator string) error {
	if o.Status != RefundRequested {
		return fmt.Errorf("order status must be RefundRequested, current: %s", o.Status.String())
	}

	oldStatus := o.Status
	o.Status = Refunded

	o.AddLog(operator, "Refund Approved", oldStatus.String(), Refunded.String(), "Refund has been approved")
	return nil
}

// ApplyDiscount 应用折扣
func (o *Order) ApplyDiscount(discountAmount int64, operator, reason string) error {
	if discountAmount < 0 {
		return fmt.Errorf("discount amount must be positive")
	}
	if discountAmount > o.TotalAmount {
		return fmt.Errorf("discount amount cannot exceed total amount")
	}

	o.DiscountAmount = discountAmount
	o.ActualAmount = o.TotalAmount - discountAmount

	o.AddLog(operator, "Discount Applied", "", "", fmt.Sprintf("Discount: %d, Reason: %s", discountAmount, reason))
	return nil
}

// AddLog 添加日志
func (o *Order) AddLog(operator, action, oldStatus, newStatus, remark string) {
	log := &OrderLog{
		Operator:  operator,
		Action:    action,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		Remark:    remark,
	}
	o.Logs = append(o.Logs, log)
}

// GetTotalQuantity 获取商品总数量
func (o *Order) GetTotalQuantity() int32 {
	var total int32
	for _, item := range o.Items {
		total += item.Quantity
	}
	return total
}
