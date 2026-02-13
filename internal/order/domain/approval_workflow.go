package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrApprovalWorkflowNotFound   = errors.New("approval workflow not found")
	ErrApprovalStepNotFound       = errors.New("approval step not found")
	ErrApprovalAlreadyProcessed   = errors.New("approval already processed")
	ErrApprovalNotAuthorized      = errors.New("approver not authorized")
	ErrApprovalWorkflowNotPending = errors.New("approval workflow not in pending status")
)

type ApprovalWorkflowStatus int8

const (
	ApprovalWorkflowStatusPending    ApprovalWorkflowStatus = 0
	ApprovalWorkflowStatusInProgress ApprovalWorkflowStatus = 1
	ApprovalWorkflowStatusApproved   ApprovalWorkflowStatus = 2
	ApprovalWorkflowStatusRejected   ApprovalWorkflowStatus = 3
	ApprovalWorkflowStatusCancelled  ApprovalWorkflowStatus = 4
)

type ApprovalStepStatus int8

const (
	ApprovalStepStatusPending    ApprovalStepStatus = 0
	ApprovalStepStatusApproved   ApprovalStepStatus = 1
	ApprovalStepStatusRejected   ApprovalStepStatus = 2
	ApprovalStepStatusSkipped    ApprovalStepStatus = 3
	ApprovalStepStatusCancelled  ApprovalStepStatus = 4
)

type ApprovalType string

const (
	ApprovalTypePriceAdjustment ApprovalType = "PRICE_ADJUSTMENT"
	ApprovalTypeRefund          ApprovalType = "REFUND"
	ApprovalTypeOrderCancel     ApprovalType = "ORDER_CANCEL"
	ApprovalTypeStockTransfer   ApprovalType = "STOCK_TRANSFER"
	ApprovalTypeProductAudit    ApprovalType = "PRODUCT_AUDIT"
)

type ApprovalWorkflow struct {
	ID             uint                    `json:"id"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
	WorkflowNo     string                  `json:"workflow_no"`
	ApprovalType   ApprovalType            `json:"approval_type"`
	BusinessID     uint64                  `json:"business_id"`
	BusinessNo     string                  `json:"business_no"`
	Title          string                  `json:"title"`
	Description    string                  `json:"description"`
	Status         ApprovalWorkflowStatus  `json:"status"`
	CurrentStep    int                     `json:"current_step"`
	TotalSteps     int                     `json:"total_steps"`
	InitiatorID    uint64                  `json:"initiator_id"`
	InitiatorName  string                  `json:"initiator_name"`
	ApprovedAt     *time.Time              `json:"approved_at"`
	RejectedAt     *time.Time              `json:"rejected_at"`
	CancelledAt    *time.Time              `json:"cancelled_at"`
	CompletedAt    *time.Time              `json:"completed_at"`
	Steps          []*ApprovalStep         `json:"steps"`
	Histories      []*ApprovalHistory      `json:"histories"`
	ExtraData      map[string]any          `json:"extra_data"`
}

type ApprovalStep struct {
	ID              uint               `json:"id"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at"`
	WorkflowID      uint               `json:"workflow_id"`
	StepNo          int                `json:"step_no"`
	StepName        string             `json:"step_name"`
	ApproverRole    string             `json:"approver_role"`
	ApproverID      uint64             `json:"approver_id"`
	ApproverName    string             `json:"approver_name"`
	Status          ApprovalStepStatus `json:"status"`
	Comment         string             `json:"comment"`
	ApprovedAt      *time.Time         `json:"approved_at"`
	RejectedAt      *time.Time         `json:"rejected_at"`
	DueAt           *time.Time         `json:"due_at"`
	RemindedAt      *time.Time         `json:"reminded_at"`
	DelegatedFromID uint64             `json:"delegated_from_id"`
}

type ApprovalHistory struct {
	ID          uint       `json:"id"`
	CreatedAt   time.Time  `json:"created_at"`
	WorkflowID  uint       `json:"workflow_id"`
	StepNo      int        `json:"step_no"`
	OperatorID  uint64     `json:"operator_id"`
	OperatorName string    `json:"operator_name"`
	Action      string     `json:"action"`
	Comment     string     `json:"comment"`
	OldStatus   string     `json:"old_status"`
	NewStatus   string     `json:"new_status"`
}

type ApprovalRule struct {
	ID              uint        `json:"id"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	ApprovalType    ApprovalType `json:"approval_type"`
	Name            string      `json:"name"`
	Description     string      `json:"description"`
	MinAmount       int64       `json:"min_amount"`
	MaxAmount       int64       `json:"max_amount"`
	Steps           []*ApprovalStepConfig `json:"steps"`
	Enabled         bool        `json:"enabled"`
	Priority        int         `json:"priority"`
}

type ApprovalStepConfig struct {
	StepNo       int    `json:"step_no"`
	StepName     string `json:"step_name"`
	ApproverRole string `json:"approver_role"`
	TimeoutHours int    `json:"timeout_hours"`
	AutoApprove  bool   `json:"auto_approve"`
}

func NewApprovalWorkflow(workflowNo string, approvalType ApprovalType, businessID uint64, businessNo, title, description string, initiatorID uint64, initiatorName string) *ApprovalWorkflow {
	return &ApprovalWorkflow{
		WorkflowNo:    workflowNo,
		ApprovalType:  approvalType,
		BusinessID:    businessID,
		BusinessNo:    businessNo,
		Title:         title,
		Description:   description,
		Status:        ApprovalWorkflowStatusPending,
		CurrentStep:   1,
		TotalSteps:    0,
		InitiatorID:   initiatorID,
		InitiatorName: initiatorName,
		Steps:         make([]*ApprovalStep, 0),
		Histories:     make([]*ApprovalHistory, 0),
		ExtraData:     make(map[string]any),
	}
}

func (w *ApprovalWorkflow) AddStep(stepNo int, stepName, approverRole string, dueAt *time.Time) *ApprovalStep {
	step := &ApprovalStep{
		WorkflowID:   w.ID,
		StepNo:       stepNo,
		StepName:     stepName,
		ApproverRole: approverRole,
		Status:       ApprovalStepStatusPending,
		DueAt:        dueAt,
	}
	w.Steps = append(w.Steps, step)
	w.TotalSteps = len(w.Steps)
	return step
}

func (w *ApprovalWorkflow) Approve(ctx context.Context, stepNo int, approverID uint64, approverName, comment string) error {
	if w.Status != ApprovalWorkflowStatusPending && w.Status != ApprovalWorkflowStatusInProgress {
		return ErrApprovalWorkflowNotPending
	}

	step := w.findStep(stepNo)
	if step == nil {
		return ErrApprovalStepNotFound
	}

	if step.Status != ApprovalStepStatusPending {
		return ErrApprovalAlreadyProcessed
	}

	oldStatus := w.Status.String()
	now := time.Now()

	step.Status = ApprovalStepStatusApproved
	step.ApproverID = approverID
	step.ApproverName = approverName
	step.Comment = comment
	step.ApprovedAt = &now

	w.Status = ApprovalWorkflowStatusInProgress
	w.addHistory(stepNo, approverID, approverName, "APPROVE", comment, oldStatus, w.Status.String())

	if w.isAllStepsApproved() {
		w.Status = ApprovalWorkflowStatusApproved
		w.ApprovedAt = &now
		w.CompletedAt = &now
	} else {
		w.CurrentStep = w.findNextPendingStep()
	}

	return nil
}

func (w *ApprovalWorkflow) Reject(ctx context.Context, stepNo int, approverID uint64, approverName, comment string) error {
	if w.Status != ApprovalWorkflowStatusPending && w.Status != ApprovalWorkflowStatusInProgress {
		return ErrApprovalWorkflowNotPending
	}

	step := w.findStep(stepNo)
	if step == nil {
		return ErrApprovalStepNotFound
	}

	if step.Status != ApprovalStepStatusPending {
		return ErrApprovalAlreadyProcessed
	}

	oldStatus := w.Status.String()
	now := time.Now()

	step.Status = ApprovalStepStatusRejected
	step.ApproverID = approverID
	step.ApproverName = approverName
	step.Comment = comment
	step.RejectedAt = &now

	w.Status = ApprovalWorkflowStatusRejected
	w.RejectedAt = &now
	w.CompletedAt = &now

	for _, s := range w.Steps {
		if s.StepNo > stepNo && s.Status == ApprovalStepStatusPending {
			s.Status = ApprovalStepStatusCancelled
		}
	}

	w.addHistory(stepNo, approverID, approverName, "REJECT", comment, oldStatus, w.Status.String())

	return nil
}

func (w *ApprovalWorkflow) Cancel(operatorID uint64, operatorName, reason string) error {
	if w.Status == ApprovalWorkflowStatusApproved || w.Status == ApprovalWorkflowStatusRejected {
		return ErrApprovalWorkflowNotPending
	}

	oldStatus := w.Status.String()
	now := time.Now()

	w.Status = ApprovalWorkflowStatusCancelled
	w.CancelledAt = &now
	w.CompletedAt = &now

	for _, step := range w.Steps {
		if step.Status == ApprovalStepStatusPending {
			step.Status = ApprovalStepStatusCancelled
		}
	}

	w.addHistory(0, operatorID, operatorName, "CANCEL", reason, oldStatus, w.Status.String())

	return nil
}

func (w *ApprovalWorkflow) Delegate(stepNo int, fromApproverID, toApproverID uint64, toApproverName string) error {
	step := w.findStep(stepNo)
	if step == nil {
		return ErrApprovalStepNotFound
	}

	if step.Status != ApprovalStepStatusPending {
		return ErrApprovalAlreadyProcessed
	}

	step.DelegatedFromID = fromApproverID
	step.ApproverID = toApproverID
	step.ApproverName = toApproverName

	w.addHistory(stepNo, fromApproverID, "", "DELEGATE", 
		fmt.Sprintf("Delegated to %s", toApproverName), 
		step.Status.String(), step.Status.String())

	return nil
}

func (w *ApprovalWorkflow) Remind(stepNo int) error {
	step := w.findStep(stepNo)
	if step == nil {
		return ErrApprovalStepNotFound
	}

	if step.Status != ApprovalStepStatusPending {
		return ErrApprovalAlreadyProcessed
	}

	now := time.Now()
	step.RemindedAt = &now

	return nil
}

func (w *ApprovalWorkflow) IsTimeout() bool {
	if w.Status != ApprovalWorkflowStatusPending && w.Status != ApprovalWorkflowStatusInProgress {
		return false
	}

	for _, step := range w.Steps {
		if step.Status == ApprovalStepStatusPending && step.DueAt != nil {
			return time.Now().After(*step.DueAt)
		}
	}

	return false
}

func (w *ApprovalWorkflow) GetCurrentStep() *ApprovalStep {
	for _, step := range w.Steps {
		if step.StepNo == w.CurrentStep && step.Status == ApprovalStepStatusPending {
			return step
		}
	}
	return nil
}

func (w *ApprovalWorkflow) findStep(stepNo int) *ApprovalStep {
	for _, step := range w.Steps {
		if step.StepNo == stepNo {
			return step
		}
	}
	return nil
}

func (w *ApprovalWorkflow) isAllStepsApproved() bool {
	for _, step := range w.Steps {
		if step.Status != ApprovalStepStatusApproved && step.Status != ApprovalStepStatusSkipped {
			return false
		}
	}
	return true
}

func (w *ApprovalWorkflow) findNextPendingStep() int {
	for _, step := range w.Steps {
		if step.Status == ApprovalStepStatusPending {
			return step.StepNo
		}
	}
	return w.CurrentStep
}

func (w *ApprovalWorkflow) addHistory(stepNo int, operatorID uint64, operatorName, action, comment, oldStatus, newStatus string) {
	history := &ApprovalHistory{
		WorkflowID:   w.ID,
		StepNo:       stepNo,
		OperatorID:   operatorID,
		OperatorName: operatorName,
		Action:       action,
		Comment:      comment,
		OldStatus:    oldStatus,
		NewStatus:    newStatus,
	}
	w.Histories = append(w.Histories, history)
}

func (s ApprovalWorkflowStatus) String() string {
	switch s {
	case ApprovalWorkflowStatusPending:
		return "PENDING"
	case ApprovalWorkflowStatusInProgress:
		return "IN_PROGRESS"
	case ApprovalWorkflowStatusApproved:
		return "APPROVED"
	case ApprovalWorkflowStatusRejected:
		return "REJECTED"
	case ApprovalWorkflowStatusCancelled:
		return "CANCELLED"
	default:
		return "UNKNOWN"
	}
}

func (s ApprovalStepStatus) String() string {
	switch s {
	case ApprovalStepStatusPending:
		return "PENDING"
	case ApprovalStepStatusApproved:
		return "APPROVED"
	case ApprovalStepStatusRejected:
		return "REJECTED"
	case ApprovalStepStatusSkipped:
		return "SKIPPED"
	case ApprovalStepStatusCancelled:
		return "CANCELLED"
	default:
		return "UNKNOWN"
	}
}

type ApprovalWorkflowRepository interface {
	Save(ctx context.Context, workflow *ApprovalWorkflow) error
	FindByID(ctx context.Context, id uint64) (*ApprovalWorkflow, error)
	FindByWorkflowNo(ctx context.Context, workflowNo string) (*ApprovalWorkflow, error)
	FindByBusinessID(ctx context.Context, approvalType ApprovalType, businessID uint64) (*ApprovalWorkflow, error)
	FindPendingByApprover(ctx context.Context, approverID uint64, limit, offset int) ([]*ApprovalWorkflow, error)
	FindTimeoutWorkflows(ctx context.Context) ([]*ApprovalWorkflow, error)
	Update(ctx context.Context, workflow *ApprovalWorkflow) error
}

type ApprovalRuleRepository interface {
	FindMatchingRule(ctx context.Context, approvalType ApprovalType, amount int64) (*ApprovalRule, error)
	FindAll(ctx context.Context) ([]*ApprovalRule, error)
	Save(ctx context.Context, rule *ApprovalRule) error
}

type ApprovalWorkflowService interface {
	InitiateWorkflow(ctx context.Context, approvalType ApprovalType, businessID uint64, businessNo, title, description string, initiatorID uint64, initiatorName string, amount int64) (*ApprovalWorkflow, error)
	Approve(ctx context.Context, workflowID uint64, stepNo int, approverID uint64, approverName, comment string) error
	Reject(ctx context.Context, workflowID uint64, stepNo int, approverID uint64, approverName, comment string) error
	Cancel(ctx context.Context, workflowID uint64, operatorID uint64, operatorName, reason string) error
	Delegate(ctx context.Context, workflowID uint64, stepNo int, fromApproverID, toApproverID uint64, toApproverName string) error
	GetPendingApprovals(ctx context.Context, approverID uint64, limit, offset int) ([]*ApprovalWorkflow, error)
	ProcessTimeoutWorkflows(ctx context.Context) error
}
