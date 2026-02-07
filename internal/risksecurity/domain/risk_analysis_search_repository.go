package domain

import "context"

// RiskAnalysisSearchRepository 定义风险分析搜索仓储接口（Elasticsearch）。
type RiskAnalysisSearchRepository interface {
	Index(ctx context.Context, result *RiskAnalysisResult) error
	Delete(ctx context.Context, resultID uint64) error
	Search(ctx context.Context, query *RiskAnalysisQuery, offset, limit int) ([]*RiskAnalysisResult, int64, error)
}
