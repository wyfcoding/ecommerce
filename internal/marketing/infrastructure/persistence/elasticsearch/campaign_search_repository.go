// 生成摘要：实现营销活动搜索仓储（Elasticsearch）。
package elasticsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/marketing/domain"
	"github.com/wyfcoding/pkg/search"
)

type campaignSearchRepository struct {
	client *search.Client
	index  string
}

// NewCampaignSearchRepository 创建营销活动搜索仓储实现。
func NewCampaignSearchRepository(client *search.Client, index string) domain.CampaignSearchRepository {
	if index == "" {
		index = "marketing_campaigns"
	}
	return &campaignSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *campaignSearchRepository) Index(ctx context.Context, campaign *domain.Campaign) error {
	if campaign == nil {
		return nil
	}
	docID := fmt.Sprintf("%d", campaign.ID)
	return r.client.Index(ctx, r.index, docID, campaign)
}

func (r *campaignSearchRepository) Delete(ctx context.Context, id uint64) error {
	docID := fmt.Sprintf("%d", id)
	return r.client.Delete(ctx, r.index, docID)
}

func (r *campaignSearchRepository) Search(ctx context.Context, status *domain.CampaignStatus, keyword string, offset, limit int) ([]*domain.Campaign, int64, error) {
	filters := make([]any, 0)
	must := make([]any, 0)

	if status != nil {
		filters = append(filters, map[string]any{
			"term": map[string]any{"status": *status},
		})
	}

	keyword = strings.TrimSpace(keyword)
	if keyword != "" {
		must = append(must, map[string]any{
			"multi_match": map[string]any{
				"query":  keyword,
				"fields": []string{"name", "description"},
			},
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

	boolQuery := map[string]any{}
	if len(filters) > 0 {
		boolQuery["filter"] = filters
	}
	if len(must) > 0 {
		boolQuery["must"] = must
	}

	if len(boolQuery) == 0 {
		query["query"] = map[string]any{"match_all": map[string]any{}}
	} else {
		query["query"] = map[string]any{"bool": boolQuery}
	}

	var searchRes struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source domain.Campaign `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, query, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	items := make([]*domain.Campaign, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		c := hit.Source
		items[i] = &c
	}

	return items, searchRes.Hits.Total.Value, nil
}
