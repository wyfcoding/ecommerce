package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/payment/v1"
	"github.com/wyfcoding/pkg/eventsourcing"
	"github.com/wyfcoding/pkg/fsm"
	"github.com/wyfcoding/pkg/idgen"
)

// --- Payment Basic Types ---

type PaymentStatus = pb.PaymentStatus

var (
	ErrInvalidParameter = errors.New("invalid parameter")
)

// --- Payment Aggregates ---

type Payment struct {
	eventsourcing.AggregateRoot
	ID             uint          `json:"id"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	PaymentNo      string        `json:"payment_no"`
	OrderID        uint64        `json:"order_id"`
	OrderNo        string        `json:"order_no"`
	UserID         uint64        `json:"user_id"`
	Amount         int64         `json:"amount"`
	CapturedAmount int64         `json:"captured_amount"`
	Currency       string        `json:"currency"`
	PaymentMethod  string        `json:"payment_method"`
	GatewayType    GatewayType   `json:"gateway_type"`
	Status         PaymentStatus `json:"status"`
	TransactionID  string        `json:"transaction_id"`
	ThirdPartyNo   string        `json:"third_party_no"`
	CallbackData   string        `json:"callback_data"`
	FailureReason  string        `json:"failure_reason"`
	PaidAt         *time.Time    `json:"paid_at"`
	CancelledAt    *time.Time    `json:"cancelled_at"`
	RefundedAt     *time.Time    `json:"refunded_at"`
	PersistenceVer int64         `json:"version"`

	fsm     *fsm.Machine[string, string] `json:"-"`
	Logs    []*PaymentLog                `json:"logs"`
	Refunds []*Refund                    `json:"refunds"`

	// 世界级特性：分账信息 (用于平台抽佣、多商家结算)
	Splits []PaymentSplit `json:"splits"`
}

// GetID 返回聚合标识。
func (p *Payment) GetID() string {
	return p.AggregateRoot.ID()
}

// Apply 实现 eventsourcing.EventApplier 接口。
func (p *Payment) Apply(event eventsourcing.DomainEvent) {
	switch e := event.(type) {
	case *PaymentInitiatedEvent:
		if p.GetID() == "" {
			p.SetID(e.AggregateID())
		}
		p.PaymentNo = e.AggregateID()
		p.OrderID = e.OrderID
		p.OrderNo = e.OrderNo
		p.UserID = e.UserID
		p.Amount = e.Amount
		p.PaymentMethod = e.PaymentMethod
		p.Status = pb.PaymentStatus_PENDING
	case *PaymentAuthorizedEvent:
		p.Status = pb.PaymentStatus_AUTHORIZED
		p.TransactionID = e.TransactionID
	case *PaymentCapturedEvent:
		p.Status = pb.PaymentStatus_SUCCESS
		p.CapturedAmount = e.Amount
		p.PaidAt = &e.PaidAt
	case *PaymentPaidEvent:
		p.Status = pb.PaymentStatus_SUCCESS
		p.Amount = e.Amount
		t := time.Unix(e.PaidAt, 0)
		p.PaidAt = &t
	case *RefundFinishedEvent:
		p.Status = pb.PaymentStatus_REFUNDED
		p.RefundedAt = &e.RefundedAt
	case *PaymentClosedEvent:
		p.Status = pb.PaymentStatus_CLOSED
		p.CancelledAt = &e.ClosedAt
	}
	p.SetVersion(event.Version())
}

// PaymentSplit 定义了资金流向的拆分详情
type PaymentSplit struct {
	ID            uint      `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	PaymentID     uint64    `json:"payment_id"`
	RecipientID   uint64    `json:"recipient_id"`
	RecipientType string    `json:"recipient_type"`
	Amount        int64     `json:"amount"`
	Status        string    `json:"status"`
}

// AccountingEntry 影子账本分录 (复式记账原则)
type AccountingEntry struct {
	ID            uint      `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	PaymentID     uint64    `json:"payment_id"`
	AccountNo     string    `json:"account_no"`
	EntryType     string    `json:"entry_type"`
	Amount        int64     `json:"amount"`
	BalanceBefore int64     `json:"balance_before"`
	BalanceAfter  int64     `json:"balance_after"`
}

type Refund struct {
	ID              uint          `json:"id"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
	RefundNo        string        `json:"refund_no"`
	PaymentID       uint64        `json:"payment_id"`
	PaymentNo       string        `json:"payment_no"`
	OrderID         uint64        `json:"order_id"`
	OrderNo         string        `json:"order_no"`
	UserID          uint64        `json:"user_id"`
	RefundAmount    int64         `json:"refund_amount"`
	Reason          string        `json:"reason"`
	Status          PaymentStatus `json:"status"`
	ThirdPartyNo    string        `json:"third_party_no"`
	GatewayRefundID string        `json:"gateway_refund_id"`
	FailureReason   string        `json:"failure_reason"`
	RefundedAt      *time.Time    `json:"refunded_at"`
}

type PaymentLog struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	PaymentID uint64    `json:"payment_id"`
	UserID    uint64    `json:"user_id"`
	Action    string    `json:"action"`
	OldStatus string    `json:"old_status"`
	NewStatus string    `json:"new_status"`
	Remark    string    `json:"remark"`
}

func NewPayment(orderID uint64, orderNo string, userID uint64, amount int64, paymentMethod string, gatewayType GatewayType, idGenerator idgen.Generator) *Payment {
	p := &Payment{
		OrderID:        orderID,
		OrderNo:        orderNo,
		UserID:         userID,
		Amount:         amount,
		PaymentMethod:  paymentMethod,
		GatewayType:    gatewayType,
		PersistenceVer: 1,
	}
	paymentNo := fmt.Sprintf("PAY%d", idGenerator.Generate())
	p.SetID(paymentNo)

	event := &PaymentInitiatedEvent{
		BaseEvent:     eventsourcing.NewBaseEvent(PaymentInitiatedEventType, paymentNo, 0),
		OrderID:       orderID,
		OrderNo:       orderNo,
		UserID:        userID,
		Amount:        amount,
		PaymentMethod: paymentMethod,
	}
	p.ApplyChange(event)
	p.Apply(event)

	p.initFSM()
	p.AddLog("INIT", "", pb.PaymentStatus_PENDING.String(), "Payment created")
	return p
}

func (p *Payment) initFSM() {
	m := fsm.NewMachine[string, string](p.Status.String())

	m.AddTransition(pb.PaymentStatus_PENDING.String(), "AUTH", pb.PaymentStatus_AUTHORIZED.String())
	m.AddTransition(pb.PaymentStatus_AUTHORIZED.String(), "CAPTURE", pb.PaymentStatus_SUCCESS.String())
	m.AddTransition(pb.PaymentStatus_PENDING.String(), "PAY_DIRECT", pb.PaymentStatus_SUCCESS.String())
	m.AddTransition(pb.PaymentStatus_PENDING.String(), "CANCEL", pb.PaymentStatus_CLOSED.String())
	m.AddTransition(pb.PaymentStatus_AUTHORIZED.String(), "VOID", pb.PaymentStatus_CLOSED.String())
	m.AddTransition(pb.PaymentStatus_SUCCESS.String(), "REFUND_REQ", pb.PaymentStatus_REFUNDING.String())
	m.AddTransition(pb.PaymentStatus_REFUNDING.String(), "REFUND_FINISH", pb.PaymentStatus_REFUNDED.String())
	m.AddTransition(pb.PaymentStatus_SUCCESS.String(), "RECONCILE", pb.PaymentStatus_RECONCILED.String())
	m.AddTransition(pb.PaymentStatus_SUCCESS.String(), "RECONCILE_FAIL", pb.PaymentStatus_RECONCILE_ERROR.String())

	p.fsm = m
}

// InitFSM 确保状态机已初始化。
func (p *Payment) InitFSM() {
	if p.fsm == nil {
		p.initFSM()
	}
}

func (p *Payment) Trigger(ctx context.Context, event string, remark string) error {
	if p.fsm == nil {
		p.initFSM()
	}
	oldStatus := p.Status
	if err := p.fsm.Trigger(ctx, event); err != nil {
		return err
	}

	newStatusStr := p.fsm.Current()
	v, _ := pb.PaymentStatus_value[newStatusStr]
	newStatus := pb.PaymentStatus(v)

	var domainEv eventsourcing.DomainEvent
	switch event {
	case "AUTH":
		domainEv = &PaymentAuthorizedEvent{
			BaseEvent:     eventsourcing.NewBaseEvent(PaymentAuthorizedEventType, p.GetID(), p.AggregateRoot.Version()),
			PaymentNo:     p.PaymentNo,
			OrderID:       p.OrderID,
			UserID:        p.UserID,
			TransactionID: p.TransactionID,
		}
	case "CAPTURE":
		domainEv = &PaymentCapturedEvent{
			BaseEvent: eventsourcing.NewBaseEvent(PaymentCapturedEventType, p.GetID(), p.AggregateRoot.Version()),
			PaymentNo: p.PaymentNo,
			OrderNo:   p.OrderNo,
			OrderID:   p.OrderID,
			UserID:    p.UserID,
			Amount:    p.CapturedAmount,
			PaidAt:    time.Now(),
		}
	case "PAY_DIRECT":
		domainEv = &PaymentPaidEvent{
			BaseEvent: eventsourcing.NewBaseEvent(PaymentPaidEventType, p.GetID(), p.AggregateRoot.Version()),
			PaymentNo: p.PaymentNo,
			OrderNo:   p.OrderNo,
			OrderID:   p.OrderID,
			UserID:    p.UserID,
			Amount:    p.Amount,
			PaidAt:    time.Now().Unix(),
		}
	case "REFUND_FINISH":
		domainEv = &RefundFinishedEvent{
			BaseEvent:    eventsourcing.NewBaseEvent(RefundFinishedEventType, p.GetID(), p.AggregateRoot.Version()),
			PaymentNo:    p.PaymentNo,
			OrderNo:      p.OrderNo,
			OrderID:      p.OrderID,
			UserID:       p.UserID,
			RefundAmount: p.Amount,
			RefundedAt:   time.Now(),
		}
	case "CANCEL", "VOID":
		domainEv = &PaymentClosedEvent{
			BaseEvent: eventsourcing.NewBaseEvent(PaymentClosedEventType, p.GetID(), p.AggregateRoot.Version()),
			PaymentNo: p.PaymentNo,
			OrderID:   p.OrderID,
			UserID:    p.UserID,
			ClosedAt:  time.Now(),
		}
	}

	if domainEv != nil {
		p.ApplyChange(domainEv)
		p.Apply(domainEv)
	} else {
		p.Status = newStatus
	}

	p.AddLog(event, oldStatus.String(), p.Status.String(), remark)
	return nil
}

func (p *Payment) AddLog(action, oldStatus, newStatus, remark string) {
	p.Logs = append(p.Logs, &PaymentLog{
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
	ChannelTypeAlipay  ChannelType = "alipay"
	ChannelTypeWechat  ChannelType = "wechat"
	ChannelTypeStripe  ChannelType = "stripe"
	ChannelTypeTrading ChannelType = "trading"
)

type GatewayType string

const (
	GatewayTypeAlipay  GatewayType = "alipay"
	GatewayTypeWechat  GatewayType = "wechat"
	GatewayTypeStripe  GatewayType = "stripe"
	GatewayTypeMock    GatewayType = "mock"
	GatewayTypeTrading GatewayType = "trading"
)

type PaymentGatewayRequest struct {
	OrderID     string
	UserID      uint64
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

type EventStore interface {
	Save(ctx context.Context, events []eventsourcing.DomainEvent) error
	GetHistory(ctx context.Context, aggregateID string) ([]eventsourcing.DomainEvent, error)
}

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
	GetUserIDByPaymentNo(ctx context.Context, paymentNo string) (uint64, error)
	FindSuccessPaymentsByDate(ctx context.Context, date time.Time) ([]*Payment, error)
	SaveReconciliationRecord(ctx context.Context, record *ReconciliationRecord) error
	ExecWithBarrier(ctx context.Context, barrier any, fn func(ctx context.Context) error) error
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

type ChannelConfig struct {
	ID          uint        `json:"id"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	Code        string      `json:"code"`
	Type        ChannelType `json:"type"`
	Name        string      `json:"name"`
	Priority    int         `json:"priority"`
	Enabled     bool        `json:"enabled"`
	ConfigJSON  string      `json:"config_json"`
	RatePercent float64     `json:"rate_percent"`
	Description string      `json:"description"`
}

type RiskService interface {
	CheckPrePayment(ctx context.Context, riskCtx *RiskContext) (*RiskResult, error)
	RecordTransaction(ctx context.Context, riskCtx *RiskContext) error
}

type RiskResult struct {
	Action      RiskAction
	Reason      string
	RuleID      string
	Description string
}

type RiskAction string

const (
	RiskActionBlock     RiskAction = "BLOCK"
	RiskActionChallenge RiskAction = "CHALLENGE"
	RiskActionPass      RiskAction = "PASS"
)

type RiskContext struct {
	UserID        uint64
	IP            string
	DeviceID      string
	Amount        int64
	PaymentMethod string
	OrderID       uint64
}

type ReconciliationRecord struct {
	ID            uint      `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	OrderNo       string    `json:"order_no"`
	PaymentID     uint64    `json:"payment_id"`
	GatewayAmount int64     `json:"gateway_amount"`
	SystemAmount  int64     `json:"system_amount"`
	DiffAmount    int64     `json:"diff_amount"`
	Status        string    `json:"status"`
	Remark        string    `json:"remark"`
}
