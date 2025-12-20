package application

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/risk_security/domain"
)

// RiskService acts as a facade for risk security operations.
type RiskService struct {
	manager *RiskManager
	query   *RiskQuery
}

// NewRiskService creates a new RiskService facade.
func NewRiskService(manager *RiskManager, query *RiskQuery) *RiskService {
	return &RiskService{
		manager: manager,
		query:   query,
	}
}

// --- Write Operations (Delegated to Manager) ---

func (s *RiskService) EvaluateRisk(ctx context.Context, userID uint64, ip, deviceID string, amount int64) (*domain.RiskAnalysisResult, error) {
	return s.manager.EvaluateRisk(ctx, userID, ip, deviceID, amount)
}

func (s *RiskService) AddToBlacklist(ctx context.Context, bType string, value, reason string, duration time.Duration) error {
	return s.manager.AddToBlacklist(ctx, bType, value, reason, duration)
}

func (s *RiskService) RemoveFromBlacklist(ctx context.Context, id uint64) error {
	return s.manager.RemoveFromBlacklist(ctx, id)
}

func (s *RiskService) RecordUserBehavior(ctx context.Context, userID uint64, ip, deviceID string) error {
	return s.manager.RecordUserBehavior(ctx, userID, ip, deviceID)
}

// --- Read Operations (Delegated to Query) ---

func (s *RiskService) GetRiskAnalysisResult(ctx context.Context, userID uint64) (*domain.RiskAnalysisResult, error) {
	return s.query.GetRiskAnalysisResult(ctx, userID)
}

func (s *RiskService) CheckBlacklist(ctx context.Context, bType string, value string) (*domain.Blacklist, error) {
	return s.query.CheckBlacklist(ctx, bType, value)
}

func (s *RiskService) GetUserBehavior(ctx context.Context, userID uint64) (*domain.UserBehavior, error) {
	return s.query.GetUserBehavior(ctx, userID)
}
