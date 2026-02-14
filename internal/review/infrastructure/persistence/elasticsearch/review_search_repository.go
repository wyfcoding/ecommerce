// 生成摘要：实现评论搜索仓储（Elasticsearch），支持分页与条件过滤。
// 假设：索引字段与 domain.Review 的 JSON 映射一致，created_at 可用于排序。
package elasticsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/review/domain"
	"github.com/wyfcoding/pkg/search"
)

type reviewSearchRepository struct {
	client *search.Client
	index  string
}

// NewReviewSearchRepository 创建评论搜索仓储实现。
func NewReviewSearchRepository(client *search.Client, index string) domain.ReviewSearchRepository {
	if index == "" {
		index = "reviews"
	}
	return &reviewSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *reviewSearchRepository) Index(ctx context.Context, review *domain.Review) error {
	if review == nil {
		return nil
	}
	docID := fmt.Sprintf("%d", review.ID)
	return r.client.Index(ctx, r.index, docID, review)
}

func (r *reviewSearchRepository) Delete(ctx context.Context, reviewID uint64) error {
	docID := fmt.Sprintf("%d", reviewID)
	return r.client.Delete(ctx, r.index, docID)
}

func (r *reviewSearchRepository) Search(ctx context.Context, productID *uint64, userID *uint64, status *domain.ReviewStatus, offset, limit int, sortBy string) ([]*domain.Review, int64, error) {
	filters := make([]any, 0)
	if productID != nil {
		filters = append(filters, map[string]any{"term": map[string]any{"product_id": *productID}})
	}
	if userID != nil {
		filters = append(filters, map[string]any{"term": map[string]any{"user_id": *userID}})
	}
	if status != nil {
		filters = append(filters, map[string]any{"term": map[string]any{"status": *status}})
	}
	if limit <= 0 {
		limit = 10
	}

	sortField, desc := parseReviewSort(sortBy)
	orderDir := "desc"
	if !desc {
		orderDir = "asc"
	}

	query := map[string]any{
		"from":             offset,
		"size":             limit,
		"track_total_hits": true,
		"sort": []any{
			map[string]any{sortField: map[string]any{"order": orderDir}},
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
				Source domain.Review `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, query, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	reviews := make([]*domain.Review, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		review := hit.Source
		reviews[i] = &review
	}
	return reviews, searchRes.Hits.Total.Value, nil
}

func parseReviewSort(sortBy string) (string, bool) {
	allowed := map[string]string{
		"created_at": "created_at",
		"updated_at": "updated_at",
		"rating":     "rating",
		"like_count": "like_count",
	}

	sortBy = strings.TrimSpace(strings.ToLower(sortBy))
	if sortBy == "" {
		return "created_at", true
	}

	desc := true
	if after, ok := strings.CutPrefix(sortBy, "-"); ok {
		sortBy = after
		desc = true
	}

	parts := strings.Fields(sortBy)
	if len(parts) > 0 {
		sortBy = parts[0]
	}
	if len(parts) > 1 {
		switch parts[1] {
		case "asc":
			desc = false
		case "desc":
			desc = true
		}
	}

	if col, ok := allowed[sortBy]; ok {
		return col, desc
	}
	return "created_at", true
}
