package domain

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

var (
	ErrSplitRuleNotFound     = errors.New("split rule not found")
	ErrSplitRuleDisabled     = errors.New("split rule is disabled")
	ErrInvalidSplitRatio     = errors.New("invalid split ratio")
	ErrSplitAmountMismatch   = errors.New("split amount mismatch")
	ErrNoMatchingRule        = errors.New("no matching split rule")
)

type SplitRuleType int8

const (
	SplitRuleTypePercentage SplitRuleType = 1
	SplitRuleTypeFixed      SplitRuleType = 2
	SplitRuleTypeTiered     SplitRuleType = 3
	SplitRuleTypeDynamic    SplitRuleType = 4
)

func (t SplitRuleType) String() string {
	switch t {
	case SplitRuleTypePercentage:
		return "PERCENTAGE"
	case SplitRuleTypeFixed:
		return "FIXED"
	case SplitRuleTypeTiered:
		return "TIERED"
	case SplitRuleTypeDynamic:
		return "DYNAMIC"
	default:
		return "UNKNOWN"
	}
}

type SplitRuleStatus int8

const (
	SplitRuleStatusEnabled  SplitRuleStatus = 1
	SplitRuleStatusDisabled SplitRuleStatus = 0
)

type SplitTargetType int8

const (
	SplitTargetPlatform   SplitTargetType = 1
	SplitTargetMerchant   SplitTargetType = 2
	SplitTargetAgent      SplitTargetType = 3
	SplitTargetPromoter   SplitTargetType = 4
	SplitTargetCharity    SplitTargetType = 5
	SplitTargetReserve    SplitTargetType = 6
)

func (t SplitTargetType) String() string {
	switch t {
	case SplitTargetPlatform:
		return "PLATFORM"
	case SplitTargetMerchant:
		return "MERCHANT"
	case SplitTargetAgent:
		return "AGENT"
	case SplitTargetPromoter:
		return "PROMOTER"
	case SplitTargetCharity:
		return "CHARITY"
	case SplitTargetReserve:
		return "RESERVE"
	default:
		return "UNKNOWN"
	}
}

type SplitRule struct {
	ID           uint              `json:"id"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	RuleNo       string            `json:"rule_no"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	RuleType     SplitRuleType     `json:"rule_type"`
	Status       SplitRuleStatus   `json:"status"`
	Priority     int               `json:"priority"`
	Conditions   []*SplitCondition `json:"conditions"`
	Allocations  []*SplitAllocation `json:"allocations"`
	MinAmount    int64             `json:"min_amount"`
	MaxAmount    int64             `json:"max_amount"`
	CategoryIDs  []uint64          `json:"category_ids"`
	MerchantIDs  []uint64          `json:"merchant_ids"`
	StartTime    *time.Time        `json:"start_time"`
	EndTime      *time.Time        `json:"end_time"`
	Version      int               `json:"version"`
}

type SplitCondition struct {
	ID         uint     `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	RuleID     uint     `json:"rule_id"`
	Field      string   `json:"field"`
	Operator   string   `json:"operator"`
	Value      string   `json:"value"`
	LogicType  string   `json:"logic_type"`
}

type SplitAllocation struct {
	ID           uint            `json:"id"`
	CreatedAt    time.Time       `json:"created_at"`
	RuleID       uint            `json:"rule_id"`
	TargetType   SplitTargetType `json:"target_type"`
	TargetID     uint64          `json:"target_id"`
	TargetName   string          `json:"target_name"`
	Value        decimal.Decimal `json:"value"`
	MinAmount    int64           `json:"min_amount"`
	MaxAmount    int64           `json:"max_amount"`
	Priority     int             `json:"priority"`
	Description  string          `json:"description"`
}

type SplitRuleTier struct {
	ID          uint            `json:"id"`
	CreatedAt   time.Time       `json:"created_at"`
	RuleID      uint            `json:"rule_id"`
	MinAmount   int64           `json:"min_amount"`
	MaxAmount   int64           `json:"max_amount"`
	Allocations []*SplitAllocation `json:"allocations"`
}

type SplitResult struct {
	OrderID        uint64            `json:"order_id"`
	OrderNo        string            `json:"order_no"`
	PaymentID      uint64            `json:"payment_id"`
	PaymentNo      string            `json:"payment_no"`
	TotalAmount    int64             `json:"total_amount"`
	Currency       string            `json:"currency"`
	RuleID         uint              `json:"rule_id"`
	RuleName       string            `json:"rule_name"`
	Allocations    []*SplitResultItem `json:"allocations"`
	AllocatedAt    time.Time         `json:"allocated_at"`
	Status         SplitStatus       `json:"status"`
	FailureReason  string            `json:"failure_reason"`
}

type SplitResultItem struct {
	ID           uint            `json:"id"`
	CreatedAt    time.Time       `json:"created_at"`
	ResultID     uint            `json:"result_id"`
	TargetType   SplitTargetType `json:"target_type"`
	TargetID     uint64          `json:"target_id"`
	TargetName   string          `json:"target_name"`
	Amount       int64           `json:"amount"`
	Ratio        decimal.Decimal `json:"ratio"`
	Status       SplitStatus     `json:"status"`
	ProcessedAt  *time.Time      `json:"processed_at"`
}

type SplitStatus int8

const (
	SplitStatusPending   SplitStatus = 0
	SplitStatusProcessing SplitStatus = 1
	SplitStatusSuccess   SplitStatus = 2
	SplitStatusFailed    SplitStatus = 3
	SplitStatusCancelled SplitStatus = 4
)

func (s SplitStatus) String() string {
	switch s {
	case SplitStatusPending:
		return "PENDING"
	case SplitStatusProcessing:
		return "PROCESSING"
	case SplitStatusSuccess:
		return "SUCCESS"
	case SplitStatusFailed:
		return "FAILED"
	case SplitStatusCancelled:
		return "CANCELLED"
	default:
		return "UNKNOWN"
	}
}

type SplitContext struct {
	OrderID      uint64
	OrderNo      string
	PaymentID    uint64
	PaymentNo    string
	TotalAmount  int64
	Currency     string
	MerchantID   uint64
	CategoryID   uint64
	ProductIDs   []uint64
	UserID       uint64
	Channel      string
	PromoterID   uint64
	AgentID      uint64
	ExtraData    map[string]any
}

func NewSplitRule(ruleNo, name string, ruleType SplitRuleType) *SplitRule {
	return &SplitRule{
		RuleNo:      ruleNo,
		Name:        name,
		RuleType:    ruleType,
		Status:      SplitRuleStatusEnabled,
		Priority:    0,
		Conditions:  make([]*SplitCondition, 0),
		Allocations: make([]*SplitAllocation, 0),
		CategoryIDs: make([]uint64, 0),
		MerchantIDs: make([]uint64, 0),
		Version:     1,
	}
}

func (r *SplitRule) AddCondition(field, operator, value, logicType string) *SplitCondition {
	condition := &SplitCondition{
		RuleID:    r.ID,
		Field:     field,
		Operator:  operator,
		Value:     value,
		LogicType: logicType,
	}
	r.Conditions = append(r.Conditions, condition)
	return condition
}

func (r *SplitRule) AddAllocation(targetType SplitTargetType, targetID uint64, targetName string, value decimal.Decimal, priority int) *SplitAllocation {
	allocation := &SplitAllocation{
		RuleID:     r.ID,
		TargetType: targetType,
		TargetID:   targetID,
		TargetName: targetName,
		Value:      value,
		Priority:   priority,
	}
	r.Allocations = append(r.Allocations, allocation)
	return allocation
}

func (r *SplitRule) Matches(ctx context.Context, splitCtx *SplitContext) bool {
	if r.Status != SplitRuleStatusEnabled {
		return false
	}

	if r.StartTime != nil && time.Now().Before(*r.StartTime) {
		return false
	}

	if r.EndTime != nil && time.Now().After(*r.EndTime) {
		return false
	}

	if r.MinAmount > 0 && splitCtx.TotalAmount < r.MinAmount {
		return false
	}

	if r.MaxAmount > 0 && splitCtx.TotalAmount > r.MaxAmount {
		return false
	}

	if len(r.MerchantIDs) > 0 {
		found := false
		for _, mid := range r.MerchantIDs {
			if mid == splitCtx.MerchantID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if len(r.CategoryIDs) > 0 {
		found := false
		for _, cid := range r.CategoryIDs {
			if cid == splitCtx.CategoryID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	for _, condition := range r.Conditions {
		if !r.evaluateCondition(condition, splitCtx) {
			return false
		}
	}

	return true
}

func (r *SplitRule) evaluateCondition(condition *SplitCondition, splitCtx *SplitContext) bool {
	var fieldValue any
	switch condition.Field {
	case "merchant_id":
		fieldValue = splitCtx.MerchantID
	case "category_id":
		fieldValue = splitCtx.CategoryID
	case "user_id":
		fieldValue = splitCtx.UserID
	case "channel":
		fieldValue = splitCtx.Channel
	case "amount":
		fieldValue = splitCtx.TotalAmount
	default:
		if splitCtx.ExtraData != nil {
			fieldValue = splitCtx.ExtraData[condition.Field]
		}
	}

	return r.compareValue(fieldValue, condition.Operator, condition.Value)
}

func (r *SplitRule) compareValue(fieldValue any, operator, conditionValue string) bool {
	if fieldValue == nil {
		return false
	}

	switch operator {
	case "eq", "==":
		return fmt.Sprintf("%v", fieldValue) == conditionValue
	case "ne", "!=":
		return fmt.Sprintf("%v", fieldValue) != conditionValue
	case "gt", ">":
		return fmt.Sprintf("%v", fieldValue) > conditionValue
	case "gte", ">=":
		return fmt.Sprintf("%v", fieldValue) >= conditionValue
	case "lt", "<":
		return fmt.Sprintf("%v", fieldValue) < conditionValue
	case "lte", "<=":
		return fmt.Sprintf("%v", fieldValue) <= conditionValue
	case "in":
		return contains(conditionValue, fmt.Sprintf("%v", fieldValue))
	case "not_in":
		return !contains(conditionValue, fmt.Sprintf("%v", fieldValue))
	default:
		return false
	}
}

func (r *SplitRule) Enable() {
	r.Status = SplitRuleStatusEnabled
}

func (r *SplitRule) Disable() {
	r.Status = SplitRuleStatusDisabled
}

func (r *SplitRule) IsEnabled() bool {
	return r.Status == SplitRuleStatusEnabled
}

type SplitRuleEngine struct {
	rules      []*SplitRule
	repository SplitRuleRepository
}

func NewSplitRuleEngine(repo SplitRuleRepository) *SplitRuleEngine {
	return &SplitRuleEngine{
		rules:      make([]*SplitRule, 0),
		repository: repo,
	}
}

func (e *SplitRuleEngine) LoadRules(ctx context.Context) error {
	rules, err := e.repository.FindEnabled(ctx)
	if err != nil {
		return err
	}
	e.rules = rules
	return nil
}

func (e *SplitRuleEngine) AddRule(rule *SplitRule) {
	e.rules = append(e.rules, rule)
}

func (e *SplitRuleEngine) FindMatchingRule(ctx context.Context, splitCtx *SplitContext) (*SplitRule, error) {
	for _, rule := range e.rules {
		if rule.Matches(ctx, splitCtx) {
			return rule, nil
		}
	}

	rule, err := e.repository.FindDefaultRule(ctx)
	if err != nil {
		return nil, ErrNoMatchingRule
	}

	return rule, nil
}

func (e *SplitRuleEngine) Calculate(ctx context.Context, splitCtx *SplitContext) (*SplitResult, error) {
	rule, err := e.FindMatchingRule(ctx, splitCtx)
	if err != nil {
		return nil, err
	}

	return e.calculateWithRule(rule, splitCtx)
}

func (e *SplitRuleEngine) calculateWithRule(rule *SplitRule, splitCtx *SplitContext) (*SplitResult, error) {
	result := &SplitResult{
		OrderID:     splitCtx.OrderID,
		OrderNo:     splitCtx.OrderNo,
		PaymentID:   splitCtx.PaymentID,
		PaymentNo:   splitCtx.PaymentNo,
		TotalAmount: splitCtx.TotalAmount,
		Currency:    splitCtx.Currency,
		RuleID:      rule.ID,
		RuleName:    rule.Name,
		Allocations: make([]*SplitResultItem, 0),
		AllocatedAt: time.Now(),
		Status:      SplitStatusPending,
	}

	totalAmount := decimal.NewFromInt(splitCtx.TotalAmount)
	var allocatedAmount decimal.Decimal

	for _, allocation := range rule.Allocations {
		var amount decimal.Decimal

		switch rule.RuleType {
		case SplitRuleTypePercentage:
			amount = totalAmount.Mul(allocation.Value).Div(decimal.NewFromInt(100))
		case SplitRuleTypeFixed:
			amount = allocation.Value
		case SplitRuleTypeTiered:
			amount = e.calculateTieredAmount(splitCtx.TotalAmount, allocation)
		case SplitRuleTypeDynamic:
			amount = e.calculateDynamicAmount(splitCtx, allocation)
		default:
			amount = totalAmount.Mul(allocation.Value).Div(decimal.NewFromInt(100))
		}

		if allocation.MinAmount > 0 && amount.LessThan(decimal.NewFromInt(allocation.MinAmount)) {
			amount = decimal.NewFromInt(allocation.MinAmount)
		}
		if allocation.MaxAmount > 0 && amount.GreaterThan(decimal.NewFromInt(allocation.MaxAmount)) {
			amount = decimal.NewFromInt(allocation.MaxAmount)
		}

		allocatedAmount = allocatedAmount.Add(amount)

		result.Allocations = append(result.Allocations, &SplitResultItem{
			TargetType: allocation.TargetType,
			TargetID:   allocation.TargetID,
			TargetName: allocation.TargetName,
			Amount:     amount.IntPart(),
			Ratio:      allocation.Value,
			Status:     SplitStatusPending,
		})
	}

	remainder := totalAmount.Sub(allocatedAmount)
	if !remainder.IsZero() && len(result.Allocations) > 0 {
		result.Allocations[0].Amount += remainder.IntPart()
	}

	return result, nil
}

func (e *SplitRuleEngine) calculateTieredAmount(totalAmount int64, allocation *SplitAllocation) decimal.Decimal {
	return decimal.NewFromInt(totalAmount).Mul(allocation.Value).Div(decimal.NewFromInt(100))
}

func (e *SplitRuleEngine) calculateDynamicAmount(splitCtx *SplitContext, allocation *SplitAllocation) decimal.Decimal {
	return decimal.NewFromInt(splitCtx.TotalAmount).Mul(allocation.Value).Div(decimal.NewFromInt(100))
}

func (e *SplitRuleEngine) Validate(result *SplitResult) error {
	var totalAllocated int64
	for _, item := range result.Allocations {
		totalAllocated += item.Amount
	}

	if totalAllocated != result.TotalAmount {
		return fmt.Errorf("%w: expected %d, got %d", ErrSplitAmountMismatch, result.TotalAmount, totalAllocated)
	}

	return nil
}

func contains(list, item string) bool {
	return len(list) > 0 && (list == item || len(list) > len(item) && list[:len(item)] == item)
}

type SplitRuleRepository interface {
	Save(ctx context.Context, rule *SplitRule) error
	FindByID(ctx context.Context, id uint64) (*SplitRule, error)
	FindByRuleNo(ctx context.Context, ruleNo string) (*SplitRule, error)
	FindEnabled(ctx context.Context) ([]*SplitRule, error)
	FindDefaultRule(ctx context.Context) (*SplitRule, error)
	FindByMerchantID(ctx context.Context, merchantID uint64) ([]*SplitRule, error)
	Update(ctx context.Context, rule *SplitRule) error
	Delete(ctx context.Context, id uint64) error
}

type SplitResultRepository interface {
	Save(ctx context.Context, result *SplitResult) error
	FindByID(ctx context.Context, id uint64) (*SplitResult, error)
	FindByOrderID(ctx context.Context, orderID uint64) (*SplitResult, error)
	FindByPaymentID(ctx context.Context, paymentID uint64) (*SplitResult, error)
	FindPending(ctx context.Context, limit int) ([]*SplitResult, error)
	Update(ctx context.Context, result *SplitResult) error
}
