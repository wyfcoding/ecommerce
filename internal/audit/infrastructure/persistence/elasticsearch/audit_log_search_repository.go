// 生成摘要：实现审计日志搜索仓储（Elasticsearch）。
package elasticsearch

import (
	"context"
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/audit/domain"
	"github.com/wyfcoding/pkg/search"
)

type auditLogSearchRepository struct {
	client *search.Client
	index  string
}

// NewAuditLogSearchRepository 创建审计日志搜索仓储实现。
func NewAuditLogSearchRepository(client *search.Client, index string) domain.AuditLogSearchRepository {
	if index == "" {
		index = "audit_logs"
	}
	return &auditLogSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *auditLogSearchRepository) Index(ctx context.Context, log *domain.AuditLog) error {
	if log == nil {
		return nil
	}
	docID := fmt.Sprintf("%d", log.ID)
	return r.client.Index(ctx, r.index, docID, log)
}

func (r *auditLogSearchRepository) Delete(ctx context.Context, logID uint64) error {
	docID := fmt.Sprintf("%d", logID)
	return r.client.Delete(ctx, r.index, docID)
}

func (r *auditLogSearchRepository) Search(ctx context.Context, query *domain.AuditLogQuery, offset, limit int) ([]*domain.AuditLog, int64, error) {
	filters := make([]any, 0)

	if query != nil {
		if query.UserID > 0 {
			filters = append(filters, map[string]any{"term": map[string]any{"user_id": query.UserID}})
		}
		if query.EventType != "" {
			filters = append(filters, map[string]any{"term": map[string]any{"event_type": query.EventType}})
		}
		if query.Module != "" {
			filters = append(filters, map[string]any{"term": map[string]any{"module": query.Module}})
		}
		if query.ResourceType != "" {
			filters = append(filters, map[string]any{"term": map[string]any{"resource_type": query.ResourceType}})
		}
		if !query.StartTime.IsZero() || !query.EndTime.IsZero() {
			rangeQuery := map[string]any{}
			if !query.StartTime.IsZero() {
				rangeQuery["gte"] = query.StartTime
			}
			if !query.EndTime.IsZero() {
				rangeQuery["lte"] = query.EndTime
			}
			filters = append(filters, map[string]any{"range": map[string]any{"timestamp": rangeQuery}})
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
				Source domain.AuditLog `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, queryBody, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	items := make([]*domain.AuditLog, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		log := hit.Source
		items[i] = &log
	}
	return items, searchRes.Hits.Total.Value, nil
}
