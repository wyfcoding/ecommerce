// 变更说明：新增组合支付功能，支持多种支付方式组合（余额+银行卡、优惠券+现金、礼品卡+现金、积分抵扣）。
// 假设：组合支付扣款顺序为 积分 → 优惠券 → 礼品卡 → 余额 → 银行卡。
package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/payment/v1"
	"github.com/wyfcoding/pkg/eventsourcing"
	"github.com/wyfcoding/pkg/idgen"
)

// --- 支付组件类型 ---

// PaymentComponentType 支付组件类型
type PaymentComponentType int

const (
	PaymentComponentBalance  PaymentComponentType = 1 // 余额支付
	PaymentComponentCard     PaymentComponentType = 2 // 银行卡支付
	PaymentComponentCoupon   PaymentComponentType = 3 // 优惠券抵扣
	PaymentComponentPoints   PaymentComponentType = 4 // 积分抵扣
	PaymentComponentGiftCard PaymentComponentType = 5 // 礼品卡支付
	PaymentComponentCredit   PaymentComponentType = 6 // 信用支付（先用后付）
	PaymentComponentWallet   PaymentComponentType = 7 // 第三方钱包（支付宝/微信）
)

// PaymentComponentStatus 支付组件状态
type PaymentComponentStatus int

const (
	ComponentStatusPending    PaymentComponentStatus = 1 // 待支付
	ComponentStatusProcessing PaymentComponentStatus = 2 // 支付中
	ComponentStatusSuccess    PaymentComponentStatus = 3 // 支付成功
	ComponentStatusFailed     PaymentComponentStatus = 4 // 支付失败
	ComponentStatusRefunded   PaymentComponentStatus = 5 // 已退款
)

// --- 组合支付 ---

// PaymentComponent 支付组件
type PaymentComponent struct {
	ID             uint64                 `json:"id"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	ComboPaymentID uint64                 `json:"combo_payment_id"`
	Type           PaymentComponentType   `json:"type"`   // 支付组件类型
	Amount         int64                  `json:"amount"` // 该组件支付金额（分）
	RefID          string                 `json:"ref_id"` // 关联ID（卡号、券码、积分账户等）
	Status         PaymentComponentStatus `json:"status"`
	TransactionID  string                 `json:"transaction_id"` // 第三方交易号
	FailureReason  string                 `json:"failure_reason"` // 失败原因
	Priority       int                    `json:"priority"`       // 扣款优先级（数字越小越优先）
	PaidAt         *time.Time             `json:"paid_at"`
}

// ComboPayment 组合支付聚合根
type ComboPayment struct {
	eventsourcing.AggregateRoot
	ID             uint64              `json:"id"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
	ComboPaymentNo string              `json:"combo_payment_no"` // 组合支付单号
	OrderID        uint64              `json:"order_id"`
	OrderNo        string              `json:"order_no"`
	UserID         uint64              `json:"user_id"`
	TotalAmount    int64               `json:"total_amount"` // 订单总金额
	PaidAmount     int64               `json:"paid_amount"`  // 已支付金额
	Components     []*PaymentComponent `json:"components"`   // 支付组件列表
	Status         pb.PaymentStatus    `json:"status"`
	PersistenceVer int64               `json:"version"`
}

// NewComboPayment 创建组合支付
func NewComboPayment(orderID uint64, orderNo string, userID uint64, totalAmount int64, idGenerator idgen.Generator) *ComboPayment {
	comboNo := fmt.Sprintf("COMBO%d", idGenerator.Generate())

	cp := &ComboPayment{
		OrderID:        orderID,
		OrderNo:        orderNo,
		UserID:         userID,
		TotalAmount:    totalAmount,
		PaidAmount:     0,
		Components:     make([]*PaymentComponent, 0),
		Status:         pb.PaymentStatus_PENDING,
		PersistenceVer: 1,
	}
	cp.SetID(comboNo)
	cp.ComboPaymentNo = comboNo
	return cp
}

// AddComponent 添加支付组件
func (cp *ComboPayment) AddComponent(compType PaymentComponentType, amount int64, refID string, priority int) error {
	if amount <= 0 {
		return errors.New("component amount must be positive")
	}

	// 检查是否超过待支付金额
	remainingAmount := cp.TotalAmount - cp.getPendingAmount()
	if amount > remainingAmount {
		return fmt.Errorf("component amount %d exceeds remaining amount %d", amount, remainingAmount)
	}

	component := &PaymentComponent{
		Type:     compType,
		Amount:   amount,
		RefID:    refID,
		Status:   ComponentStatusPending,
		Priority: priority,
	}
	cp.Components = append(cp.Components, component)
	return nil
}

// getPendingAmount 获取待支付组件总金额
func (cp *ComboPayment) getPendingAmount() int64 {
	var pending int64
	for _, comp := range cp.Components {
		if comp.Status == ComponentStatusPending || comp.Status == ComponentStatusProcessing {
			pending += comp.Amount
		}
	}
	return pending
}

// ProcessPayment 执行组合支付
func (cp *ComboPayment) ProcessPayment(ctx context.Context, processors map[PaymentComponentType]ComponentProcessor) error {
	// 按优先级排序组件
	sortedComponents := cp.getSortedComponents()

	for _, comp := range sortedComponents {
		processor, ok := processors[comp.Type]
		if !ok {
			comp.Status = ComponentStatusFailed
			comp.FailureReason = "no processor found"
			continue
		}

		comp.Status = ComponentStatusProcessing
		result, err := processor.Process(ctx, comp)
		if err != nil {
			comp.Status = ComponentStatusFailed
			comp.FailureReason = err.Error()
			// 组合支付中任一组件失败，需要回滚已成功的组件
			cp.Status = pb.PaymentStatus_FAILED
			return cp.rollbackSuccessComponents(ctx, processors)
		}

		comp.Status = ComponentStatusSuccess
		comp.TransactionID = result.TransactionID
		now := time.Now()
		comp.PaidAt = &now
		cp.PaidAmount += comp.Amount
	}

	// 检查是否全部支付完成
	if cp.PaidAmount >= cp.TotalAmount {
		cp.Status = pb.PaymentStatus_SUCCESS
	}
	return nil
}

// rollbackSuccessComponents 回滚已成功的组件
func (cp *ComboPayment) rollbackSuccessComponents(ctx context.Context, processors map[PaymentComponentType]ComponentProcessor) error {
	var rollbackErrors []error
	for _, comp := range cp.Components {
		if comp.Status == ComponentStatusSuccess {
			processor := processors[comp.Type]
			if err := processor.Rollback(ctx, comp); err != nil {
				rollbackErrors = append(rollbackErrors, err)
			} else {
				comp.Status = ComponentStatusRefunded
			}
		}
	}
	if len(rollbackErrors) > 0 {
		return fmt.Errorf("rollback errors: %v", rollbackErrors)
	}
	return nil
}

// getSortedComponents 按优先级排序组件
func (cp *ComboPayment) getSortedComponents() []*PaymentComponent {
	// 简单冒泡排序
	sorted := make([]*PaymentComponent, len(cp.Components))
	copy(sorted, cp.Components)
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-1-i; j++ {
			if sorted[j].Priority > sorted[j+1].Priority {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}
	return sorted
}

// RefundComponent 退款单个组件
func (cp *ComboPayment) RefundComponent(ctx context.Context, componentID uint64, processor ComponentProcessor) error {
	for _, comp := range cp.Components {
		if comp.ID == componentID {
			if comp.Status != ComponentStatusSuccess {
				return errors.New("can only refund success component")
			}
			if err := processor.Rollback(ctx, comp); err != nil {
				return err
			}
			comp.Status = ComponentStatusRefunded
			cp.PaidAmount -= comp.Amount
			return nil
		}
	}
	return errors.New("component not found")
}

// --- 支付组件处理器接口 ---

// ComponentProcessResult 组件处理结果
type ComponentProcessResult struct {
	TransactionID string
	Success       bool
	Message       string
}

// ComponentProcessor 支付组件处理器接口
type ComponentProcessor interface {
	// Process 处理支付
	Process(ctx context.Context, component *PaymentComponent) (*ComponentProcessResult, error)
	// Rollback 回滚支付
	Rollback(ctx context.Context, component *PaymentComponent) error
}

// --- 余额支付处理器 ---

// BalanceProcessor 余额支付处理器
type BalanceProcessor struct {
	BalanceService BalanceService
}

// BalanceService 余额服务接口
type BalanceService interface {
	Deduct(ctx context.Context, userID uint64, amount int64, reason string) (string, error)
	Refund(ctx context.Context, userID uint64, amount int64, transactionID string) error
	GetBalance(ctx context.Context, userID uint64) (int64, error)
}

// Process 处理余额支付
func (p *BalanceProcessor) Process(ctx context.Context, component *PaymentComponent) (*ComponentProcessResult, error) {
	// 从 RefID 解析用户ID（实际实现中可能从上下文获取）
	transactionID, err := p.BalanceService.Deduct(ctx, 0, component.Amount, "combo payment")
	if err != nil {
		return nil, err
	}
	return &ComponentProcessResult{
		TransactionID: transactionID,
		Success:       true,
	}, nil
}

// Rollback 回滚余额支付
func (p *BalanceProcessor) Rollback(ctx context.Context, component *PaymentComponent) error {
	return p.BalanceService.Refund(ctx, 0, component.Amount, component.TransactionID)
}

// --- 积分支付处理器 ---

// PointsProcessor 积分支付处理器
type PointsProcessor struct {
	PointsService PointsService
}

// PointsService 积分服务接口
type PointsService interface {
	Deduct(ctx context.Context, userID uint64, points int64, reason string) (string, error)
	Refund(ctx context.Context, userID uint64, points int64, transactionID string) error
	GetPoints(ctx context.Context, userID uint64) (int64, error)
	GetExchangeRate(ctx context.Context) float64 // 积分兑换比例（多少积分=1分钱）
}

// Process 处理积分支付
func (p *PointsProcessor) Process(ctx context.Context, component *PaymentComponent) (*ComponentProcessResult, error) {
	// 将金额转换为积分
	rate := p.PointsService.GetExchangeRate(ctx)
	pointsNeeded := int64(float64(component.Amount) * rate)

	transactionID, err := p.PointsService.Deduct(ctx, 0, pointsNeeded, "combo payment")
	if err != nil {
		return nil, err
	}
	return &ComponentProcessResult{
		TransactionID: transactionID,
		Success:       true,
	}, nil
}

// Rollback 回滚积分支付
func (p *PointsProcessor) Rollback(ctx context.Context, component *PaymentComponent) error {
	rate := p.PointsService.GetExchangeRate(ctx)
	pointsToRefund := int64(float64(component.Amount) * rate)
	return p.PointsService.Refund(ctx, 0, pointsToRefund, component.TransactionID)
}

// --- 优惠券支付处理器 ---

// CouponProcessor 优惠券支付处理器
type CouponProcessor struct {
	CouponService CouponService
}

// CouponService 优惠券服务接口
type CouponService interface {
	Use(ctx context.Context, userID uint64, couponCode string, orderID uint64) error
	Restore(ctx context.Context, userID uint64, couponCode string) error
	GetCouponValue(ctx context.Context, couponCode string) (int64, error)
}

// Process 处理优惠券支付
func (p *CouponProcessor) Process(ctx context.Context, component *PaymentComponent) (*ComponentProcessResult, error) {
	err := p.CouponService.Use(ctx, 0, component.RefID, 0)
	if err != nil {
		return nil, err
	}
	return &ComponentProcessResult{
		TransactionID: component.RefID,
		Success:       true,
	}, nil
}

// Rollback 回滚优惠券支付
func (p *CouponProcessor) Rollback(ctx context.Context, component *PaymentComponent) error {
	return p.CouponService.Restore(ctx, 0, component.RefID)
}

// --- 礼品卡支付处理器 ---

// GiftCardProcessor 礼品卡支付处理器
type GiftCardProcessor struct {
	GiftCardService GiftCardService
}

// GiftCardService 礼品卡服务接口
type GiftCardService interface {
	Deduct(ctx context.Context, cardNo string, amount int64) (string, error)
	Refund(ctx context.Context, cardNo string, amount int64, transactionID string) error
	GetBalance(ctx context.Context, cardNo string) (int64, error)
}

// Process 处理礼品卡支付
func (p *GiftCardProcessor) Process(ctx context.Context, component *PaymentComponent) (*ComponentProcessResult, error) {
	transactionID, err := p.GiftCardService.Deduct(ctx, component.RefID, component.Amount)
	if err != nil {
		return nil, err
	}
	return &ComponentProcessResult{
		TransactionID: transactionID,
		Success:       true,
	}, nil
}

// Rollback 回滚礼品卡支付
func (p *GiftCardProcessor) Rollback(ctx context.Context, component *PaymentComponent) error {
	return p.GiftCardService.Refund(ctx, component.RefID, component.Amount, component.TransactionID)
}

// --- 组合支付仓储接口 ---

// ComboPaymentRepository 组合支付仓储接口
type ComboPaymentRepository interface {
	Save(ctx context.Context, payment *ComboPayment) error
	Update(ctx context.Context, payment *ComboPayment) error
	FindByID(ctx context.Context, id uint64) (*ComboPayment, error)
	FindByComboPaymentNo(ctx context.Context, comboPaymentNo string) (*ComboPayment, error)
	FindByOrderID(ctx context.Context, orderID uint64) (*ComboPayment, error)
}

// --- 组合支付事件 ---

// ComboPaymentCreatedEvent 组合支付创建事件
type ComboPaymentCreatedEvent struct {
	eventsourcing.BaseEvent
	OrderID     uint64 `json:"order_id"`
	UserID      uint64 `json:"user_id"`
	TotalAmount int64  `json:"total_amount"`
}

// ComboPaymentCompletedEvent 组合支付完成事件
type ComboPaymentCompletedEvent struct {
	eventsourcing.BaseEvent
	OrderID    uint64 `json:"order_id"`
	UserID     uint64 `json:"user_id"`
	PaidAmount int64  `json:"paid_amount"`
}

// ComboPaymentFailedEvent 组合支付失败事件
type ComboPaymentFailedEvent struct {
	eventsourcing.BaseEvent
	OrderID       uint64 `json:"order_id"`
	UserID        uint64 `json:"user_id"`
	FailureReason string `json:"failure_reason"`
}
