package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/risksecurity/domain"
	riskv1 "github.com/wyfcoding/financialtrading/goapi/risk/v1"
	"github.com/wyfcoding/pkg/algorithm"
)

// RiskManager 处理风控安全的写操作。
type RiskManager struct {
	repo          domain.RiskRepository
	logger        *slog.Logger
	detector      *algorithm.AntiBotDetector
	remoteRiskCli riskv1.RiskServiceClient
}

// NewRiskManager creates a new RiskManager instance.
func NewRiskManager(repo domain.RiskRepository, logger *slog.Logger) *RiskManager {
	return &RiskManager{
		repo:     repo,
		logger:   logger,
		detector: algorithm.NewAntiBotDetector(),
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
	// 1. 检查黑名单
	isIPBlacklisted, err := m.repo.IsBlacklisted(ctx, domain.BlacklistTypeIP, ip)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to check IP blacklist", "ip", ip, "error", err)
	}
	if isIPBlacklisted {
		return m.createResult(ctx, userID, domain.RiskLevelCritical, 100, "IP address found in blacklist")
	}

	isUserBlacklisted, err := m.repo.IsBlacklisted(ctx, domain.BlacklistTypeUser, fmt.Sprintf("%d", userID))
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to check user blacklist", "user_id", userID, "error", err)
	}
	if isUserBlacklisted {
		return m.createResult(ctx, userID, domain.RiskLevelCritical, 100, "User ID found in blacklist")
	}

	// 2. 评估风险规则 (基础规则)
	score := int32(0)
	if amount > 1000000 {
		score += 20
	}

	// 3. 评估用户行为 (历史行为)
	behaviorData, err := m.repo.GetUserBehavior(ctx, userID)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to get user behavior", "user_id", userID, "error", err)
	}
	if behaviorData != nil && behaviorData.LastLoginIP != "" && behaviorData.LastLoginIP != ip {
		score += 30
	}

	// 4. 实时反爬虫/机器人检测 (算法模型)
	currentBehavior := algorithm.UserBehavior{
		UserID:    userID,
		IP:        ip,
		UserAgent: "Unknown", // TODO: 应该从Context或上层传入UserAgent
		Timestamp: time.Now(),
		Action:    "transaction",
	}

	isBot, reason := m.detector.IsBot(currentBehavior)
	botScore := m.detector.GetRiskScore(currentBehavior)

	// 融合算法评分
	if isBot {
		score += 50
		m.logger.WarnContext(ctx, "bot activity detected", "user_id", userID, "reason", reason)
	}
	score += int32(botScore)

	// 5. 远程风控评估 (Cross-Project Interaction)
	if m.remoteRiskCli != nil {
		remoteResp, err := m.remoteRiskCli.AssessRisk(ctx, &riskv1.AssessRiskRequest{
			UserId:   fmt.Sprintf("%d", userID),
			Symbol:   "RETAIL/PAYMENT", // 模拟零售支付场景
			Side:     "OUT",
			Quantity: "1",
			Price:    fmt.Sprintf("%d", amount),
		})
		if err != nil {
			m.logger.WarnContext(ctx, "remote risk assessment failed, skipping", "error", err)
		} else {
			m.logger.InfoContext(ctx, "remote financial risk check completed", "score", remoteResp.RiskScore, "is_allowed", remoteResp.IsAllowed)
			if !remoteResp.IsAllowed {
				score += 40 // 叠加远程高风险分数
			}
		}
	}

	// 6. 根据风险分数确定风险等级
	if score > 100 {
		score = 100
	}
	level := domain.RiskLevelLow
	if score > 80 {
		level = domain.RiskLevelCritical
	} else if score > 60 {
		level = domain.RiskLevelHigh
	} else if score > 40 {
		level = domain.RiskLevelMedium
	}

	return m.createResult(ctx, userID, level, score, fmt.Sprintf("Risk evaluation completed. Bot check: %v (%s)", isBot, reason))
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
