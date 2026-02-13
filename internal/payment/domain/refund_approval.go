package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrRefundRequestNotFound    = errors.New("refund request not found")
	ErrRefundAlreadyProcessed   = errors.New("refund request already processed")
	ErrRefundNotAuthorized      = errors.New("not authorized to process refund")
	ErrRefundAmountExceeded     = errors.New("refund amount exceeds order amount")
	ErrRefundTimeExpired        = errors.New("refund time expired")
	ErrInvalidRefundStatus      = errors.New("invalid refund status for operation")
)

type RefundType int8

const (
	RefundTypeFull      RefundType = 1
	RefundTypePartial   RefundType = 2
	RefundTypeOnlyGoods RefundType = 3
	RefundTypeOnlyMoney RefundType = 4
)

func (t RefundType) String() string {
	switch t {
	case RefundTypeFull:
		return "FULL"
	case RefundTypePartial:
		return "PARTIAL"
	case RefundTypeOnlyGoods:
		return "ONLY_GOODS"
	case RefundTypeOnlyMoney:
		return "ONLY_MONEY"
	default:
		return "UNKNOWN"
	}
}

type RefundReason int8

const (
	RefundReasonUserRequest     RefundReason = 1
	RefundReasonQualityIssue    RefundReason = 2
	RefundReasonWrongItem       RefundReason = 3
	RefundReasonDamage          RefundReason = 4
	RefundReasonLateDelivery    RefundReason = 5
	RefundReasonPriceError      RefundReason = 6
	RefundReasonSystemError     RefundReason = 7
	RefundReasonMerchantAgree   RefundReason = 8
	RefundReasonPlatformIntervene RefundReason = 9
)

func (r RefundReason) String() string {
	switch r {
	case RefundReasonUserRequest:
		return "USER_REQUEST"
	case RefundReasonQualityIssue:
		return "QUALITY_ISSUE"
	case RefundReasonWrongItem:
		return "WRONG_ITEM"
	case RefundReasonDamage:
		return "DAMAGE"
	case RefundReasonLateDelivery:
		return "LATE_DELIVERY"
	case RefundReasonPriceError:
		return "PRICE_ERROR"
	case RefundReasonSystemError:
		return "SYSTEM_ERROR"
	case RefundReasonMerchantAgree:
		return "MERCHANT_AGREE"
	case RefundReasonPlatformIntervene:
		return "PLATFORM_INTERVENE"
	default:
		return "UNKNOWN"
	}
}

type RefundRequestStatus int8

const (
	RefundStatusPending      RefundRequestStatus = 1
	RefundStatusMerchantReview RefundRequestStatus = 2
	RefundStatusPlatformReview RefundRequestStatus = 3
	RefundStatusApproved     RefundRequestStatus = 4
	RefundStatusRejected     RefundRequestStatus = 5
	RefundStatusProcessing   RefundRequestStatus = 6
	RefundStatusSuccess      RefundRequestStatus = 7
	RefundStatusFailed       RefundRequestStatus = 8
	RefundStatusCancelled    RefundRequestStatus = 9
)

func (s RefundRequestStatus) String() string {
	switch s {
	case RefundStatusPending:
		return "PENDING"
	case RefundStatusMerchantReview:
		return "MERCHANT_REVIEW"
	case RefundStatusPlatformReview:
		return "PLATFORM_REVIEW"
	case RefundStatusApproved:
		return "APPROVED"
	case RefundStatusRejected:
		return "REJECTED"
	case RefundStatusProcessing:
		return "PROCESSING"
	case RefundStatusSuccess:
		return "SUCCESS"
	case RefundStatusFailed:
		return "FAILED"
	case RefundStatusCancelled:
		return "CANCELLED"
	default:
		return "UNKNOWN"
	}
}

type RefundRequest struct {
	ID                 uint                 `json:"id"`
	CreatedAt          time.Time            `json:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at"`
	RefundNo           string               `json:"refund_no"`
	OrderID            uint64               `json:"order_id"`
	OrderNo            string               `json:"order_no"`
	PaymentID          uint64               `json:"payment_id"`
	PaymentNo          string               `json:"payment_no"`
	UserID             uint64               `json:"user_id"`
	MerchantID         uint64               `json:"merchant_id"`
	RefundType         RefundType           `json:"refund_type"`
	RefundReason       RefundReason         `json:"refund_reason"`
	ReasonDescription  string               `json:"reason_description"`
	Status             RefundRequestStatus  `json:"status"`
	OriginalAmount     int64                `json:"original_amount"`
	RequestedAmount    int64                `json:"requested_amount"`
	ApprovedAmount     int64                `json:"approved_amount"`
	ActualRefundAmount int64                `json:"actual_refund_amount"`
	Currency           string               `json:"currency"`
	Items              []*RefundItem        `json:"items"`
	Evidence           []*RefundEvidence    `json:"evidence"`
	ApprovalFlow       *RefundApprovalFlow  `json:"approval_flow"`
	Histories          []*RefundHistory     `json:"histories"`
	TimeoutAt          *time.Time           `json:"timeout_at"`
	ProcessedAt        *time.Time           `json:"processed_at"`
	CompletedAt        *time.Time           `json:"completed_at"`
	Channel            string               `json:"channel"`
	ChannelRefundNo    string               `json:"channel_refund_no"`
	FailureReason      string               `json:"failure_reason"`
}

type RefundItem struct {
	ID            uint    `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	RefundID      uint    `json:"refund_id"`
	OrderItemID   uint64  `json:"order_item_id"`
	ProductID     uint64  `json:"product_id"`
	SkuID         uint64  `json:"sku_id"`
	ProductName   string  `json:"product_name"`
	SkuName       string  `json:"sku_name"`
	Quantity      int32   `json:"quantity"`
	OriginalPrice int64   `json:"original_price"`
	RefundAmount  int64   `json:"refund_amount"`
}

type RefundEvidence struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	RefundID    uint      `json:"refund_id"`
	Type        string    `json:"type"`
	URL         string    `json:"url"`
	Description string    `json:"description"`
}

type RefundApprovalFlow struct {
	ID               uint                   `json:"id"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	RefundID         uint                   `json:"refund_id"`
	CurrentStep      int                    `json:"current_step"`
	TotalSteps       int                    `json:"total_steps"`
	Steps            []*RefundApprovalStep  `json:"steps"`
	AutoApprove      bool                   `json:"auto_approve"`
	AutoApproveAfter time.Duration          `json:"auto_approve_after"`
}

type RefundApprovalStep struct {
	ID           uint                  `json:"id"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
	FlowID       uint                  `json:"flow_id"`
	StepNo       int                   `json:"step_no"`
	StepName     string                `json:"step_name"`
	ApproverRole string                `json:"approver_role"`
	ApproverID   uint64                `json:"approver_id"`
	ApproverName string                `json:"approver_name"`
	Status       RefundApprovalStatus  `json:"status"`
	Comment      string                `json:"comment"`
	ApprovedAt   *time.Time            `json:"approved_at"`
	RejectedAt   *time.Time            `json:"rejected_at"`
	DueAt        *time.Time            `json:"due_at"`
}

type RefundApprovalStatus int8

const (
	RefundApprovalPending  RefundApprovalStatus = 0
	RefundApprovalApproved RefundApprovalStatus = 1
	RefundApprovalRejected RefundApprovalStatus = 2
)

type RefundHistory struct {
	ID           uint      `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	RefundID     uint      `json:"refund_id"`
	OperatorID   uint64    `json:"operator_id"`
	OperatorName string    `json:"operator_name"`
	OperatorRole string    `json:"operator_role"`
	Action       string    `json:"action"`
	OldStatus    string    `json:"old_status"`
	NewStatus    string    `json:"new_status"`
	Comment      string    `json:"comment"`
}

type RefundConfig struct {
	MaxRefundDays        int           `json:"max_refund_days"`
	AutoApproveDays      int           `json:"auto_approve_days"`
	MerchantReviewHours  int           `json:"merchant_review_hours"`
	PlatformReviewHours  int           `json:"platform_review_hours"`
	MaxRefundPercent     float64       `json:"max_refund_percent"`
	RequireEvidence      bool          `json:"require_evidence"`
	MinEvidenceCount     int           `json:"min_evidence_count"`
}

func DefaultRefundConfig() *RefundConfig {
	return &RefundConfig{
		MaxRefundDays:        30,
		AutoApproveDays:      7,
		MerchantReviewHours:  48,
		PlatformReviewHours:  24,
		MaxRefundPercent:     100,
		RequireEvidence:      true,
		MinEvidenceCount:     1,
	}
}

func NewRefundRequest(refundNo string, orderID uint64, orderNo string, paymentID uint64, paymentNo string, userID, merchantID uint64, refundType RefundType, refundReason RefundReason, reasonDescription string, originalAmount, requestedAmount int64, currency string) *RefundRequest {
	return &RefundRequest{
		RefundNo:          refundNo,
		OrderID:           orderID,
		OrderNo:           orderNo,
		PaymentID:         paymentID,
		PaymentNo:         paymentNo,
		UserID:            userID,
		MerchantID:        merchantID,
		RefundType:        refundType,
		RefundReason:      refundReason,
		ReasonDescription: reasonDescription,
		Status:            RefundStatusPending,
		OriginalAmount:    originalAmount,
		RequestedAmount:   requestedAmount,
		Currency:          currency,
		Items:             make([]*RefundItem, 0),
		Evidence:          make([]*RefundEvidence, 0),
		Histories:         make([]*RefundHistory, 0),
	}
}

func (r *RefundRequest) AddItem(item *RefundItem) {
	r.Items = append(r.Items, item)
}

func (r *RefundRequest) AddEvidence(evidence *RefundEvidence) {
	r.Evidence = append(r.Evidence, evidence)
}

func (r *RefundRequest) SetTimeout(timeout time.Duration) {
	t := time.Now().Add(timeout)
	r.TimeoutAt = &t
}

func (r *RefundRequest) IsTimeout() bool {
	if r.TimeoutAt == nil {
		return false
	}
	return time.Now().After(*r.TimeoutAt)
}

func (r *RefundRequest) CanApprove() bool {
	return r.Status == RefundStatusPending || r.Status == RefundStatusMerchantReview || r.Status == RefundStatusPlatformReview
}

func (r *RefundRequest) CanReject() bool {
	return r.Status == RefundStatusPending || r.Status == RefundStatusMerchantReview || r.Status == RefundStatusPlatformReview
}

func (r *RefundRequest) CanCancel() bool {
	return r.Status == RefundStatusPending || r.Status == RefundStatusMerchantReview
}

func (r *RefundRequest) CanProcess() bool {
	return r.Status == RefundStatusApproved
}

func (r *RefundRequest) SubmitToMerchant() error {
	if r.Status != RefundStatusPending {
		return ErrInvalidRefundStatus
	}

	oldStatus := r.Status.String()
	r.Status = RefundStatusMerchantReview
	r.addHistory(0, "SYSTEM", "SYSTEM", "SUBMIT_TO_MERCHANT", oldStatus, r.Status.String(), "提交商家审核")

	return nil
}

func (r *RefundRequest) MerchantApprove(approverID uint64, approverName string, approvedAmount int64, comment string) error {
	if r.Status != RefundStatusMerchantReview {
		return ErrInvalidRefundStatus
	}

	if approvedAmount > r.RequestedAmount {
		return ErrRefundAmountExceeded
	}

	oldStatus := r.Status.String()
	now := time.Now()

	r.ApprovedAmount = approvedAmount
	r.Status = RefundStatusApproved
	r.CompletedAt = &now

	r.addHistory(approverID, approverName, "MERCHANT", "MERCHANT_APPROVE", oldStatus, r.Status.String(), comment)

	return nil
}

func (r *RefundRequest) MerchantReject(approverID uint64, approverName string, reason string) error {
	if r.Status != RefundStatusMerchantReview {
		return ErrInvalidRefundStatus
	}

	oldStatus := r.Status.String()

	r.Status = RefundStatusPlatformReview
	r.FailureReason = reason

	r.addHistory(approverID, approverName, "MERCHANT", "MERCHANT_REJECT", oldStatus, r.Status.String(), reason)

	return nil
}

func (r *RefundRequest) PlatformApprove(approverID uint64, approverName string, approvedAmount int64, comment string) error {
	if r.Status != RefundStatusPlatformReview {
		return ErrInvalidRefundStatus
	}

	if approvedAmount > r.RequestedAmount {
		return ErrRefundAmountExceeded
	}

	oldStatus := r.Status.String()
	now := time.Now()

	r.ApprovedAmount = approvedAmount
	r.Status = RefundStatusApproved
	r.CompletedAt = &now

	r.addHistory(approverID, approverName, "PLATFORM", "PLATFORM_APPROVE", oldStatus, r.Status.String(), comment)

	return nil
}

func (r *RefundRequest) PlatformReject(approverID uint64, approverName string, reason string) error {
	if r.Status != RefundStatusPlatformReview {
		return ErrInvalidRefundStatus
	}

	oldStatus := r.Status.String()
	now := time.Now()

	r.Status = RefundStatusRejected
	r.FailureReason = reason
	r.CompletedAt = &now

	r.addHistory(approverID, approverName, "PLATFORM", "PLATFORM_REJECT", oldStatus, r.Status.String(), reason)

	return nil
}

func (r *RefundRequest) StartProcessing() error {
	if r.Status != RefundStatusApproved {
		return ErrInvalidRefundStatus
	}

	oldStatus := r.Status.String()
	now := time.Now()

	r.Status = RefundStatusProcessing
	r.ProcessedAt = &now

	r.addHistory(0, "SYSTEM", "SYSTEM", "START_PROCESSING", oldStatus, r.Status.String(), "开始处理退款")

	return nil
}

func (r *RefundRequest) MarkSuccess(channelRefundNo string, actualAmount int64) error {
	if r.Status != RefundStatusProcessing {
		return ErrInvalidRefundStatus
	}

	oldStatus := r.Status.String()
	now := time.Now()

	r.Status = RefundStatusSuccess
	r.ChannelRefundNo = channelRefundNo
	r.ActualRefundAmount = actualAmount
	r.CompletedAt = &now

	r.addHistory(0, "SYSTEM", "SYSTEM", "REFUND_SUCCESS", oldStatus, r.Status.String(), fmt.Sprintf("退款成功，金额: %d", actualAmount))

	return nil
}

func (r *RefundRequest) MarkFailed(reason string) error {
	if r.Status != RefundStatusProcessing {
		return ErrInvalidRefundStatus
	}

	oldStatus := r.Status.String()
	now := time.Now()

	r.Status = RefundStatusFailed
	r.FailureReason = reason
	r.CompletedAt = &now

	r.addHistory(0, "SYSTEM", "SYSTEM", "REFUND_FAILED", oldStatus, r.Status.String(), reason)

	return nil
}

func (r *RefundRequest) Cancel(operatorID uint64, operatorName, reason string) error {
	if !r.CanCancel() {
		return ErrInvalidRefundStatus
	}

	oldStatus := r.Status.String()
	now := time.Now()

	r.Status = RefundStatusCancelled
	r.FailureReason = reason
	r.CompletedAt = &now

	r.addHistory(operatorID, operatorName, "USER", "CANCEL", oldStatus, r.Status.String(), reason)

	return nil
}

func (r *RefundRequest) AutoApprove() error {
	if r.Status != RefundStatusMerchantReview {
		return ErrInvalidRefundStatus
	}

	oldStatus := r.Status.String()
	now := time.Now()

	r.ApprovedAmount = r.RequestedAmount
	r.Status = RefundStatusApproved
	r.CompletedAt = &now

	r.addHistory(0, "SYSTEM", "SYSTEM", "AUTO_APPROVE", oldStatus, r.Status.String(), "超时自动同意")

	return nil
}

func (r *RefundRequest) addHistory(operatorID uint64, operatorName, operatorRole, action, oldStatus, newStatus, comment string) {
	history := &RefundHistory{
		RefundID:     r.ID,
		OperatorID:   operatorID,
		OperatorName: operatorName,
		OperatorRole: operatorRole,
		Action:       action,
		OldStatus:    oldStatus,
		NewStatus:    newStatus,
		Comment:      comment,
	}
	r.Histories = append(r.Histories, history)
}

func (r *RefundRequest) InitApprovalFlow(config *RefundConfig) *RefundApprovalFlow {
	flow := &RefundApprovalFlow{
		RefundID:    r.ID,
		CurrentStep: 1,
		TotalSteps:  2,
		Steps:       make([]*RefundApprovalStep, 0),
		AutoApprove: true,
		AutoApproveAfter: time.Duration(config.MerchantReviewHours) * time.Hour,
	}

	merchantDueAt := time.Now().Add(time.Duration(config.MerchantReviewHours) * time.Hour)
	flow.Steps = append(flow.Steps, &RefundApprovalStep{
		StepNo:       1,
		StepName:     "商家审核",
		ApproverRole: "MERCHANT",
		Status:       RefundApprovalPending,
		DueAt:        &merchantDueAt,
	})

	platformDueAt := time.Now().Add(time.Duration(config.MerchantReviewHours+config.PlatformReviewHours) * time.Hour)
	flow.Steps = append(flow.Steps, &RefundApprovalStep{
		StepNo:       2,
		StepName:     "平台审核",
		ApproverRole: "PLATFORM",
		Status:       RefundApprovalPending,
		DueAt:        &platformDueAt,
	})

	r.ApprovalFlow = flow
	return flow
}

type RefundRequestRepository interface {
	Save(ctx context.Context, request *RefundRequest) error
	FindByID(ctx context.Context, id uint64) (*RefundRequest, error)
	FindByRefundNo(ctx context.Context, refundNo string) (*RefundRequest, error)
	FindByOrderID(ctx context.Context, orderID uint64) ([]*RefundRequest, error)
	FindByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*RefundRequest, error)
	FindByMerchantID(ctx context.Context, merchantID uint64, limit, offset int) ([]*RefundRequest, error)
	FindPending(ctx context.Context, limit int) ([]*RefundRequest, error)
	FindTimeout(ctx context.Context) ([]*RefundRequest, error)
	Update(ctx context.Context, request *RefundRequest) error
}

type RefundApprovalService interface {
	SubmitRefund(ctx context.Context, request *RefundRequest) error
	MerchantApprove(ctx context.Context, refundID uint64, approverID uint64, approverName string, approvedAmount int64, comment string) error
	MerchantReject(ctx context.Context, refundID uint64, approverID uint64, approverName string, reason string) error
	PlatformApprove(ctx context.Context, refundID uint64, approverID uint64, approverName string, approvedAmount int64, comment string) error
	PlatformReject(ctx context.Context, refundID uint64, approverID uint64, approverName string, reason string) error
	ProcessTimeout(ctx context.Context) error
}
