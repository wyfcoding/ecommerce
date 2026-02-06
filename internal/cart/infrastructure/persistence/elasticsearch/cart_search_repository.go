// 生成摘要：实现购物车搜索仓储（Elasticsearch），支持按用户检索。
// 假设：索引字段与 domain.Cart 的 JSON 映射一致。
package elasticsearch

import (
	"context"
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/cart/domain"
	"github.com/wyfcoding/pkg/search"
)

type cartSearchRepository struct {
	client *search.Client
	index  string
}

// NewCartSearchRepository 创建购物车搜索仓储实现。
func NewCartSearchRepository(client *search.Client, index string) domain.CartSearchRepository {
	if index == "" {
		index = "carts"
	}
	return &cartSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *cartSearchRepository) Index(ctx context.Context, cart *domain.Cart) error {
	if cart == nil {
		return nil
	}
	return r.client.Index(ctx, r.index, r.documentID(cart.UserID), cart)
}

func (r *cartSearchRepository) Delete(ctx context.Context, userID uint64) error {
	if userID == 0 {
		return nil
	}
	return r.client.Delete(ctx, r.index, r.documentID(userID))
}

func (r *cartSearchRepository) Search(ctx context.Context, userID uint64) (*domain.Cart, error) {
	if userID == 0 {
		return nil, nil
	}
	query := map[string]any{
		"size": 1,
		"query": map[string]any{
			"term": map[string]any{"user_id": userID},
		},
	}

	var searchRes struct {
		Hits struct {
			Hits []struct {
				Source domain.Cart `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, query, &searchRes); err != nil {
		return nil, fmt.Errorf("es search failed: %w", err)
	}
	if len(searchRes.Hits.Hits) == 0 {
		return nil, nil
	}
	item := searchRes.Hits.Hits[0].Source
	return &item, nil
}

func (r *cartSearchRepository) documentID(userID uint64) string {
	return fmt.Sprintf("%d", userID)
}
