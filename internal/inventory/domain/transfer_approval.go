package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrTransferApprovalNotFound    = errors.New("transfer approval not found")
	ErrTransferApprovalProcessed   = errors.New("transfer approval already processed")
	ErrTransferApprovalNotPending  = errors.New("transfer approval not in pending status")
	ErrTransferApprovalNotAuthorized = errors.New("not authorized for transfer approval")
)

type TransferApprovalStatus int8

const (
	TransferApprovalPending    TransferApprovalStatus = 0
	TransferApprovalApproved   TransferApprovalStatus = 1
	TransferApprovalRejected   TransferApprovalStatus = 2
	TransferApprovalCancelled  TransferApprovalStatus = 3
	TransferApprovalExpired    TransferApprovalStatus = 4
)

func (s TransferApprovalStatus) String() string {
	switch s {
	case TransferApprovalPending:
		return "PENDING"
	case TransferApprovalApproved:
		return "APPROVED"
	case TransferApprovalRejected:
		return "REJECTED"
	case TransferApprovalCancelled:
		return "CANCELLED"
	case TransferApprovalExpired:
		return "EXPIRED"
	default:
		return "UNKNOWN"
	}
}

type TransferApprovalLevel int8

const (
	ApprovalLevelWarehouse TransferApprovalLevel = 1
	ApprovalLevelManager   TransferApprovalLevel = 2
	ApprovalLevelDirector  TransferApprovalLevel = 3
	ApprovalLevelFinance   TransferApprovalLevel = 4
)

func (l TransferApprovalLevel) String() string {
	switch l {
	case ApprovalLevelWarehouse:
		return "WAREHOUSE"
	case ApprovalLevelManager:
		return "MANAGER"
	case ApprovalLevelDirector:
		return "DIRECTOR"
	case ApprovalLevelFinance:
		return "FINANCE"
	default:
		return "UNKNOWN"
	}
}

type TransferApprovalRequest struct {
	ID              uint                    `json:"id"`
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
	RequestNo       string                  `json:"request_no"`
	TransferID      uint64                  `json:"transfer_id"`
	TransferNo      string                  `json:"transfer_no"`
	SkuID           uint64                  `json:"sku_id"`
	ProductID       uint64                  `json:"product_id"`
	ProductName     string                  `json:"product_name"`
	FromWarehouseID uint64                  `json:"from_warehouse_id"`
	FromWarehouseName string                `json:"from_warehouse_name"`
	ToWarehouseID   uint64                  `json:"to_warehouse_id"`
	ToWarehouseName string                  `json:"to_warehouse_name"`
	Quantity        int32                   `json:"quantity"`
	EstimatedValue  int64                   `json:"estimated_value"`
	TransferType    TransferType            `json:"transfer_type"`
	Reason          string                  `json:"reason"`
	RequesterID     uint64                  `json:"requester_id"`
	RequesterName   string                  `json:"requester_name"`
	Status          TransferApprovalStatus  `json:"status"`
	CurrentLevel    TransferApprovalLevel   `json:"current_level"`
	TotalLevels     int                     `json:"total_levels"`
	ApprovalSteps   []*TransferApprovalStep `json:"approval_steps"`
	Histories       []*TransferApprovalHistory `json:"histories"`
	TimeoutAt       *time.Time              `json:"timeout_at"`
	ApprovedAt      *time.Time              `json:"approved_at"`
	RejectedAt      *time.Time              `json:"rejected_at"`
	CompletedAt     *time.Time              `json:"completed_at"`
}

type TransferApprovalStep struct {
	ID            uint                   `json:"id"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	RequestID     uint                   `json:"request_id"`
	Level         TransferApprovalLevel  `json:"level"`
	LevelName     string                 `json:"level_name"`
	ApproverRole  string                 `json:"approver_role"`
	ApproverID    uint64                 `json:"approver_id"`
	ApproverName  string                 `json:"approver_name"`
	Status        TransferApprovalStatus `json:"status"`
	Comment       string                 `json:"comment"`
	ApprovedAt    *time.Time             `json:"approved_at"`
	RejectedAt    *time.Time             `json:"rejected_at"`
	DueAt         *time.Time             `json:"due_at"`
	RemindedAt    *time.Time             `json:"reminded_at"`
	DelegatedFrom uint64                 `json:"delegated_from"`
}

type TransferApprovalHistory struct {
	ID           uint      `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	RequestID    uint      `json:"request_id"`
	Level        TransferApprovalLevel `json:"level"`
	OperatorID   uint64    `json:"operator_id"`
	OperatorName string    `json:"operator_name"`
	Action       string    `json:"action"`
	OldStatus    string    `json:"old_status"`
	NewStatus    string    `json:"new_status"`
	Comment      string    `json:"comment"`
}

type TransferApprovalRule struct {
	ID                 uint                   `json:"id"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
	Name               string                 `json:"name"`
	Description        string                 `json:"description"`
	TransferType       TransferType           `json:"transfer_type"`
	MinQuantity        int32                  `json:"min_quantity"`
	MaxQuantity        int32                  `json:"max_quantity"`
	MinValue           int64                  `json:"min_value"`
	MaxValue           int64                  `json:"max_value"`
	ApprovalLevels     []TransferApprovalLevel `json:"approval_levels"`
	TimeoutHours       int                    `json:"timeout_hours"`
	AutoApproveEnabled bool                   `json:"auto_approve_enabled"`
	Enabled            bool                   `json:"enabled"`
	Priority           int                    `json:"priority"`
}

type TransferApprovalConfig struct {
	DefaultTimeoutHours  int     `json:"default_timeout_hours"`
	RemindBeforeHours    int     `json:"remind_before_hours"`
	MaxApprovalLevels    int     `json:"max_approval_levels"`
	AutoApproveThreshold int32   `json:"auto_approve_threshold"`
	RequireReason        bool    `json:"require_reason"`
}

func DefaultTransferApprovalConfig() *TransferApprovalConfig {
	return &TransferApprovalConfig{
		DefaultTimeoutHours:  48,
		RemindBeforeHours:    4,
		MaxApprovalLevels:    4,
		AutoApproveThreshold: 10,
		RequireReason:        true,
	}
}

func NewTransferApprovalRequest(requestNo string, transferID uint64, transferNo string, skuID, productID uint64, productName string, fromWarehouseID, toWarehouseID uint64, fromWarehouseName, toWarehouseName string, quantity int32, estimatedValue int64, transferType TransferType, reason string, requesterID uint64, requesterName string) *TransferApprovalRequest {
	return &TransferApprovalRequest{
		RequestNo:        requestNo,
		TransferID:       transferID,
		TransferNo:       transferNo,
		SkuID:            skuID,
		ProductID:        productID,
		ProductName:      productName,
		FromWarehouseID:  fromWarehouseID,
		FromWarehouseName: fromWarehouseName,
		ToWarehouseID:    toWarehouseID,
		ToWarehouseName:  toWarehouseName,
		Quantity:         quantity,
		EstimatedValue:   estimatedValue,
		TransferType:     transferType,
		Reason:           reason,
		RequesterID:      requesterID,
		RequesterName:    requesterName,
		Status:           TransferApprovalPending,
		CurrentLevel:     ApprovalLevelWarehouse,
		TotalLevels:      1,
		ApprovalSteps:    make([]*TransferApprovalStep, 0),
		Histories:        make([]*TransferApprovalHistory, 0),
	}
}

func (r *TransferApprovalRequest) AddApprovalStep(level TransferApprovalLevel, approverRole string, dueAt *time.Time) *TransferApprovalStep {
	step := &TransferApprovalStep{
		RequestID:    r.ID,
		Level:        level,
		LevelName:    level.String(),
		ApproverRole: approverRole,
		Status:       TransferApprovalPending,
		DueAt:        dueAt,
	}
	r.ApprovalSteps = append(r.ApprovalSteps, step)
	r.TotalLevels = len(r.ApprovalSteps)
	return step
}

func (r *TransferApprovalRequest) SetTimeout(timeout time.Duration) {
	t := time.Now().Add(timeout)
	r.TimeoutAt = &t
}

func (r *TransferApprovalRequest) IsTimeout() bool {
	if r.TimeoutAt == nil {
		return false
	}
	return time.Now().After(*r.TimeoutAt)
}

func (r *TransferApprovalRequest) GetCurrentStep() *TransferApprovalStep {
	for _, step := range r.ApprovalSteps {
		if step.Status == TransferApprovalPending {
			return step
		}
	}
	return nil
}

func (r *TransferApprovalRequest) Approve(ctx context.Context, level TransferApprovalLevel, approverID uint64, approverName, comment string) error {
	if r.Status != TransferApprovalPending {
		return ErrTransferApprovalNotPending
	}

	step := r.findStepByLevel(level)
	if step == nil {
		return ErrTransferApprovalNotFound
	}

	if step.Status != TransferApprovalPending {
		return ErrTransferApprovalProcessed
	}

	oldStatus := r.Status.String()
	now := time.Now()

	step.Status = TransferApprovalApproved
	step.ApproverID = approverID
	step.ApproverName = approverName
	step.Comment = comment
	step.ApprovedAt = &now

	r.addHistory(level, approverID, approverName, "APPROVE", oldStatus, r.Status.String(), comment)

	if r.isAllStepsApproved() {
		r.Status = TransferApprovalApproved
		r.ApprovedAt = &now
		r.CompletedAt = &now
	} else {
		r.CurrentLevel = r.findNextPendingLevel()
	}

	return nil
}

func (r *TransferApprovalRequest) Reject(ctx context.Context, level TransferApprovalLevel, approverID uint64, approverName, reason string) error {
	if r.Status != TransferApprovalPending {
		return ErrTransferApprovalNotPending
	}

	step := r.findStepByLevel(level)
	if step == nil {
		return ErrTransferApprovalNotFound
	}

	if step.Status != TransferApprovalPending {
		return ErrTransferApprovalProcessed
	}

	oldStatus := r.Status.String()
	now := time.Now()

	step.Status = TransferApprovalRejected
	step.ApproverID = approverID
	step.ApproverName = approverName
	step.Comment = reason
	step.RejectedAt = &now

	r.Status = TransferApprovalRejected
	r.RejectedAt = &now
	r.CompletedAt = &now

	for _, s := range r.ApprovalSteps {
		if s.Status == TransferApprovalPending {
			s.Status = TransferApprovalCancelled
		}
	}

	r.addHistory(level, approverID, approverName, "REJECT", oldStatus, r.Status.String(), reason)

	return nil
}

func (r *TransferApprovalRequest) Cancel(operatorID uint64, operatorName, reason string) error {
	if r.Status == TransferApprovalApproved || r.Status == TransferApprovalRejected {
		return ErrTransferApprovalProcessed
	}

	oldStatus := r.Status.String()
	now := time.Now()

	r.Status = TransferApprovalCancelled
	r.CompletedAt = &now

	for _, step := range r.ApprovalSteps {
		if step.Status == TransferApprovalPending {
			step.Status = TransferApprovalCancelled
		}
	}

	r.addHistory(r.CurrentLevel, operatorID, operatorName, "CANCEL", oldStatus, r.Status.String(), reason)

	return nil
}

func (r *TransferApprovalRequest) Delegate(level TransferApprovalLevel, fromApproverID, toApproverID uint64, toApproverName string) error {
	step := r.findStepByLevel(level)
	if step == nil {
		return ErrTransferApprovalNotFound
	}

	if step.Status != TransferApprovalPending {
		return ErrTransferApprovalProcessed
	}

	step.DelegatedFrom = fromApproverID
	step.ApproverID = toApproverID
	step.ApproverName = toApproverName

	r.addHistory(level, fromApproverID, "", "DELEGATE", step.Status.String(), step.Status.String(),
		fmt.Sprintf("Delegated to %s", toApproverName))

	return nil
}

func (r *TransferApprovalRequest) Remind(level TransferApprovalLevel) error {
	step := r.findStepByLevel(level)
	if step == nil {
		return ErrTransferApprovalNotFound
	}

	if step.Status != TransferApprovalPending {
		return ErrTransferApprovalProcessed
	}

	now := time.Now()
	step.RemindedAt = &now

	return nil
}

func (r *TransferApprovalRequest) findStepByLevel(level TransferApprovalLevel) *TransferApprovalStep {
	for _, step := range r.ApprovalSteps {
		if step.Level == level {
			return step
		}
	}
	return nil
}

func (r *TransferApprovalRequest) isAllStepsApproved() bool {
	for _, step := range r.ApprovalSteps {
		if step.Status != TransferApprovalApproved {
			return false
		}
	}
	return true
}

func (r *TransferApprovalRequest) findNextPendingLevel() TransferApprovalLevel {
	for _, step := range r.ApprovalSteps {
		if step.Status == TransferApprovalPending {
			return step.Level
		}
	}
	return r.CurrentLevel
}

func (r *TransferApprovalRequest) addHistory(level TransferApprovalLevel, operatorID uint64, operatorName, action, oldStatus, newStatus, comment string) {
	history := &TransferApprovalHistory{
		RequestID:    r.ID,
		Level:        level,
		OperatorID:   operatorID,
		OperatorName: operatorName,
		Action:       action,
		OldStatus:    oldStatus,
		NewStatus:    newStatus,
		Comment:      comment,
	}
	r.Histories = append(r.Histories, history)
}

func NewTransferApprovalRule(name string, transferType TransferType) *TransferApprovalRule {
	return &TransferApprovalRule{
		Name:           name,
		TransferType:   transferType,
		ApprovalLevels: make([]TransferApprovalLevel, 0),
		Enabled:        true,
	}
}

func (rule *TransferApprovalRule) Matches(quantity int32, value int64, transferType TransferType) bool {
	if !rule.Enabled {
		return false
	}

	if rule.TransferType != transferType && rule.TransferType != 0 {
		return false
	}

	if rule.MinQuantity > 0 && quantity < rule.MinQuantity {
		return false
	}

	if rule.MaxQuantity > 0 && quantity > rule.MaxQuantity {
		return false
	}

	if rule.MinValue > 0 && value < rule.MinValue {
		return false
	}

	if rule.MaxValue > 0 && value > rule.MaxValue {
		return false
	}

	return true
}

func (rule *TransferApprovalRule) AddApprovalLevel(level TransferApprovalLevel) {
	rule.ApprovalLevels = append(rule.ApprovalLevels, level)
}

type TransferApprovalRepository interface {
	Save(ctx context.Context, request *TransferApprovalRequest) error
	FindByID(ctx context.Context, id uint64) (*TransferApprovalRequest, error)
	FindByRequestNo(ctx context.Context, requestNo string) (*TransferApprovalRequest, error)
	FindByTransferID(ctx context.Context, transferID uint64) (*TransferApprovalRequest, error)
	FindPending(ctx context.Context, limit, offset int) ([]*TransferApprovalRequest, error)
	FindPendingByApprover(ctx context.Context, approverID uint64, limit, offset int) ([]*TransferApprovalRequest, error)
	FindTimeout(ctx context.Context) ([]*TransferApprovalRequest, error)
	Update(ctx context.Context, request *TransferApprovalRequest) error
}

type TransferApprovalRuleRepository interface {
	FindByID(ctx context.Context, id uint64) (*TransferApprovalRule, error)
	FindMatchingRule(ctx context.Context, transferType TransferType, quantity int32, value int64) (*TransferApprovalRule, error)
	FindAll(ctx context.Context) ([]*TransferApprovalRule, error)
	Save(ctx context.Context, rule *TransferApprovalRule) error
}

type TransferApprovalService interface {
	InitiateApproval(ctx context.Context, transfer *StockTransfer) (*TransferApprovalRequest, error)
	Approve(ctx context.Context, requestID uint64, level TransferApprovalLevel, approverID uint64, approverName, comment string) error
	Reject(ctx context.Context, requestID uint64, level TransferApprovalLevel, approverID uint64, approverName, reason string) error
	Cancel(ctx context.Context, requestID uint64, operatorID uint64, operatorName, reason string) error
	Delegate(ctx context.Context, requestID uint64, level TransferApprovalLevel, fromApproverID, toApproverID uint64, toApproverName string) error
	GetPendingApprovals(ctx context.Context, approverID uint64, limit, offset int) ([]*TransferApprovalRequest, error)
	ProcessTimeout(ctx context.Context) error
}
