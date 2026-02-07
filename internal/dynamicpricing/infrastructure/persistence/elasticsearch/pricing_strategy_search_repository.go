package elasticsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/dynamicpricing/domain"
	"github.com/wyfcoding/pkg/search"
)

type pricingStrategySearchRepository struct {
	client *search.Client
	index  string
}

// NewPricingStrategySearchRepository 创建定价策略搜索仓储实现。
func NewPricingStrategySearchRepository(client *search.Client, index string) domain.PricingStrategySearchRepository {
	if index == "" {
		index = "pricing_strategies"
	}
	return &pricingStrategySearchRepository{
		client: client,
		index:  index,
	}
}

func (r *pricingStrategySearchRepository) Index(ctx context.Context, strategy *domain.PricingStrategy) error {
	if strategy == nil {
		return nil
	}
	return r.client.Index(ctx, r.index, r.documentID(strategy.SKUID), strategy)
}

func (r *pricingStrategySearchRepository) Delete(ctx context.Context, skuID uint64) error {
	if skuID == 0 {
		return nil
	}
	return r.client.Delete(ctx, r.index, r.documentID(skuID))
}

func (r *pricingStrategySearchRepository) Search(ctx context.Context, query *domain.PricingStrategyQuery, offset, limit int) ([]*domain.PricingStrategy, int64, error) {
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
		if query.SKUID > 0 {
			filters = append(filters, map[string]any{"term": map[string]any{"sku_id": query.SKUID}})
		}
		if query.Enabled != nil {
			filters = append(filters, map[string]any{"term": map[string]any{"enabled": *query.Enabled}})
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
				Source domain.PricingStrategy `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, esQuery, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	items := make([]*domain.PricingStrategy, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		item := hit.Source
		items[i] = &item
	}
	return items, searchRes.Hits.Total.Value, nil
}

func (r *pricingStrategySearchRepository) documentID(skuID uint64) string {
	return strings.TrimSpace(fmt.Sprintf("%d", skuID))
}
