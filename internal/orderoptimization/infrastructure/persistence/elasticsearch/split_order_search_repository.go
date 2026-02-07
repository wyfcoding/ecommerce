// 生成摘要：实现拆分订单搜索仓储（Elasticsearch）。
package elasticsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/orderoptimization/domain"
	"github.com/wyfcoding/pkg/search"
)

type splitOrderSearchRepository struct {
	client *search.Client
	index  string
}

// NewSplitOrderSearchRepository 创建拆分订单搜索仓储实现。
func NewSplitOrderSearchRepository(client *search.Client, index string) domain.SplitOrderSearchRepository {
	if index == "" {
		index = "order_optimization_splits"
	}
	return &splitOrderSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *splitOrderSearchRepository) Index(ctx context.Context, order *domain.SplitOrder) error {
	if order == nil {
		return nil
	}
	return r.client.Index(ctx, r.index, r.documentID(uint64(order.ID)), order)
}

func (r *splitOrderSearchRepository) Delete(ctx context.Context, splitOrderID uint64) error {
	if splitOrderID == 0 {
		return nil
	}
	return r.client.Delete(ctx, r.index, r.documentID(splitOrderID))
}

func (r *splitOrderSearchRepository) SearchByOriginalOrderID(ctx context.Context, originalOrderID uint64, offset, limit int) ([]*domain.SplitOrder, int64, error) {
	esQuery := map[string]any{
		"from":             offset,
		"size":             limit,
		"track_total_hits": true,
		"sort": []any{
			map[string]any{"created_at": map[string]any{"order": "asc"}},
		},
		"query": map[string]any{
			"term": map[string]any{"original_order_id": originalOrderID},
		},
	}

	var searchRes struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source domain.SplitOrder `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, esQuery, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	items := make([]*domain.SplitOrder, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		item := hit.Source
		items[i] = &item
	}
	return items, searchRes.Hits.Total.Value, nil
}

func (r *splitOrderSearchRepository) documentID(id uint64) string {
	return strings.TrimSpace(fmt.Sprintf("%d", id))
}
