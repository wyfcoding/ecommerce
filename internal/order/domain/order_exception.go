package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/order/v1"
)

var (
	ErrOrderExceptionNotFound     = errors.New("order exception not found")
	ErrExceptionAlreadyResolved   = errors.New("exception already resolved")
	ErrExceptionCannotBeResolved  = errors.New("exception cannot be resolved in current status")
	ErrInvalidExceptionType       = errors.New("invalid exception type")
	ErrResolutionNotApplicable    = errors.New("resolution not applicable for this exception type")
)

type ExceptionType int8

const (
	ExceptionTypeStockInsufficient  ExceptionType = 1
	ExceptionTypeAddressInvalid     ExceptionType = 2
	ExceptionTypePaymentFailed      ExceptionType = 3
	ExceptionTypePriceMismatch      ExceptionType = 4
	ExceptionTypePromotionExpired   ExceptionType = 5
	ExceptionTypeCouponInvalid      ExceptionType = 6
	ExceptionTypeUserRiskBlocked    ExceptionType = 7
	ExceptionTypeMerchantSuspended  ExceptionType = 8
	ExceptionTypeProductOffline     ExceptionType = 9
	ExceptionTypeLogisticsException ExceptionType = 10
	ExceptionTypeSystemError        ExceptionType = 11
	ExceptionTypeOther              ExceptionType = 99
)

func (t ExceptionType) String() string {
	switch t {
	case ExceptionTypeStockInsufficient:
		return "STOCK_INSUFFICIENT"
	case ExceptionTypeAddressInvalid:
		return "ADDRESS_INVALID"
	case ExceptionTypePaymentFailed:
		return "PAYMENT_FAILED"
	case ExceptionTypePriceMismatch:
		return "PRICE_MISMATCH"
	case ExceptionTypePromotionExpired:
		return "PROMOTION_EXPIRED"
	case ExceptionTypeCouponInvalid:
		return "COUPON_INVALID"
	case ExceptionTypeUserRiskBlocked:
		return "USER_RISK_BLOCKED"
	case ExceptionTypeMerchantSuspended:
		return "MERCHANT_SUSPENDED"
	case ExceptionTypeProductOffline:
		return "PRODUCT_OFFLINE"
	case ExceptionTypeLogisticsException:
		return "LOGISTICS_EXCEPTION"
	case ExceptionTypeSystemError:
		return "SYSTEM_ERROR"
	default:
		return "OTHER"
	}
}

type ExceptionSeverity int8

const (
	ExceptionSeverityLow      ExceptionSeverity = 1
	ExceptionSeverityMedium   ExceptionSeverity = 2
	ExceptionSeverityHigh     ExceptionSeverity = 3
	ExceptionSeverityCritical ExceptionSeverity = 4
)

func (s ExceptionSeverity) String() string {
	switch s {
	case ExceptionSeverityLow:
		return "LOW"
	case ExceptionSeverityMedium:
		return "MEDIUM"
	case ExceptionSeverityHigh:
		return "HIGH"
	case ExceptionSeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

type ExceptionStatus int8

const (
	ExceptionStatusPending    ExceptionStatus = 1
	ExceptionStatusProcessing ExceptionStatus = 2
	ExceptionStatusResolved   ExceptionStatus = 3
	ExceptionStatusEscalated  ExceptionStatus = 4
	ExceptionStatusClosed     ExceptionStatus = 5
)

func (s ExceptionStatus) String() string {
	switch s {
	case ExceptionStatusPending:
		return "PENDING"
	case ExceptionStatusProcessing:
		return "PROCESSING"
	case ExceptionStatusResolved:
		return "RESOLVED"
	case ExceptionStatusEscalated:
		return "ESCALATED"
	case ExceptionStatusClosed:
		return "CLOSED"
	default:
		return "UNKNOWN"
	}
}

type ResolutionType int8

const (
	ResolutionTypeAutoCancel    ResolutionType = 1
	ResolutionTypeManualCancel  ResolutionType = 2
	ResolutionTypeRetry         ResolutionType = 3
	ResolutionTypeContactUser   ResolutionType = 4
	ResolutionTypeContactMerchant ResolutionType = 5
	ResolutionTypePartialShip   ResolutionType = 6
	ResolutionTypeReplaceItem   ResolutionType = 7
	ResolutionTypeAdjustPrice   ResolutionType = 8
	ResolutionTypeManualFix     ResolutionType = 9
	ResolutionTypeIgnore        ResolutionType = 10
)

func (r ResolutionType) String() string {
	switch r {
	case ResolutionTypeAutoCancel:
		return "AUTO_CANCEL"
	case ResolutionTypeManualCancel:
		return "MANUAL_CANCEL"
	case ResolutionTypeRetry:
		return "RETRY"
	case ResolutionTypeContactUser:
		return "CONTACT_USER"
	case ResolutionTypeContactMerchant:
		return "CONTACT_MERCHANT"
	case ResolutionTypePartialShip:
		return "PARTIAL_SHIP"
	case ResolutionTypeReplaceItem:
		return "REPLACE_ITEM"
	case ResolutionTypeAdjustPrice:
		return "ADJUST_PRICE"
	case ResolutionTypeManualFix:
		return "MANUAL_FIX"
	case ResolutionTypeIgnore:
		return "IGNORE"
	default:
		return "UNKNOWN"
	}
}

type OrderException struct {
	ID               uint              `json:"id"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
	ExceptionNo      string            `json:"exception_no"`
	OrderID          uint64            `json:"order_id"`
	OrderNo          string            `json:"order_no"`
	UserID           uint64            `json:"user_id"`
	ExceptionType    ExceptionType     `json:"exception_type"`
	Severity         ExceptionSeverity `json:"severity"`
	Status           ExceptionStatus   `json:"status"`
	Title            string            `json:"title"`
	Description      string            `json:"description"`
	Source           string            `json:"source"`
	SourceData       map[string]any    `json:"source_data"`
	AffectedItems    []uint64          `json:"affected_items"`
	HandlerID        uint64            `json:"handler_id"`
	HandlerName      string            `json:"handler_name"`
	ResolutionType   ResolutionType    `json:"resolution_type"`
	ResolutionDetail string            `json:"resolution_detail"`
	ResolvedAt       *time.Time        `json:"resolved_at"`
	EscalatedAt      *time.Time        `json:"escalated_at"`
	EscalatedTo      uint64            `json:"escalated_to"`
	ClosedAt         *time.Time        `json:"closed_at"`
	ProcessingAt     *time.Time        `json:"processing_at"`
	DueAt            *time.Time        `json:"due_at"`
	RemindedAt       *time.Time        `json:"reminded_at"`
	Attempts         int               `json:"attempts"`
	MaxAttempts      int               `json:"max_attempts"`
	Histories        []*ExceptionHistory `json:"histories"`
}

type ExceptionHistory struct {
	ID           uint      `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	ExceptionID  uint      `json:"exception_id"`
	OperatorID   uint64    `json:"operator_id"`
	OperatorName string    `json:"operator_name"`
	Action       string    `json:"action"`
	OldStatus    string    `json:"old_status"`
	NewStatus    string    `json:"new_status"`
	Comment      string    `json:"comment"`
}

type ExceptionRule struct {
	ID                uint              `json:"id"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	ExceptionType     ExceptionType     `json:"exception_type"`
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	AutoResolution    ResolutionType    `json:"auto_resolution"`
	AutoResolve       bool              `json:"auto_resolve"`
	Severity          ExceptionSeverity `json:"severity"`
	TimeoutMinutes    int               `json:"timeout_minutes"`
	EscalationEnabled bool              `json:"escalation_enabled"`
	EscalationMinutes int               `json:"escalation_minutes"`
	NotifyRoles       []string          `json:"notify_roles"`
	Enabled           bool              `json:"enabled"`
}

func NewOrderException(exceptionNo string, orderID uint64, orderNo string, userID uint64, exceptionType ExceptionType, severity ExceptionSeverity, title, description, source string) *OrderException {
	return &OrderException{
		ExceptionNo:   exceptionNo,
		OrderID:       orderID,
		OrderNo:       orderNo,
		UserID:        userID,
		ExceptionType: exceptionType,
		Severity:      severity,
		Status:        ExceptionStatusPending,
		Title:         title,
		Description:   description,
		Source:        source,
		SourceData:    make(map[string]any),
		AffectedItems: make([]uint64, 0),
		Attempts:      0,
		MaxAttempts:   3,
		Histories:     make([]*ExceptionHistory, 0),
	}
}

func (e *OrderException) SetSourceData(data map[string]any) {
	e.SourceData = data
}

func (e *OrderException) SetAffectedItems(itemIDs []uint64) {
	e.AffectedItems = itemIDs
}

func (e *OrderException) SetDueAt(dueAt time.Time) {
	e.DueAt = &dueAt
}

func (e *OrderException) Assign(handlerID uint64, handlerName string) error {
	if e.Status == ExceptionStatusResolved || e.Status == ExceptionStatusClosed {
		return ErrExceptionAlreadyResolved
	}

	oldStatus := e.Status.String()
	now := time.Now()

	e.HandlerID = handlerID
	e.HandlerName = handlerName
	e.Status = ExceptionStatusProcessing
	e.ProcessingAt = &now

	e.addHistory(handlerID, handlerName, "ASSIGN", oldStatus, e.Status.String(), fmt.Sprintf("Assigned to %s", handlerName))

	return nil
}

func (e *OrderException) Resolve(resolutionType ResolutionType, detail string, operatorID uint64, operatorName string) error {
	if e.Status == ExceptionStatusResolved || e.Status == ExceptionStatusClosed {
		return ErrExceptionAlreadyResolved
	}

	oldStatus := e.Status.String()
	now := time.Now()

	e.ResolutionType = resolutionType
	e.ResolutionDetail = detail
	e.Status = ExceptionStatusResolved
	e.ResolvedAt = &now

	e.addHistory(operatorID, operatorName, "RESOLVE", oldStatus, e.Status.String(), fmt.Sprintf("Resolution: %s, Detail: %s", resolutionType.String(), detail))

	return nil
}

func (e *OrderException) Escalate(toUserID uint64, reason string, operatorID uint64, operatorName string) error {
	if e.Status == ExceptionStatusResolved || e.Status == ExceptionStatusClosed {
		return ErrExceptionAlreadyResolved
	}

	oldStatus := e.Status.String()
	now := time.Now()

	e.Status = ExceptionStatusEscalated
	e.EscalatedAt = &now
	e.EscalatedTo = toUserID

	e.addHistory(operatorID, operatorName, "ESCALATE", oldStatus, e.Status.String(), fmt.Sprintf("Escalated to user %d, Reason: %s", toUserID, reason))

	return nil
}

func (e *OrderException) Close(reason string, operatorID uint64, operatorName string) error {
	if e.Status == ExceptionStatusClosed {
		return ErrExceptionAlreadyResolved
	}

	oldStatus := e.Status.String()
	now := time.Time{}

	e.Status = ExceptionStatusClosed
	e.ClosedAt = &now

	e.addHistory(operatorID, operatorName, "CLOSE", oldStatus, e.Status.String(), reason)

	return nil
}

func (e *OrderException) Retry(operatorID uint64, operatorName string) error {
	if e.Status == ExceptionStatusResolved || e.Status == ExceptionStatusClosed {
		return ErrExceptionAlreadyResolved
	}

	if e.Attempts >= e.MaxAttempts {
		return errors.New("max retry attempts exceeded")
	}

	e.Attempts++
	oldStatus := e.Status.String()

	e.addHistory(operatorID, operatorName, "RETRY", oldStatus, e.Status.String(), fmt.Sprintf("Retry attempt %d/%d", e.Attempts, e.MaxAttempts))

	return nil
}

func (e *OrderException) Remind() error {
	if e.Status == ExceptionStatusResolved || e.Status == ExceptionStatusClosed {
		return ErrExceptionAlreadyResolved
	}

	now := time.Now()
	e.RemindedAt = &now

	return nil
}

func (e *OrderException) IsTimeout() bool {
	if e.Status == ExceptionStatusResolved || e.Status == ExceptionStatusClosed {
		return false
	}

	if e.DueAt == nil {
		return false
	}

	return time.Now().After(*e.DueAt)
}

func (e *OrderException) IsCritical() bool {
	return e.Severity == ExceptionSeverityCritical
}

func (e *OrderException) CanAutoResolve() bool {
	return e.Status == ExceptionStatusPending && e.Attempts < e.MaxAttempts
}

func (e *OrderException) addHistory(operatorID uint64, operatorName, action, oldStatus, newStatus, comment string) {
	history := &ExceptionHistory{
		ExceptionID:  e.ID,
		OperatorID:   operatorID,
		OperatorName: operatorName,
		Action:       action,
		OldStatus:    oldStatus,
		NewStatus:    newStatus,
		Comment:      comment,
	}
	e.Histories = append(e.Histories, history)
}

type ExceptionHandler interface {
	CanHandle(exceptionType ExceptionType) bool
	Handle(ctx context.Context, exception *OrderException) (*ResolutionResult, error)
	GetPriority() int
}

type ResolutionResult struct {
	Success      bool
	Resolution   ResolutionType
	Detail       string
	ShouldRetry  bool
	RetryAfter   time.Duration
	ShouldEscalate bool
}

type ExceptionHandlerChain struct {
	handlers []ExceptionHandler
}

func NewExceptionHandlerChain() *ExceptionHandlerChain {
	return &ExceptionHandlerChain{
		handlers: make([]ExceptionHandler, 0),
	}
}

func (c *ExceptionHandlerChain) AddHandler(handler ExceptionHandler) {
	c.handlers = append(c.handlers, handler)
}

func (c *ExceptionHandlerChain) Handle(ctx context.Context, exception *OrderException) (*ResolutionResult, error) {
	for _, handler := range c.handlers {
		if handler.CanHandle(exception.ExceptionType) {
			return handler.Handle(ctx, exception)
		}
	}
	return &ResolutionResult{
		Success: false,
		Detail:  "no suitable handler found",
	}, nil
}

type StockInsufficientHandler struct{}

func (h *StockInsufficientHandler) CanHandle(exceptionType ExceptionType) bool {
	return exceptionType == ExceptionTypeStockInsufficient
}

func (h *StockInsufficientHandler) Handle(ctx context.Context, exception *OrderException) (*ResolutionResult, error) {
	if len(exception.AffectedItems) > 0 {
		return &ResolutionResult{
			Success:      true,
			Resolution:   ResolutionTypePartialShip,
			Detail:       "partial shipment for available items",
			ShouldRetry:  false,
		}, nil
	}

	return &ResolutionResult{
		Success:      true,
		Resolution:   ResolutionTypeAutoCancel,
		Detail:       "auto cancel due to stock unavailable",
		ShouldRetry:  false,
	}, nil
}

func (h *StockInsufficientHandler) GetPriority() int { return 100 }

type PaymentFailedHandler struct{}

func (h *PaymentFailedHandler) CanHandle(exceptionType ExceptionType) bool {
	return exceptionType == ExceptionTypePaymentFailed
}

func (h *PaymentFailedHandler) Handle(ctx context.Context, exception *OrderException) (*ResolutionResult, error) {
	if exception.Attempts < exception.MaxAttempts {
		return &ResolutionResult{
			Success:      true,
			Resolution:   ResolutionTypeRetry,
			Detail:       "retry payment",
			ShouldRetry:  true,
			RetryAfter:   time.Minute * time.Duration(exception.Attempts+1),
		}, nil
	}

	return &ResolutionResult{
		Success:      true,
		Resolution:   ResolutionTypeContactUser,
		Detail:       "contact user for alternative payment",
		ShouldRetry:  false,
	}, nil
}

func (h *PaymentFailedHandler) GetPriority() int { return 100 }

type AddressInvalidHandler struct{}

func (h *AddressInvalidHandler) CanHandle(exceptionType ExceptionType) bool {
	return exceptionType == ExceptionTypeAddressInvalid
}

func (h *AddressInvalidHandler) Handle(ctx context.Context, exception *OrderException) (*ResolutionResult, error) {
	return &ResolutionResult{
		Success:      true,
		Resolution:   ResolutionTypeContactUser,
		Detail:       "contact user to update address",
		ShouldRetry:  false,
	}, nil
}

func (h *AddressInvalidHandler) GetPriority() int { return 90 }

type UserRiskBlockedHandler struct{}

func (h *UserRiskBlockedHandler) CanHandle(exceptionType ExceptionType) bool {
	return exceptionType == ExceptionTypeUserRiskBlocked
}

func (h *UserRiskBlockedHandler) Handle(ctx context.Context, exception *OrderException) (*ResolutionResult, error) {
	return &ResolutionResult{
		Success:        true,
		Resolution:     ResolutionTypeAutoCancel,
		Detail:         "auto cancel due to risk block",
		ShouldRetry:    false,
		ShouldEscalate: true,
	}, nil
}

func (h *UserRiskBlockedHandler) GetPriority() int { return 200 }

func (o *Order) CreateException(exceptionType ExceptionType, severity ExceptionSeverity, title, description, source string) *OrderException {
	exceptionNo := fmt.Sprintf("EX%s%04d", time.Now().Format("20060102150405"), o.ID%10000)
	return NewOrderException(exceptionNo, uint64(o.ID), o.OrderNo, o.UserID, exceptionType, severity, title, description, source)
}

func (o *Order) HasCriticalException() bool {
	return false
}

func (o *Order) CanProceedWithException() bool {
	return o.Status == pb.OrderStatus_PENDING_PAYMENT || o.Status == pb.OrderStatus_ALLOCATING
}

type OrderExceptionRepository interface {
	Save(ctx context.Context, exception *OrderException) error
	FindByID(ctx context.Context, id uint64) (*OrderException, error)
	FindByExceptionNo(ctx context.Context, exceptionNo string) (*OrderException, error)
	FindByOrderID(ctx context.Context, orderID uint64) ([]*OrderException, error)
	FindPending(ctx context.Context, limit, offset int) ([]*OrderException, error)
	FindByHandler(ctx context.Context, handlerID uint64, limit, offset int) ([]*OrderException, error)
	FindTimeout(ctx context.Context) ([]*OrderException, error)
	FindByType(ctx context.Context, exceptionType ExceptionType, limit, offset int) ([]*OrderException, error)
	Update(ctx context.Context, exception *OrderException) error
}

type ExceptionRuleRepository interface {
	FindByType(ctx context.Context, exceptionType ExceptionType) (*ExceptionRule, error)
	FindAll(ctx context.Context) ([]*ExceptionRule, error)
	Save(ctx context.Context, rule *ExceptionRule) error
}

type ExceptionNotificationService interface {
	NotifyExceptionCreated(ctx context.Context, exception *OrderException) error
	NotifyExceptionResolved(ctx context.Context, exception *OrderException) error
	NotifyExceptionEscalated(ctx context.Context, exception *OrderException) error
	NotifyExceptionTimeout(ctx context.Context, exception *OrderException) error
}
