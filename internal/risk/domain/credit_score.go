// 生成摘要：
// - 从 internal/riskanalyzer 服务合并而来，补充信用分模型与额度管理能力
// - 支持"先用后付"等信用支付场景的决策支撑
// - 新增 CreditProfile（信用档案）、CreditRecord（信用轨迹）、
//   CreditEvaluator（信用评分引擎接口）、CreditProfileRepository 等领域对象
// - 信用分范围 300-850 分，额度根据信用等级与收入模型自动计算

package domain

import (
	"context"
	"time"
)

// CreditLevel 信用等级
type CreditLevel string

const (
	CreditLevelPoor      CreditLevel = "POOR"      // 300-500
	CreditLevelFair      CreditLevel = "FAIR"      // 500-650
	CreditLevelGood      CreditLevel = "GOOD"      // 650-750
	CreditLevelExcellent CreditLevel = "EXCELLENT" // 750-850
)

// ScoreFactor 评分因子
type ScoreFactor string

const (
	FactorPaymentHistory ScoreFactor = "PAYMENT_HISTORY" // 还款历史（权重35%）
	FactorDebtRatio      ScoreFactor = "DEBT_RATIO"      // 负债比率（权重30%）
	FactorAccountAge     ScoreFactor = "ACCOUNT_AGE"     // 账户时长（权重15%）
	FactorBehaviorData   ScoreFactor = "BEHAVIOR_DATA"   // 购物行为稳定性（权重10%）
	FactorIdentityAuth   ScoreFactor = "IDENTITY_AUTH"   // 身份完整度/KYC（权重10%）
)

// CreditRecordType 信用记录类型
type CreditRecordType int

const (
	RecordOntimeRepay  CreditRecordType = 1 // 按时还款（加分）
	RecordLateRepay    CreditRecordType = 2 // 逾期（减分）
	RecordDefault      CreditRecordType = 3 // 违约（重度减分）
	RecordLimitUpdate  CreditRecordType = 4 // 额度调整
	RecordScoreRefresh CreditRecordType = 5 // 定期评分更新
)

// CreditProfile 用户信用档案聚合
// 包含信用分数、等级、额度信息以及各维度评分因子
type CreditProfile struct {
	ID             uint64         `json:"id"`
	UserID         uint64         `json:"user_id"`
	Score          int            `json:"score"`
	Level          CreditLevel    `json:"level"`
	TotalLimit     int64          `json:"total_limit"`     // 总信用额度（分）
	UsedLimit      int64          `json:"used_limit"`      // 已用额度（分）
	AvailableLimit int64          `json:"available_limit"` // 剩余可用额度（分）
	Factors        []*FactorScore `json:"factors"`         // 各项分值
	LastAssessedAt time.Time      `json:"last_assessed_at"`
	Status         string         `json:"status"` // ACTIVE/FROZEN/CLOSED
}

// FactorScore 单项因子得分
type FactorScore struct {
	Factor   ScoreFactor `json:"factor"`
	Score    int         `json:"score"`
	Comments string      `json:"comments"`
}

// CreditRecord 信用轨迹记录
// 追踪每一次信用变化事件
type CreditRecord struct {
	ID          uint64           `json:"id"`
	ProfileID   uint64           `json:"profile_id"`
	Type        CreditRecordType `json:"type"`
	ScoreChange int              `json:"score_change"`
	LimitChange int64            `json:"limit_change"`
	RelatedID   string           `json:"related_id"` // 关联订单/还款单号
	Reason      string           `json:"reason"`
	HappenedAt  time.Time        `json:"happened_at"`
}

// DetermineLevel 根据分数更新信用等级
func (p *CreditProfile) DetermineLevel() {
	switch {
	case p.Score >= 750:
		p.Level = CreditLevelExcellent
	case p.Score >= 650:
		p.Level = CreditLevelGood
	case p.Score >= 500:
		p.Level = CreditLevelFair
	default:
		p.Level = CreditLevelPoor
	}
}

// FreezeLimit 冻结额度（用于"先用后付"下单时预占额度）
func (p *CreditProfile) FreezeLimit(amount int64) error {
	if amount > p.AvailableLimit {
		return ErrInsufficientCreditLimit
	}
	p.UsedLimit += amount
	p.AvailableLimit -= amount
	return nil
}

// ReleaseLimit 释放额度（还款或取消时归还额度）
func (p *CreditProfile) ReleaseLimit(amount int64) {
	p.UsedLimit -= amount
	p.AvailableLimit += amount
}

// CreditEvaluator 信用评分服务接口
type CreditEvaluator interface {
	// Evaluate 执行全面信用评分
	Evaluate(ctx context.Context, userID uint64) (*CreditProfile, error)
	// CalculateLimit 测算可用信用额度
	CalculateLimit(ctx context.Context, profile *CreditProfile) (int64, error)
}

// CreditProfileRepository 信用档案仓储
type CreditProfileRepository interface {
	FindByUserID(ctx context.Context, userID uint64) (*CreditProfile, error)
	Save(ctx context.Context, profile *CreditProfile) error
	AddRecord(ctx context.Context, record *CreditRecord) error
	FindRecords(ctx context.Context, profileID uint64, limit int) ([]*CreditRecord, error)
}
