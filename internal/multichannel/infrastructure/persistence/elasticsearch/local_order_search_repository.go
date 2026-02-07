// 生成摘要：实现渠道订单搜索仓储（Elasticsearch）。
package elasticsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/multichannel/domain"
	"github.com/wyfcoding/pkg/search"
)

type localOrderSearchRepository struct {
	client *search.Client
	index  string
}

// NewLocalOrderSearchRepository 创建渠道订单搜索仓储实现。
func NewLocalOrderSearchRepository(client *search.Client, index string) domain.LocalOrderSearchRepository {
	if index == "" {
		index = "multichannel_orders"
	}
	return &localOrderSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *localOrderSearchRepository) Index(ctx context.Context, order *domain.LocalOrder) error {
	if order == nil {
		return nil
	}
	return r.client.Index(ctx, r.index, r.documentID(uint64(order.ID)), order)
}

func (r *localOrderSearchRepository) Delete(ctx context.Context, orderID uint64) error {
	if orderID == 0 {
		return nil
	}
	return r.client.Delete(ctx, r.index, r.documentID(orderID))
}

func (r *localOrderSearchRepository) Search(ctx context.Context, query *domain.LocalOrderQuery, offset, limit int) ([]*domain.LocalOrder, int64, error) {
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
		if query.ChannelID > 0 {
			filters = append(filters, map[string]any{"term": map[string]any{"channel_id": query.ChannelID}})
		}
		if query.Status != "" {
			filters = append(filters, map[string]any{"term": map[string]any{"status": query.Status}})
		}
		if query.ChannelOrderID != "" {
			filters = append(filters, map[string]any{"term": map[string]any{"channel_order_id": query.ChannelOrderID}})
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
				Source domain.LocalOrder `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, esQuery, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	items := make([]*domain.LocalOrder, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		item := hit.Source
		items[i] = &item
	}
	return items, searchRes.Hits.Total.Value, nil
}

func (r *localOrderSearchRepository) documentID(id uint64) string {
	return strings.TrimSpace(fmt.Sprintf("%d", id))
}
