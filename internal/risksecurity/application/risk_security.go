package application

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/risksecurity/domain"
	riskv1 "github.com/wyfcoding/financialtrading/goapi/risk/v1"
)

// RiskService 作为风控安全操作的门面。
type RiskService struct {
	manager *RiskManager
	query   *RiskQuery
}

// NewRiskService 创建风控服务门面实例。
func NewRiskService(manager *RiskManager, query *RiskQuery) *RiskService {
	return &RiskService{
		manager: manager,
		query:   query,
	}
}

func (s *RiskService) SetRemoteRiskClient(cli riskv1.RiskServiceClient) {
	s.manager.SetRemoteRiskClient(cli)
}

// --- 写操作（委托给 Manager）---

// EvaluateRisk 评估交易风险。
func (s *RiskService) EvaluateRisk(ctx context.Context, userID uint64, ip, deviceID string, amount int64) (*domain.RiskAnalysisResult, error) {
	return s.manager.EvaluateRisk(ctx, userID, ip, deviceID, amount)
}

// AddToBlacklist 将用户或IP加入黑名单。
func (s *RiskService) AddToBlacklist(ctx context.Context, bType string, value, reason string, duration time.Duration) error {
	return s.manager.AddToBlacklist(ctx, bType, value, reason, duration)
}

// RemoveFromBlacklist 从黑名单中移除。
func (s *RiskService) RemoveFromBlacklist(ctx context.Context, id uint64) error {
	return s.manager.RemoveFromBlacklist(ctx, id)
}

// RecordUserBehavior 记录用户行为数据。
func (s *RiskService) RecordUserBehavior(ctx context.Context, userID uint64, ip, deviceID string) error {
	return s.manager.RecordUserBehavior(ctx, userID, ip, deviceID)
}

// --- 读操作（委托给 Query）---

// GetRiskAnalysisResult 获取指定用户的风险评估结果。
func (s *RiskService) GetRiskAnalysisResult(ctx context.Context, userID uint64) (*domain.RiskAnalysisResult, error) {
	return s.query.GetRiskAnalysisResult(ctx, userID)
}

// CheckBlacklist 检查是否在黑名单中。
func (s *RiskService) CheckBlacklist(ctx context.Context, bType string, value string) (*domain.Blacklist, error) {
	return s.query.CheckBlacklist(ctx, bType, value)
}

// GetUserBehavior 获取用户的行为记录。
func (s *RiskService) GetUserBehavior(ctx context.Context, userID uint64) (*domain.UserBehavior, error) {
	return s.query.GetUserBehavior(ctx, userID)
}
