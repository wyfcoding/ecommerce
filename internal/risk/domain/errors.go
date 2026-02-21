// 生成摘要：risk 领域统一错误定义

package domain

import "errors"

var (
	// ErrRiskAssessmentNotFound 风险评估记录不存在
	ErrRiskAssessmentNotFound = errors.New("risk assessment not found")
	// ErrBlacklistEntryNotFound 黑名单条目不存在
	ErrBlacklistEntryNotFound = errors.New("blacklist entry not found")
	// ErrRuleNotFound 风控规则不存在
	ErrRuleNotFound = errors.New("risk rule not found")
	// ErrInsufficientCreditLimit 信用额度不足
	ErrInsufficientCreditLimit = errors.New("insufficient credit limit")
	// ErrCreditProfileNotFound 信用档案不存在
	ErrCreditProfileNotFound = errors.New("credit profile not found")
	// ErrCreditProfileFrozen 信用档案已冻结
	ErrCreditProfileFrozen = errors.New("credit profile frozen")
)
