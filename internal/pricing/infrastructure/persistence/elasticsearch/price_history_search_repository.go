// 生成摘要：实现价格历史搜索仓储（Elasticsearch）。
package elasticsearch

import (
	"context"
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/pricing/domain"
	"github.com/wyfcoding/pkg/search"
)

type priceHistorySearchRepository struct {
	client *search.Client
	index  string
}

// NewPriceHistorySearchRepository 创建价格历史搜索仓储实现。
func NewPriceHistorySearchRepository(client *search.Client, index string) domain.PriceHistorySearchRepository {
	if index == "" {
		index = "pricing_histories"
	}
	return &priceHistorySearchRepository{
		client: client,
		index:  index,
	}
}

func (r *priceHistorySearchRepository) Index(ctx context.Context, history *domain.PriceHistory) error {
	if history == nil {
		return nil
	}
	docID := fmt.Sprintf("%d", history.ID)
	return r.client.Index(ctx, r.index, docID, history)
}

func (r *priceHistorySearchRepository) Delete(ctx context.Context, id uint64) error {
	docID := fmt.Sprintf("%d", id)
	return r.client.Delete(ctx, r.index, docID)
}

func (r *priceHistorySearchRepository) Search(ctx context.Context, productID, skuID uint64, offset, limit int) ([]*domain.PriceHistory, int64, error) {
	filters := make([]any, 0)
	if productID > 0 {
		filters = append(filters, map[string]any{
			"term": map[string]any{"product_id": productID},
		})
	}
	if skuID > 0 {
		filters = append(filters, map[string]any{
			"term": map[string]any{"sku_id": skuID},
		})
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
				Source domain.PriceHistory `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, query, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	items := make([]*domain.PriceHistory, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		h := hit.Source
		items[i] = &h
	}

	return items, searchRes.Hits.Total.Value, nil
}
