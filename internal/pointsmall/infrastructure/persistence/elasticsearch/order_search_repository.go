// 生成摘要：实现积分订单搜索仓储（Elasticsearch）。
package elasticsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/pointsmall/domain"
	"github.com/wyfcoding/pkg/search"
)

type orderSearchRepository struct {
	client *search.Client
	index  string
}

// NewOrderSearchRepository 创建积分订单搜索仓储实现。
func NewOrderSearchRepository(client *search.Client, index string) domain.PointsOrderSearchRepository {
	if index == "" {
		index = "points_orders"
	}
	return &orderSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *orderSearchRepository) Index(ctx context.Context, order *domain.PointsOrder) error {
	if order == nil {
		return nil
	}
	return r.client.Index(ctx, r.index, r.documentID(uint64(order.ID)), order)
}

func (r *orderSearchRepository) Delete(ctx context.Context, orderID uint64) error {
	if orderID == 0 {
		return nil
	}
	return r.client.Delete(ctx, r.index, r.documentID(orderID))
}

func (r *orderSearchRepository) Search(ctx context.Context, query *domain.PointsOrderQuery, offset, limit int) ([]*domain.PointsOrder, int64, error) {
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
		if query.Status != nil {
			filters = append(filters, map[string]any{"term": map[string]any{"status": *query.Status}})
		}
		if query.OrderNo != "" {
			filters = append(filters, map[string]any{"term": map[string]any{"order_no": query.OrderNo}})
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
				Source domain.PointsOrder `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, esQuery, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	items := make([]*domain.PointsOrder, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		item := hit.Source
		items[i] = &item
	}
	return items, searchRes.Hits.Total.Value, nil
}

func (r *orderSearchRepository) documentID(id uint64) string {
	return strings.TrimSpace(fmt.Sprintf("%d", id))
}
