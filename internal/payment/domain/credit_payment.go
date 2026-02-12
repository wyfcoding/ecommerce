// 变更说明：新增信用支付功能，支持先用后付、授信额度管理、还款计划、逾期处理。
// 假设：信用支付默认账期30天，逾期利率为日万五。
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

// --- 信用支付状态 ---

// CreditPaymentStatus 信用支付状态
type CreditPaymentStatus int

const (
	CreditStatusPending     CreditPaymentStatus = 1 // 待确认
	CreditStatusApproved    CreditPaymentStatus = 2 // 已批准（等待使用）
	CreditStatusUsed        CreditPaymentStatus = 3 // 已使用（待还款）
	CreditStatusPartialPaid CreditPaymentStatus = 4 // 部分还款
	CreditStatusPaid        CreditPaymentStatus = 5 // 已还清
	CreditStatusOverdue     CreditPaymentStatus = 6 // 已逾期
	CreditStatusDefault     CreditPaymentStatus = 7 // 违约
	CreditStatusClosed      CreditPaymentStatus = 8 // 已关闭
)

// --- 信用账户 ---

// CreditAccount 信用账户
type CreditAccount struct {
	ID             uint64    `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	UserID         uint64    `json:"user_id"`
	TotalLimit     int64     `json:"total_limit"`     // 总授信额度（分）
	UsedLimit      int64     `json:"used_limit"`      // 已使用额度（分）
	AvailableLimit int64     `json:"available_limit"` // 可用额度（分）
	OverdueAmount  int64     `json:"overdue_amount"`  // 逾期金额（分）
	Status         string    `json:"status"`          // ACTIVE/FROZEN/CLOSED
	CreditScore    int       `json:"credit_score"`    // 信用评分
	BillingDay     int       `json:"billing_day"`     // 账单日（每月几号）
	RepaymentDay   int       `json:"repayment_day"`   // 还款日（每月几号）
	OverdueRate    float64   `json:"overdue_rate"`    // 逾期利率（日万分比）
}

// NewCreditAccount 创建信用账户
func NewCreditAccount(userID uint64, totalLimit int64, billingDay, repaymentDay int) *CreditAccount {
	return &CreditAccount{
		UserID:         userID,
		TotalLimit:     totalLimit,
		UsedLimit:      0,
		AvailableLimit: totalLimit,
		OverdueAmount:  0,
		Status:         "ACTIVE",
		CreditScore:    600, // 默认信用分
		BillingDay:     billingDay,
		RepaymentDay:   repaymentDay,
		OverdueRate:    0.0005, // 默认日万五
	}
}

// UseCredit 使用信用额度
func (ca *CreditAccount) UseCredit(amount int64) error {
	if ca.Status != "ACTIVE" {
		return errors.New("credit account is not active")
	}
	if amount > ca.AvailableLimit {
		return fmt.Errorf("insufficient credit limit: available=%d, required=%d", ca.AvailableLimit, amount)
	}

	ca.UsedLimit += amount
	ca.AvailableLimit -= amount
	return nil
}

// RestoreCredit 恢复信用额度（还款后）
func (ca *CreditAccount) RestoreCredit(amount int64) {
	ca.UsedLimit -= amount
	if ca.UsedLimit < 0 {
		ca.UsedLimit = 0
	}
	ca.AvailableLimit = ca.TotalLimit - ca.UsedLimit
}

// AdjustLimit 调整授信额度
func (ca *CreditAccount) AdjustLimit(newLimit int64, reason string) error {
	if newLimit < ca.UsedLimit {
		return fmt.Errorf("new limit %d is less than used limit %d", newLimit, ca.UsedLimit)
	}
	ca.TotalLimit = newLimit
	ca.AvailableLimit = newLimit - ca.UsedLimit
	return nil
}

// Freeze 冻结账户
func (ca *CreditAccount) Freeze(reason string) {
	ca.Status = "FROZEN"
}

// Unfreeze 解冻账户
func (ca *CreditAccount) Unfreeze() {
	ca.Status = "ACTIVE"
}

// --- 信用支付 ---

// CreditPayment 信用支付聚合根
type CreditPayment struct {
	eventsourcing.AggregateRoot
	ID               uint64              `json:"id"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
	CreditPaymentNo  string              `json:"credit_payment_no"` // 信用支付单号
	OrderID          uint64              `json:"order_id"`
	OrderNo          string              `json:"order_no"`
	UserID           uint64              `json:"user_id"`
	CreditAccountID  uint64              `json:"credit_account_id"`
	Principal        int64               `json:"principal"`        // 本金（分）
	Interest         int64               `json:"interest"`         // 利息（分）
	Penalty          int64               `json:"penalty"`          // 逾期罚息（分）
	TotalAmount      int64               `json:"total_amount"`     // 应还总额（分）
	PaidAmount       int64               `json:"paid_amount"`      // 已还金额（分）
	RemainingAmount  int64               `json:"remaining_amount"` // 待还金额（分）
	Status           CreditPaymentStatus `json:"status"`
	PaymentStatus    pb.PaymentStatus    `json:"payment_status"`
	UsedAt           *time.Time          `json:"used_at"`           // 使用时间
	DueDate          *time.Time          `json:"due_date"`          // 还款到期日
	PaidAt           *time.Time          `json:"paid_at"`           // 还清时间
	OverdueDays      int                 `json:"overdue_days"`      // 逾期天数
	RepaymentRecords []*RepaymentRecord  `json:"repayment_records"` // 还款记录
	Logs             []*CreditPaymentLog `json:"logs"`              // 操作日志
	PersistenceVer   int64               `json:"version"`

	// 配置参数
	InterestRate    float64 `json:"interest_rate"`     // 利率（日万分比）
	GracePeriodDays int     `json:"grace_period_days"` // 免息期（天）
}

// RepaymentRecord 还款记录
type RepaymentRecord struct {
	ID              uint64    `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	CreditPaymentID uint64    `json:"credit_payment_id"`
	Amount          int64     `json:"amount"`         // 还款金额
	Principal       int64     `json:"principal"`      // 本金部分
	Interest        int64     `json:"interest"`       // 利息部分
	Penalty         int64     `json:"penalty"`        // 罚息部分
	PaymentMethod   string    `json:"payment_method"` // 还款方式
	TransactionID   string    `json:"transaction_id"` // 交易号
}

// CreditPaymentLog 信用支付操作日志
type CreditPaymentLog struct {
	ID              uint64    `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	CreditPaymentID uint64    `json:"credit_payment_id"`
	Action          string    `json:"action"`
	Operator        string    `json:"operator"`
	OldStatus       string    `json:"old_status"`
	NewStatus       string    `json:"new_status"`
	Amount          int64     `json:"amount"`
	Remark          string    `json:"remark"`
}

// NewCreditPayment 创建信用支付
func NewCreditPayment(orderID uint64, orderNo string, userID, creditAccountID uint64, principal int64, gracePeriodDays int, interestRate float64, idGenerator idgen.Generator) *CreditPayment {
	if gracePeriodDays <= 0 {
		gracePeriodDays = 30 // 默认30天免息期
	}
	if interestRate <= 0 {
		interestRate = 0.0005 // 默认日万五
	}

	creditNo := fmt.Sprintf("CREDIT%d", idGenerator.Generate())

	cp := &CreditPayment{
		OrderID:          orderID,
		OrderNo:          orderNo,
		UserID:           userID,
		CreditAccountID:  creditAccountID,
		Principal:        principal,
		Interest:         0,
		Penalty:          0,
		TotalAmount:      principal,
		PaidAmount:       0,
		RemainingAmount:  principal,
		Status:           CreditStatusPending,
		PaymentStatus:    pb.PaymentStatus_PENDING,
		RepaymentRecords: make([]*RepaymentRecord, 0),
		Logs:             make([]*CreditPaymentLog, 0),
		InterestRate:     interestRate,
		GracePeriodDays:  gracePeriodDays,
		PersistenceVer:   1,
	}
	cp.SetID(creditNo)
	cp.CreditPaymentNo = creditNo
	cp.AddLog("CREATE", "System", "", CreditStatusPending.String(), 0, "Credit payment created")
	return cp
}

// Use 使用信用额度完成支付
func (cp *CreditPayment) Use(ctx context.Context) error {
	if cp.Status != CreditStatusPending && cp.Status != CreditStatusApproved {
		return errors.New("invalid status for use")
	}

	now := time.Now()
	dueDate := now.AddDate(0, 0, cp.GracePeriodDays)

	cp.Status = CreditStatusUsed
	cp.PaymentStatus = pb.PaymentStatus_SUCCESS
	cp.UsedAt = &now
	cp.DueDate = &dueDate

	cp.AddLog("USE", "System", CreditStatusPending.String(), CreditStatusUsed.String(), cp.Principal, fmt.Sprintf("Due date: %s", dueDate.Format("2006-01-02")))
	return nil
}

// CalculateInterest 计算利息（免息期后）
func (cp *CreditPayment) CalculateInterest() int64 {
	if cp.DueDate == nil || cp.UsedAt == nil {
		return 0
	}

	now := time.Now()
	dueDate := *cp.DueDate

	// 免息期内不计息
	if now.Before(dueDate) {
		return 0
	}

	// 计算超过免息期的天数
	overdueDays := int(now.Sub(dueDate).Hours() / 24)
	if overdueDays <= 0 {
		return 0
	}

	cp.OverdueDays = overdueDays

	// 计算利息 = 本金 * 日利率 * 超期天数
	interest := int64(float64(cp.RemainingAmount) * cp.InterestRate * float64(overdueDays))
	cp.Interest = interest
	cp.TotalAmount = cp.Principal + interest
	cp.RemainingAmount = cp.TotalAmount - cp.PaidAmount

	return interest
}

// CalculatePenalty 计算逾期罚息
func (cp *CreditPayment) CalculatePenalty(penaltyRate float64) int64 {
	if cp.DueDate == nil {
		return 0
	}

	now := time.Now()
	if now.Before(*cp.DueDate) {
		return 0
	}

	// 计算逾期天数
	overdueDays := int(now.Sub(*cp.DueDate).Hours() / 24)
	if overdueDays <= 0 {
		return 0
	}

	cp.OverdueDays = overdueDays
	cp.Status = CreditStatusOverdue

	// 罚息 = 剩余本金 * 罚息利率 * 逾期天数
	penalty := int64(float64(cp.RemainingAmount) * penaltyRate * float64(overdueDays))
	cp.Penalty = penalty
	cp.TotalAmount = cp.Principal + cp.Interest + penalty
	cp.RemainingAmount = cp.TotalAmount - cp.PaidAmount

	return penalty
}

// Repay 还款
func (cp *CreditPayment) Repay(ctx context.Context, amount int64, paymentMethod, transactionID string) error {
	if cp.Status != CreditStatusUsed && cp.Status != CreditStatusPartialPaid && cp.Status != CreditStatusOverdue {
		return errors.New("invalid status for repayment")
	}

	// 先更新利息和罚息
	cp.CalculateInterest()
	cp.CalculatePenalty(cp.InterestRate * 2) // 罚息率为正常利率的2倍

	if amount > cp.RemainingAmount {
		return fmt.Errorf("repayment amount %d exceeds remaining amount %d", amount, cp.RemainingAmount)
	}

	// 还款优先级：罚息 → 利息 → 本金
	var principalPart, interestPart, penaltyPart int64
	remaining := amount

	// 先还罚息
	if cp.Penalty > 0 && remaining > 0 {
		if remaining >= cp.Penalty {
			penaltyPart = cp.Penalty
			remaining -= cp.Penalty
			cp.Penalty = 0
		} else {
			penaltyPart = remaining
			cp.Penalty -= remaining
			remaining = 0
		}
	}

	// 再还利息
	if cp.Interest > 0 && remaining > 0 {
		if remaining >= cp.Interest {
			interestPart = cp.Interest
			remaining -= cp.Interest
			cp.Interest = 0
		} else {
			interestPart = remaining
			cp.Interest -= remaining
			remaining = 0
		}
	}

	// 最后还本金
	if remaining > 0 {
		principalPart = remaining
		cp.Principal -= remaining
	}

	cp.PaidAmount += amount
	cp.RemainingAmount = cp.TotalAmount - cp.PaidAmount

	// 创建还款记录
	record := &RepaymentRecord{
		Amount:        amount,
		Principal:     principalPart,
		Interest:      interestPart,
		Penalty:       penaltyPart,
		PaymentMethod: paymentMethod,
		TransactionID: transactionID,
	}
	cp.RepaymentRecords = append(cp.RepaymentRecords, record)

	// 更新状态
	oldStatus := cp.Status
	if cp.RemainingAmount <= 0 {
		cp.Status = CreditStatusPaid
		now := time.Now()
		cp.PaidAt = &now
	} else {
		cp.Status = CreditStatusPartialPaid
	}

	cp.AddLog("REPAY", "System", oldStatus.String(), cp.Status.String(), amount, fmt.Sprintf("Principal: %d, Interest: %d, Penalty: %d", principalPart, interestPart, penaltyPart))
	return nil
}

// MarkDefault 标记违约
func (cp *CreditPayment) MarkDefault(ctx context.Context, reason string) error {
	if cp.Status != CreditStatusOverdue {
		return errors.New("can only mark default in overdue status")
	}

	cp.Status = CreditStatusDefault
	cp.AddLog("DEFAULT", "System", CreditStatusOverdue.String(), CreditStatusDefault.String(), cp.RemainingAmount, reason)
	return nil
}

// AddLog 添加操作日志
func (cp *CreditPayment) AddLog(action, operator, oldStatus, newStatus string, amount int64, remark string) {
	log := &CreditPaymentLog{
		Action:    action,
		Operator:  operator,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		Amount:    amount,
		Remark:    remark,
	}
	cp.Logs = append(cp.Logs, log)
}

// String 返回状态字符串
func (s CreditPaymentStatus) String() string {
	switch s {
	case CreditStatusPending:
		return "PENDING"
	case CreditStatusApproved:
		return "APPROVED"
	case CreditStatusUsed:
		return "USED"
	case CreditStatusPartialPaid:
		return "PARTIAL_PAID"
	case CreditStatusPaid:
		return "PAID"
	case CreditStatusOverdue:
		return "OVERDUE"
	case CreditStatusDefault:
		return "DEFAULT"
	case CreditStatusClosed:
		return "CLOSED"
	default:
		return "UNKNOWN"
	}
}

// --- 信用账户仓储接口 ---

// CreditAccountRepository 信用账户仓储接口
type CreditAccountRepository interface {
	Save(ctx context.Context, account *CreditAccount) error
	Update(ctx context.Context, account *CreditAccount) error
	FindByID(ctx context.Context, id uint64) (*CreditAccount, error)
	FindByUserID(ctx context.Context, userID uint64) (*CreditAccount, error)
}

// CreditPaymentRepository 信用支付仓储接口
type CreditPaymentRepository interface {
	Save(ctx context.Context, payment *CreditPayment) error
	Update(ctx context.Context, payment *CreditPayment) error
	FindByID(ctx context.Context, id uint64) (*CreditPayment, error)
	FindByCreditPaymentNo(ctx context.Context, creditPaymentNo string) (*CreditPayment, error)
	FindByOrderID(ctx context.Context, orderID uint64) (*CreditPayment, error)
	FindOverduePayments(ctx context.Context, before time.Time) ([]*CreditPayment, error)
	FindPendingRepayments(ctx context.Context, userID uint64) ([]*CreditPayment, error)
}

// --- 信用支付事件 ---

// CreditUsedEvent 信用使用事件
type CreditUsedEvent struct {
	eventsourcing.BaseEvent
	OrderID uint64    `json:"order_id"`
	UserID  uint64    `json:"user_id"`
	Amount  int64     `json:"amount"`
	DueDate time.Time `json:"due_date"`
}

// CreditRepaidEvent 信用还款事件
type CreditRepaidEvent struct {
	eventsourcing.BaseEvent
	OrderID   uint64 `json:"order_id"`
	UserID    uint64 `json:"user_id"`
	Amount    int64  `json:"amount"`
	Remaining int64  `json:"remaining"`
}

// CreditOverdueEvent 信用逾期事件
type CreditOverdueEvent struct {
	eventsourcing.BaseEvent
	OrderID     uint64 `json:"order_id"`
	UserID      uint64 `json:"user_id"`
	OverdueDays int    `json:"overdue_days"`
	Penalty     int64  `json:"penalty"`
}
