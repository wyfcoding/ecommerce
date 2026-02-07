// 生成摘要：实现拼团活动搜索仓储（Elasticsearch）。
package elasticsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/groupbuy/domain"
	"github.com/wyfcoding/pkg/search"
)

type groupbuySearchRepository struct {
	client *search.Client
	index  string
}

// NewGroupbuySearchRepository 创建拼团活动搜索仓储实现。
func NewGroupbuySearchRepository(client *search.Client, index string) domain.GroupbuySearchRepository {
	if index == "" {
		index = "groupbuys"
	}
	return &groupbuySearchRepository{
		client: client,
		index:  index,
	}
}

func (r *groupbuySearchRepository) Index(ctx context.Context, groupbuy *domain.Groupbuy) error {
	if groupbuy == nil {
		return nil
	}
	return r.client.Index(ctx, r.index, r.documentID(uint64(groupbuy.ID)), groupbuy)
}

func (r *groupbuySearchRepository) Delete(ctx context.Context, groupbuyID uint64) error {
	if groupbuyID == 0 {
		return nil
	}
	return r.client.Delete(ctx, r.index, r.documentID(groupbuyID))
}

func (r *groupbuySearchRepository) Search(ctx context.Context, query *domain.GroupbuyQuery, offset, limit int) ([]*domain.Groupbuy, int64, error) {
	esQuery := map[string]any{
		"from":             offset,
		"size":             limit,
		"track_total_hits": true,
		"sort": []any{
			map[string]any{"created_at": map[string]any{"order": "desc"}},
		},
	}

	filters := make([]any, 0, 4)
	if query != nil {
		if query.Status != nil {
			filters = append(filters, map[string]any{"term": map[string]any{"status": *query.Status}})
		}
		if query.ProductID > 0 {
			filters = append(filters, map[string]any{"term": map[string]any{"product_id": query.ProductID}})
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
				Source domain.Groupbuy `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, esQuery, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	items := make([]*domain.Groupbuy, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		item := hit.Source
		items[i] = &item
	}
	return items, searchRes.Hits.Total.Value, nil
}

func (r *groupbuySearchRepository) documentID(id uint64) string {
	return strings.TrimSpace(fmt.Sprintf("%d", id))
}
