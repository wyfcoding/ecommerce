package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/wyfcoding/ecommerce/internal/risksecurity/domain"
	riskv1 "github.com/wyfcoding/financialtrading/goapi/risk/v1"
	"github.com/wyfcoding/pkg/algorithm"
	"github.com/wyfcoding/pkg/utils/ctxutil"
)

// RiskManager 处理风控安全的写操作。
type RiskManager struct {
	repo          domain.RiskRepository
	logger        *slog.Logger
	detector      *algorithm.AntiBotDetector
	calculator    *algorithm.RiskCalculator
	remoteRiskCli riskv1.RiskServiceClient
}

// NewRiskManager creates a new RiskManager instance.
func NewRiskManager(repo domain.RiskRepository, logger *slog.Logger) *RiskManager {
	return &RiskManager{
		repo:       repo,
		logger:     logger,
		detector:   algorithm.NewAntiBotDetector(),
		calculator: algorithm.NewRiskCalculator(),
	}
}

func (m *RiskManager) SetRemoteRiskClient(cli riskv1.RiskServiceClient) {
	m.remoteRiskCli = cli
}

// UserRelation 用户关联关系
type UserRelation struct {
	FromUserID int
	ToUserID   int
}

// DetectFraudGroups 检测欺诈团伙
func (m *RiskManager) DetectFraudGroups(ctx context.Context, numUsers int, relations []UserRelation) [][]int {
	m.logger.InfoContext(ctx, "starting fraud group detection", "nodes", numUsers, "relations", len(relations))

	// 1. 构建图
	g := algorithm.NewGraph(numUsers)
	for _, rel := range relations {
		g.AddEdge(rel.FromUserID, rel.ToUserID)
	}

	// 2. 运行 Tarjan 算法
	scc := algorithm.NewTarjanSCC(g)
	groups := scc.Run()

	// 3. 过滤出规模大于 1 的团伙
	fraudGroups := make([][]int, 0)
	for _, group := range groups {
		if len(group) > 1 {
			fraudGroups = append(fraudGroups, group)
		}
	}

	m.logger.InfoContext(ctx, "fraud group detection completed", "detected_groups", len(fraudGroups))
	return fraudGroups
}

// EvaluateRisk 评估指定用户操作的风险。
func (m *RiskManager) EvaluateRisk(ctx context.Context, userID uint64, ip, deviceID string, amount int64) (*domain.RiskAnalysisResult, error) {
	// 1. 检查黑名单 (一票否决)
	isIPBlacklisted, _ := m.repo.IsBlacklisted(ctx, domain.BlacklistTypeIP, ip)
	isUserBlacklisted, _ := m.repo.IsBlacklisted(ctx, domain.BlacklistTypeUser, fmt.Sprintf("%d", userID))

	if isIPBlacklisted || isUserBlacklisted {
		return m.createResult(ctx, userID, domain.RiskLevelCritical, 100, "Entity found in blacklist")
	}

	// 2. 准备多准则评分因子 (0-1)
	factors := make(map[string]float64)

	// 2.1 金额风险 (假设 50000 以上为高风险)
	if amount > 5000000 {
		factors["amount_risk"] = 1.0
	} else if amount > 1000000 {
		factors["amount_risk"] = 0.5
	}

	// 2.2 异地登录风险
	behaviorData, _ := m.repo.GetUserBehavior(ctx, userID)
	if behaviorData != nil && behaviorData.LastLoginIP != "" && behaviorData.LastLoginIP != ip {
		factors["location_risk"] = 0.8
	}

	// 2.3 机器人检测风险
	currentBehavior := algorithm.UserBehavior{
		UserID:    userID,
		IP:        ip,
		UserAgent: ctxutil.GetUserAgent(ctx),
		Timestamp: time.Now(),
		Action:    "transaction",
	}
	isBot, botReason := m.detector.IsBot(currentBehavior)
	if isBot {
		factors["bot_risk"] = 1.0
	} else {
		factors["bot_risk"] = float64(m.detector.GetRiskScore(currentBehavior)) / 100.0
	}

	// 2.4 远程金融风控风险
	if m.remoteRiskCli != nil {
		remoteResp, err := m.remoteRiskCli.AssessRisk(ctx, &riskv1.AssessRiskRequest{
			UserId: fmt.Sprintf("%d", userID),
			Symbol: "PAYMENT",
			Side:   "OUT",
			Price:  fmt.Sprintf("%d", amount),
		})
		if err == nil {
			score, _ := strconv.ParseFloat(remoteResp.RiskScore, 64)
			factors["financial_risk"] = score / 100.0
		}
	}

	// 3. 执行加权聚合评分
	weights := map[string]float64{
		"bot_risk":       0.4,
		"financial_risk": 0.3,
		"location_risk":  0.2,
		"amount_risk":    0.1,
	}

	finalScore := m.calculator.EvaluateFraudScore(factors, weights)
	intScore := int32(finalScore * 100)

	// 4. 判定风险等级
	level := domain.RiskLevelLow
	if intScore > 80 {
		level = domain.RiskLevelCritical
	} else if intScore > 60 {
		level = domain.RiskLevelHigh
	} else if intScore > 40 {
		level = domain.RiskLevelMedium
	}

	return m.createResult(ctx, userID, level, intScore, fmt.Sprintf("Weighted analysis completed. Bot detected: %v (%s)", isBot, botReason))
}

// createResult 是一个辅助函数，用于创建并保存 RiskAnalysisResult 实体。
func (m *RiskManager) createResult(ctx context.Context, userID uint64, level domain.RiskLevel, score int32, reason string) (*domain.RiskAnalysisResult, error) {
	items := []*domain.RiskItem{
		{
			Type:      domain.RiskTypeAnomalousTransaction,
			Level:     level,
			Score:     score,
			Reason:    reason,
			Timestamp: time.Now(),
		},
	}
	itemsJSON, err := json.Marshal(items)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to marshal risk items", "user_id", userID, "error", err)
		return nil, err
	}

	result := &domain.RiskAnalysisResult{
		UserID:    userID,
		RiskScore: score,
		RiskLevel: level,
		RiskItems: string(itemsJSON),
	}

	if err := m.repo.SaveAnalysisResult(ctx, result); err != nil {
		m.logger.ErrorContext(ctx, "failed to save risk analysis result", "user_id", userID, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "risk analysis result saved", "user_id", userID, "risk_level", level)
	return result, nil
}

// AddToBlacklist 将指定类型和值的实体添加到黑名单。
func (m *RiskManager) AddToBlacklist(ctx context.Context, bType string, value, reason string, duration time.Duration) error {
	blacklist := &domain.Blacklist{
		Type:      domain.BlacklistType(bType),
		Value:     value,
		Reason:    reason,
		ExpiresAt: time.Now().Add(duration),
	}
	if err := m.repo.SaveBlacklist(ctx, blacklist); err != nil {
		m.logger.ErrorContext(ctx, "failed to save blacklist entry", "type", bType, "value", value, "error", err)
		return err
	}
	m.logger.InfoContext(ctx, "added to blacklist", "type", bType, "value", value)
	return nil
}

// RemoveFromBlacklist 从黑名单中移除指定ID的条目。
func (m *RiskManager) RemoveFromBlacklist(ctx context.Context, id uint64) error {
	if err := m.repo.DeleteBlacklist(ctx, id); err != nil {
		m.logger.ErrorContext(ctx, "failed to delete blacklist entry", "id", id, "error", err)
		return err
	}
	m.logger.InfoContext(ctx, "removed from blacklist", "id", id)
	return nil
}

// RecordUserBehavior 记录或更新用户的行为数据。
func (m *RiskManager) RecordUserBehavior(ctx context.Context, userID uint64, ip, deviceID string) error {
	behavior, err := m.repo.GetUserBehavior(ctx, userID)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to get user behavior for recording", "user_id", userID, "error", err)
		return err
	}

	if behavior == nil {
		behavior = &domain.UserBehavior{
			UserID: userID,
		}
	}

	behavior.LastLoginIP = ip
	behavior.LastLoginDevice = deviceID
	behavior.LastLoginTime = time.Now()

	if err := m.repo.SaveUserBehavior(ctx, behavior); err != nil {
		m.logger.ErrorContext(ctx, "failed to save user behavior", "user_id", userID, "error", err)
		return err
	}
	m.logger.InfoContext(ctx, "user behavior recorded", "user_id", userID, "ip", ip)
	return nil
}
