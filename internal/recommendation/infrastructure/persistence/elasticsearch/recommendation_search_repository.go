// 生成摘要：实现推荐搜索仓储（Elasticsearch），支持分页与条件过滤。
// 假设：索引字段与 domain.Recommendation 的 JSON 映射一致，score 可用于排序。
package elasticsearch

import (
	"context"
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/recommendation/domain"
	"github.com/wyfcoding/pkg/search"
)

type recommendationSearchRepository struct {
	client *search.Client
	index  string
}

// NewRecommendationSearchRepository 创建推荐搜索仓储实现。
func NewRecommendationSearchRepository(client *search.Client, index string) domain.RecommendationSearchRepository {
	if index == "" {
		index = "recommendations"
	}
	return &recommendationSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *recommendationSearchRepository) Index(ctx context.Context, rec *domain.Recommendation) error {
	if rec == nil {
		return nil
	}
	docID := r.documentID(rec.UserID, rec.RecommendationType, rec.ProductID)
	return r.client.Index(ctx, r.index, docID, rec)
}

func (r *recommendationSearchRepository) Delete(ctx context.Context, documentID string) error {
	if documentID == "" {
		return nil
	}
	return r.client.Delete(ctx, r.index, documentID)
}

func (r *recommendationSearchRepository) DeleteByUserAndType(ctx context.Context, userID uint64, recType *domain.RecommendationType) error {
	query := map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"filter": r.buildFilters(userID, recType),
			},
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

func (r *recommendationSearchRepository) Search(ctx context.Context, userID uint64, recType *domain.RecommendationType, offset, limit int) ([]*domain.Recommendation, int64, error) {
	filters := r.buildFilters(userID, recType)

	if limit <= 0 {
		limit = 10
	}

	query := map[string]any{
		"from":             offset,
		"size":             limit,
		"track_total_hits": true,
		"sort": []any{
			map[string]any{"score": map[string]any{"order": "desc"}},
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
				Source domain.Recommendation `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, query, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	results := make([]*domain.Recommendation, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		rec := hit.Source
		results[i] = &rec
	}

	return results, searchRes.Hits.Total.Value, nil
}

func (r *recommendationSearchRepository) buildFilters(userID uint64, recType *domain.RecommendationType) []any {
	filters := make([]any, 0, 2)
	if userID > 0 {
		filters = append(filters, map[string]any{
			"term": map[string]any{"user_id": userID},
		})
	}
	if recType != nil && *recType != "" {
		filters = append(filters, map[string]any{
			"term": map[string]any{"recommendation_type": *recType},
		})
	}
	return filters
}

func (r *recommendationSearchRepository) documentID(userID uint64, recType domain.RecommendationType, productID uint64) string {
	return fmt.Sprintf("%d:%s:%d", userID, recType, productID)
}
