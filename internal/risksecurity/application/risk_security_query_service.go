package application

import (
	"context"
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/risksecurity/domain"
)

// RiskSecurityQueryService 处理风控安全的读操作。
type RiskSecurityQueryService struct {
	repo domain.RiskRepository
}

// NewRiskSecurityQueryService creates a new RiskSecurityQueryService instance.
func NewRiskSecurityQueryService(repo domain.RiskRepository) *RiskSecurityQueryService {
	return &RiskSecurityQueryService{
		repo: repo,
	}
}

// GetRiskAnalysisResult 获取指定用户的最新风险分析结果。
func (q *RiskSecurityQueryService) GetRiskAnalysisResult(ctx context.Context, userID uint64) (*domain.RiskAnalysisResult, error) {
	results, err := q.repo.ListAnalysisResults(ctx, userID, 1)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no risk analysis result found for user %d", userID)
	}
	return results[0], nil
}

// CheckBlacklist 检查指定类型和值是否在黑名单中。
func (q *RiskSecurityQueryService) CheckBlacklist(ctx context.Context, bType string, value string) (*domain.Blacklist, error) {
	return q.repo.GetBlacklist(ctx, domain.BlacklistType(bType), value)
}

// GetUserBehavior 获取用户行为数据。
func (q *RiskSecurityQueryService) GetUserBehavior(ctx context.Context, userID uint64) (*domain.UserBehavior, error) {
	return q.repo.GetUserBehavior(ctx, userID)
}
