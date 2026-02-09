// 变更说明：实现跨项目风控联动逻辑，支持电商侧用户画像与交易侧交易风控指标的同步与联动评分。
// 假设：电商侧的高风险操作（如频繁退款、盗号预警）会同步给交易系统限制出金或下单；交易侧的异常波动也会反馈给电商系统锁定余额。
package domain

import (
	"context"
	"fmt"
	"time"
)

// RiskSynergyScope 风险联动领域
type RiskSynergyScope string

const (
	ScopeEcommerce RiskSynergyScope = "ECOMMERCE"
	ScopeTrading   RiskSynergyScope = "TRADING"
)

// SharedRiskSignal 共享风险信号
type SharedRiskSignal struct {
	SignalID   string
	UserID     string
	Source     RiskSynergyScope
	RiskType   string // e.g., FRAUD, LAUNDERING, ABNORMAL_TRADING
	Level      RiskLevel
	Score      float64 // 风险评分 (0-100)
	Details    string
	DetectedAt time.Time
}

// RiskSynergyService 跨项目风控协同服务
type RiskSynergyService interface {
	ReportSignal(ctx context.Context, signal *SharedRiskSignal) error
	GetUnifiedRiskScore(ctx context.Context, userID string) (float64, error)
	ApplyRestriction(ctx context.Context, userID string, level RiskLevel) error
}

// RiskBridge 风险桥接器
type RiskBridge struct {
	service RiskSynergyService
}

func NewRiskBridge(s RiskSynergyService) *RiskBridge {
	return &RiskBridge{service: s}
}

// HandleEcommerceRisk 处理由电商侧触发的风险联动
func (b *RiskBridge) HandleEcommerceRisk(ctx context.Context, userID, riskType string, score float64) error {
	level := RiskLow
	if score > 80 {
		level = RiskCritical
	} else if score > 60 {
		level = RiskHigh
	} else if score > 40 {
		level = RiskMedium
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
	if level >= RiskHigh {
		return b.service.ApplyRestriction(ctx, userID, level)
	}

	return nil
}
