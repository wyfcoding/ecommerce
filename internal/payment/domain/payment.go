package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/wyfcoding/pkg/fsm"
	"github.com/wyfcoding/pkg/idgen"
	"gorm.io/gorm"
)

// --- Payment Basic Types ---

type PaymentStatus int

const (
	PaymentPending        PaymentStatus = 1  // 待支付
	PaymentAuthorized     PaymentStatus = 10 // 已授权 (Pre-auth)
	PaymentSuccess        PaymentStatus = 2  // 支付成功 (Captured)
	PaymentFailed         PaymentStatus = 3  // 支付失败
	PaymentCancelled      PaymentStatus = 4  // 已取消 (Voided)
	PaymentRefunding      PaymentStatus = 5  // 退款中
	PaymentRefunded       PaymentStatus = 6  // 已退款
	PaymentReconciled     PaymentStatus = 20 // 已对账
	PaymentReconcileError PaymentStatus = 21 // 对账异常
)

func (s PaymentStatus) String() string {
	names := map[PaymentStatus]string{
		PaymentPending:        "Pending",
		PaymentAuthorized:     "Authorized",
		PaymentSuccess:        "Success",
		PaymentFailed:         "Failed",
		PaymentCancelled:      "Cancelled",
		PaymentRefunding:      "Refunding",
		PaymentRefunded:       "Refunded",
		PaymentReconciled:     "Reconciled",
		PaymentReconcileError: "ReconcileError",
	}
	return names[s]
}

// --- Payment Aggregates ---

type Payment struct {
	gorm.Model
	PaymentNo      string      `gorm:"uniqueIndex;size:64"`
	OrderID        uint64      `gorm:"index"`
	OrderNo        string      `gorm:"size:64"`
	UserID         uint64      `gorm:"index"`
	Amount         int64       `gorm:"not null"`
	CapturedAmount int64       `gorm:"default:0"`
	PaymentMethod  string      `gorm:"size:32"`
	GatewayType    GatewayType `gorm:"size:32"`
	Status         PaymentStatus
	TransactionID  string `gorm:"size:128"`
	ThirdPartyNo   string `gorm:"size:128"`
	CallbackData   string `gorm:"type:text"`
	FailureReason  string `gorm:"size:255"`
	PaidAt         *time.Time
	CancelledAt    *time.Time
	RefundedAt     *time.Time

	fsm     *fsm.Machine  `gorm:"-"`
	Logs    []*PaymentLog `gorm:"foreignKey:PaymentID"`
	Refunds []*Refund     `gorm:"foreignKey:PaymentID"`
}

type Refund struct {
	gorm.Model
	RefundNo        string `gorm:"uniqueIndex;size:64"`
	PaymentID       uint64 `gorm:"index"`
	PaymentNo       string `gorm:"size:64"`
	OrderID         uint64 `gorm:"index"`
	OrderNo         string `gorm:"size:64"`
	UserID          uint64 `gorm:"index"`
	RefundAmount    int64  `gorm:"not null"`
	Reason          string `gorm:"size:255"`
	Status          PaymentStatus
	ThirdPartyNo    string `gorm:"size:128"`
	GatewayRefundID string `gorm:"size:128"`
	FailureReason   string `gorm:"size:255"`
	RefundedAt      *time.Time
}

type PaymentLog struct {
	gorm.Model
	PaymentID uint64 `gorm:"index"`
	UserID    uint64 `gorm:"index"`
	Action    string `gorm:"size:64"`
	OldStatus string `gorm:"size:32"`
	NewStatus string `gorm:"size:32"`
	Remark    string `gorm:"size:255"`
}

func NewPayment(orderID uint64, orderNo string, userID uint64, amount int64, paymentMethod string, gatewayType GatewayType, idGenerator idgen.Generator) *Payment {
	p := &Payment{
		PaymentNo:     fmt.Sprintf("PAY%d", idGenerator.Generate()),
		OrderID:       orderID,
		OrderNo:       orderNo,
		UserID:        userID,
		Amount:        amount,
		PaymentMethod: paymentMethod,
		GatewayType:   gatewayType,
		Status:        PaymentPending,
	}
	p.initFSM()
	p.AddLog("INIT", "", PaymentPending.String(), "Payment created")
	return p
}

func (p *Payment) initFSM() {
	m := fsm.NewMachine(fsm.State(p.Status.String()))

	// 标准支付流
	m.AddTransition(fsm.State(PaymentPending.String()), "AUTH", fsm.State(PaymentAuthorized.String()))
	m.AddTransition(fsm.State(PaymentAuthorized.String()), "CAPTURE", fsm.State(PaymentSuccess.String()))
	m.AddTransition(fsm.State(PaymentPending.String()), "PAY_DIRECT", fsm.State(PaymentSuccess.String()))

	// 逆向流
	m.AddTransition(fsm.State(PaymentPending.String()), "CANCEL", fsm.State(PaymentCancelled.String()))
	m.AddTransition(fsm.State(PaymentAuthorized.String()), "VOID", fsm.State(PaymentCancelled.String()))

	// 退款流
	m.AddTransition(fsm.State(PaymentSuccess.String()), "REFUND_REQ", fsm.State(PaymentRefunding.String()))
	m.AddTransition(fsm.State(PaymentRefunding.String()), "REFUND_FINISH", fsm.State(PaymentRefunded.String()))

	// 对账流
	m.AddTransition(fsm.State(PaymentSuccess.String()), "RECONCILE", fsm.State(PaymentReconciled.String()))
	m.AddTransition(fsm.State(PaymentSuccess.String()), "RECONCILE_FAIL", fsm.State(PaymentReconcileError.String()))

	p.fsm = m
}

func (p *Payment) Trigger(ctx context.Context, event string, remark string) error {
	if p.fsm == nil {
		p.initFSM()
	}
	oldStatus := p.Status
	if err := p.fsm.Trigger(ctx, fsm.Event(event)); err != nil {
		return err
	}

	newStatusStr := string(p.fsm.Current())
	for s, name := range statusNamesMap {
		if name == newStatusStr {
			p.Status = s
			break
		}
	}
	p.AddLog(event, oldStatus.String(), p.Status.String(), remark)
	return nil
}

var statusNamesMap = map[PaymentStatus]string{
	PaymentPending:        "Pending",
	PaymentAuthorized:     "Authorized",
	PaymentSuccess:        "Success",
	PaymentFailed:         "Failed",
	PaymentCancelled:      "Cancelled",
	PaymentRefunding:      "Refunding",
	PaymentRefunded:       "Refunded",
	PaymentReconciled:     "Reconciled",
	PaymentReconcileError: "ReconcileError",
}

func (p *Payment) AddLog(action, oldStatus, newStatus, remark string) {
	p.Logs = append(p.Logs, &PaymentLog{
		PaymentID: uint64(p.ID),
		UserID:    p.UserID,
		Action:    action,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		Remark:    remark,
	})
}

// --- Channel & Gateway Definitions ---

type ChannelType string

const (
	ChannelTypeAlipay ChannelType = "alipay"
	ChannelTypeWechat ChannelType = "wechat"
	ChannelTypeStripe ChannelType = "stripe"
)

type ChannelConfig struct {
	gorm.Model
	Code        string      `gorm:"uniqueIndex;size:32;not null" json:"code"`
	Type        ChannelType `gorm:"size:32;not null" json:"type"`
	Name        string      `gorm:"size:64" json:"name"`
	Priority    int         `gorm:"default:0" json:"priority"`
	Enabled     bool        `gorm:"default:true" json:"enabled"`
	ConfigJSON  string      `gorm:"type:text" json:"config_json"`
	RatePercent float64     `gorm:"type:decimal(5,2)" json:"rate_percent"`
	Description string      `gorm:"size:255" json:"description"`
}

type GatewayType string

const (
	GatewayTypeAlipay GatewayType = "alipay"
	GatewayTypeWechat GatewayType = "wechat"
	GatewayTypeStripe GatewayType = "stripe"
	GatewayTypeMock   GatewayType = "mock"
)

type PaymentGatewayRequest struct {
	OrderID     string
	Amount      int64
	Currency    string
	Description string
}

type PaymentGatewayResponse struct {
	TransactionID string
	PaymentURL    string
	RawResponse   string
}

type PaymentGateway interface {
	PreAuth(ctx context.Context, req *PaymentGatewayRequest) (*PaymentGatewayResponse, error)
	Capture(ctx context.Context, transactionID string, amount int64) (*PaymentGatewayResponse, error)
	Void(ctx context.Context, transactionID string) error
	Refund(ctx context.Context, transactionID string, amount int64) error
	// DownloadBill 下载指定日期的对账单数据
	DownloadBill(ctx context.Context, date time.Time) ([]*GatewayBillItem, error)
}

type GatewayBillItem struct {
	TransactionID string
	PaymentNo     string
	Amount        int64
	Status        string
	PaidAt        time.Time
}

// --- Repositories ---

type PaymentRepository interface {
	FindByID(ctx context.Context, userID uint64, id uint64) (*Payment, error)
	FindByPaymentNo(ctx context.Context, userID uint64, paymentNo string) (*Payment, error)
	FindByOrderID(ctx context.Context, userID uint64, orderID uint64) (*Payment, error)
	Save(ctx context.Context, payment *Payment) error
	Update(ctx context.Context, payment *Payment) error
	SaveLog(ctx context.Context, log *PaymentLog) error
	FindLogsByPaymentID(ctx context.Context, userID uint64, paymentID uint64) ([]*PaymentLog, error)
	Transaction(ctx context.Context, userID uint64, fn func(tx any) error) error
	WithTx(tx any) PaymentRepository

	// 辅助查询
	GetUserIDByPaymentNo(ctx context.Context, paymentNo string) (uint64, error)

	// 对账相关
	FindSuccessPaymentsByDate(ctx context.Context, date time.Time) ([]*Payment, error)
	SaveReconciliationRecord(ctx context.Context, record *ReconciliationRecord) error
}

type RefundRepository interface {
	FindByID(ctx context.Context, userID uint64, id uint64) (*Refund, error)
	FindByRefundNo(ctx context.Context, userID uint64, refundNo string) (*Refund, error)
	Save(ctx context.Context, refund *Refund) error
	Transaction(ctx context.Context, userID uint64, fn func(tx any) error) error
	WithTx(tx any) RefundRepository
}

type ChannelRepository interface {
	FindByCode(ctx context.Context, code string) (*ChannelConfig, error)
	ListEnabledByType(ctx context.Context, channelType ChannelType) ([]*ChannelConfig, error)
	Save(ctx context.Context, channel *ChannelConfig) error
	Transaction(ctx context.Context, fn func(tx any) error) error
	WithTx(tx any) ChannelRepository
}

// --- Risk Services ---

type RiskAction string

const (
	RiskActionPass      RiskAction = "PASS"
	RiskActionBlock     RiskAction = "BLOCK"
	RiskActionChallenge RiskAction = "CHALLENGE"
)

type RiskResult struct {
	Action      RiskAction
	Reason      string
	RuleID      string
	Description string
}

type RiskContext struct {
	UserID        uint64
	IP            string
	DeviceID      string
	Amount        int64
	PaymentMethod string
	OrderID       uint64
}

type RiskService interface {
	CheckPrePayment(ctx context.Context, riskCtx *RiskContext) (*RiskResult, error)
	RecordTransaction(ctx context.Context, riskCtx *RiskContext) error
}

// --- Reconciliation ---

type ReconciliationRecord struct {
	gorm.Model
	PaymentID     uint64 `gorm:"index"`
	OrderNo       string `gorm:"index"`
	SystemAmount  int64
	GatewayAmount int64
	DiffAmount    int64
	Status        string // MATCH, MISMATCH, MISSING_SYSTEM, MISSING_GATEWAY
	Remark        string
}
