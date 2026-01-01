package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/wyfcoding/pkg/fsm"
	"gorm.io/gorm"
)

// OrderStatus 定义了订单的生命周期状态。
type OrderStatus int

const (
	PendingPayment  OrderStatus = 1  // 待支付：订单已创建，等待用户支付。
	Allocating      OrderStatus = 10 // 分配中：新增状态，表示正在进行库存分配（Saga事务中）。
	Paid            OrderStatus = 2  // 已支付：订单已完成支付。
	Shipped         OrderStatus = 3  // 已发货：商品已从仓库发出。
	Delivered       OrderStatus = 4  // 已送达：商品已送达买家。
	Completed       OrderStatus = 5  // 已完成：订单已签收，交易完成。
	Cancelled       OrderStatus = 6  // 已取消：订单被取消。
	RefundRequested OrderStatus = 7  // 退款中：买家已申请退款，等待处理。
	Refunded        OrderStatus = 8  // 已退款：退款已处理完成。
	Closed          OrderStatus = 9  // 已关闭：订单最终关闭（可能因超时未支付或取消）。
)

// String 返回 OrderStatus 的字符串表示。
func (s OrderStatus) String() string {
	names := map[OrderStatus]string{
		PendingPayment:  "PendingPayment",
		Allocating:      "Allocating",
		Paid:            "Paid",
		Shipped:         "Shipped",
		Delivered:       "Delivered",
		Completed:       "Completed",
		Cancelled:       "Cancelled",
		RefundRequested: "RefundRequested",
		Refunded:        "Refunded",
		Closed:          "Closed",
	}
	return names[s] // 根据状态值返回对应的字符串。
}

// Order 实体是订单模块的聚合根。
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
	fsm             *fsm.Machine     `gorm:"-" json:"-"`
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
	RecipientName   string `gorm:"type:varchar(64);comment:收货人姓名" json:"recipient_name"`
	PhoneNumber     string `gorm:"type:varchar(20);comment:手机号" json:"phone_number"`
	Province        string `gorm:"type:varchar(64);comment:省份" json:"province"`
	City            string `gorm:"type:varchar(64);comment:城市" json:"city"`
	District        string `gorm:"type:varchar(64);comment:区县" json:"district"`
	DetailedAddress string `gorm:"type:varchar(255);comment:详细地址" json:"detailed_address"`
	PostalCode      string `gorm:"type:varchar(20);comment:邮政编码" json:"postal_code"`
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

// TimeoutScheduler 定义了超时调度的接口，用于处理订单超时取消等逻辑。
type TimeoutScheduler interface {
	ScheduleTimeout(orderID string, timeout time.Duration, callback func(orderID string)) error
	Start()
	Stop()
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
	order.initFSM()
	return order
}

func (o *Order) initFSM() {
	m := fsm.NewMachine(fsm.State(o.Status.String()))

	// 定义转换规则
	m.AddTransition(fsm.State(PendingPayment.String()), "PAY", fsm.State(Paid.String()))
	m.AddTransition(fsm.State(Paid.String()), "SHIP", fsm.State(Shipped.String()))
	m.AddTransition(fsm.State(Shipped.String()), "DELIVER", fsm.State(Delivered.String()))
	m.AddTransition(fsm.State(Delivered.String()), "COMPLETE", fsm.State(Completed.String()))

	// 取消与退款
	m.AddTransition(fsm.State(PendingPayment.String()), "CANCEL", fsm.State(Cancelled.String()))
	m.AddTransition(fsm.State(Paid.String()), "CANCEL", fsm.State(Cancelled.String()))
	m.AddTransition(fsm.State(Paid.String()), "REFUND_REQ", fsm.State(RefundRequested.String()))
	m.AddTransition(fsm.State(Shipped.String()), "REFUND_REQ", fsm.State(RefundRequested.String()))
	m.AddTransition(fsm.State(Delivered.String()), "REFUND_REQ", fsm.State(RefundRequested.String()))
	m.AddTransition(fsm.State(RefundRequested.String()), "REFUND_APPROVE", fsm.State(Refunded.String()))

	o.fsm = m
}

// AfterFind GORM 钩子，加载后初始化状态机
func (o *Order) AfterFind(tx *gorm.DB) error {
	o.initFSM()
	return nil
}

// Trigger 触发状态变更
func (o *Order) Trigger(ctx context.Context, event string, operator string, remark string, args ...interface{}) error {
	if o.fsm == nil {
		o.initFSM()
	}

	oldStatus := o.Status
	err := o.fsm.Trigger(ctx, fsm.Event(event), args...)
	if err != nil {
		return err
	}

	newStatusStr := string(o.fsm.Current())
	for s, name := range statusNames {
		if name == newStatusStr {
			o.Status = s
			break
		}
	}

	o.AddLog(operator, event, oldStatus.String(), o.Status.String(), remark)
	return nil
}

var statusNames = map[OrderStatus]string{
	PendingPayment:  "PendingPayment",
	Allocating:      "Allocating",
	Paid:            "Paid",
	Shipped:         "Shipped",
	Delivered:       "Delivered",
	Completed:       "Completed",
	Cancelled:       "Cancelled",
	RefundRequested: "RefundRequested",
	Refunded:        "Refunded",
	Closed:          "Closed",
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
