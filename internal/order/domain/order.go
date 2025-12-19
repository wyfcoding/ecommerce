package domain

import (
	"fmt"  // 导入格式化库，用于错误信息。
	"time" // 导入时间库。

	"gorm.io/gorm" // 导入GORM库。
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
// 它代表一个订单的完整信息，包含了订单编号、用户、商品项、状态和日志等。
type Order struct {
	gorm.Model                       // 嵌入gorm.Model，包含ID, CreatedAt, UpdatedAt, DeletedAt等通用字段。
	OrderNo         string           `gorm:"type:varchar(64);uniqueIndex;not null;comment:订单编号" json:"order_no"` // 订单编号，唯一索引，不允许为空。
	UserID          uint64           `gorm:"index;not null;comment:用户ID" json:"user_id"`                         // 用户ID，索引字段。
	Status          OrderStatus      `gorm:"type:tinyint;not null;default:1;comment:订单状态" json:"status"`         // 订单状态，默认为待支付。
	TotalAmount     int64            `gorm:"not null;comment:订单总金额(分)" json:"total_amount"`                      // 订单商品总金额（单位：分）。
	ActualAmount    int64            `gorm:"not null;comment:实际支付金额(分)" json:"actual_amount"`                    // 用户实际支付的金额（单位：分）。
	ShippingFee     int64            `gorm:"not null;default:0;comment:运费(分)" json:"shipping_fee"`               // 运费（单位：分）。
	DiscountAmount  int64            `gorm:"not null;default:0;comment:优惠金额(分)" json:"discount_amount"`          // 订单优惠金额（单位：分）。
	PaymentMethod   string           `gorm:"type:varchar(32);comment:支付方式" json:"payment_method"`                // 支付方式。
	Remark          string           `gorm:"type:varchar(255);comment:订单备注" json:"remark"`                       // 订单备注。
	ShippingAddress *ShippingAddress `gorm:"embedded;embeddedPrefix:shipping_" json:"shipping_address"`          // 嵌入的收货地址信息。
	Items           []*OrderItem     `gorm:"foreignKey:OrderID" json:"items"`                                    // 关联的订单商品项列表，一对多关系。
	Logs            []*OrderLog      `gorm:"foreignKey:OrderID" json:"logs"`                                     // 关联的订单操作日志列表，一对多关系。
	PaidAt          *time.Time       `gorm:"comment:支付时间" json:"paid_at"`                                        // 支付时间。
	ShippedAt       *time.Time       `gorm:"comment:发货时间" json:"shipped_at"`                                     // 发货时间。
	DeliveredAt     *time.Time       `gorm:"comment:送达时间" json:"delivered_at"`                                   // 送达时间。
	CompletedAt     *time.Time       `gorm:"comment:完成时间" json:"completed_at"`                                   // 完成时间。
	CancelledAt     *time.Time       `gorm:"comment:取消时间" json:"cancelled_at"`                                   // 取消时间。
}

// OrderItem 实体代表订单中的一个商品项。
type OrderItem struct {
	gorm.Model             // 嵌入gorm.Model。
	OrderID         uint64 `gorm:"index;not null;comment:订单ID" json:"order_id"`                 // 关联的订单ID，索引字段。
	ProductID       uint64 `gorm:"not null;comment:商品ID" json:"product_id"`                     // 商品ID。
	SkuID           uint64 `gorm:"not null;comment:SKU ID" json:"sku_id"`                       // 商品SKU ID。
	ProductName     string `gorm:"type:varchar(255);not null;comment:商品名称" json:"product_name"` // 商品名称。
	SkuName         string `gorm:"type:varchar(255);not null;comment:SKU名称" json:"sku_name"`    // SKU名称。
	ProductImageURL string `gorm:"type:varchar(255);comment:商品图片URL" json:"product_image_url"`  // 商品图片URL。
	Price           int64  `gorm:"not null;comment:单价(分)" json:"price"`                         // 商品单价（单位：分）。
	Quantity        int32  `gorm:"not null;comment:数量" json:"quantity"`                         // 购买数量。
	TotalPrice      int64  `gorm:"not null;comment:总价(分)" json:"total_price"`                   // 商品总价（单价 * 数量，单位：分）。
}

// ShippingAddress 值对象定义了订单的收货地址信息。
// 作为一个嵌入式结构，它不会有自己的ID和CreatedAt/UpdatedAt字段。
type ShippingAddress struct {
	RecipientName   string `gorm:"type:varchar(64);comment:收货人姓名" json:"recipient_name"`   // 收货人姓名。
	PhoneNumber     string `gorm:"type:varchar(20);comment:手机号" json:"phone_number"`       // 收货人手机号。
	Province        string `gorm:"type:varchar(64);comment:省份" json:"province"`            // 省份。
	City            string `gorm:"type:varchar(64);comment:城市" json:"city"`                // 城市。
	District        string `gorm:"type:varchar(64);comment:区县" json:"district"`            // 区县。
	DetailedAddress string `gorm:"type:varchar(255);comment:详细地址" json:"detailed_address"` // 详细地址。
	PostalCode      string `gorm:"type:varchar(20);comment:邮政编码" json:"postal_code"`       // 邮政编码。
}

// OrderLog 值对象定义了订单的操作日志记录。
type OrderLog struct {
	gorm.Model        // 嵌入gorm.Model。
	OrderID    uint64 `gorm:"index;not null;comment:订单ID" json:"order_id"`           // 关联的订单ID，索引字段。
	Operator   string `gorm:"type:varchar(64);not null;comment:操作人" json:"operator"` // 操作人员或系统。
	Action     string `gorm:"type:varchar(64);not null;comment:操作动作" json:"action"`  // 操作描述。
	OldStatus  string `gorm:"type:varchar(32);comment:旧状态" json:"old_status"`        // 操作前的订单状态。
	NewStatus  string `gorm:"type:varchar(32);comment:新状态" json:"new_status"`        // 操作后的订单状态。
	Remark     string `gorm:"type:varchar(255);comment:备注" json:"remark"`            // 备注信息。
}

// NewOrder 创建并返回一个新的 Order 实体实例。
// orderNo: 订单编号。
// userID: 用户ID。
// items: 订单商品项列表。
// shippingAddr: 收货地址。
func NewOrder(orderNo string, userID uint64, items []*OrderItem, shippingAddr *ShippingAddress) *Order {
	var totalAmount int64
	for _, item := range items {
		item.TotalPrice = item.Price * int64(item.Quantity) // 计算每个商品项的总价。
		totalAmount += item.TotalPrice                      // 累加计算订单总金额。
	}

	order := &Order{
		OrderNo:         orderNo,
		UserID:          userID,
		Status:          PendingPayment, // 新订单的初始状态为待支付。
		TotalAmount:     totalAmount,
		ActualAmount:    totalAmount, // 实际支付金额默认与总金额相同。
		ShippingFee:     0,           // 默认运费为0。
		DiscountAmount:  0,           // 默认优惠金额为0。
		ShippingAddress: shippingAddr,
		Items:           items,
		Logs:            []*OrderLog{}, // 初始化订单日志列表。
	}

	// 添加订单创建日志。
	order.AddLog("System", "Order Created", "", PendingPayment.String(), "Initial order creation")
	return order
}

// CanPay 检查订单是否可以进行支付操作。
func (o *Order) CanPay() error {
	if o.Status != PendingPayment {
		return fmt.Errorf("order status must be PendingPayment, current: %s", o.Status.String())
	}
	return nil
}

// Pay 支付订单。
// paymentMethod: 支付方式。
// operator: 操作人。
func (o *Order) Pay(paymentMethod string, operator string) error {
	if err := o.CanPay(); err != nil {
		return err
	}

	oldStatus := o.Status
	o.Status = Paid // 状态变更为已支付。
	o.PaymentMethod = paymentMethod
	now := time.Now()
	o.PaidAt = &now // 记录支付时间。

	o.AddLog(operator, "Order Paid", oldStatus.String(), Paid.String(), fmt.Sprintf("Payment method: %s", paymentMethod))
	return nil
}

// CanShip 检查订单是否可以进行发货操作。
func (o *Order) CanShip() error {
	if o.Status != Paid {
		return fmt.Errorf("order status must be Paid, current: %s", o.Status.String())
	}
	return nil
}

// Ship 发货订单。
// operator: 操作人。
func (o *Order) Ship(operator string) error {
	if err := o.CanShip(); err != nil {
		return err
	}

	oldStatus := o.Status
	o.Status = Shipped // 状态变更为已发货。
	now := time.Now()
	o.ShippedAt = &now // 记录发货时间。

	o.AddLog(operator, "Order Shipped", oldStatus.String(), Shipped.String(), "Order has been shipped")
	return nil
}

// CanDeliver 检查订单是否可以进行送达操作。
func (o *Order) CanDeliver() error {
	if o.Status != Shipped {
		return fmt.Errorf("order status must be Shipped, current: %s", o.Status.String())
	}
	return nil
}

// Deliver 送达订单。
// operator: 操作人。
func (o *Order) Deliver(operator string) error {
	if err := o.CanDeliver(); err != nil {
		return err
	}

	oldStatus := o.Status
	o.Status = Delivered // 状态变更为已送达。
	now := time.Now()
	o.DeliveredAt = &now // 记录送达时间。

	o.AddLog(operator, "Order Delivered", oldStatus.String(), Delivered.String(), "Order has been delivered")
	return nil
}

// CanComplete 检查订单是否可以进行完成操作。
func (o *Order) CanComplete() error {
	if o.Status != Delivered {
		return fmt.Errorf("order status must be Delivered, current: %s", o.Status.String())
	}
	return nil
}

// Complete 完成订单。
// operator: 操作人。
func (o *Order) Complete(operator string) error {
	if err := o.CanComplete(); err != nil {
		return err
	}

	oldStatus := o.Status
	o.Status = Completed // 状态变更为已完成。
	now := time.Now()
	o.CompletedAt = &now // 记录完成时间。

	o.AddLog(operator, "Order Completed", oldStatus.String(), Completed.String(), "Order has been completed")
	return nil
}

// CanCancel 检查订单是否可以进行取消操作。
func (o *Order) CanCancel() error {
	// 只有待支付和已支付状态的订单可以取消。
	if o.Status != PendingPayment && o.Status != Paid {
		return fmt.Errorf("order cannot be cancelled in current status: %s", o.Status.String())
	}
	return nil
}

// Cancel 取消订单。
// operator: 操作人。
// reason: 取消原因。
func (o *Order) Cancel(operator, reason string) error {
	if err := o.CanCancel(); err != nil {
		return err
	}

	oldStatus := o.Status
	o.Status = Cancelled // 状态变更为已取消。
	now := time.Now()
	o.CancelledAt = &now // 记录取消时间。

	o.AddLog(operator, "Order Cancelled", oldStatus.String(), Cancelled.String(), reason)
	return nil
}

// CanRequestRefund 检查订单是否可以进行申请退款操作。
func (o *Order) CanRequestRefund() error {
	// 只有已支付、已发货和已送达状态的订单可以申请退款。
	if o.Status != Paid && o.Status != Shipped && o.Status != Delivered {
		return fmt.Errorf("order cannot request refund in current status: %s", o.Status.String())
	}
	return nil
}

// RequestRefund 申请退款。
// operator: 操作人。
// reason: 退款原因。
func (o *Order) RequestRefund(operator, reason string) error {
	if err := o.CanRequestRefund(); err != nil {
		return err
	}

	oldStatus := o.Status
	o.Status = RefundRequested // 状态变更为退款中。

	o.AddLog(operator, "Refund Requested", oldStatus.String(), RefundRequested.String(), reason)
	return nil
}

// ApproveRefund 批准退款。
// operator: 操作人。
func (o *Order) ApproveRefund(operator string) error {
	if o.Status != RefundRequested {
		return fmt.Errorf("order status must be RefundRequested, current: %s", o.Status.String())
	}

	oldStatus := o.Status
	o.Status = Refunded // 状态变更为已退款。

	o.AddLog(operator, "Refund Approved", oldStatus.String(), Refunded.String(), "Refund has been approved")
	return nil
}

// ApplyDiscount 应用折扣。
// discountAmount: 折扣金额。
// operator: 操作人。
// reason: 折扣原因。
func (o *Order) ApplyDiscount(discountAmount int64, operator, reason string) error {
	if discountAmount < 0 {
		return fmt.Errorf("discount amount must be positive")
	}
	if discountAmount > o.TotalAmount {
		return fmt.Errorf("discount amount cannot exceed total amount")
	}

	o.DiscountAmount = discountAmount               // 设置优惠金额。
	o.ActualAmount = o.TotalAmount - discountAmount // 重新计算实际支付金额。

	o.AddLog(operator, "Discount Applied", "", "", fmt.Sprintf("Discount: %d, Reason: %s", discountAmount, reason))
	return nil
}

// AddLog 添加订单操作日志。
// operator: 操作人。
// action: 操作动作。
// oldStatus: 旧状态。
// newStatus: 新状态。
// remark: 备注信息。
func (o *Order) AddLog(operator, action, oldStatus, newStatus, remark string) {
	log := &OrderLog{
		Operator:  operator,
		Action:    action,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		Remark:    remark,
	}
	o.Logs = append(o.Logs, log) // 将新日志添加到日志列表中。
}

// GetTotalQuantity 获取订单中所有商品的总数量。
func (o *Order) GetTotalQuantity() int32 {
	var total int32
	for _, item := range o.Items {
		total += item.Quantity
	}
	return total
}
