// 生成摘要：实现审计日志搜索仓储（Elasticsearch）。
package elasticsearch

import (
	"context"
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/admin/domain"
	"github.com/wyfcoding/pkg/search"
)

type auditLogSearchRepository struct {
	client *search.Client
	index  string
}

// NewAuditLogSearchRepository 创建审计日志搜索仓储实现。
func NewAuditLogSearchRepository(client *search.Client, index string) domain.AuditLogSearchRepository {
	if index == "" {
		index = "admin_audit_logs"
	}
	return &auditLogSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *auditLogSearchRepository) Index(ctx context.Context, logEntry *domain.AuditLog) error {
	if logEntry == nil {
		return nil
	}
	docID := fmt.Sprintf("%d", logEntry.ID)
	return r.client.Index(ctx, r.index, docID, logEntry)
}

func (r *auditLogSearchRepository) Delete(ctx context.Context, id uint) error {
	docID := fmt.Sprintf("%d", id)
	return r.client.Delete(ctx, r.index, docID)
}

func (r *auditLogSearchRepository) Search(ctx context.Context, userID *uint, action, resource *string, offset, limit int) ([]*domain.AuditLog, int64, error) {
	filters := make([]any, 0)
	if userID != nil {
		filters = append(filters, map[string]any{"term": map[string]any{"user_id": *userID}})
	}
	if action != nil && *action != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"action": *action}})
	}
	if resource != nil && *resource != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"resource": *resource}})
	}

	query := map[string]any{
		"from":             offset,
		"size":             limit,
		"track_total_hits": true,
		"sort": []any{
			map[string]any{"created_at": map[string]any{"order": "desc"}},
		},
	}

	if len(filters) == 0 {
		query["query"] = map[string]any{"match_all": map[string]any{}}
	} else {
		query["query"] = map[string]any{
			"bool": map[string]any{
				"filter": filters,
			},
		}
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

	if err := r.client.Search(ctx, r.index, query, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	items := make([]*domain.AuditLog, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		logEntry := hit.Source
		items[i] = &logEntry
	}

	return items, searchRes.Hits.Total.Value, nil
}
