package domain

import (
	"errors"
	"time"
)

var (
	ErrApprovalNotFound      = errors.New("approval not found")
	ErrApprovalAlreadyExists = errors.New("approval already exists")
	ErrApprovalNotPending    = errors.New("approval is not pending")
	ErrApprovalAlreadyHandled = errors.New("approval already handled")
	ErrInvalidApprovalAction = errors.New("invalid approval action")
)

type ApprovalStatus string

const (
	ApprovalStatusPending   ApprovalStatus = "PENDING"
	ApprovalStatusApproved  ApprovalStatus = "APPROVED"
	ApprovalStatusRejected  ApprovalStatus = "REJECTED"
	ApprovalStatusCancelled ApprovalStatus = "CANCELLED"
	ApprovalStatusExpired   ApprovalStatus = "EXPIRED"
)

type ApprovalType string

const (
	ApprovalTypeCampaign     ApprovalType = "CAMPAIGN"
	ApprovalTypeCoupon       ApprovalType = "COUPON"
	ApprovalTypeFlashsale    ApprovalType = "FLASHSALE"
	ApprovalTypeGroupbuy     ApprovalType = "GROUPBUY"
	ApprovalTypeBanner       ApprovalType = "BANNER"
	ApprovalTypePromotion    ApprovalType = "PROMOTION"
)

type ApprovalPriority string

const (
	ApprovalPriorityLow    ApprovalPriority = "LOW"
	ApprovalPriorityMedium ApprovalPriority = "MEDIUM"
	ApprovalPriorityHigh   ApprovalPriority = "HIGH"
	ApprovalPriorityUrgent ApprovalPriority = "URGENT"
)

type CampaignApproval struct {
	ID             uint             `json:"id"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
	ApprovalNo     string           `json:"approval_no"`
	ApprovalType   ApprovalType     `json:"approval_type"`
	TargetID       uint64           `json:"target_id"`
	TargetName     string           `json:"target_name"`
	Title          string           `json:"title"`
	Description    string           `json:"description"`
	Priority       ApprovalPriority `json:"priority"`
	Status         ApprovalStatus   `json:"status"`
	RequesterID    uint64           `json:"requester_id"`
	RequesterName  string           `json:"requester_name"`
	Department     string           `json:"department"`
	Budget         int64            `json:"budget"`
	StartTime      time.Time        `json:"start_time"`
	EndTime        time.Time        `json:"end_time"`
	Attachments    []string         `json:"attachments"`
	ApprovalFlow   *ApprovalFlow    `json:"approval_flow"`
	CurrentStep    int              `json:"current_step"`
	ApprovalHistory []*ApprovalRecord `json:"approval_history"`
	ExpiresAt      *time.Time       `json:"expires_at"`
	ApprovedAt     *time.Time       `json:"approved_at"`
	RejectedAt     *time.Time       `json:"rejected_at"`
	CancelledAt    *time.Time       `json:"cancelled_at"`
	CancelReason   string           `json:"cancel_reason"`
	Remarks        string           `json:"remarks"`
}

type ApprovalFlow struct {
	ID        uint            `json:"id"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	Name      string          `json:"name"`
	Type      ApprovalType    `json:"type"`
	Steps     []*ApprovalStep `json:"steps"`
	Enabled   bool            `json:"enabled"`
	Version   int             `json:"version"`
}

type ApprovalStep struct {
	ID           uint             `json:"id"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`
	FlowID       uint             `json:"flow_id"`
	StepNo       int              `json:"step_no"`
	Name         string           `json:"name"`
	Description  string           `json:"description"`
	Approvers    []*Approver      `json:"approvers"`
	ApproveType  string           `json:"approve_type"`
	MinApprovals int              `json:"min_approvals"`
	Timeout      int              `json:"timeout"`
	TimeoutAction string          `json:"timeout_action"`
}

type Approver struct {
	ID         uint   `json:"id"`
	UserID     uint64 `json:"user_id"`
	UserName   string `json:"user_name"`
	RoleID     uint64 `json:"role_id"`
	RoleName   string `json:"role_name"`
	Department string `json:"department"`
}

type ApprovalRecord struct {
	ID          uint           `json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	ApprovalID  uint           `json:"approval_id"`
	StepNo      int            `json:"step_no"`
	ApproverID  uint64         `json:"approver_id"`
	ApproverName string        `json:"approver_name"`
	Action      string         `json:"action"`
	Comment     string         `json:"comment"`
	Attachments []string       `json:"attachments"`
	ApprovedAt  time.Time      `json:"approved_at"`
}

type ApprovalNotification struct {
	ID           uint           `json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	ApprovalID   uint           `json:"approval_id"`
	ApprovalNo   string         `json:"approval_no"`
	ApproverID   uint64         `json:"approver_id"`
	Title        string         `json:"title"`
	Content      string         `json:"content"`
	Type         string         `json:"type"`
	Read         bool           `json:"read"`
	ReadAt       *time.Time     `json:"read_at"`
	SentAt       *time.Time     `json:"sent_at"`
}

func NewCampaignApproval(approvalType ApprovalType, targetID uint64, targetName, title, description string,
	requesterID uint64, requesterName, department string, budget int64, startTime, endTime time.Time) *CampaignApproval {
	return &CampaignApproval{
		ApprovalNo:     generateApprovalNo(),
		ApprovalType:   approvalType,
		TargetID:       targetID,
		TargetName:     targetName,
		Title:          title,
		Description:    description,
		Priority:       ApprovalPriorityMedium,
		Status:         ApprovalStatusPending,
		RequesterID:    requesterID,
		RequesterName:  requesterName,
		Department:     department,
		Budget:         budget,
		StartTime:      startTime,
		EndTime:        endTime,
		Attachments:    make([]string, 0),
		ApprovalHistory: make([]*ApprovalRecord, 0),
		CurrentStep:    1,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

func generateApprovalNo() string {
	return "APR" + time.Now().Format("20060102150405")
}

func (a *CampaignApproval) SetApprovalFlow(flow *ApprovalFlow) {
	a.ApprovalFlow = flow
	a.UpdatedAt = time.Now()
}

func (a *CampaignApproval) SetPriority(priority ApprovalPriority) {
	a.Priority = priority
	a.UpdatedAt = time.Now()
}

func (a *CampaignApproval) SetExpiresAt(expiresAt time.Time) {
	a.ExpiresAt = &expiresAt
	a.UpdatedAt = time.Now()
}

func (a *CampaignApproval) AddAttachment(url string) {
	a.Attachments = append(a.Attachments, url)
	a.UpdatedAt = time.Now()
}

func (a *CampaignApproval) Approve(approverID uint64, approverName, comment string) error {
	if a.Status != ApprovalStatusPending {
		return ErrApprovalNotPending
	}
	
	record := &ApprovalRecord{
		ApprovalID:   a.ID,
		StepNo:       a.CurrentStep,
		ApproverID:   approverID,
		ApproverName: approverName,
		Action:       "APPROVE",
		Comment:      comment,
		ApprovedAt:   time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	a.ApprovalHistory = append(a.ApprovalHistory, record)
	
	if a.ApprovalFlow != nil && a.CurrentStep < len(a.ApprovalFlow.Steps) {
		a.CurrentStep++
	} else {
		a.Status = ApprovalStatusApproved
		now := time.Now()
		a.ApprovedAt = &now
	}
	
	a.UpdatedAt = time.Now()
	return nil
}

func (a *CampaignApproval) Reject(approverID uint64, approverName, reason string) error {
	if a.Status != ApprovalStatusPending {
		return ErrApprovalNotPending
	}
	
	record := &ApprovalRecord{
		ApprovalID:   a.ID,
		StepNo:       a.CurrentStep,
		ApproverID:   approverID,
		ApproverName: approverName,
		Action:       "REJECT",
		Comment:      reason,
		ApprovedAt:   time.Now(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	a.ApprovalHistory = append(a.ApprovalHistory, record)
	
	a.Status = ApprovalStatusRejected
	now := time.Now()
	a.RejectedAt = &now
	a.UpdatedAt = now
	return nil
}

func (a *CampaignApproval) Cancel(reason string) error {
	if a.Status != ApprovalStatusPending {
		return ErrApprovalNotPending
	}
	
	a.Status = ApprovalStatusCancelled
	a.CancelReason = reason
	now := time.Now()
	a.CancelledAt = &now
	a.UpdatedAt = now
	return nil
}

func (a *CampaignApproval) Expire() {
	a.Status = ApprovalStatusExpired
	a.UpdatedAt = time.Now()
}

func (a *CampaignApproval) IsExpired() bool {
	if a.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*a.ExpiresAt)
}

func (a *CampaignApproval) CanModify() bool {
	return a.Status == ApprovalStatusPending && a.CurrentStep == 1
}

func (a *CampaignApproval) GetCurrentApprovers() []*Approver {
	if a.ApprovalFlow == nil || a.CurrentStep > len(a.ApprovalFlow.Steps) {
		return nil
	}
	return a.ApprovalFlow.Steps[a.CurrentStep-1].Approvers
}

func NewApprovalFlow(name string, approvalType ApprovalType) *ApprovalFlow {
	return &ApprovalFlow{
		Name:      name,
		Type:      approvalType,
		Steps:     make([]*ApprovalStep, 0),
		Enabled:   true,
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (f *ApprovalFlow) AddStep(step *ApprovalStep) {
	step.FlowID = f.ID
	step.StepNo = len(f.Steps) + 1
	f.Steps = append(f.Steps, step)
	f.UpdatedAt = time.Now()
}

func (f *ApprovalFlow) Enable() {
	f.Enabled = true
	f.UpdatedAt = time.Now()
}

func (f *ApprovalFlow) Disable() {
	f.Enabled = false
	f.UpdatedAt = time.Now()
}

func (f *ApprovalFlow) NewVersion() {
	f.Version++
	f.UpdatedAt = time.Now()
}

func NewApprovalStep(name, description string, approveType string, minApprovals, timeout int) *ApprovalStep {
	return &ApprovalStep{
		Name:         name,
		Description:  description,
		Approvers:    make([]*Approver, 0),
		ApproveType:  approveType,
		MinApprovals: minApprovals,
		Timeout:      timeout,
		TimeoutAction: "ESCALATE",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

func (s *ApprovalStep) AddApprover(approver *Approver) {
	s.Approvers = append(s.Approvers, approver)
	s.UpdatedAt = time.Now()
}

func (s *ApprovalStep) RemoveApprover(userID uint64) {
	for i, a := range s.Approvers {
		if a.UserID == userID {
			s.Approvers = append(s.Approvers[:i], s.Approvers[i+1:]...)
			break
		}
	}
	s.UpdatedAt = time.Now()
}

type ApprovalRepository interface {
	Save(ctx interface{}, approval *CampaignApproval) error
	Update(ctx interface{}, approval *CampaignApproval) error
	FindByID(ctx interface{}, id uint) (*CampaignApproval, error)
	FindByApprovalNo(ctx interface{}, approvalNo string) (*CampaignApproval, error)
	FindByTargetID(ctx interface{}, approvalType ApprovalType, targetID uint64) ([]*CampaignApproval, error)
	FindPendingByRequester(ctx interface{}, requesterID uint64) ([]*CampaignApproval, error)
	FindPendingByApprover(ctx interface{}, approverID uint64) ([]*CampaignApproval, error)
	FindExpired(ctx interface{}) ([]*CampaignApproval, error)
	
	SaveFlow(ctx interface{}, flow *ApprovalFlow) error
	UpdateFlow(ctx interface{}, flow *ApprovalFlow) error
	FindFlowByID(ctx interface{}, id uint) (*ApprovalFlow, error)
	FindFlowByType(ctx interface{}, approvalType ApprovalType) (*ApprovalFlow, error)
	FindEnabledFlows(ctx interface{}) ([]*ApprovalFlow, error)
	
	SaveNotification(ctx interface{}, notification *ApprovalNotification) error
	FindNotificationByID(ctx interface{}, id uint) (*ApprovalNotification, error)
	FindUnreadNotifications(ctx interface{}, approverID uint64) ([]*ApprovalNotification, error)
	MarkNotificationRead(ctx interface{}, id uint) error
}

type ApprovalService interface {
	CreateApproval(ctx interface{}, approvalType ApprovalType, targetID uint64, 
		title, description string, requesterID uint64) (*CampaignApproval, error)
	Approve(ctx interface{}, approvalID uint, approverID uint64, comment string) error
	Reject(ctx interface{}, approvalID uint, approverID uint64, reason string) error
	Cancel(ctx interface{}, approvalID uint, reason string) error
	GetPendingApprovals(ctx interface{}, approverID uint64) ([]*CampaignApproval, error)
	GetApprovalHistory(ctx interface{}, approvalID uint) ([]*ApprovalRecord, error)
	ExpireApprovals(ctx interface{}) error
}
