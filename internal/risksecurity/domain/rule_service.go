package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/pkg/ruleengine"
)

// RuleEvaluationService 负责执行动态风控规则
type RuleEvaluationService struct {
	engine *ruleengine.Engine
	repo   RiskRepository // 需要访问存储以获取历史数据 (e.g. 过去1小时交易量)
}

func NewRuleEvaluationService(repo RiskRepository, logger *slog.Logger) *RuleEvaluationService {
	return &RuleEvaluationService{
		engine: ruleengine.NewEngine(logger),
		repo:   repo,
	}
}

// LoadRules 从数据库加载并编译所有规则
func (s *RuleEvaluationService) LoadRules(ctx context.Context) error {
	rules, err := s.repo.ListEnabledRules(ctx)
	if err != nil {
		return err
	}

	for _, r := range rules {
		err := s.engine.AddRule(ruleengine.Rule{
			ID:         fmt.Sprintf("%d", r.ID),
			Name:       r.Name,
			Expression: r.Condition,
			Metadata: map[string]any{
				"score": r.Score,
				"type":  string(r.Type),
			},
		})
		if err != nil {
			// Log error but continue loading other rules
			continue
		}
	}
	return nil
}

// EvaluateTransaction 评估一笔交易
func (s *RuleEvaluationService) EvaluateTransaction(ctx context.Context, riskCtx *RiskContext) (*RiskAnalysisResult, error) {
	// 1. 准备事实数据 (Facts)
	facts := map[string]any{
		"amount":         riskCtx.Amount,
		"currency":       "CNY", // 假设
		"user_id":        riskCtx.UserID,
		"ip":             riskCtx.IP,
		"device_id":      riskCtx.DeviceID,
		"payment_method": riskCtx.PaymentMethod,
		"hour_of_day":    time.Now().Hour(),
	}

	// 2. 增强事实数据：计算速度指标 (Velocity Metrics)
	// 这是一个非常“业务”的逻辑：需要查 Redis/DB 统计过去 X 时间窗口的聚合值
	// 规则示例： tx_count_1h > 10
	metrics, err := s.repo.GetVelocityMetrics(ctx, riskCtx.UserID)
	if err == nil {
		facts["tx_count_1h"] = metrics.TxCount1h
		facts["tx_amount_1h"] = metrics.TxAmount1h
		facts["tx_count_24h"] = metrics.TxCount24h
		facts["failed_tx_count_1h"] = metrics.FailedTxCount1h
	}

	// 3. 执行规则引擎
	results, err := s.engine.ExecuteAll(ctx, facts)
	if err != nil {
		return nil, err
	}

	// 4. 聚合结果
	totalScore := int32(0)
	var maxLevel RiskLevel = RiskLevelVeryLow
	triggeredItems := make([]*RiskItem, 0)

	for _, res := range results {
		scoreVal, _ := res.Metadata["score"].(int32)
		// 注意：Metadata 中的数字可能是 float64 (JSON/Expr 默认行为)，需转换
		if s, ok := res.Metadata["score"].(float64); ok {
			scoreVal = int32(s)
		} else if s, ok := res.Metadata["score"].(int); ok {
			scoreVal = int32(s)
		}

		typeVal, _ := res.Metadata["type"].(string)

		totalScore += scoreVal

		item := &RiskItem{
			Type:      RiskType(typeVal),
			Score:     scoreVal,
			Reason:    fmt.Sprintf("Rule matched: %s", res.RuleID), // 实际应使用规则名称
			Timestamp: time.Now(),
		}
		triggeredItems = append(triggeredItems, item)
	}

	// 5. 判定等级
	if totalScore > 100 {
		maxLevel = RiskLevelCritical
	} else if totalScore > 80 {
		maxLevel = RiskLevelHigh
	} else if totalScore > 50 {
		maxLevel = RiskLevelMedium
	} else if totalScore > 20 {
		maxLevel = RiskLevelLow
	}

	// 序列化
	itemsJSON, _ := json.Marshal(triggeredItems)

	return &RiskAnalysisResult{
		UserID:    riskCtx.UserID,
		RiskScore: totalScore,
		RiskLevel: maxLevel,
		RiskItems: string(itemsJSON),
	}, nil
}
