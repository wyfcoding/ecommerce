package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/wyfcoding/ecommerce/internal/frauddetection/domain"
)

type RuleEngineImpl struct {
	velocityTracker VelocityTracker
}

type VelocityTracker interface {
	GetTransactionCount(ctx context.Context, userID uint64, duration time.Duration) (int, error)
	GetAmountSum(ctx context.Context, userID uint64, duration time.Duration) (int64, error)
}

func NewRuleEngine(velocityTracker VelocityTracker) domain.RuleEngine {
	return &RuleEngineImpl{
		velocityTracker: velocityTracker,
	}
}

func (e *RuleEngineImpl) Evaluate(ctx context.Context, rule *domain.RiskRule, txCtx *domain.TransactionContext) (bool, int, error) {
	switch rule.Type {
	case domain.RuleTypeVelocity:
		return e.evaluateVelocityRule(ctx, rule, txCtx)
	case domain.RuleTypeAmount:
		return e.evaluateAmountRule(rule, txCtx)
	case domain.RuleTypeGeography:
		return e.evaluateGeographyRule(rule, txCtx)
	case domain.RuleTypeBehavior:
		return e.evaluateBehaviorRule(rule, txCtx)
	case domain.RuleTypeDevice:
		return e.evaluateDeviceRule(rule, txCtx)
	case domain.RuleTypeCustom:
		return e.evaluateCustomRule(rule, txCtx)
	default:
		return false, 0, fmt.Errorf("unknown rule type: %s", rule.Type)
	}
}

type VelocityCondition struct {
	Type     string `json:"type"`     // "count" or "amount"
	Duration int    `json:"duration"` // in minutes
	Threshold int   `json:"threshold"`
}

func (e *RuleEngineImpl) evaluateVelocityRule(ctx context.Context, rule *domain.RiskRule, txCtx *domain.TransactionContext) (bool, int, error) {
	var cond VelocityCondition
	if err := json.Unmarshal([]byte(rule.Condition), &cond); err != nil {
		return false, 0, fmt.Errorf("invalid velocity condition: %w", err)
	}

	duration := time.Duration(cond.Duration) * time.Minute

	switch cond.Type {
	case "count":
		count, err := e.velocityTracker.GetTransactionCount(ctx, txCtx.UserID, duration)
		if err != nil {
			return false, 0, err
		}
		if count >= cond.Threshold {
			score := rule.RiskWeight
			if count > cond.Threshold*2 {
				score = rule.RiskWeight * 2
			}
			return true, score, nil
		}
	case "amount":
		sum, err := e.velocityTracker.GetAmountSum(ctx, txCtx.UserID, duration)
		if err != nil {
			return false, 0, err
		}
		if sum >= int64(cond.Threshold) {
			score := rule.RiskWeight
			if sum >= int64(cond.Threshold)*2 {
				score = rule.RiskWeight * 2
			}
			return true, score, nil
		}
	}

	return false, 0, nil
}

type AmountCondition struct {
	MinAmount int64 `json:"min_amount"`
	MaxAmount int64 `json:"max_amount"`
}

func (e *RuleEngineImpl) evaluateAmountRule(rule *domain.RiskRule, txCtx *domain.TransactionContext) (bool, int, error) {
	var cond AmountCondition
	if err := json.Unmarshal([]byte(rule.Condition), &cond); err != nil {
		return false, 0, fmt.Errorf("invalid amount condition: %w", err)
	}

	if cond.MinAmount > 0 && txCtx.Amount < cond.MinAmount {
		return false, 0, nil
	}
	if cond.MaxAmount > 0 && txCtx.Amount > cond.MaxAmount {
		return true, rule.RiskWeight, nil
	}

	return false, 0, nil
}

type GeographyCondition struct {
	BlockedCountries []string `json:"blocked_countries"`
	AllowedCountries []string `json:"allowed_countries"`
}

func (e *RuleEngineImpl) evaluateGeographyRule(rule *domain.RiskRule, txCtx *domain.TransactionContext) (bool, int, error) {
	var cond GeographyCondition
	if err := json.Unmarshal([]byte(rule.Condition), &cond); err != nil {
		return false, 0, fmt.Errorf("invalid geography condition: %w", err)
	}

	if txCtx.Country == "" {
		return false, 0, nil
	}

	for _, blocked := range cond.BlockedCountries {
		if txCtx.Country == blocked {
			return true, rule.RiskWeight, nil
		}
	}

	if len(cond.AllowedCountries) > 0 {
		for _, allowed := range cond.AllowedCountries {
			if txCtx.Country == allowed {
				return false, 0, nil
			}
		}
		return true, rule.RiskWeight, nil
	}

	return false, 0, nil
}

type BehaviorCondition struct {
	Pattern string `json:"pattern"`
}

func (e *RuleEngineImpl) evaluateBehaviorRule(rule *domain.RiskRule, txCtx *domain.TransactionContext) (bool, int, error) {
	var cond BehaviorCondition
	if err := json.Unmarshal([]byte(rule.Condition), &cond); err != nil {
		return false, 0, fmt.Errorf("invalid behavior condition: %w", err)
	}

	switch cond.Pattern {
	case "new_user_high_amount":
		return e.checkNewUserHighAmount(txCtx, rule.RiskWeight)
	case "suspicious_email":
		return e.checkSuspiciousEmail(txCtx, rule.RiskWeight)
	case "mismatched_billing_shipping":
		return false, 0, nil
	default:
		return false, 0, nil
	}
}

func (e *RuleEngineImpl) checkNewUserHighAmount(txCtx *domain.TransactionContext, weight int) (bool, int, error) {
	if txCtx.Amount > 100000 && txCtx.Metadata != nil {
		if createdAt, ok := txCtx.Metadata["user_created_at"]; ok {
			t, err := time.Parse(time.RFC3339, createdAt)
			if err == nil && time.Since(t) < 24*time.Hour {
				return true, weight, nil
			}
		}
	}
	return false, 0, nil
}

func (e *RuleEngineImpl) checkSuspiciousEmail(txCtx *domain.TransactionContext, weight int) (bool, int, error) {
	if txCtx.Email == "" {
		return false, 0, nil
	}

	suspiciousPatterns := []string{
		`^[a-z]{10,}@`,
		`^\d{8,}@`,
		`@temp`,
		`@fake`,
		`@test`,
	}

	for _, pattern := range suspiciousPatterns {
		matched, _ := regexp.MatchString(pattern, txCtx.Email)
		if matched {
			return true, weight, nil
		}
	}

	return false, 0, nil
}

type DeviceCondition struct {
	RequireFingerprint bool     `json:"require_fingerprint"`
	BlockedDevices     []string `json:"blocked_devices"`
}

func (e *RuleEngineImpl) evaluateDeviceRule(rule *domain.RiskRule, txCtx *domain.TransactionContext) (bool, int, error) {
	var cond DeviceCondition
	if err := json.Unmarshal([]byte(rule.Condition), &cond); err != nil {
		return false, 0, fmt.Errorf("invalid device condition: %w", err)
	}

	if cond.RequireFingerprint && txCtx.DeviceFingerprint == "" {
		return true, rule.RiskWeight, nil
	}

	for _, blocked := range cond.BlockedDevices {
		if txCtx.DeviceFingerprint == blocked {
			return true, rule.RiskWeight, nil
		}
	}

	return false, 0, nil
}

func (e *RuleEngineImpl) evaluateCustomRule(rule *domain.RiskRule, txCtx *domain.TransactionContext) (bool, int, error) {
	conditions := make(map[string]interface{})
	if err := json.Unmarshal([]byte(rule.Condition), &conditions); err != nil {
		return false, 0, fmt.Errorf("invalid custom condition: %w", err)
	}

	for key, value := range conditions {
		switch key {
		case "ip_range":
			if ipRange, ok := value.(string); ok {
				if e.matchIPRange(txCtx.IPAddress, ipRange) {
					return true, rule.RiskWeight, nil
				}
			}
		case "amount_multiplier":
			if multiplier, ok := value.(float64); ok {
				if e.checkAmountMultiplier(txCtx, multiplier) {
					return true, rule.RiskWeight, nil
				}
			}
		case "metadata_match":
			if metadataCond, ok := value.(map[string]interface{}); ok {
				if e.matchMetadata(txCtx.Metadata, metadataCond) {
					return true, rule.RiskWeight, nil
				}
			}
		}
	}

	return false, 0, nil
}

func (e *RuleEngineImpl) matchIPRange(ip, ipRange string) bool {
	return ip != "" && ipRange != "" && ip[:len(ipRange)] == ipRange
}

func (e *RuleEngineImpl) checkAmountMultiplier(txCtx *domain.TransactionContext, multiplier float64) bool {
	if txCtx.Metadata == nil {
		return false
	}
	avgAmountStr, ok := txCtx.Metadata["avg_transaction_amount"]
	if !ok {
		return false
	}
	avgAmount, err := strconv.ParseFloat(avgAmountStr, 64)
	if err != nil {
		return false
	}
	return float64(txCtx.Amount) > avgAmount*multiplier
}

func (e *RuleEngineImpl) matchMetadata(metadata map[string]string, conditions map[string]interface{}) bool {
	for key, expectedValue := range conditions {
		if actualValue, ok := metadata[key]; ok {
			if actualValue != fmt.Sprintf("%v", expectedValue) {
				return false
			}
		} else {
			return false
		}
	}
	return true
}
