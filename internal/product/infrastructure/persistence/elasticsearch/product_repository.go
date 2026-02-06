package elasticsearch

import (
	"context"
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/product/domain"
	"github.com/wyfcoding/pkg/search"
)

type productSearchRepository struct {
	client *search.Client
	index  string
}

// NewProductSearchRepository 创建商品搜索仓储实现。
func NewProductSearchRepository(client *search.Client, index string) domain.ProductSearchRepository {
	if index == "" {
		index = "products"
	}
	return &productSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *productSearchRepository) Index(ctx context.Context, product *domain.Product) error {
	if product == nil {
		return nil
	}
	docID := fmt.Sprintf("%d", product.ID)
	return r.client.Index(ctx, r.index, docID, product)
}

func (r *productSearchRepository) Search(ctx context.Context, query string, limit int) ([]*domain.Product, error) {
	esQuery := map[string]any{
		"query": map[string]any{
			"multi_match": map[string]any{
				"query":  query,
				"fields": []string{"name^3", "description", "category_name"},
			},
		},
		"size": limit,
	}

	var searchRes struct {
		Hits struct {
			Hits []struct {
				Source domain.Product `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, esQuery, &searchRes); err != nil {
		return nil, fmt.Errorf("es search failed: %w", err)
	}

	products := make([]*domain.Product, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		p := hit.Source
		products[i] = &p
	}

	return products, nil
}

func (r *productSearchRepository) Delete(ctx context.Context, productID uint64) error {
	docID := fmt.Sprintf("%d", productID)
	return r.client.Delete(ctx, r.index, docID)
}
