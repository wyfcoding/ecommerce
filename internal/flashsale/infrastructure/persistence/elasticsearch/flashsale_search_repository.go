package elasticsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/flashsale/domain"
	"github.com/wyfcoding/pkg/search"
)

type flashsaleSearchRepository struct {
	client *search.Client
	index  string
}

// NewFlashsaleSearchRepository 创建秒杀活动搜索仓储实现。
func NewFlashsaleSearchRepository(client *search.Client, index string) domain.FlashsaleSearchRepository {
	if index == "" {
		index = "flashsale_events"
	}
	return &flashsaleSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *flashsaleSearchRepository) Index(ctx context.Context, flashsale *domain.Flashsale) error {
	if flashsale == nil {
		return nil
	}
	return r.client.Index(ctx, r.index, r.documentID(uint64(flashsale.ID)), flashsale)
}

func (r *flashsaleSearchRepository) Delete(ctx context.Context, flashsaleID uint64) error {
	if flashsaleID == 0 {
		return nil
	}
	return r.client.Delete(ctx, r.index, r.documentID(flashsaleID))
}

func (r *flashsaleSearchRepository) Search(ctx context.Context, query *domain.FlashsaleQuery, offset, limit int) ([]*domain.Flashsale, int64, error) {
	esQuery := map[string]any{
		"from":             offset,
		"size":             limit,
		"track_total_hits": true,
		"sort": []any{
			map[string]any{"start_time": map[string]any{"order": "asc"}},
		},
	}

	filters := make([]any, 0, 4)
	if query != nil {
		if query.Status != nil {
			filters = append(filters, map[string]any{"term": map[string]any{"status": *query.Status}})
		}
		if query.ProductID > 0 {
			filters = append(filters, map[string]any{"term": map[string]any{"product_id": query.ProductID}})
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
				Source domain.Flashsale `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, esQuery, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	items := make([]*domain.Flashsale, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		item := hit.Source
		items[i] = &item
	}
	return items, searchRes.Hits.Total.Value, nil
}

func (r *flashsaleSearchRepository) documentID(id uint64) string {
	return strings.TrimSpace(fmt.Sprintf("%d", id))
}
