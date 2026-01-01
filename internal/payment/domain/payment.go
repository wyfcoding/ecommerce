package domain

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// --- Payment Basic Types ---

// PaymentStatus 支付状态枚举
type PaymentStatus int

const (
	PaymentPending   PaymentStatus = 1 // 待支付
	PaymentSuccess   PaymentStatus = 2 // 支付成功
	PaymentFailed    PaymentStatus = 3 // 支付失败
	PaymentCancelled PaymentStatus = 4 // 已取消
	PaymentRefunding PaymentStatus = 5 // 退款中
	PaymentRefunded  PaymentStatus = 6 // 已退款
)

func (s PaymentStatus) String() string {
	switch s {
	case PaymentPending:
		return "Pending"
	case PaymentSuccess:
		return "Success"
	case PaymentFailed:
		return "Failed"
	case PaymentCancelled:
		return "Cancelled"
	case PaymentRefunding:
		return "Refunding"
	case PaymentRefunded:
		return "Refunded"
	default:
		return "Unknown"
	}
}

// --- Payment Aggregates ---

// Payment 支付聚合根
type Payment struct {
	ID            uint64
	PaymentNo     string
	OrderID       uint64
	OrderNo       string
	UserID        uint64
	Amount        int64
	PaymentMethod string
	GatewayType   GatewayType
	Status        PaymentStatus
	TransactionID string
	ThirdPartyNo  string
	CallbackData  string
	FailureReason string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	PaidAt        *time.Time
	CancelledAt   *time.Time
	RefundedAt    *time.Time
	Logs          []*PaymentLog
	Refunds       []*Refund
}

// Refund 退款实体
type Refund struct {
	ID              uint64
	RefundNo        string
	PaymentID       uint64
	PaymentNo       string
	OrderID         uint64
	OrderNo         string
	UserID          uint64
	RefundAmount    int64
	Reason          string
	Status          PaymentStatus
	ThirdPartyNo    string
	GatewayRefundID string
	FailureReason   string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	RefundedAt      *time.Time
}

// PaymentLog 支付操作日志
type PaymentLog struct {
	ID        uint64
	PaymentID uint64
	Action    string
	OldStatus string
	NewStatus string
	Remark    string
	CreatedAt time.Time
}

func NewPayment(orderID uint64, orderNo string, userID uint64, amount int64, paymentMethod string, gatewayType GatewayType) *Payment {
	now := time.Now()
	payment := &Payment{
		PaymentNo:     generatePaymentNo(),
		OrderID:       orderID,
		OrderNo:       orderNo,
		UserID:        userID,
		Amount:        amount,
		PaymentMethod: paymentMethod,
		GatewayType:   gatewayType,
		Status:        PaymentPending,
		CreatedAt:     now,
		UpdatedAt:     now,
		Logs:          make([]*PaymentLog, 0),
		Refunds:       make([]*Refund, 0),
	}
	payment.AddLog("Payment Initiated", "", PaymentPending.String(), fmt.Sprintf("Amount: %d, Gateway: %s", amount, gatewayType))
	return payment
}

func (p *Payment) CanProcess() error {
	if p.Status != PaymentPending {
		return fmt.Errorf("invalid status for processing: %s", p.Status)
	}
	return nil
}

func (p *Payment) Process(success bool, transactionID, thirdPartyNo string) error {
	if p.Status == PaymentSuccess && p.TransactionID == transactionID {
		return nil
	}
	if err := p.CanProcess(); err != nil {
		return err
	}
	oldStatus := p.Status
	p.TransactionID = transactionID
	p.ThirdPartyNo = thirdPartyNo
	p.UpdatedAt = time.Now()
	if success {
		p.Status = PaymentSuccess
		now := time.Now()
		p.PaidAt = &now
		p.AddLog("Payment Success", oldStatus.String(), p.Status.String(), "TransID: "+transactionID)
	} else {
		p.Status = PaymentFailed
		p.FailureReason = "Payment processing reported failure"
		p.AddLog("Payment Failed", oldStatus.String(), p.Status.String(), p.FailureReason)
	}
	return nil
}

func (p *Payment) CanRefund() error {
	if p.Status != PaymentSuccess && p.Status != PaymentRefunding && p.Status != PaymentRefunded {
		return fmt.Errorf("payment not successful, current status: %s", p.Status)
	}
	if p.Status == PaymentRefunded {
		return fmt.Errorf("payment fully refunded already")
	}
	return nil
}

func (p *Payment) CreateRefund(refundAmount int64, reason string) (*Refund, error) {
	if err := p.CanRefund(); err != nil {
		return nil, err
	}
	if refundAmount <= 0 || refundAmount > p.Amount {
		return nil, fmt.Errorf("invalid refund amount: %d", refundAmount)
	}
	refund := &Refund{
		RefundNo:     generateRefundNo(),
		PaymentID:    p.ID,
		PaymentNo:    p.PaymentNo,
		OrderID:      p.OrderID,
		OrderNo:      p.OrderNo,
		UserID:       p.UserID,
		RefundAmount: refundAmount,
		Reason:       reason,
		Status:       PaymentRefunding,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	p.Refunds = append(p.Refunds, refund)
	p.Status = PaymentRefunding
	p.UpdatedAt = time.Now()
	p.AddLog("Refund Created", "", "", fmt.Sprintf("Amt: %d, Reason: %s", refundAmount, reason))
	return refund, nil
}

func (p *Payment) ProcessRefund(refundNo string, success bool, gatewayRefundID string) error {
	var refund *Refund
	for _, r := range p.Refunds {
		if r.RefundNo == refundNo {
			refund = r
			break
		}
	}
	if refund == nil {
		return fmt.Errorf("refund not found: %s", refundNo)
	}
	if refund.Status == PaymentRefunded && success {
		return nil
	}
	refund.UpdatedAt = time.Now()
	refund.GatewayRefundID = gatewayRefundID
	if success {
		refund.Status = PaymentRefunded
		now := time.Now()
		refund.RefundedAt = &now
		p.Status = PaymentRefunded
		p.RefundedAt = &now
		p.AddLog("Refund Success", "", PaymentRefunded.String(), "RefundNo: "+refundNo)
	} else {
		refund.Status = PaymentFailed
		refund.FailureReason = "Gateway rejected refund"
		p.Status = PaymentSuccess
		p.AddLog("Refund Failed", "", PaymentFailed.String(), "RefundNo: "+refundNo)
	}
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Payment) Cancel(reason string) error {
	if p.Status != PaymentPending {
		return fmt.Errorf("cannot cancel from status: %s", p.Status)
	}
	oldStatus := p.Status
	p.Status = PaymentCancelled
	now := time.Now()
	p.CancelledAt = &now
	p.UpdatedAt = now
	p.FailureReason = reason
	p.AddLog("Payment Cancelled", oldStatus.String(), p.Status.String(), reason)
	return nil
}

func (p *Payment) AddLog(action, oldStatus, newStatus, remark string) {
	p.Logs = append(p.Logs, &PaymentLog{
		PaymentID: p.ID,
		Action:    action,
		OldStatus: oldStatus,
		NewStatus: newStatus,
		Remark:    remark,
		CreatedAt: time.Now(),
	})
}

func generatePaymentNo() string {
	return fmt.Sprintf("P%s", time.Now().Format("20060102150405"))
}

func generateRefundNo() string {
	return fmt.Sprintf("R%s", time.Now().Format("20060102150405"))
}

// --- Channel Config ---

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

type ChannelRepository interface {
	FindByCode(ctx context.Context, code string) (*ChannelConfig, error)
	ListEnabledByType(ctx context.Context, channelType ChannelType) ([]*ChannelConfig, error)
	Save(ctx context.Context, channel *ChannelConfig) error
}

// --- Payment Gateway ---

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
	ClientIP    string
	ReturnURL   string
	NotifyURL   string
	ExtraData   map[string]string
}

type PaymentGatewayResponse struct {
	TransactionID string
	PaymentURL    string
	QRCode        string
	AppParam      string
	RawResponse   string
}

type RefundGatewayRequest struct {
	PaymentID     string
	TransactionID string
	RefundID      string
	Amount        int64
	Reason        string
}

type RefundGatewayResponse struct {
	RefundID    string
	Status      string
	RawResponse string
}

type PaymentGateway interface {
	Pay(ctx context.Context, req *PaymentGatewayRequest) (*PaymentGatewayResponse, error)
	Query(ctx context.Context, transactionID string) (*PaymentGatewayResponse, error)
	Refund(ctx context.Context, req *RefundGatewayRequest) (*RefundGatewayResponse, error)
	QueryRefund(ctx context.Context, refundID string) (*RefundGatewayResponse, error)
	VerifyCallback(ctx context.Context, data map[string]string) (bool, error)
	GetType() GatewayType
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
