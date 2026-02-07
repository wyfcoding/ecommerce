// 生成摘要：实现积分商品搜索仓储（Elasticsearch）。
package elasticsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/pointsmall/domain"
	"github.com/wyfcoding/pkg/search"
)

type productSearchRepository struct {
	client *search.Client
	index  string
}

// NewProductSearchRepository 创建积分商品搜索仓储实现。
func NewProductSearchRepository(client *search.Client, index string) domain.PointsProductSearchRepository {
	if index == "" {
		index = "points_products"
	}
	return &productSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *productSearchRepository) Index(ctx context.Context, product *domain.PointsProduct) error {
	if product == nil {
		return nil
	}
	return r.client.Index(ctx, r.index, r.documentID(uint64(product.ID)), product)
}

func (r *productSearchRepository) Delete(ctx context.Context, productID uint64) error {
	if productID == 0 {
		return nil
	}
	return r.client.Delete(ctx, r.index, r.documentID(productID))
}

func (r *productSearchRepository) Search(ctx context.Context, query *domain.PointsProductQuery, offset, limit int) ([]*domain.PointsProduct, int64, error) {
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
	}

	var must []any
	if query != nil && query.Keyword != "" {
		must = append(must, map[string]any{
			"multi_match": map[string]any{
				"query":  query.Keyword,
				"fields": []string{"name", "description"},
			},
		})
	}

	if len(filters) == 0 && len(must) == 0 {
		esQuery["query"] = map[string]any{"match_all": map[string]any{}}
	} else {
		boolQuery := map[string]any{}
		if len(filters) > 0 {
			boolQuery["filter"] = filters
		}
		if len(must) > 0 {
			boolQuery["must"] = must
		}
		esQuery["query"] = map[string]any{"bool": boolQuery}
	}

	var searchRes struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source domain.PointsProduct `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, esQuery, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	items := make([]*domain.PointsProduct, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		item := hit.Source
		items[i] = &item
	}
	return items, searchRes.Hits.Total.Value, nil
}

func (r *productSearchRepository) documentID(id uint64) string {
	return strings.TrimSpace(fmt.Sprintf("%d", id))
}
