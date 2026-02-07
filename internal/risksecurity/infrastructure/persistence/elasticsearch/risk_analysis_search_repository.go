// 生成摘要：实现风险分析搜索仓储（Elasticsearch）。
package elasticsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/risksecurity/domain"
	"github.com/wyfcoding/pkg/search"
)

type riskAnalysisSearchRepository struct {
	client *search.Client
	index  string
}

// NewRiskAnalysisSearchRepository 创建风险分析搜索仓储实现。
func NewRiskAnalysisSearchRepository(client *search.Client, index string) domain.RiskAnalysisSearchRepository {
	if index == "" {
		index = "risk_analysis_results"
	}
	return &riskAnalysisSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *riskAnalysisSearchRepository) Index(ctx context.Context, result *domain.RiskAnalysisResult) error {
	if result == nil {
		return nil
	}
	return r.client.Index(ctx, r.index, r.documentID(uint64(result.ID)), result)
}

func (r *riskAnalysisSearchRepository) Delete(ctx context.Context, resultID uint64) error {
	if resultID == 0 {
		return nil
	}
	return r.client.Delete(ctx, r.index, r.documentID(resultID))
}

func (r *riskAnalysisSearchRepository) Search(ctx context.Context, query *domain.RiskAnalysisQuery, offset, limit int) ([]*domain.RiskAnalysisResult, int64, error) {
	esQuery := map[string]any{
		"from":             offset,
		"size":             limit,
		"track_total_hits": true,
		"sort": []any{
			map[string]any{"created_at": map[string]any{"order": "desc"}},
		},
	}

	filters := make([]any, 0, 6)
	if query != nil {
		if query.UserID > 0 {
			filters = append(filters, map[string]any{"term": map[string]any{"user_id": query.UserID}})
		}
		if query.Level != nil {
			filters = append(filters, map[string]any{"term": map[string]any{"risk_level": *query.Level}})
		}
		if !query.StartTime.IsZero() || !query.EndTime.IsZero() {
			rangeQuery := map[string]any{}
			if !query.StartTime.IsZero() {
				rangeQuery["gte"] = query.StartTime
			}
			if !query.EndTime.IsZero() {
				rangeQuery["lte"] = query.EndTime
			}
			filters = append(filters, map[string]any{"range": map[string]any{"created_at": rangeQuery}})
		}
	}

	if len(filters) == 0 {
		esQuery["query"] = map[string]any{"match_all": map[string]any{}}
	} else {
		esQuery["query"] = map[string]any{"bool": map[string]any{"filter": filters}}
	}

	var searchRes struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source domain.RiskAnalysisResult `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, esQuery, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	items := make([]*domain.RiskAnalysisResult, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		result := hit.Source
		items[i] = &result
	}
	return items, searchRes.Hits.Total.Value, nil
}

func (r *riskAnalysisSearchRepository) documentID(id uint64) string {
	return strings.TrimSpace(fmt.Sprintf("%d", id))
}
