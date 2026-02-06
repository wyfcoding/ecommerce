// 生成摘要：实现收藏夹搜索仓储（Elasticsearch），支持分页与条件过滤。
// 假设：索引字段与 domain.Wishlist 的 JSON 映射一致，created_at 可用于排序。
package elasticsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/wishlist/domain"
	"github.com/wyfcoding/pkg/search"
)

type wishlistSearchRepository struct {
	client *search.Client
	index  string
}

// NewWishlistSearchRepository 创建收藏夹搜索仓储实现。
func NewWishlistSearchRepository(client *search.Client, index string) domain.WishlistSearchRepository {
	if index == "" {
		index = "wishlists"
	}
	return &wishlistSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *wishlistSearchRepository) Index(ctx context.Context, item *domain.Wishlist) error {
	if item == nil {
		return nil
	}
	docID := r.documentID(item.UserID, item.SkuID)
	return r.client.Index(ctx, r.index, docID, item)
}

func (r *wishlistSearchRepository) Delete(ctx context.Context, documentID string) error {
	if documentID == "" {
		return nil
	}
	return r.client.Delete(ctx, r.index, documentID)
}

func (r *wishlistSearchRepository) DeleteByUser(ctx context.Context, userID uint64) error {
	query := map[string]any{
		"query": map[string]any{
			"term": map[string]any{"user_id": userID},
		},
		"size": 1000,
	}

	var searchRes struct {
		Hits struct {
			Hits []struct {
				ID string `json:"_id"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, query, &searchRes); err != nil {
		return fmt.Errorf("es search failed: %w", err)
	}

	for _, hit := range searchRes.Hits.Hits {
		if hit.ID == "" {
			continue
		}
		if err := r.client.Delete(ctx, r.index, hit.ID); err != nil {
			return err
		}
	}
	return nil
}

func (r *wishlistSearchRepository) Search(ctx context.Context, userID uint64, offset, limit int) ([]*domain.Wishlist, int64, error) {
	if limit <= 0 {
		limit = 10
	}

	query := map[string]any{
		"from":             offset,
		"size":             limit,
		"track_total_hits": true,
		"sort": []any{
			map[string]any{parseWishlistSort(""): map[string]any{"order": "desc"}},
		},
		"query": map[string]any{
			"term": map[string]any{"user_id": userID},
		},
	}

	var searchRes struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source domain.Wishlist `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, query, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	results := make([]*domain.Wishlist, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		item := hit.Source
		results[i] = &item
	}

	return results, searchRes.Hits.Total.Value, nil
}

func (r *wishlistSearchRepository) documentID(userID, skuID uint64) string {
	return fmt.Sprintf("%d:%d", userID, skuID)
}

func parseWishlistSort(sortBy string) string {
	allowed := map[string]string{
		"created_at": "created_at",
		"updated_at": "updated_at",
	}
	sortBy = strings.TrimSpace(strings.ToLower(sortBy))
	if col, ok := allowed[sortBy]; ok {
		return col
	}
	return "created_at"
}
