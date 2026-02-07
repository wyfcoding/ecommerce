// 生成摘要：实现内容审核记录搜索仓储（Elasticsearch）。
package elasticsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/contentmoderation/domain"
	"github.com/wyfcoding/pkg/search"
)

type moderationRecordSearchRepository struct {
	client *search.Client
	index  string
}

// NewModerationRecordSearchRepository 创建审核记录搜索仓储实现。
func NewModerationRecordSearchRepository(client *search.Client, index string) domain.ModerationRecordSearchRepository {
	if index == "" {
		index = "moderation_records"
	}
	return &moderationRecordSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *moderationRecordSearchRepository) Index(ctx context.Context, record *domain.ModerationRecord) error {
	if record == nil {
		return nil
	}
	return r.client.Index(ctx, r.index, r.documentID(uint64(record.ID)), record)
}

func (r *moderationRecordSearchRepository) Delete(ctx context.Context, recordID uint64) error {
	if recordID == 0 {
		return nil
	}
	return r.client.Delete(ctx, r.index, r.documentID(recordID))
}

func (r *moderationRecordSearchRepository) Search(ctx context.Context, query *domain.ModerationRecordQuery, offset, limit int) ([]*domain.ModerationRecord, int64, error) {
	esQuery := map[string]any{
		"from":             offset,
		"size":             limit,
		"track_total_hits": true,
		"sort": []any{
			map[string]any{"created_at": map[string]any{"order": "desc"}},
		},
	}

	filters := make([]any, 0, 8)
	if query != nil {
		if query.UserID > 0 {
			filters = append(filters, map[string]any{"term": map[string]any{"user_id": query.UserID}})
		}
		if query.ContentType != "" {
			filters = append(filters, map[string]any{"term": map[string]any{"content_type": query.ContentType}})
		}
		if query.ContentID > 0 {
			filters = append(filters, map[string]any{"term": map[string]any{"content_id": query.ContentID}})
		}
		if query.Status != nil {
			filters = append(filters, map[string]any{"term": map[string]any{"status": *query.Status}})
		}
		if query.ModeratorID > 0 {
			filters = append(filters, map[string]any{"term": map[string]any{"moderator_id": query.ModeratorID}})
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
		esQuery["query"] = map[string]any{
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
				Source domain.ModerationRecord `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, esQuery, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	items := make([]*domain.ModerationRecord, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		record := hit.Source
		items[i] = &record
	}
	return items, searchRes.Hits.Total.Value, nil
}

func (r *moderationRecordSearchRepository) documentID(id uint64) string {
	return strings.TrimSpace(fmt.Sprintf("%d", id))
}
