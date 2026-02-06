// 生成摘要：实现AI模型搜索仓储（Elasticsearch）。
package elasticsearch

import (
	"context"
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/aimodel/domain"
	"github.com/wyfcoding/pkg/search"
)

type aiModelSearchRepository struct {
	client *search.Client
	index  string
}

// NewAIModelSearchRepository 创建AI模型搜索仓储实现。
func NewAIModelSearchRepository(client *search.Client, index string) domain.AIModelSearchRepository {
	if index == "" {
		index = "aimodel_models"
	}
	return &aiModelSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *aiModelSearchRepository) Index(ctx context.Context, model *domain.AIModel) error {
	if model == nil {
		return nil
	}
	docID := fmt.Sprintf("%d", model.ID)
	return r.client.Index(ctx, r.index, docID, model)
}

func (r *aiModelSearchRepository) Delete(ctx context.Context, modelID uint64) error {
	docID := fmt.Sprintf("%d", modelID)
	return r.client.Delete(ctx, r.index, docID)
}

func (r *aiModelSearchRepository) Search(ctx context.Context, status *domain.ModelStatus, modelType, algorithm string, creatorID *uint64, offset, limit int) ([]*domain.AIModel, int64, error) {
	filters := make([]any, 0)

	if status != nil {
		filters = append(filters, map[string]any{
			"term": map[string]any{"status": *status},
		})
	}
	if modelType != "" {
		filters = append(filters, map[string]any{
			"term": map[string]any{"type": modelType},
		})
	}
	if algorithm != "" {
		filters = append(filters, map[string]any{
			"term": map[string]any{"algorithm": algorithm},
		})
	}
	if creatorID != nil {
		filters = append(filters, map[string]any{
			"term": map[string]any{"creator_id": *creatorID},
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
		query["query"] = map[string]any{"bool": map[string]any{"filter": filters}}
	}

	var searchRes struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source domain.AIModel `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, query, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	items := make([]*domain.AIModel, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		model := hit.Source
		items[i] = &model
	}
	return items, searchRes.Hits.Total.Value, nil
}
