package application

import (
	"context"
	"ecommerce/internal/risk_security/domain/entity"
	"ecommerce/internal/risk_security/domain/repository"
	"encoding/json"
	"time"

	"log/slog"
)

type RiskService struct {
	repo   repository.RiskRepository
	logger *slog.Logger
}

func NewRiskService(repo repository.RiskRepository, logger *slog.Logger) *RiskService {
	return &RiskService{
		repo:   repo,
		logger: logger,
	}
}

// EvaluateRisk 评估风险
func (s *RiskService) EvaluateRisk(ctx context.Context, userID uint64, ip, deviceID string, amount int64) (*entity.RiskAnalysisResult, error) {
	// 1. Check Blacklist
	isIPBlacklisted, _ := s.repo.IsBlacklisted(ctx, entity.BlacklistTypeIP, ip)
	if isIPBlacklisted {
		return s.createResult(ctx, userID, entity.RiskLevelCritical, 100, "IP in blacklist")
	}

	isUserBlacklisted, _ := s.repo.IsBlacklisted(ctx, entity.BlacklistTypeUser, string(rune(userID))) // Simplified conversion
	if isUserBlacklisted {
		return s.createResult(ctx, userID, entity.RiskLevelCritical, 100, "User in blacklist")
	}

	// 2. Check Rules (Mock logic)
	score := int32(0)
	if amount > 1000000 { // > 10000.00
		score += 20
	}

	// 3. Check Behavior (Mock)
	behavior, _ := s.repo.GetUserBehavior(ctx, userID)
	if behavior != nil && behavior.LastLoginIP != "" && behavior.LastLoginIP != ip {
		score += 30
	}

	level := entity.RiskLevelLow
	if score > 80 {
		level = entity.RiskLevelCritical
	} else if score > 60 {
		level = entity.RiskLevelHigh
	} else if score > 40 {
		level = entity.RiskLevelMedium
	}

	return s.createResult(ctx, userID, level, score, "Risk evaluation completed")
}

func (s *RiskService) createResult(ctx context.Context, userID uint64, level entity.RiskLevel, score int32, reason string) (*entity.RiskAnalysisResult, error) {
	items := []*entity.RiskItem{
		{
			Type:      entity.RiskTypeAnomalousTransaction, // Default for now
			Level:     level,
			Score:     score,
			Reason:    reason,
			Timestamp: time.Now(),
		},
	}
	itemsJSON, _ := json.Marshal(items)

	result := &entity.RiskAnalysisResult{
		UserID:    userID,
		RiskScore: score,
		RiskLevel: level,
		RiskItems: string(itemsJSON),
	}

	if err := s.repo.SaveAnalysisResult(ctx, result); err != nil {
		return nil, err
	}
	return result, nil
}

// AddToBlacklist 添加到黑名单
func (s *RiskService) AddToBlacklist(ctx context.Context, bType string, value, reason string, duration time.Duration) error {
	blacklist := &entity.Blacklist{
		Type:      entity.BlacklistType(bType),
		Value:     value,
		Reason:    reason,
		ExpiresAt: time.Now().Add(duration),
	}
	return s.repo.SaveBlacklist(ctx, blacklist)
}

// RemoveFromBlacklist 移除黑名单
func (s *RiskService) RemoveFromBlacklist(ctx context.Context, id uint64) error {
	return s.repo.DeleteBlacklist(ctx, id)
}

// RecordUserBehavior 记录用户行为
func (s *RiskService) RecordUserBehavior(ctx context.Context, userID uint64, ip, deviceID string) error {
	behavior, err := s.repo.GetUserBehavior(ctx, userID)
	if err != nil {
		return err
	}

	if behavior == nil {
		behavior = &entity.UserBehavior{
			UserID: userID,
		}
	}

	behavior.LastLoginIP = ip
	behavior.LastLoginDevice = deviceID
	behavior.LastLoginTime = time.Now()

	return s.repo.SaveUserBehavior(ctx, behavior)
}
