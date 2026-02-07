package domain

import "context"

// RiskAnalysisReadRepository 定义风险分析读模型仓储接口（Redis）。
type RiskAnalysisReadRepository interface {
	SaveLatest(ctx context.Context, userID uint64, result *RiskAnalysisResult) error
	GetLatestByUser(ctx context.Context, userID uint64) (*RiskAnalysisResult, error)
	DeleteLatestByUser(ctx context.Context, userID uint64) error
}
