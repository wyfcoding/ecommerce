package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/risksecurity/domain"
)

// RiskManager 处理风控安全的写操作。
type RiskManager struct {
	repo   domain.RiskRepository
	logger *slog.Logger
}

// NewRiskManager creates a new RiskManager instance.
func NewRiskManager(repo domain.RiskRepository, logger *slog.Logger) *RiskManager {
	return &RiskManager{
		repo:   repo,
		logger: logger,
	}
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

	// 2. 评估风险规则
	score := int32(0)
	if amount > 1000000 {
		score += 20
	}

	// 3. 评估用户行为
	behavior, err := m.repo.GetUserBehavior(ctx, userID)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to get user behavior", "user_id", userID, "error", err)
	}
	if behavior != nil && behavior.LastLoginIP != "" && behavior.LastLoginIP != ip {
		score += 30
	}

	// 4. 根据风险分数确定风险等级
	level := domain.RiskLevelLow
	if score > 80 {
		level = domain.RiskLevelCritical
	} else if score > 60 {
		level = domain.RiskLevelHigh
	} else if score > 40 {
		level = domain.RiskLevelMedium
	}

	return m.createResult(ctx, userID, level, score, "Risk evaluation completed")
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
