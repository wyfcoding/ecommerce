// 变更说明：新增订单增强功能，包括服务类订单、周期购订单、分期订单、礼品卡订单等高级订单类型。
// 假设：服务类订单支持预约时段管理，周期购支持按周/月/季度配送，分期订单支持最多24期。
package domain

import (
	"context"
	"fmt"
	"time"

	pb "github.com/wyfcoding/ecommerce/goapi/order/v1"
	"github.com/wyfcoding/pkg/fsm"
)

// --- 订单类型扩展常量 ---

// OrderTypeExtended 扩展订单类型
type OrderTypeExtended int

const (
	OrderTypeService      OrderTypeExtended = 10 // 服务类订单：美容、家政、维修等到店/上门服务
	OrderTypeSubscription OrderTypeExtended = 11 // 周期购订单：定期配送商品
	OrderTypeGiftCard     OrderTypeExtended = 12 // 礼品卡订单：购买礼品卡
	OrderTypeInstallment  OrderTypeExtended = 13 // 分期付款订单
	OrderTypeCOD          OrderTypeExtended = 14 // 货到付款订单
	OrderTypeProxy        OrderTypeExtended = 15 // 代购订单：海外代购
)

// --- 服务类订单 ---

// ServiceBooking 服务预约信息值对象
type ServiceBooking struct {
	ID               uint64     `json:"id"`
	OrderID          uint64     `json:"order_id"`
	ServiceType      string     `json:"service_type"`       // 服务类型：到店/上门
	BookingDate      time.Time  `json:"booking_date"`       // 预约日期
	TimeSlotStart    string     `json:"time_slot_start"`    // 预约时段开始（如 "09:00"）
	TimeSlotEnd      string     `json:"time_slot_end"`      // 预约时段结束（如 "11:00"）
	ServiceAddress   string     `json:"service_address"`    // 服务地址（上门服务）
	ServiceLat       float64    `json:"service_lat"`        // 服务地址纬度
	ServiceLon       float64    `json:"service_lon"`        // 服务地址经度
	ServiceStaffID   uint64     `json:"service_staff_id"`   // 服务人员ID
	ServiceStaffName string     `json:"service_staff_name"` // 服务人员姓名
	ConfirmedAt      *time.Time `json:"confirmed_at"`       // 商家确认时间
	ServicedAt       *time.Time `json:"serviced_at"`        // 服务完成时间
	Remark           string     `json:"remark"`             // 预约备注
}

// ServiceBookingStatus 服务预约状态
type ServiceBookingStatus int

const (
	ServiceBookingPending   ServiceBookingStatus = 1 // 待确认
	ServiceBookingConfirmed ServiceBookingStatus = 2 // 已确认
	ServiceBookingInService ServiceBookingStatus = 3 // 服务中
	ServiceBookingCompleted ServiceBookingStatus = 4 // 已完成
	ServiceBookingCancelled ServiceBookingStatus = 5 // 已取消
)

// ServiceOrder 服务类订单聚合
type ServiceOrder struct {
	*Order
	Booking       *ServiceBooking      `json:"booking"`
	BookingStatus ServiceBookingStatus `json:"booking_status"`
}

// NewServiceOrder 创建服务类订单
func NewServiceOrder(orderNo string, userID uint64, items []*OrderItem, booking *ServiceBooking) *ServiceOrder {
	// 服务类订单不需要收货地址，使用服务地址
	order := NewOrder(orderNo, userID, pb.OrderType_ORDER_TYPE_UNSPECIFIED, items, nil)
	order.OrderType = pb.OrderType(OrderTypeService)

	so := &ServiceOrder{
		Order:         order,
		Booking:       booking,
		BookingStatus: ServiceBookingPending,
	}
	so.initServiceFSM()
	return so
}

// initServiceFSM 初始化服务订单状态机
func (so *ServiceOrder) initServiceFSM() {
	m := fsm.NewMachine[string, string](so.Status.String())

	// 服务订单特殊流程
	m.AddTransition(pb.OrderStatus_PENDING_PAYMENT.String(), "PAY", "SERVICE_PENDING_CONFIRM")
	m.AddTransition("SERVICE_PENDING_CONFIRM", "CONFIRM_BOOKING", "SERVICE_CONFIRMED")
	m.AddTransition("SERVICE_CONFIRMED", "START_SERVICE", "SERVICE_IN_PROGRESS")
	m.AddTransition("SERVICE_IN_PROGRESS", "COMPLETE_SERVICE", pb.OrderStatus_COMPLETED.String())
	m.AddTransition("SERVICE_PENDING_CONFIRM", "CANCEL", pb.OrderStatus_CANCELLED.String())
	m.AddTransition("SERVICE_CONFIRMED", "CANCEL", pb.OrderStatus_CANCELLED.String())

	so.fsm = m
}

// ConfirmBooking 商家确认预约
func (so *ServiceOrder) ConfirmBooking(ctx context.Context, operator string) error {
	if err := so.Trigger(ctx, "CONFIRM_BOOKING", operator, "Booking confirmed by merchant"); err != nil {
		return err
	}
	so.BookingStatus = ServiceBookingConfirmed
	now := time.Now()
	so.Booking.ConfirmedAt = &now
	return nil
}

// StartService 开始服务
func (so *ServiceOrder) StartService(ctx context.Context, operator string) error {
	if err := so.Trigger(ctx, "START_SERVICE", operator, "Service started"); err != nil {
		return err
	}
	so.BookingStatus = ServiceBookingInService
	return nil
}

// CompleteService 完成服务
func (so *ServiceOrder) CompleteService(ctx context.Context, operator string) error {
	if err := so.Trigger(ctx, "COMPLETE_SERVICE", operator, "Service completed"); err != nil {
		return err
	}
	so.BookingStatus = ServiceBookingCompleted
	now := time.Now()
	so.Booking.ServicedAt = &now
	so.CompletedAt = &now
	return nil
}

// --- 周期购订单 ---

// SubscriptionCycle 周期类型
type SubscriptionCycle int

const (
	SubscriptionCycleWeekly    SubscriptionCycle = 1 // 每周
	SubscriptionCycleBiweekly  SubscriptionCycle = 2 // 每两周
	SubscriptionCycleMonthly   SubscriptionCycle = 3 // 每月
	SubscriptionCycleQuarterly SubscriptionCycle = 4 // 每季度
)

// SubscriptionPlan 周期购计划值对象
type SubscriptionPlan struct {
	ID               uint64            `json:"id"`
	OrderID          uint64            `json:"order_id"`
	Cycle            SubscriptionCycle `json:"cycle"`              // 配送周期
	TotalPeriods     int32             `json:"total_periods"`      // 总期数，0表示无限
	CompletedPeriods int32             `json:"completed_periods"`  // 已完成期数
	NextDeliveryDate time.Time         `json:"next_delivery_date"` // 下次配送日期
	IsPaused         bool              `json:"is_paused"`          // 是否暂停
	PausedAt         *time.Time        `json:"paused_at"`          // 暂停时间
	ResumedAt        *time.Time        `json:"resumed_at"`         // 恢复时间
	CancelledAt      *time.Time        `json:"cancelled_at"`       // 取消时间
}

// SubscriptionOrder 周期购订单聚合
type SubscriptionOrder struct {
	*Order
	Plan        *SubscriptionPlan `json:"plan"`
	ChildOrders []*Order          `json:"child_orders"` // 子订单列表（每期配送生成一个）
}

// NewSubscriptionOrder 创建周期购订单
func NewSubscriptionOrder(orderNo string, userID uint64, items []*OrderItem, addr *ShippingAddress, plan *SubscriptionPlan) *SubscriptionOrder {
	order := NewOrder(orderNo, userID, pb.OrderType(OrderTypeSubscription), items, addr)

	return &SubscriptionOrder{
		Order:       order,
		Plan:        plan,
		ChildOrders: []*Order{},
	}
}

// GenerateNextDelivery 生成下一期配送订单
func (so *SubscriptionOrder) GenerateNextDelivery(orderNoGenerator func() string) (*Order, error) {
	if so.Plan.IsPaused {
		return nil, fmt.Errorf("subscription is paused")
	}
	if so.Plan.TotalPeriods > 0 && so.Plan.CompletedPeriods >= so.Plan.TotalPeriods {
		return nil, fmt.Errorf("subscription periods completed")
	}

	// 创建子订单
	childOrder := NewOrder(
		orderNoGenerator(),
		so.UserID,
		pb.OrderType_NORMAL,
		so.Items,
		so.ShippingAddress,
	)
	childOrder.Remark = fmt.Sprintf("周期购第%d期", so.Plan.CompletedPeriods+1)

	so.ChildOrders = append(so.ChildOrders, childOrder)
	so.Plan.CompletedPeriods++
	so.updateNextDeliveryDate()

	return childOrder, nil
}

// PauseSubscription 暂停周期购
func (so *SubscriptionOrder) PauseSubscription() {
	so.Plan.IsPaused = true
	now := time.Now()
	so.Plan.PausedAt = &now
}

// ResumeSubscription 恢复周期购
func (so *SubscriptionOrder) ResumeSubscription() {
	so.Plan.IsPaused = false
	now := time.Now()
	so.Plan.ResumedAt = &now
	so.updateNextDeliveryDate()
}

// updateNextDeliveryDate 更新下次配送日期
func (so *SubscriptionOrder) updateNextDeliveryDate() {
	var interval time.Duration
	switch so.Plan.Cycle {
	case SubscriptionCycleWeekly:
		interval = 7 * 24 * time.Hour
	case SubscriptionCycleBiweekly:
		interval = 14 * 24 * time.Hour
	case SubscriptionCycleMonthly:
		interval = 30 * 24 * time.Hour
	case SubscriptionCycleQuarterly:
		interval = 90 * 24 * time.Hour
	}
	so.Plan.NextDeliveryDate = time.Now().Add(interval)
}

// --- 分期付款订单 ---

// InstallmentStatus 分期状态
type InstallmentStatus int

const (
	InstallmentStatusPending   InstallmentStatus = 1 // 待支付
	InstallmentStatusPaid      InstallmentStatus = 2 // 已支付
	InstallmentStatusOverdue   InstallmentStatus = 3 // 已逾期
	InstallmentStatusCancelled InstallmentStatus = 4 // 已取消
)

// InstallmentPlan 分期计划值对象
type InstallmentPlan struct {
	ID            uint64            `json:"id"`
	OrderID       uint64            `json:"order_id"`
	PeriodNumber  int32             `json:"period_number"` // 期数（第几期）
	Amount        int64             `json:"amount"`        // 本期金额（分）
	Principal     int64             `json:"principal"`     // 本金（分）
	Interest      int64             `json:"interest"`      // 利息/手续费（分）
	DueDate       time.Time         `json:"due_date"`      // 还款日期
	PaidAt        *time.Time        `json:"paid_at"`       // 实际支付时间
	Status        InstallmentStatus `json:"status"`
	PaymentMethod string            `json:"payment_method"` // 支付方式
	TransactionID string            `json:"transaction_id"` // 支付交易号
}

// InstallmentOrder 分期订单聚合
type InstallmentOrder struct {
	*Order
	TotalPeriods     int32              `json:"total_periods"` // 总分期数
	InterestRate     float64            `json:"interest_rate"` // 利率（年化）
	InstallmentPlans []*InstallmentPlan `json:"installment_plans"`
}

// NewInstallmentOrder 创建分期订单
func NewInstallmentOrder(orderNo string, userID uint64, items []*OrderItem, addr *ShippingAddress, totalPeriods int32, interestRate float64) *InstallmentOrder {
	order := NewOrder(orderNo, userID, pb.OrderType(OrderTypeInstallment), items, addr)

	io := &InstallmentOrder{
		Order:        order,
		TotalPeriods: totalPeriods,
		InterestRate: interestRate,
	}
	io.generateInstallmentPlans()
	return io
}

// generateInstallmentPlans 生成分期计划
func (io *InstallmentOrder) generateInstallmentPlans() {
	totalAmount := io.ActualAmount
	principalPerPeriod := totalAmount / int64(io.TotalPeriods)
	remainder := totalAmount - principalPerPeriod*int64(io.TotalPeriods)

	io.InstallmentPlans = make([]*InstallmentPlan, io.TotalPeriods)
	for i := int32(0); i < io.TotalPeriods; i++ {
		principal := principalPerPeriod
		if i == io.TotalPeriods-1 {
			principal += remainder // 最后一期包含余数
		}

		// 简化利息计算：月利率 = 年利率 / 12
		interest := int64(float64(principal) * io.InterestRate / 12)

		io.InstallmentPlans[i] = &InstallmentPlan{
			PeriodNumber: i + 1,
			Principal:    principal,
			Interest:     interest,
			Amount:       principal + interest,
			DueDate:      time.Now().AddDate(0, int(i+1), 0), // 每月一期
			Status:       InstallmentStatusPending,
		}
	}
}

// PayInstallment 支付某一期
func (io *InstallmentOrder) PayInstallment(periodNumber int32, paymentMethod, transactionID string) error {
	if periodNumber < 1 || periodNumber > io.TotalPeriods {
		return fmt.Errorf("invalid period number: %d", periodNumber)
	}

	plan := io.InstallmentPlans[periodNumber-1]
	if plan.Status == InstallmentStatusPaid {
		return fmt.Errorf("installment already paid")
	}

	plan.Status = InstallmentStatusPaid
	now := time.Now()
	plan.PaidAt = &now
	plan.PaymentMethod = paymentMethod
	plan.TransactionID = transactionID

	// 检查是否所有分期都已支付
	allPaid := true
	for _, p := range io.InstallmentPlans {
		if p.Status != InstallmentStatusPaid {
			allPaid = false
			break
		}
	}
	if allPaid {
		io.PaymentStatus = pb.PaymentStatus_SUCCESS
	}

	return nil
}

// GetOverdueInstallments 获取逾期分期列表
func (io *InstallmentOrder) GetOverdueInstallments() []*InstallmentPlan {
	var overdue []*InstallmentPlan
	now := time.Now()
	for _, plan := range io.InstallmentPlans {
		if plan.Status == InstallmentStatusPending && plan.DueDate.Before(now) {
			plan.Status = InstallmentStatusOverdue
			overdue = append(overdue, plan)
		}
	}
	return overdue
}

// --- 礼品卡订单 ---

// GiftCardStatus 礼品卡状态
type GiftCardStatus int

const (
	GiftCardStatusInactive  GiftCardStatus = 1 // 未激活
	GiftCardStatusActive    GiftCardStatus = 2 // 已激活
	GiftCardStatusUsed      GiftCardStatus = 3 // 已使用完
	GiftCardStatusExpired   GiftCardStatus = 4 // 已过期
	GiftCardStatusCancelled GiftCardStatus = 5 // 已作废
)

// GiftCard 礼品卡值对象
type GiftCard struct {
	ID           uint64         `json:"id"`
	OrderID      uint64         `json:"order_id"`
	CardNo       string         `json:"card_no"`       // 卡号
	CardPassword string         `json:"card_password"` // 卡密（加密存储）
	FaceValue    int64          `json:"face_value"`    // 面值（分）
	Balance      int64          `json:"balance"`       // 余额（分）
	OwnerUserID  uint64         `json:"owner_user_id"` // 持有人用户ID（绑定后）
	Status       GiftCardStatus `json:"status"`
	ExpiryDate   time.Time      `json:"expiry_date"`  // 有效期
	ActivatedAt  *time.Time     `json:"activated_at"` // 激活时间
	BoundAt      *time.Time     `json:"bound_at"`     // 绑定时间
}

// GiftCardOrder 礼品卡订单聚合
type GiftCardOrder struct {
	*Order
	GiftCards      []*GiftCard `json:"gift_cards"`      // 购买的礼品卡列表
	RecipientEmail string      `json:"recipient_email"` // 接收人邮箱（电子礼品卡）
	RecipientPhone string      `json:"recipient_phone"` // 接收人手机
	GiftMessage    string      `json:"gift_message"`    // 祝福语
}

// NewGiftCardOrder 创建礼品卡订单
func NewGiftCardOrder(orderNo string, userID uint64, faceValue int64, quantity int32, expiryDate time.Time) *GiftCardOrder {
	// 礼品卡作为虚拟商品
	item := &OrderItem{
		ProductName: "礼品卡",
		ProductType: pb.ProductType_VIRTUAL,
		Price:       faceValue,
		Quantity:    quantity,
		TotalPrice:  faceValue * int64(quantity),
	}

	order := NewOrder(orderNo, userID, pb.OrderType(OrderTypeGiftCard), []*OrderItem{item}, nil)

	gco := &GiftCardOrder{
		Order:     order,
		GiftCards: make([]*GiftCard, quantity),
	}

	// 预生成礼品卡（支付后激活）
	for i := int32(0); i < quantity; i++ {
		gco.GiftCards[i] = &GiftCard{
			FaceValue:  faceValue,
			Balance:    faceValue,
			Status:     GiftCardStatusInactive,
			ExpiryDate: expiryDate,
		}
	}

	return gco
}

// ActivateGiftCards 激活礼品卡（支付成功后调用）
func (gco *GiftCardOrder) ActivateGiftCards(cardNoGenerator func() string, passwordGenerator func() string) {
	now := time.Now()
	for _, card := range gco.GiftCards {
		card.CardNo = cardNoGenerator()
		card.CardPassword = passwordGenerator()
		card.Status = GiftCardStatusActive
		card.ActivatedAt = &now
	}
}

// BindGiftCard 绑定礼品卡到用户账户
func (gc *GiftCard) BindGiftCard(userID uint64, cardPassword string) error {
	if gc.CardPassword != cardPassword {
		return fmt.Errorf("invalid card password")
	}
	if gc.Status != GiftCardStatusActive {
		return fmt.Errorf("gift card is not active")
	}
	if gc.OwnerUserID != 0 {
		return fmt.Errorf("gift card already bound")
	}

	gc.OwnerUserID = userID
	now := time.Now()
	gc.BoundAt = &now
	return nil
}

// UseGiftCard 使用礼品卡
func (gc *GiftCard) UseGiftCard(amount int64) error {
	if gc.Status != GiftCardStatusActive {
		return fmt.Errorf("gift card is not active")
	}
	if gc.Balance < amount {
		return fmt.Errorf("insufficient gift card balance")
	}
	if time.Now().After(gc.ExpiryDate) {
		gc.Status = GiftCardStatusExpired
		return fmt.Errorf("gift card has expired")
	}

	gc.Balance -= amount
	if gc.Balance == 0 {
		gc.Status = GiftCardStatusUsed
	}
	return nil
}

// --- 货到付款订单 ---

// CODStatus 货到付款状态
type CODStatus int

const (
	CODStatusPending   CODStatus = 1 // 待配送
	CODStatusDelivered CODStatus = 2 // 已送达待付款
	CODStatusCollected CODStatus = 3 // 已收款
	CODStatusRefused   CODStatus = 4 // 拒收
	CODStatusReturned  CODStatus = 5 // 已退回
)

// CODOrder 货到付款订单聚合
type CODOrder struct {
	*Order
	CODStatus       CODStatus  `json:"cod_status"`
	CollectedAmount int64      `json:"collected_amount"` // 实际收款金额
	CollectedAt     *time.Time `json:"collected_at"`     // 收款时间
	CollectorID     string     `json:"collector_id"`     // 收款人（配送员）ID
	CollectorName   string     `json:"collector_name"`   // 收款人姓名
	RefusedReason   string     `json:"refused_reason"`   // 拒收原因
}

// NewCODOrder 创建货到付款订单
func NewCODOrder(orderNo string, userID uint64, items []*OrderItem, addr *ShippingAddress) *CODOrder {
	order := NewOrder(orderNo, userID, pb.OrderType(OrderTypeCOD), items, addr)
	order.PaymentMethod = "COD"

	return &CODOrder{
		Order:     order,
		CODStatus: CODStatusPending,
	}
}

// MarkDelivered 标记已送达（等待付款）
func (co *CODOrder) MarkDelivered(ctx context.Context, operator string) error {
	if co.CODStatus != CODStatusPending {
		return fmt.Errorf("invalid COD status for delivery")
	}
	co.CODStatus = CODStatusDelivered
	co.ShippingStatus = pb.ShippingStatus_SHIPPING_DELIVERED
	now := time.Now()
	co.DeliveredAt = &now
	co.AddLog(operator, "COD_DELIVERED", co.Status.String(), co.Status.String(), "Package delivered, awaiting payment")
	return nil
}

// CollectPayment 收取货款
func (co *CODOrder) CollectPayment(ctx context.Context, collectorID, collectorName string, amount int64) error {
	if co.CODStatus != CODStatusDelivered {
		return fmt.Errorf("can only collect payment after delivery")
	}
	if amount != co.ActualAmount {
		return fmt.Errorf("collected amount mismatch: expected %d, got %d", co.ActualAmount, amount)
	}

	co.CODStatus = CODStatusCollected
	co.CollectedAmount = amount
	now := time.Now()
	co.CollectedAt = &now
	co.PaidAt = &now
	co.CollectorID = collectorID
	co.CollectorName = collectorName
	co.PaymentStatus = pb.PaymentStatus_SUCCESS
	co.Status = pb.OrderStatus_COMPLETED
	co.CompletedAt = &now

	co.AddLog(collectorName, "COD_COLLECTED", co.Status.String(), pb.OrderStatus_COMPLETED.String(), fmt.Sprintf("Payment collected: %d", amount))
	return nil
}

// RefuseDelivery 拒收
func (co *CODOrder) RefuseDelivery(ctx context.Context, reason string, operator string) error {
	if co.CODStatus != CODStatusDelivered && co.CODStatus != CODStatusPending {
		return fmt.Errorf("invalid status for refuse")
	}

	co.CODStatus = CODStatusRefused
	co.RefusedReason = reason
	co.Status = pb.OrderStatus_CANCELLED
	now := time.Now()
	co.CancelledAt = &now

	co.AddLog(operator, "COD_REFUSED", co.Status.String(), pb.OrderStatus_CANCELLED.String(), reason)
	return nil
}

// --- 代购订单 ---

// ProxyPurchaseStatus 代购状态
type ProxyPurchaseStatus int

const (
	ProxyStatusPending    ProxyPurchaseStatus = 1 // 待采购
	ProxyStatusPurchasing ProxyPurchaseStatus = 2 // 采购中
	ProxyStatusPurchased  ProxyPurchaseStatus = 3 // 已采购
	ProxyStatusShipping   ProxyPurchaseStatus = 4 // 国际运输中
	ProxyStatusCustoms    ProxyPurchaseStatus = 5 // 清关中
	ProxyStatusDomestic   ProxyPurchaseStatus = 6 // 国内配送中
	ProxyStatusDelivered  ProxyPurchaseStatus = 7 // 已送达
)

// ProxyPurchaseOrder 代购订单聚合
type ProxyPurchaseOrder struct {
	*Order
	ProxyStatus          ProxyPurchaseStatus `json:"proxy_status"`
	PurchaseCountry      string              `json:"purchase_country"`       // 采购国家
	PurchaseCurrency     string              `json:"purchase_currency"`      // 采购币种
	PurchasePrice        int64               `json:"purchase_price"`         // 采购价（原币种，分）
	ExchangeRate         float64             `json:"exchange_rate"`          // 汇率
	ServiceFee           int64               `json:"service_fee"`            // 代购服务费（分）
	InternationalFee     int64               `json:"international_fee"`      // 国际运费（分）
	CustomsDuty          int64               `json:"customs_duty"`           // 关税（分）
	InternationalTrackNo string              `json:"international_track_no"` // 国际物流单号
	CustomsStatus        string              `json:"customs_status"`         // 清关状态
	EstimatedArrival     *time.Time          `json:"estimated_arrival"`      // 预计到货时间
}

// NewProxyPurchaseOrder 创建代购订单
func NewProxyPurchaseOrder(orderNo string, userID uint64, items []*OrderItem, addr *ShippingAddress, purchaseCountry, purchaseCurrency string, purchasePrice int64, exchangeRate float64, serviceFee, internationalFee int64) *ProxyPurchaseOrder {
	order := NewOrder(orderNo, userID, pb.OrderType(OrderTypeProxy), items, addr)

	// 计算总价 = 采购价 * 汇率 + 服务费 + 国际运费
	totalInCNY := int64(float64(purchasePrice)*exchangeRate) + serviceFee + internationalFee
	order.TotalAmount = totalInCNY
	order.ActualAmount = totalInCNY

	return &ProxyPurchaseOrder{
		Order:            order,
		ProxyStatus:      ProxyStatusPending,
		PurchaseCountry:  purchaseCountry,
		PurchaseCurrency: purchaseCurrency,
		PurchasePrice:    purchasePrice,
		ExchangeRate:     exchangeRate,
		ServiceFee:       serviceFee,
		InternationalFee: internationalFee,
	}
}

// StartPurchasing 开始采购
func (po *ProxyPurchaseOrder) StartPurchasing(ctx context.Context, operator string) error {
	if po.ProxyStatus != ProxyStatusPending {
		return fmt.Errorf("invalid proxy status")
	}
	po.ProxyStatus = ProxyStatusPurchasing
	po.AddLog(operator, "PROXY_PURCHASING", po.Status.String(), po.Status.String(), "Started purchasing overseas")
	return nil
}

// ConfirmPurchased 确认采购完成
func (po *ProxyPurchaseOrder) ConfirmPurchased(ctx context.Context, internationalTrackNo string, operator string) error {
	if po.ProxyStatus != ProxyStatusPurchasing {
		return fmt.Errorf("invalid proxy status")
	}
	po.ProxyStatus = ProxyStatusPurchased
	po.InternationalTrackNo = internationalTrackNo
	po.AddLog(operator, "PROXY_PURCHASED", po.Status.String(), po.Status.String(), fmt.Sprintf("International tracking: %s", internationalTrackNo))
	return nil
}

// UpdateCustomsStatus 更新清关状态
func (po *ProxyPurchaseOrder) UpdateCustomsStatus(ctx context.Context, status string, duty int64, operator string) error {
	po.ProxyStatus = ProxyStatusCustoms
	po.CustomsStatus = status
	po.CustomsDuty = duty
	// 更新总价（加上关税）
	po.ActualAmount += duty
	po.AddLog(operator, "CUSTOMS_UPDATE", po.Status.String(), po.Status.String(), fmt.Sprintf("Customs: %s, Duty: %d", status, duty))
	return nil
}
