package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/risksecurity/domain"
)

// RiskSecurityQueryService 处理风控安全的读操作。
type RiskSecurityQueryService struct {
	repo               domain.RiskRepository
	analysisReadRepo   domain.RiskAnalysisReadRepository
	blacklistReadRepo  domain.BlacklistReadRepository
	behaviorReadRepo   domain.UserBehaviorReadRepository
	analysisSearchRepo domain.RiskAnalysisSearchRepository
	logger             *slog.Logger
}

// NewRiskSecurityQueryService creates a new RiskSecurityQueryService instance.
func NewRiskSecurityQueryService(
	repo domain.RiskRepository,
	analysisReadRepo domain.RiskAnalysisReadRepository,
	blacklistReadRepo domain.BlacklistReadRepository,
	behaviorReadRepo domain.UserBehaviorReadRepository,
	analysisSearchRepo domain.RiskAnalysisSearchRepository,
	logger *slog.Logger,
) *RiskSecurityQueryService {
	return &RiskSecurityQueryService{
		repo:               repo,
		analysisReadRepo:   analysisReadRepo,
		blacklistReadRepo:  blacklistReadRepo,
		behaviorReadRepo:   behaviorReadRepo,
		analysisSearchRepo: analysisSearchRepo,
		logger:             logger,
	}
}

// GetRiskAnalysisResult 获取指定用户的最新风险分析结果。
func (q *RiskSecurityQueryService) GetRiskAnalysisResult(ctx context.Context, userID uint64) (*domain.RiskAnalysisResult, error) {
	if q.analysisReadRepo != nil {
		if cached, err := q.analysisReadRepo.GetLatestByUser(ctx, userID); err == nil && cached != nil {
			return cached, nil
		}
	}

	query := &domain.RiskAnalysisQuery{
		UserID:   userID,
		Page:     1,
		PageSize: 1,
	}

	if q.analysisSearchRepo != nil {
		list, _, err := q.analysisSearchRepo.Search(ctx, query, 0, 1)
		if err == nil && len(list) > 0 {
			return list[0], nil
		}
		if err != nil {
			q.logger.WarnContext(ctx, "risk analysis search fallback to mysql", "error", err)
		}
	}

	results, _, err := q.repo.ListAnalysisResults(ctx, query)
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
	if q.blacklistReadRepo != nil {
		if cached, err := q.blacklistReadRepo.GetByTypeValue(ctx, domain.BlacklistType(bType), value); err == nil && cached != nil {
			return cached, nil
		}
	}
	return q.repo.GetBlacklist(ctx, domain.BlacklistType(bType), value)
}

// GetUserBehavior 获取用户行为数据。
func (q *RiskSecurityQueryService) GetUserBehavior(ctx context.Context, userID uint64) (*domain.UserBehavior, error) {
	if q.behaviorReadRepo != nil {
		if cached, err := q.behaviorReadRepo.GetByUserID(ctx, userID); err == nil && cached != nil {
			return cached, nil
		}
	}
	return q.repo.GetUserBehavior(ctx, userID)
}
