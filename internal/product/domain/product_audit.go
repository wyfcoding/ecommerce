package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrProductAuditNotFound     = errors.New("product audit not found")
	ErrProductAuditProcessed    = errors.New("product audit already processed")
	ErrProductAuditNotPending   = errors.New("product audit not in pending status")
	ErrProductAuditNotAuthorized = errors.New("not authorized for product audit")
)

type ProductAuditType int8

const (
	ProductAuditTypeCreate    ProductAuditType = 1
	ProductAuditTypeUpdate    ProductAuditType = 2
	ProductAuditTypePrice     ProductAuditType = 3
	ProductAuditTypeOffline   ProductAuditType = 4
	ProductAuditTypeOnline    ProductAuditType = 5
	ProductAuditTypeDelete    ProductAuditType = 6
)

func (t ProductAuditType) String() string {
	switch t {
	case ProductAuditTypeCreate:
		return "CREATE"
	case ProductAuditTypeUpdate:
		return "UPDATE"
	case ProductAuditTypePrice:
		return "PRICE"
	case ProductAuditTypeOffline:
		return "OFFLINE"
	case ProductAuditTypeOnline:
		return "ONLINE"
	case ProductAuditTypeDelete:
		return "DELETE"
	default:
		return "UNKNOWN"
	}
}

type ProductAuditStatus int8

const (
	ProductAuditStatusPending    ProductAuditStatus = 1
	ProductAuditStatusApproved   ProductAuditStatus = 2
	ProductAuditStatusRejected   ProductAuditStatus = 3
	ProductAuditStatusCancelled  ProductAuditStatus = 4
)

func (s ProductAuditStatus) String() string {
	switch s {
	case ProductAuditStatusPending:
		return "PENDING"
	case ProductAuditStatusApproved:
		return "APPROVED"
	case ProductAuditStatusRejected:
		return "REJECTED"
	case ProductAuditStatusCancelled:
		return "CANCELLED"
	default:
		return "UNKNOWN"
	}
}

type ProductAudit struct {
	ID              uint               `json:"id"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
	AuditNo         string             `json:"audit_no"`
	ProductID       uint64             `json:"product_id"`
	ProductName     string             `json:"product_name"`
	MerchantID      uint64             `json:"merchant_id"`
	MerchantName    string             `json:"merchant_name"`
	AuditType       ProductAuditType   `json:"audit_type"`
	Status          ProductAuditStatus `json:"status"`
	OldData         string             `json:"old_data"`
	NewData         string             `json:"new_data"`
	Changes         []*ProductChange   `json:"changes"`
	Reason          string             `json:"reason"`
	RequesterID     uint64             `json:"requester_id"`
	RequesterName   string             `json:"requester_name"`
	ApproverID      uint64             `json:"approver_id"`
	ApproverName    string             `json:"approver_name"`
	ApprovedAt      *time.Time         `json:"approved_at"`
	RejectedAt      *time.Time         `json:"rejected_at"`
	RejectionReason string             `json:"rejection_reason"`
	CompletedAt     *time.Time         `json:"completed_at"`
	TimeoutAt       *time.Time         `json:"timeout_at"`
	Priority        int                `json:"priority"`
	Comment         string             `json:"comment"`
	Histories       []*ProductAuditHistory `json:"histories"`
}

type ProductChange struct {
	Field      string `json:"field"`
	OldValue   string `json:"old_value"`
	NewValue   string `json:"new_value"`
	ChangeType string `json:"change_type"`
}

type ProductAuditHistory struct {
	ID           uint      `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	AuditID      uint      `json:"audit_id"`
	OperatorID   uint64    `json:"operator_id"`
	OperatorName string    `json:"operator_name"`
	Action       string    `json:"action"`
	OldStatus    string    `json:"old_status"`
	NewStatus    string    `json:"new_status"`
	Comment      string    `json:"comment"`
}

type ProductAuditRule struct {
	ID              uint            `json:"id"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	AuditType       ProductAuditType `json:"audit_type"`
	MinPrice        int64           `json:"min_price"`
	MaxPrice        int64           `json:"max_price"`
	CategoryIDs     []uint64        `json:"category_ids"`
	RequireApproval bool            `json:"require_approval"`
	AutoApprove     bool            `json:"auto_approve"`
	TimeoutHours    int             `json:"timeout_hours"`
	Enabled         bool            `json:"enabled"`
	Priority        int             `json:"priority"`
}

type ProductAuditConfig struct {
	DefaultTimeoutHours int `json:"default_timeout_hours"`
	MaxPendingAudits    int `json:"max_pending_audits"`
	AutoApproveEnabled  bool `json:"auto_approve_enabled"`
	RequireReason       bool `json:"require_reason"`
}

func DefaultProductAuditConfig() *ProductAuditConfig {
	return &ProductAuditConfig{
		DefaultTimeoutHours: 24,
		MaxPendingAudits:    1000,
		AutoApproveEnabled:  false,
		RequireReason:       true,
	}
}

func NewProductAudit(auditNo string, productID uint64, productName string, merchantID uint64, merchantName string, auditType ProductAuditType, oldData, newData string, changes []*ProductChange, reason string, requesterID uint64, requesterName string) *ProductAudit {
	return &ProductAudit{
		AuditNo:       auditNo,
		ProductID:     productID,
		ProductName:   productName,
		MerchantID:    merchantID,
		MerchantName:  merchantName,
		AuditType:     auditType,
		Status:        ProductAuditStatusPending,
		OldData:       oldData,
		NewData:       newData,
		Changes:       changes,
		Reason:        reason,
		RequesterID:   requesterID,
		RequesterName: requesterName,
		Priority:      0,
		Histories:     make([]*ProductAuditHistory, 0),
	}
}

func (a *ProductAudit) SetTimeout(timeout time.Duration) {
	t := time.Now().Add(timeout)
	a.TimeoutAt = &t
}

func (a *ProductAudit) IsTimeout() bool {
	if a.TimeoutAt == nil {
		return false
	}
	return time.Now().After(*a.TimeoutAt)
}

func (a *ProductAudit) IsPending() bool {
	return a.Status == ProductAuditStatusPending
}

func (a *ProductAudit) Approve(approverID uint64, approverName, comment string) error {
	if a.Status != ProductAuditStatusPending {
		return ErrProductAuditProcessed
	}

	oldStatus := a.Status.String()
	now := time.Now()

	a.Status = ProductAuditStatusApproved
	a.ApproverID = approverID
	a.ApproverName = approverName
	a.ApprovedAt = &now
	a.CompletedAt = &now
	a.Comment = comment

	a.addHistory(approverID, approverName, "APPROVE", oldStatus, a.Status.String(), comment)

	return nil
}

func (a *ProductAudit) Reject(approverID uint64, approverName, reason string) error {
	if a.Status != ProductAuditStatusPending {
		return ErrProductAuditProcessed
	}

	oldStatus := a.Status.String()
	now := time.Now()

	a.Status = ProductAuditStatusRejected
	a.ApproverID = approverID
	a.ApproverName = approverName
	a.RejectionReason = reason
	a.RejectedAt = &now
	a.CompletedAt = &now

	a.addHistory(approverID, approverName, "REJECT", oldStatus, a.Status.String(), reason)

	return nil
}

func (a *ProductAudit) Cancel(operatorID uint64, operatorName, reason string) error {
	if a.Status != ProductAuditStatusPending {
		return ErrProductAuditProcessed
	}

	oldStatus := a.Status.String()
	now := time.Now()

	a.Status = ProductAuditStatusCancelled
	a.CompletedAt = &now

	a.addHistory(operatorID, operatorName, "CANCEL", oldStatus, a.Status.String(), reason)

	return nil
}

func (a *ProductAudit) addHistory(operatorID uint64, operatorName, action, oldStatus, newStatus, comment string) {
	history := &ProductAuditHistory{
		AuditID:      a.ID,
		OperatorID:   operatorID,
		OperatorName: operatorName,
		Action:       action,
		OldStatus:    oldStatus,
		NewStatus:    newStatus,
		Comment:      comment,
	}
	a.Histories = append(a.Histories, history)
}

type ProductAuditRepository interface {
	Save(ctx context.Context, audit *ProductAudit) error
	FindByID(ctx context.Context, id uint64) (*ProductAudit, error)
	FindByAuditNo(ctx context.Context, auditNo string) (*ProductAudit, error)
	FindByProductID(ctx context.Context, productID uint64) ([]*ProductAudit, error)
	FindByMerchantID(ctx context.Context, merchantID uint64, limit, offset int) ([]*ProductAudit, error)
	FindPending(ctx context.Context, limit, offset int) ([]*ProductAudit, error)
	FindTimeout(ctx context.Context) ([]*ProductAudit, error)
	Update(ctx context.Context, audit *ProductAudit) error
}

type ProductAuditRuleRepository interface {
	FindByID(ctx context.Context, id uint64) (*ProductAuditRule, error)
	FindMatchingRule(ctx context.Context, auditType ProductAuditType, price int64, categoryID uint64) (*ProductAuditRule, error)
	FindAll(ctx context.Context) ([]*ProductAuditRule, error)
	Save(ctx context.Context, rule *ProductAuditRule) error
}
