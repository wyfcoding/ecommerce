// 生成摘要：
// - 从 internal/riskanalyzer 服务合并而来，补充反欺诈引擎、风险评估聚合根等能力
// - 与 risksecurity.go 互补：risksecurity 定义基础风险类型与黑名单，fraud_engine 负责反欺诈分析与规则引擎
// - 新增 RiskAssessment（风险评估聚合根）、MatchedRule（命中规则明细）、
//   FraudEngine（反欺诈引擎接口）、RiskBridge（跨项目风控协同桥接器）等领域对象

package domain

import (
	"context"
	"fmt"
	"time"
)

// RiskRuleType 风险规则类型（细分类别）
type RiskRuleType int

const (
	RuleVelocity       RiskRuleType = 1 // 频次限制（如1分钟内支付10次）
	RuleDeviceFinger   RiskRuleType = 2 // 设备指纹异常
	RuleGeoMismatch    RiskRuleType = 3 // 地理位置异常（如IP与收货地距离过远）
	RuleLargeAmount    RiskRuleType = 4 // 大额订单异常
	RuleBlacklistMatch RiskRuleType = 5 // 黑名单匹配
	RuleScriptBehavior RiskRuleType = 6 // 脚本/机器人行为特征
)

// FraudRiskAssessmentDecision 风险评估决策
type FraudRiskAssessmentDecision string

const (
	DecisionPass   FraudRiskAssessmentDecision = "PASS"
	DecisionReview FraudRiskAssessmentDecision = "REVIEW"
	DecisionReject FraudRiskAssessmentDecision = "REJECT"
)

// FraudRiskRule 反欺诈风险规则实体
// 比 RiskRule 更详细，包含权重和拦截标志
type FraudRiskRule struct {
	ID         uint64       `json:"id"`
	Name       string       `json:"name"`
	RuleType   RiskRuleType `json:"rule_type"`
	Config     string       `json:"config"`      // JSON 格式规则配置
	Weight     int          `json:"weight"`      // 权重（分值贡献上限）
	IsBlocking bool         `json:"is_blocking"` // 是否直接拦截
	IsActive   bool         `json:"is_active"`
}

// RiskAssessment 风险评估聚合根
// 表示对某个目标（订单/支付/登录）的一次完整风险评估
type RiskAssessment struct {
	ID           uint64                      `json:"id"`
	CreatedAt    time.Time                   `json:"created_at"`
	TargetID     string                      `json:"target_id"`   // 目标 ID（OrderID/PaymentID/UserID）
	TargetType   string                      `json:"target_type"` // ORDER/PAYMENT/LOGIN
	UserID       uint64                      `json:"user_id"`
	TotalScore   int                         `json:"total_score"` // 最终风险得分
	Level        RiskLevel                   `json:"level"`
	MatchedRules []*MatchedRule              `json:"matched_rules"` // 命中的细则
	Decision     FraudRiskAssessmentDecision `json:"decision"`      // PASS/REVIEW/REJECT
	DeviceFinger string                      `json:"device_finger"`
	IPAddress    string                      `json:"ip_address"`
	ReviewerID   string                      `json:"reviewer_id"` // 后续人工审核人
	ReviewedAt   *time.Time                  `json:"reviewed_at"`
	ReviewNotes  string                      `json:"review_notes"`
}

// MatchedRule 命中的细则
type MatchedRule struct {
	RuleID      uint64 `json:"rule_id"`
	RuleName    string `json:"rule_name"`
	Score       int    `json:"score"`
	Reason      string `json:"reason"`
	RawEvidence string `json:"raw_evidence"` // 原始证据数据
}

// DetermineDecision 根据得分决定处理行为
// 假设：95 分以上直接拦截，80 分以上人工审核，其余放行
func (a *RiskAssessment) DetermineDecision() {
	if a.TotalScore >= 95 {
		a.Level = RiskLevelCritical
		a.Decision = DecisionReject
	} else if a.TotalScore >= 80 {
		a.Level = RiskLevelHigh
		a.Decision = DecisionReview
	} else if a.TotalScore >= 50 {
		a.Level = RiskLevelMedium
		a.Decision = DecisionPass
	} else {
		a.Level = RiskLevelLow
		a.Decision = DecisionPass
	}
}

// FraudEngine 反欺诈引擎服务接口
type FraudEngine interface {
	// Analyze 执行全面风险分析
	Analyze(ctx context.Context, assessment *RiskAssessment) (*RiskAssessment, error)
	// GetDeviceTrustScore 获取设备信任分
	GetDeviceTrustScore(ctx context.Context, deviceFinger string) (int, error)
}

// RiskAssessmentRepository 风险评估仓储接口
type RiskAssessmentRepository interface {
	SaveAssessment(ctx context.Context, assessment *RiskAssessment) error
	FindAssessment(ctx context.Context, targetID string) (*RiskAssessment, error)
	FindActiveRules(ctx context.Context, targetType string) ([]*FraudRiskRule, error)
}

// --- 跨项目风控协同 ---

// RiskSynergyScope 风险联动领域
type RiskSynergyScope string

const (
	ScopeEcommerce RiskSynergyScope = "ECOMMERCE"
	ScopeTrading   RiskSynergyScope = "TRADING"
)

// SharedRiskSignal 共享风险信号
// 用于电商与交易系统之间的风险联动
type SharedRiskSignal struct {
	SignalID   string           `json:"signal_id"`
	UserID     string           `json:"user_id"`
	Source     RiskSynergyScope `json:"source"`
	RiskType   string           `json:"risk_type"` // FRAUD/LAUNDERING/ABNORMAL_TRADING
	Level      RiskLevel        `json:"level"`
	Score      float64          `json:"score"` // 风险评分 (0-100)
	Details    string           `json:"details"`
	DetectedAt time.Time        `json:"detected_at"`
}

// RiskSynergyService 跨项目风控协同服务接口
type RiskSynergyService interface {
	ReportSignal(ctx context.Context, signal *SharedRiskSignal) error
	GetUnifiedRiskScore(ctx context.Context, userID string) (float64, error)
	ApplyRestriction(ctx context.Context, userID string, level RiskLevel) error
}

// RiskBridge 风险桥接器
// 负责处理跨项目风险联动逻辑
type RiskBridge struct {
	service RiskSynergyService
}

// NewRiskBridge 创建风险桥接器
func NewRiskBridge(s RiskSynergyService) *RiskBridge {
	return &RiskBridge{service: s}
}

// HandleEcommerceRisk 处理由电商侧触发的风险联动
// 将电商侧的高风险操作（如频繁退款、盗号预警）同步给交易系统限制出金或下单
func (b *RiskBridge) HandleEcommerceRisk(ctx context.Context, userID, riskType string, score float64) error {
	level := RiskLevelLow
	if score > 80 {
		level = RiskLevelCritical
	} else if score > 60 {
		level = RiskLevelHigh
	} else if score > 40 {
		level = RiskLevelMedium
	}

	signal := &SharedRiskSignal{
		SignalID:   fmt.Sprintf("RISK-E-%d", time.Now().UnixNano()),
		UserID:     userID,
		Source:     ScopeEcommerce,
		RiskType:   riskType,
		Level:      level,
		Score:      score,
		DetectedAt: time.Now(),
	}

	// 1. 上报风险信号
	if err := b.service.ReportSignal(ctx, signal); err != nil {
		return err
	}

	// 2. 如果是高风险，自动应用交易限制
	if level >= RiskLevelHigh {
		return b.service.ApplyRestriction(ctx, userID, level)
	}

	return nil
}
