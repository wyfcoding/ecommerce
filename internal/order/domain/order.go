// 变更说明：新增订单事件版本字段 Version，用于事件溯源并发控制与读模型回放。
// 假设：事件流以订单ID作为聚合根ID，版本从 0 递增。
package domain

import (
	"context"
	"fmt"
	"time"

	pb "github.com/wyfcoding/ecommerce/goapi/order/v1"
	"github.com/wyfcoding/pkg/fsm"
)

// TimeoutScheduler 定义了超时调度的接口，用于处理订单超时取消等逻辑。
type TimeoutScheduler interface {
	ScheduleTimeout(orderID string, timeout time.Duration, callback func(orderID string)) error
	Start()
	Stop()
}

// Order 实体是订单模块的聚合根。
type Order struct {
	ID                   uint                         `json:"id"`
	CreatedAt            time.Time                    `json:"created_at"`
	UpdatedAt            time.Time                    `json:"updated_at"`
	OrderNo              string                       `json:"order_no"`
	Version              int64                        `json:"version"`
	UserID               uint64                       `json:"user_id"`
	Status               pb.OrderStatus               `json:"status"`
	TotalAmount          int64                        `json:"total_amount"`
	ActualAmount         int64                        `json:"actual_amount"`
	ShippingFee          int64                        `json:"shipping_fee"`
	DiscountAmount       int64                        `json:"discount_amount"`
	PaymentMethod        string                       `json:"payment_method"`
	PaymentTransactionID string                       `json:"payment_transaction_id"`
	Remark               string                       `json:"remark"`
	TrackingNumber       string                       `json:"tracking_number"`
	LogisticsCompany     string                       `json:"logistics_company"`
	RefundAmount         int64                        `json:"refund_amount"`
	RefundReason         string                       `json:"refund_reason"`
	ShippingAddress      *ShippingAddress             `json:"shipping_address"`
	Items                []*OrderItem                 `json:"items"`
	Logs                 []*OrderLog                  `json:"logs"`
	PaidAt               *time.Time                   `json:"paid_at"`
	ShippedAt            *time.Time                   `json:"shipped_at"`
	DeliveredAt          *time.Time                   `json:"delivered_at"`
	CompletedAt          *time.Time                   `json:"completed_at"`
	CancelledAt          *time.Time                   `json:"cancelled_at"`
	fsm                  *fsm.Machine[string, string] `json:"-"`
}

// OrderItem 实体代表订单中的一个商品项。
type OrderItem struct {
	ID              uint      `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	OrderID         uint64    `json:"order_id"`
	ProductID       uint64    `json:"product_id"`
	SkuID           uint64    `json:"sku_id"`
	ProductName     string    `json:"product_name"`
	SkuName         string    `json:"sku_name"`
	ProductImageURL string    `json:"product_image_url"`
	Price           int64     `json:"price"`
	Quantity        int32     `json:"quantity"`
	TotalPrice      int64     `json:"total_price"`
}

// ShippingAddress 值对象定义了订单的收货地址信息。
type ShippingAddress struct {
	RecipientName   string  `json:"recipient_name"`
	PhoneNumber     string  `json:"phone_number"`
	Province        string  `json:"province"`
	City            string  `json:"city"`
	District        string  `json:"district"`
	DetailedAddress string  `json:"detailed_address"`
	PostalCode      string  `json:"postal_code"`
	Lat             float64 `json:"lat"`
	Lon             float64 `json:"lon"`
}

// OrderLog 值对象定义了订单的操作日志记录。
type OrderLog struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	OrderID   uint64    `json:"order_id"`
	Operator  string    `json:"operator"`
	Action    string    `json:"action"`
	OldStatus string    `json:"old_status"`
	NewStatus string    `json:"new_status"`
	Remark    string    `json:"remark"`
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

// InitFSM 确保状态机已初始化。
func (o *Order) InitFSM() {
	if o.fsm == nil {
		o.initFSM()
	}
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
