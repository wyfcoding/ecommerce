// 生成摘要：实现指标搜索仓储（Elasticsearch）。
package elasticsearch

import (
	"context"
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/analytics/domain"
	"github.com/wyfcoding/pkg/search"
)

type metricSearchRepository struct {
	client *search.Client
	index  string
}

// NewMetricSearchRepository 创建指标搜索仓储实现。
func NewMetricSearchRepository(client *search.Client, index string) domain.MetricSearchRepository {
	if index == "" {
		index = "analytics_metrics"
	}
	return &metricSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *metricSearchRepository) Index(ctx context.Context, metric *domain.Metric) error {
	if metric == nil {
		return nil
	}
	docID := fmt.Sprintf("%d", metric.ID)
	return r.client.Index(ctx, r.index, docID, metric)
}

func (r *metricSearchRepository) Delete(ctx context.Context, metricID uint64) error {
	docID := fmt.Sprintf("%d", metricID)
	return r.client.Delete(ctx, r.index, docID)
}

func (r *metricSearchRepository) Search(ctx context.Context, query *domain.MetricQuery, offset, limit int) ([]*domain.Metric, int64, error) {
	filters := make([]any, 0)

	if query != nil {
		if query.MetricType != "" {
			filters = append(filters, map[string]any{
				"term": map[string]any{"metric_type": query.MetricType},
			})
		}
		if query.Granularity != "" {
			filters = append(filters, map[string]any{
				"term": map[string]any{"granularity": query.Granularity},
			})
		}
		if query.Dimension != "" {
			filters = append(filters, map[string]any{
				"term": map[string]any{"dimension": query.Dimension},
			})
		}
		if query.DimensionVal != "" {
			filters = append(filters, map[string]any{
				"term": map[string]any{"dimension_val": query.DimensionVal},
			})
		}
		if !query.StartTime.IsZero() || !query.EndTime.IsZero() {
			rangeQuery := map[string]any{}
			if !query.StartTime.IsZero() {
				rangeQuery["gte"] = query.StartTime
			}
			if !query.EndTime.IsZero() {
				rangeQuery["lte"] = query.EndTime
			}
			filters = append(filters, map[string]any{
				"range": map[string]any{"timestamp": rangeQuery},
			})
		}
	}

	queryBody := map[string]any{
		"from":             offset,
		"size":             limit,
		"track_total_hits": true,
		"sort": []any{
			map[string]any{"timestamp": map[string]any{"order": "desc"}},
		},
	}

	if len(filters) == 0 {
		queryBody["query"] = map[string]any{"match_all": map[string]any{}}
	} else {
		queryBody["query"] = map[string]any{"bool": map[string]any{"filter": filters}}
	}

	var searchRes struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source domain.Metric `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, queryBody, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	items := make([]*domain.Metric, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		metric := hit.Source
		items[i] = &metric
	}
	return items, searchRes.Hits.Total.Value, nil
}
