package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrArbitrationNotFound     = errors.New("arbitration not found")
	ErrArbitrationAlreadyEnded = errors.New("arbitration already ended")
	ErrArbitrationNotPending   = errors.New("arbitration not in pending status")
	ErrInvalidArbitrationResult = errors.New("invalid arbitration result")
)

type ArbitrationStatus int8

const (
	ArbitrationStatusPending      ArbitrationStatus = 1
	ArbitrationStatusProcessing   ArbitrationStatus = 2
	ArbitrationStatusUserWin      ArbitrationStatus = 3
	ArbitrationStatusMerchantWin  ArbitrationStatus = 4
	ArbitrationStatusCompromise   ArbitrationStatus = 5
	ArbitrationStatusCancelled    ArbitrationStatus = 6
)

type ArbitrationResult string

const (
	ArbitrationResultUserWin     ArbitrationResult = "user_win"
	ArbitrationResultMerchantWin ArbitrationResult = "merchant_win"
	ArbitrationResultCompromise  ArbitrationResult = "compromise"
)

type ArbitrationParty string

const (
	ArbitrationPartyUser     ArbitrationParty = "user"
	ArbitrationPartyMerchant ArbitrationParty = "merchant"
	ArbitrationPartyPlatform ArbitrationParty = "platform"
)

type Arbitration struct {
	ID               uint64             `json:"id"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
	ArbitrationNo    string             `json:"arbitration_no"`
	AfterSalesID     uint64             `json:"after_sales_id"`
	AfterSalesNo     string             `json:"after_sales_no"`
	OrderID          uint64             `json:"order_id"`
	OrderNo          string             `json:"order_no"`
	UserID           uint64             `json:"user_id"`
	MerchantID       uint64             `json:"merchant_id"`
	Status           ArbitrationStatus  `json:"status"`
	Reason           string             `json:"reason"`
	Description      string             `json:"description"`
	Evidence         []*ArbitrationEvidence `json:"evidence"`
	HandlerID        uint64             `json:"handler_id"`
	HandlerName      string             `json:"handler_name"`
	Result           ArbitrationResult  `json:"result"`
	ResultReason     string             `json:"result_reason"`
	ResultAmount     int64              `json:"result_amount"`
	StartedAt        *time.Time         `json:"started_at"`
	EndedAt          *time.Time         `json:"ended_at"`
	Deadline         *time.Time         `json:"deadline"`
	AppealCount      int                `json:"appeal_count"`
	MaxAppeals       int                `json:"max_appeals"`
	History          []*ArbitrationHistory `json:"history"`
}

type ArbitrationEvidence struct {
	ID            uint64           `json:"id"`
	CreatedAt     time.Time        `json:"created_at"`
	ArbitrationID uint64           `json:"arbitration_id"`
	Party         ArbitrationParty `json:"party"`
	Type          EvidenceType     `json:"type"`
	Title         string           `json:"title"`
	Description   string           `json:"description"`
	Url           string           `json:"url"`
	Verified      bool             `json:"verified"`
	VerifiedAt    *time.Time       `json:"verified_at"`
}

type EvidenceType string

const (
	EvidenceTypeImage    EvidenceType = "image"
	EvidenceTypeVideo    EvidenceType = "video"
	EvidenceTypeDocument EvidenceType = "document"
	EvidenceTypeChatLog  EvidenceType = "chat_log"
	EvidenceTypeOther    EvidenceType = "other"
)

type ArbitrationHistory struct {
	ID            uint64            `json:"id"`
	CreatedAt     time.Time         `json:"created_at"`
	ArbitrationID uint64            `json:"arbitration_id"`
	Action        string            `json:"action"`
	OldStatus     ArbitrationStatus `json:"old_status"`
	NewStatus     ArbitrationStatus `json:"new_status"`
	Operator      string            `json:"operator"`
	OperatorID    uint64            `json:"operator_id"`
	Remark        string            `json:"remark"`
}

type ArbitrationConfig struct {
	ProcessingDays    int `json:"processing_days"`
	MaxAppeals        int `json:"max_appeals"`
	AppealDays        int `json:"appeal_days"`
	AutoUserWinDays   int `json:"auto_user_win_days"`
	CompromisePercent float64 `json:"compromise_percent"`
}

func DefaultArbitrationConfig() *ArbitrationConfig {
	return &ArbitrationConfig{
		ProcessingDays:    7,
		MaxAppeals:        2,
		AppealDays:        3,
		AutoUserWinDays:   48,
		CompromisePercent: 0.5,
	}
}

func NewArbitration(arbitrationNo string, afterSalesID uint64, afterSalesNo string, orderID uint64, orderNo string, userID, merchantID uint64, reason, description string, config *ArbitrationConfig) *Arbitration {
	arb := &Arbitration{
		ArbitrationNo: arbitrationNo,
		AfterSalesID:  afterSalesID,
		AfterSalesNo:  afterSalesNo,
		OrderID:       orderID,
		OrderNo:       orderNo,
		UserID:        userID,
		MerchantID:    merchantID,
		Status:        ArbitrationStatusPending,
		Reason:        reason,
		Description:   description,
		MaxAppeals:    config.MaxAppeals,
		Evidence:      []*ArbitrationEvidence{},
		History:       []*ArbitrationHistory{},
	}

	deadline := time.Now().AddDate(0, 0, config.ProcessingDays)
	arb.Deadline = &deadline

	return arb
}

func (a *Arbitration) AddEvidence(party ArbitrationParty, evidenceType EvidenceType, title, description, url string) {
	evidence := &ArbitrationEvidence{
		ArbitrationID: a.ID,
		Party:         party,
		Type:          evidenceType,
		Title:         title,
		Description:   description,
		Url:           url,
		Verified:      false,
	}
	a.Evidence = append(a.Evidence, evidence)
}

func (a *Arbitration) VerifyEvidence(evidenceID uint64) error {
	for _, e := range a.Evidence {
		if e.ID == evidenceID {
			e.Verified = true
			now := time.Now()
			e.VerifiedAt = &now
			return nil
		}
	}
	return ErrArbitrationNotFound
}

func (a *Arbitration) StartProcessing(handlerID uint64, handlerName string) error {
	if a.Status != ArbitrationStatusPending {
		return ErrArbitrationNotPending
	}

	a.Status = ArbitrationStatusProcessing
	a.HandlerID = handlerID
	a.HandlerName = handlerName
	now := time.Now()
	a.StartedAt = &now

	a.addHistory("StartProcessing", ArbitrationStatusPending, ArbitrationStatusProcessing, handlerName, handlerID, "")
	return nil
}

func (a *Arbitration) Decide(result ArbitrationResult, reason string, amount int64) error {
	if a.Status != ArbitrationStatusProcessing {
		return ErrArbitrationAlreadyEnded
	}

	a.Result = result
	a.ResultReason = reason
	a.ResultAmount = amount

	switch result {
	case ArbitrationResultUserWin:
		a.Status = ArbitrationStatusUserWin
	case ArbitrationResultMerchantWin:
		a.Status = ArbitrationStatusMerchantWin
	case ArbitrationResultCompromise:
		a.Status = ArbitrationStatusCompromise
	default:
		return ErrInvalidArbitrationResult
	}

	now := time.Now()
	a.EndedAt = &now

	a.addHistory("Decide", ArbitrationStatusProcessing, a.Status, a.HandlerName, a.HandlerID, reason)
	return nil
}

func (a *Arbitration) Appeal(party ArbitrationParty, reason string) error {
	if a.Status != ArbitrationStatusUserWin && a.Status != ArbitrationStatusMerchantWin && a.Status != ArbitrationStatusCompromise {
		return ErrArbitrationNotPending
	}

	if a.AppealCount >= a.MaxAppeals {
		return errors.New("max appeals exceeded")
	}

	a.AppealCount++
	a.Status = ArbitrationStatusPending
	a.EndedAt = nil

	operator := string(party)
	a.addHistory("Appeal", a.Status, ArbitrationStatusPending, operator, 0, reason)
	return nil
}

func (a *Arbitration) Cancel(reason string) error {
	if a.Status == ArbitrationStatusCancelled {
		return ErrArbitrationAlreadyEnded
	}

	oldStatus := a.Status
	a.Status = ArbitrationStatusCancelled
	now := time.Now()
	a.EndedAt = &now

	a.addHistory("Cancel", oldStatus, ArbitrationStatusCancelled, "System", 0, reason)
	return nil
}

func (a *Arbitration) IsEnded() bool {
	return a.Status == ArbitrationStatusUserWin ||
		a.Status == ArbitrationStatusMerchantWin ||
		a.Status == ArbitrationStatusCompromise ||
		a.Status == ArbitrationStatusCancelled
}

func (a *Arbitration) IsUserWin() bool {
	return a.Status == ArbitrationStatusUserWin
}

func (a *Arbitration) IsMerchantWin() bool {
	return a.Status == ArbitrationStatusMerchantWin
}

func (a *Arbitration) IsCompromise() bool {
	return a.Status == ArbitrationStatusCompromise
}

func (a *Arbitration) CanAppeal() bool {
	return a.AppealCount < a.MaxAppeals && !a.IsEnded()
}

func (a *Arbitration) IsOverdue() bool {
	if a.Deadline == nil {
		return false
	}
	return time.Now().After(*a.Deadline) && !a.IsEnded()
}

func (a *Arbitration) GetEvidenceByParty(party ArbitrationParty) []*ArbitrationEvidence {
	var result []*ArbitrationEvidence
	for _, e := range a.Evidence {
		if e.Party == party {
			result = append(result, e)
		}
	}
	return result
}

func (a *Arbitration) addHistory(action string, oldStatus, newStatus ArbitrationStatus, operator string, operatorID uint64, remark string) {
	a.History = append(a.History, &ArbitrationHistory{
		ArbitrationID: a.ID,
		Action:        action,
		OldStatus:     oldStatus,
		NewStatus:     newStatus,
		Operator:      operator,
		OperatorID:    operatorID,
		Remark:        remark,
	})
}

func (s ArbitrationStatus) String() string {
	switch s {
	case ArbitrationStatusPending:
		return "PENDING"
	case ArbitrationStatusProcessing:
		return "PROCESSING"
	case ArbitrationStatusUserWin:
		return "USER_WIN"
	case ArbitrationStatusMerchantWin:
		return "MERCHANT_WIN"
	case ArbitrationStatusCompromise:
		return "COMPROMISE"
	case ArbitrationStatusCancelled:
		return "CANCELLED"
	default:
		return "UNKNOWN"
	}
}

type ArbitrationRepository interface {
	Save(ctx context.Context, arbitration *Arbitration) error
	FindByID(ctx context.Context, id uint64) (*Arbitration, error)
	FindByArbitrationNo(ctx context.Context, arbitrationNo string) (*Arbitration, error)
	FindByAfterSalesID(ctx context.Context, afterSalesID uint64) (*Arbitration, error)
	FindByOrderID(ctx context.Context, orderID uint64) (*Arbitration, error)
	FindByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*Arbitration, error)
	FindByMerchantID(ctx context.Context, merchantID uint64, limit, offset int) ([]*Arbitration, error)
	FindPending(ctx context.Context, limit, offset int) ([]*Arbitration, error)
	FindProcessing(ctx context.Context, limit, offset int) ([]*Arbitration, error)
	FindOverdue(ctx context.Context) ([]*Arbitration, error)
	Update(ctx context.Context, arbitration *Arbitration) error
}

type ArbitrationService interface {
	CreateArbitration(ctx context.Context, afterSalesID uint64, reason, description string) (*Arbitration, error)
	AddEvidence(ctx context.Context, arbitrationID uint64, party ArbitrationParty, evidenceType EvidenceType, title, description, url string) error
	StartProcessing(ctx context.Context, arbitrationID, handlerID uint64, handlerName string) error
	Decide(ctx context.Context, arbitrationID uint64, result ArbitrationResult, reason string, amount int64) error
	Appeal(ctx context.Context, arbitrationID uint64, party ArbitrationParty, reason string) error
	Cancel(ctx context.Context, arbitrationID uint64, reason string) error
}
