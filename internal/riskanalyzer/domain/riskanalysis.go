// 变更说明：新增风险分析功能，支持反欺诈引擎、风险规则引擎、实时异常拦截及黑名单管理。
// 假设：风险分为0-100分，80分以上视为高风险需要人工审核，95分以上直接拦截。
package domain

import (
	"context"
	"time"
)

// --- 风险等级 ---

// RiskLevel 风险等级
type RiskLevel string

const (
	RiskLow      RiskLevel = "LOW"
	RiskMedium   RiskLevel = "MEDIUM"
	RiskHigh     RiskLevel = "HIGH"
	RiskCritical RiskLevel = "CRITICAL"
)

// --- 风险规则 ---

// RiskRuleType 风险规则类型
type RiskRuleType int

const (
	RuleVelocity       RiskRuleType = 1 // 频次限制（如1分钟内支付10次）
	RuleDeviceFinger   RiskRuleType = 2 // 设备指纹异常
	RuleGeoMismatch    RiskRuleType = 3 // 地理位置异常（如IP与收货地距离过远）
	RuleLargeAmount    RiskRuleType = 4 // 大额订单异常
	RuleBlacklistMatch RiskRuleType = 5 // 黑名单匹配
	RuleScriptBehavior RiskRuleType = 6 // 脚本/机器人行为特征
)

// RiskRule 风险规则实体
type RiskRule struct {
	ID         uint64       `json:"id"`
	Name       string       `json:"name"`
	Type       RiskRuleType `json:"type"`
	Config     string       `json:"config"`      // JSON格式规则配置
	Weight     int          `json:"weight"`      // 权重（分值贡献上限）
	IsBlocking bool         `json:"is_blocking"` // 是否直接拦截
	IsActive   bool         `json:"is_active"`
}

// --- 风险评估结果 ---

// RiskAssessment 风险评估聚合根
type RiskAssessment struct {
	ID           uint64         `json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	TargetID     string         `json:"target_id"`   // 目标ID（OrderID/PaymentID/UserID）
	TargetType   string         `json:"target_type"` // ORDER/PAYMENT/LOGIN
	UserID       uint64         `json:"user_id"`
	TotalScore   int            `json:"total_score"` // 最终风险得分
	Level        RiskLevel      `json:"level"`
	MatchedRules []*MatchedRule `json:"matched_rules"` // 命中的细则
	Decision     string         `json:"decision"`      // PASS/REVIEW/REJECT
	DeviceFinger string         `json:"device_finger"`
	IPAddress    string         `json:"ip_address"`
	ReviewerID   string         `json:"reviewer_id"` // 后续人工审核人
	ReviewedAt   *time.Time     `json:"reviewed_at"`
	ReviewNotes  string         `json:"review_notes"`
}

// MatchedRule 命中的细则
type MatchedRule struct {
	RuleID      uint64 `json:"rule_id"`
	RuleName    string `json:"rule_name"`
	Score       int    `json:"score"`
	Reason      string `json:"reason"`
	RawEvidence string `json:"raw_evidence"` // 原始证据数据
}

// --- 黑名单 ---

// BlacklistType 黑名单类型
type BlacklistType int

const (
	BlacklistUser   BlacklistType = 1
	BlacklistIP     BlacklistType = 2
	BlacklistDevice BlacklistType = 3
	BlacklistCard   BlacklistType = 4 // 银行卡号
)

// Blacklist 黑名单项
type Blacklist struct {
	ID        uint64        `json:"id"`
	Type      BlacklistType `json:"type"`
	Value     string        `json:"value"` // 目标值（UID/IP/DeviceID）
	Reason    string        `json:"reason"`
	Source    string        `json:"source"`     // 来源：SYSTEM/MANUAL/THIRD_PARTY
	ExpiresAt *time.Time    `json:"expires_at"` // 过期时间（nil为永久）
	CreatedAt time.Time     `json:"created_at"`
}

// --- 风险分析仓库接口 ---

// RiskRepository 风险仓库接口
type RiskRepository interface {
	SaveAssessment(ctx context.Context, assessment *RiskAssessment) error
	FindAssessment(ctx context.Context, targetID string) (*RiskAssessment, error)

	FindActiveRules(ctx context.Context, targetType string) ([]*RiskRule, error)

	IsInBlacklist(ctx context.Context, blacklistType BlacklistType, value string) (bool, error)
	AddToBlacklist(ctx context.Context, item *Blacklist) error
	RemoveFromBlacklist(ctx context.Context, blacklistType BlacklistType, value string) error
}

// --- 反欺诈引擎服务接口 ---

// FraudEngine 反欺诈引擎服务接口
type FraudEngine interface {
	// Analyze 执行全面风险分析
	Analyze(ctx context.Context, assessment *RiskAssessment) (*RiskAssessment, error)
	// GetDeviceTrustScore 获取设备信任分
	GetDeviceTrustScore(ctx context.Context, deviceFinger string) (int, error)
}

// --- 业务方法示例 ---

// DetermineDecision 根据得分决定处理行为
func (a *RiskAssessment) DetermineDecision() {
	if a.TotalScore >= 95 {
		a.Level = RiskCritical
		a.Decision = "REJECT"
	} else if a.TotalScore >= 80 {
		a.Level = RiskHigh
		a.Decision = "REVIEW"
	} else if a.TotalScore >= 50 {
		a.Level = RiskMedium
		a.Decision = "PASS" // 或者强制二次验证
	} else {
		a.Level = RiskLow
		a.Decision = "PASS"
	}
}
