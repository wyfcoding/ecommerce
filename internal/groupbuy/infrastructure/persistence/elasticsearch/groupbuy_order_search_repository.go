// 生成摘要：实现拼团订单搜索仓储（Elasticsearch）。
package elasticsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/groupbuy/domain"
	"github.com/wyfcoding/pkg/search"
)

type groupbuyOrderSearchRepository struct {
	client *search.Client
	index  string
}

// NewGroupbuyOrderSearchRepository 创建拼团订单搜索仓储实现。
func NewGroupbuyOrderSearchRepository(client *search.Client, index string) domain.GroupbuyOrderSearchRepository {
	if index == "" {
		index = "groupbuy_orders"
	}
	return &groupbuyOrderSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *groupbuyOrderSearchRepository) Index(ctx context.Context, order *domain.GroupbuyOrder) error {
	if order == nil {
		return nil
	}
	return r.client.Index(ctx, r.index, r.documentID(uint64(order.ID)), order)
}

func (r *groupbuyOrderSearchRepository) Delete(ctx context.Context, orderID uint64) error {
	if orderID == 0 {
		return nil
	}
	return r.client.Delete(ctx, r.index, r.documentID(orderID))
}

func (r *groupbuyOrderSearchRepository) Search(ctx context.Context, query *domain.GroupbuyOrderQuery, offset, limit int) ([]*domain.GroupbuyOrder, int64, error) {
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
		if query.UserID > 0 {
			filters = append(filters, map[string]any{"term": map[string]any{"user_id": query.UserID}})
		}
		if query.TeamID > 0 {
			filters = append(filters, map[string]any{"term": map[string]any{"team_id": query.TeamID}})
		}
		if query.Status != nil {
			filters = append(filters, map[string]any{"term": map[string]any{"status": *query.Status}})
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
				Source domain.GroupbuyOrder `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, esQuery, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	items := make([]*domain.GroupbuyOrder, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		item := hit.Source
		items[i] = &item
	}
	return items, searchRes.Hits.Total.Value, nil
}

func (r *groupbuyOrderSearchRepository) documentID(id uint64) string {
	return strings.TrimSpace(fmt.Sprintf("%d", id))
}
