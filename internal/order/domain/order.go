// 变更说明：新增订单事件版本字段 Version，用于事件溯源并发控制与读模型回放。
// 假设：事件流以订单ID作为聚合根ID，版本从 0 递增。
package domain

import (
	"context"
	"fmt"
	"time"

	pb "github.com/wyfcoding/ecommerce/goapi/order/v1"
	"github.com/wyfcoding/pkg/fsm"

	"gorm.io/gorm"
)

// TimeoutScheduler 定义了超时调度的接口，用于处理订单超时取消等逻辑。
type TimeoutScheduler interface {
	ScheduleTimeout(orderID string, timeout time.Duration, callback func(orderID string)) error
	Start()
	Stop()
}

// Order 实体是订单模块的聚合根。
type Order struct {
	gorm.Model
	OrderNo         string                       `gorm:"type:varchar(64);uniqueIndex;not null;comment:订单编号" json:"order_no"`
	Version         int64                        `gorm:"not null;default:0;comment:事件版本号(用于事件溯源并发控制)" json:"version"`
	UserID          uint64                       `gorm:"index;not null;comment:用户ID" json:"user_id"`
	Status          pb.OrderStatus               `gorm:"type:tinyint;not null;default:1;comment:订单状态" json:"status"`
	TotalAmount     int64                        `gorm:"not null;comment:订单总金额(分)" json:"total_amount"`
	ActualAmount    int64                        `gorm:"not null;comment:实际支付金额(分)" json:"actual_amount"`
	ShippingFee     int64                        `gorm:"not null;default:0;comment:运费(分)" json:"shipping_fee"`
	DiscountAmount  int64                        `gorm:"not null;default:0;comment:优惠金额(分)" json:"discount_amount"`
	PaymentMethod   string                       `gorm:"type:varchar(32);comment:支付方式" json:"payment_method"`
	Remark          string                       `gorm:"type:varchar(255);comment:订单备注" json:"remark"`
	ShippingAddress *ShippingAddress             `gorm:"embedded;embeddedPrefix:shipping_" json:"shipping_address"`
	Items           []*OrderItem                 `gorm:"foreignKey:OrderID" json:"items"`
	Logs            []*OrderLog                  `gorm:"foreignKey:OrderID" json:"logs"`
	PaidAt          *time.Time                   `gorm:"comment:支付时间" json:"paid_at"`
	ShippedAt       *time.Time                   `gorm:"comment:发货时间" json:"shipped_at"`
	DeliveredAt     *time.Time                   `gorm:"comment:送达时间" json:"delivered_at"`
	CompletedAt     *time.Time                   `gorm:"comment:完成时间" json:"completed_at"`
	CancelledAt     *time.Time                   `gorm:"comment:取消时间" json:"cancelled_at"`
	fsm             *fsm.Machine[string, string] `gorm:"-" json:"-"`
}

// OrderItem 实体代表订单中的一个商品项。
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

// ShippingAddress 值对象定义了订单的收货地址信息。
type ShippingAddress struct {
	RecipientName   string  `gorm:"type:varchar(64);comment:收货人姓名" json:"recipient_name"`
	PhoneNumber     string  `gorm:"type:varchar(20);comment:手机号" json:"phone_number"`
	Province        string  `gorm:"type:varchar(64);comment:省份" json:"province"`
	City            string  `gorm:"type:varchar(64);comment:城市" json:"city"`
	District        string  `gorm:"type:varchar(64);comment:区县" json:"district"`
	DetailedAddress string  `gorm:"type:varchar(255);comment:详细地址" json:"detailed_address"`
	PostalCode      string  `gorm:"type:varchar(20);comment:邮政编码" json:"postal_code"`
	Lat             float64 `gorm:"type:decimal(10,6);comment:纬度" json:"lat"`
	Lon             float64 `gorm:"type:decimal(10,6);comment:经度" json:"lon"`
}

// OrderLog 值对象定义了订单的操作日志记录。
type OrderLog struct {
	gorm.Model
	OrderID   uint64 `gorm:"index;not null;comment:订单ID" json:"order_id"`
	Operator  string `gorm:"type:varchar(64);not null;comment:操作人" json:"operator"`
	Action    string `gorm:"type:varchar(64);not null;comment:操作动作" json:"action"`
	OldStatus string `gorm:"type:varchar(32);comment:旧状态" json:"old_status"`
	NewStatus string `gorm:"type:varchar(32);comment:新状态" json:"new_status"`
	Remark    string `gorm:"type:varchar(255);comment:备注" json:"remark"`
}

// NewOrder 创建并返回一个新的 Order 实体实例。
func NewOrder(orderNo string, userID uint64, items []*OrderItem, shippingAddr *ShippingAddress) *Order {
	var totalAmount int64
	for _, item := range items {
		item.TotalPrice = item.Price * int64(item.Quantity)
		totalAmount += item.TotalPrice
	}

	order := &Order{
		OrderNo:         orderNo,
		UserID:          userID,
		Status:          pb.OrderStatus_PENDING_PAYMENT,
		TotalAmount:     totalAmount,
		ActualAmount:    totalAmount,
		ShippingFee:     0,
		DiscountAmount:  0,
		ShippingAddress: shippingAddr,
		Items:           items,
		Logs:            []*OrderLog{},
	}

	order.AddLog("System", "Order Created", "", pb.OrderStatus_PENDING_PAYMENT.String(), "Initial order creation")
	order.initFSM()
	return order
}

func (o *Order) initFSM() {
	m := fsm.NewMachine[string, string](o.Status.String())

	// 定义转换规则 (使用 Proto String 表示)
	m.AddTransition(pb.OrderStatus_PENDING_PAYMENT.String(), "PAY", pb.OrderStatus_PAID.String())
	m.AddTransition(pb.OrderStatus_ALLOCATING.String(), "CONFIRM", pb.OrderStatus_PENDING_PAYMENT.String())
	m.AddTransition(pb.OrderStatus_PAID.String(), "SHIP", pb.OrderStatus_SHIPPED.String())
	m.AddTransition(pb.OrderStatus_SHIPPED.String(), "DELIVER", pb.OrderStatus_DELIVERED.String())
	m.AddTransition(pb.OrderStatus_DELIVERED.String(), "COMPLETE", pb.OrderStatus_COMPLETED.String())

	// 取消与退款
	m.AddTransition(pb.OrderStatus_PENDING_PAYMENT.String(), "CANCEL", pb.OrderStatus_CANCELLED.String())
	m.AddTransition(pb.OrderStatus_ALLOCATING.String(), "CANCEL", pb.OrderStatus_CANCELLED.String())
	m.AddTransition(pb.OrderStatus_PAID.String(), "CANCEL", pb.OrderStatus_CANCELLED.String())
	m.AddTransition(pb.OrderStatus_PAID.String(), "REFUND_REQ", pb.OrderStatus_REFUND_REQUESTED.String())
	m.AddTransition(pb.OrderStatus_SHIPPED.String(), "REFUND_REQ", pb.OrderStatus_REFUND_REQUESTED.String())
	m.AddTransition(pb.OrderStatus_DELIVERED.String(), "REFUND_REQ", pb.OrderStatus_REFUND_REQUESTED.String())
	m.AddTransition(pb.OrderStatus_REFUND_REQUESTED.String(), "REFUND_APPROVE", pb.OrderStatus_REFUNDED.String())

	o.fsm = m
}

// AfterFind GORM 钩子，加载后初始化状态机
func (o *Order) AfterFind(tx *gorm.DB) error {
	o.initFSM()
	return nil
}

// Trigger 触发状态变更
func (o *Order) Trigger(ctx context.Context, event string, operator string, remark string, args ...any) error {
	if o.fsm == nil {
		o.initFSM()
	}

	oldStatus := o.Status
	err := o.fsm.Trigger(ctx, event, args...)
	if err != nil {
		return err
	}

	newStatusStr := o.fsm.Current()
	// Reverse lookup proto enum from string
	for i := 0; i < len(pb.OrderStatus_name); i++ {
		st := pb.OrderStatus(i)
		if st.String() == newStatusStr {
			o.Status = st
			break
		}
	}

	o.AddLog(operator, event, oldStatus.String(), o.Status.String(), remark)
	return nil
}

// Pay 支付订单。
func (o *Order) Pay(ctx context.Context, paymentMethod string, operator string) error {
	if err := o.Trigger(ctx, "PAY", operator, fmt.Sprintf("Payment method: %s", paymentMethod)); err != nil {
		return err
	}
	o.PaymentMethod = paymentMethod
	now := time.Now()
	o.PaidAt = &now
	return nil
}

// Ship 发货订单。
func (o *Order) Ship(ctx context.Context, operator string) error {
	if err := o.Trigger(ctx, "SHIP", operator, "Order has been shipped"); err != nil {
		return err
	}
	now := time.Now()
	o.ShippedAt = &now
	return nil
}

// Deliver 送达订单。
func (o *Order) Deliver(ctx context.Context, operator string) error {
	if err := o.Trigger(ctx, "DELIVER", operator, "Order has been delivered"); err != nil {
		return err
	}
	now := time.Now()
	o.DeliveredAt = &now
	return nil
}

// Complete 完成订单。
func (o *Order) Complete(ctx context.Context, operator string) error {
	if err := o.Trigger(ctx, "COMPLETE", operator, "Order has been completed"); err != nil {
		return err
	}
	now := time.Now()
	o.CompletedAt = &now
	return nil
}

// Cancel 取消订单。
func (o *Order) Cancel(ctx context.Context, operator, reason string) error {
	if err := o.Trigger(ctx, "CANCEL", operator, reason); err != nil {
		return err
	}
	now := time.Now()
	o.CancelledAt = &now
	return nil
}

// RequestRefund 申请退款。
func (o *Order) RequestRefund(ctx context.Context, operator, reason string) error {
	return o.Trigger(ctx, "REFUND_REQ", operator, reason)
}

// ApproveRefund 批准退款。
func (o *Order) ApproveRefund(ctx context.Context, operator string) error {
	return o.Trigger(ctx, "REFUND_APPROVE", operator, "Refund has been approved")
}

// ApplyDiscount 应用折扣。
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

// AddLog 添加订单操作日志。
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

// GetTotalQuantity 获取订单中所有商品的总数量。
func (o *Order) GetTotalQuantity() int32 {
	var total int32
	for _, item := range o.Items {
		total += item.Quantity
	}
	return total
}
