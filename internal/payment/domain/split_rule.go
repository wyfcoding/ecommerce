package domain

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidSplitRule     = errors.New("invalid split rule")
	ErrSplitAmountExceeded  = errors.New("split amount exceeds payment amount")
	ErrSplitAlreadyExecuted = errors.New("split already executed")
	ErrSplitNotExecuted     = errors.New("split not executed")
)

type SplitStatus string

const (
	SplitStatusPending   SplitStatus = "PENDING"
	SplitStatusProcessing SplitStatus = "PROCESSING"
	SplitStatusSuccess   SplitStatus = "SUCCESS"
	SplitStatusFailed    SplitStatus = "FAILED"
	SplitStatusCancelled SplitStatus = "CANCELLED"
)

type RecipientType string

const (
	RecipientTypeMerchant RecipientType = "MERCHANT"
	RecipientTypePlatform RecipientType = "PLATFORM"
	RecipientTypeAgent    RecipientType = "AGENT"
	RecipientTypeSupplier RecipientType = "SUPPLIER"
	RecipientTypePromoter RecipientType = "PROMOTER"
)

type SplitRuleType string

const (
	SplitRuleTypeFixed    SplitRuleType = "FIXED"
	SplitRuleTypePercent  SplitRuleType = "PERCENT"
	SplitRuleTypeTiered   SplitRuleType = "TIERED"
	SplitRuleTypeDynamic  SplitRuleType = "DYNAMIC"
)

type SplitRule struct {
	ID            uint64         `json:"id"`
	RuleNo        string         `json:"rule_no"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	MerchantID    uint64         `json:"merchant_id"`
	RuleType      SplitRuleType  `json:"rule_type"`
	Priority      int            `json:"priority"`
	Enabled       bool           `json:"enabled"`
	Conditions    []SplitCondition `json:"conditions"`
	Actions       []SplitAction  `json:"actions"`
	EffectiveFrom *time.Time     `json:"effective_from,omitempty"`
	EffectiveTo   *time.Time     `json:"effective_to,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type SplitCondition struct {
	ID         uint64                 `json:"id"`
	RuleID     uint64                 `json:"rule_id"`
	Field      string                 `json:"field"`
	Operator   string                 `json:"operator"`
	Value      string                 `json:"value"`
	Extra      map[string]interface{} `json:"extra,omitempty"`
}

type SplitAction struct {
	ID           uint64       `json:"id"`
	RuleID       uint64       `json:"rule_id"`
	RecipientID   uint64      `json:"recipient_id"`
	RecipientType RecipientType `json:"recipient_type"`
	SplitType     SplitRuleType `json:"split_type"`
	FixedAmount   int64       `json:"fixed_amount,omitempty"`
	PercentRate   float64     `json:"percent_rate,omitempty"`
	MinAmount     int64       `json:"min_amount,omitempty"`
	MaxAmount     int64       `json:"max_amount,omitempty"`
	Description   string      `json:"description"`
}

type SplitRuleTier struct {
	ID            uint64  `json:"id"`
	RuleActionID  uint64  `json:"rule_action_id"`
	MinAmount     int64   `json:"min_amount"`
	MaxAmount     int64   `json:"max_amount"`
	PercentRate   float64 `json:"percent_rate"`
	FixedAmount   int64   `json:"fixed_amount"`
}

type PaymentSplitDetail struct {
	ID             uint64        `json:"id"`
	SplitNo        string        `json:"split_no"`
	PaymentID      uint64        `json:"payment_id"`
	PaymentNo      string        `json:"payment_no"`
	OrderID        uint64        `json:"order_id"`
	OrderNo        string        `json:"order_no"`
	TotalAmount    int64         `json:"total_amount"`
	SplitAmount    int64         `json:"split_amount"`
	Status         SplitStatus   `json:"status"`
	RuleID         uint64        `json:"rule_id,omitempty"`
	RuleNo         string        `json:"rule_no,omitempty"`
	Recipients     []*SplitRecipient `json:"recipients"`
	CreatedAt      time.Time     `json:"created_at"`
	ExecutedAt     *time.Time    `json:"executed_at,omitempty"`
	CompletedAt    *time.Time    `json:"completed_at,omitempty"`
	FailureReason  string        `json:"failure_reason,omitempty"`
}

type SplitRecipient struct {
	ID             uint64        `json:"id"`
	SplitID        uint64        `json:"split_id"`
	RecipientID    uint64        `json:"recipient_id"`
	RecipientType  RecipientType `json:"recipient_type"`
	RecipientName  string        `json:"recipient_name"`
	RecipientAccount string      `json:"recipient_account"`
	Amount         int64         `json:"amount"`
	Fee            int64         `json:"fee"`
	ActualAmount   int64         `json:"actual_amount"`
	Status         SplitStatus   `json:"status"`
	TransactionNo  string        `json:"transaction_no,omitempty"`
	FailureReason  string        `json:"failure_reason,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
	ProcessedAt    *time.Time    `json:"processed_at,omitempty"`
}

type SplitRuleEngine struct {
	rules []*SplitRule
}

func NewSplitRuleEngine() *SplitRuleEngine {
	return &SplitRuleEngine{
		rules: make([]*SplitRule, 0),
	}
}

func (e *SplitRuleEngine) AddRule(rule *SplitRule) {
	e.rules = append(e.rules, rule)
}

func (e *SplitRuleEngine) RemoveRule(ruleID uint64) {
	for i, rule := range e.rules {
		if rule.ID == ruleID {
			e.rules = append(e.rules[:i], e.rules[i+1:]...)
			break
		}
	}
}

func (e *SplitRuleEngine) FindApplicableRules(ctx *SplitContext) []*SplitRule {
	applicable := make([]*SplitRule, 0)
	now := time.Now()
	
	for _, rule := range e.rules {
		if !rule.Enabled {
			continue
		}
		
		if rule.EffectiveFrom != nil && now.Before(*rule.EffectiveFrom) {
			continue
		}
		if rule.EffectiveTo != nil && now.After(*rule.EffectiveTo) {
			continue
		}
		
		if e.matchConditions(rule, ctx) {
			applicable = append(applicable, rule)
		}
	}
	
	return applicable
}

func (e *SplitRuleEngine) matchConditions(rule *SplitRule, ctx *SplitContext) bool {
	for _, cond := range rule.Conditions {
		if !e.matchCondition(cond, ctx) {
			return false
		}
	}
	return true
}

func (e *SplitRuleEngine) matchCondition(cond SplitCondition, ctx *SplitContext) bool {
	var value interface{}
	switch cond.Field {
	case "merchant_id":
		value = ctx.MerchantID
	case "amount":
		value = ctx.Amount
	case "product_category":
		value = ctx.ProductCategory
	case "payment_method":
		value = ctx.PaymentMethod
	case "user_level":
		value = ctx.UserLevel
	case "order_type":
		value = ctx.OrderType
	default:
		if v, ok := ctx.Extra[cond.Field]; ok {
			value = v
		}
	}
	
	return e.compare(value, cond.Operator, cond.Value)
}

func (e *SplitRuleEngine) compare(value interface{}, operator, expected string) bool {
	strValue := fmt.Sprintf("%v", value)
	
	switch operator {
	case "=":
		return strValue == expected
	case "!=":
		return strValue != expected
	case ">":
		return strValue > expected
	case ">=":
		return strValue >= expected
	case "<":
		return strValue < expected
	case "<=":
		return strValue <= expected
	case "in":
		return contains(expected, strValue)
	case "not_in":
		return !contains(expected, strValue)
	default:
		return false
	}
}

func (e *SplitRuleEngine) CalculateSplit(rule *SplitRule, ctx *SplitContext) ([]*SplitRecipient, error) {
	recipients := make([]*SplitRecipient, 0)
	var totalSplit int64
	
	for _, action := range rule.Actions {
		amount, err := e.calculateActionAmount(action, ctx.Amount)
		if err != nil {
			return nil, err
		}
		
		if amount > 0 {
			recipient := &SplitRecipient{
				RecipientID:   action.RecipientID,
				RecipientType: action.RecipientType,
				Amount:        amount,
				Status:        SplitStatusPending,
				CreatedAt:     time.Now(),
			}
			recipients = append(recipients, recipient)
			totalSplit += amount
		}
	}
	
	if totalSplit > ctx.Amount {
		return nil, ErrSplitAmountExceeded
	}
	
	return recipients, nil
}

func (e *SplitRuleEngine) calculateActionAmount(action SplitAction, totalAmount int64) (int64, error) {
	var amount int64
	
	switch action.SplitType {
	case SplitRuleTypeFixed:
		amount = action.FixedAmount
		
	case SplitRuleTypePercent:
		amount = int64(float64(totalAmount) * action.PercentRate / 100)
		
	case SplitRuleTypeTiered:
		amount = e.calculateTieredAmount(action.ID, totalAmount)
		
	case SplitRuleTypeDynamic:
		amount = e.calculateDynamicAmount(action, totalAmount)
	}
	
	if action.MinAmount > 0 && amount < action.MinAmount {
		amount = action.MinAmount
	}
	if action.MaxAmount > 0 && amount > action.MaxAmount {
		amount = action.MaxAmount
	}
	
	return amount, nil
}

func (e *SplitRuleEngine) calculateTieredAmount(actionID uint64, totalAmount int64) int64 {
	return int64(float64(totalAmount) * 0.03)
}

func (e *SplitRuleEngine) calculateDynamicAmount(action SplitAction, totalAmount int64) int64 {
	return int64(float64(totalAmount) * 0.02)
}

type SplitContext struct {
	PaymentID       uint64                 `json:"payment_id"`
	PaymentNo       string                 `json:"payment_no"`
	OrderID         uint64                 `json:"order_id"`
	OrderNo         string                 `json:"order_no"`
	MerchantID      uint64                 `json:"merchant_id"`
	Amount          int64                  `json:"amount"`
	PaymentMethod   string                 `json:"payment_method"`
	ProductCategory string                 `json:"product_category"`
	UserLevel       int                    `json:"user_level"`
	OrderType       string                 `json:"order_type"`
	Extra           map[string]interface{} `json:"extra,omitempty"`
}

func NewSplitContext(paymentID, orderID, merchantID uint64, amount int64) *SplitContext {
	return &SplitContext{
		PaymentID:  paymentID,
		OrderID:    orderID,
		MerchantID: merchantID,
		Amount:     amount,
		Extra:      make(map[string]interface{}),
	}
}

func contains(list, item string) bool {
	return len(list) > 0 && (list == item || len(list) > len(item) && 
		(list[:len(item)+1] == item+"," || list[len(list)-len(item)-1:] == ","+item ||
			containsMiddle(list, item)))
}

func containsMiddle(list, item string) bool {
	for i := 0; i < len(list)-len(item)-1; i++ {
		if list[i:i+len(item)] == item && list[i+len(item)] == ',' {
			return true
		}
	}
	return false
}

type SplitRuleRepository interface {
	Create(rule *SplitRule) error
	Update(rule *SplitRule) error
	Delete(ruleID uint64) error
	FindByID(ruleID uint64) (*SplitRule, error)
	FindByRuleNo(ruleNo string) (*SplitRule, error)
	ListByMerchantID(merchantID uint64, enabled *bool) ([]*SplitRule, error)
	ListAll(enabled *bool) ([]*SplitRule, error)
}

type PaymentSplitRepository interface {
	Create(split *PaymentSplitDetail) error
	Update(split *PaymentSplitDetail) error
	FindByID(splitID uint64) (*PaymentSplitDetail, error)
	FindBySplitNo(splitNo string) (*PaymentSplitDetail, error)
	FindByPaymentID(paymentID uint64) (*PaymentSplitDetail, error)
	FindByOrderID(orderID uint64) ([]*PaymentSplitDetail, error)
	ListByStatus(status SplitStatus, limit int) ([]*PaymentSplitDetail, error)
	ListPendingSplits(beforeTime time.Time, limit int) ([]*PaymentSplitDetail, error)
}

type SplitRecipientRepository interface {
	Create(recipient *SplitRecipient) error
	Update(recipient *SplitRecipient) error
	FindByID(id uint64) (*SplitRecipient, error)
	FindBySplitID(splitID uint64) ([]*SplitRecipient, error)
	ListByRecipientID(recipientID uint64, recipientType RecipientType, startTime, endTime *time.Time) ([]*SplitRecipient, error)
}
