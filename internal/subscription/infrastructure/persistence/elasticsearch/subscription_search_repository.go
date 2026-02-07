// 生成摘要：实现订阅记录搜索仓储（Elasticsearch）。
package elasticsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/subscription/domain"
	"github.com/wyfcoding/pkg/search"
)

type subscriptionSearchRepository struct {
	client *search.Client
	index  string
}

// NewSubscriptionSearchRepository 创建订阅记录搜索仓储实现。
func NewSubscriptionSearchRepository(client *search.Client, index string) domain.SubscriptionSearchRepository {
	if index == "" {
		index = "subscriptions"
	}
	return &subscriptionSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *subscriptionSearchRepository) Index(ctx context.Context, sub *domain.Subscription) error {
	if sub == nil {
		return nil
	}
	return r.client.Index(ctx, r.index, r.documentID(uint64(sub.ID)), sub)
}

func (r *subscriptionSearchRepository) Delete(ctx context.Context, subscriptionID uint64) error {
	if subscriptionID == 0 {
		return nil
	}
	return r.client.Delete(ctx, r.index, r.documentID(subscriptionID))
}

func (r *subscriptionSearchRepository) Search(ctx context.Context, query *domain.SubscriptionQuery, offset, limit int) ([]*domain.Subscription, int64, error) {
	esQuery := map[string]any{
		"from":             offset,
		"size":             limit,
		"track_total_hits": true,
		"sort": []any{
			map[string]any{"created_at": map[string]any{"order": "desc"}},
		},
	}

	filters := make([]any, 0, 8)
	if query != nil {
		if query.UserID > 0 {
			filters = append(filters, map[string]any{"term": map[string]any{"user_id": query.UserID}})
		}
		if query.PlanID > 0 {
			filters = append(filters, map[string]any{"term": map[string]any{"plan_id": query.PlanID}})
		}
		if query.Status != nil {
			filters = append(filters, map[string]any{"term": map[string]any{"status": *query.Status}})
		}
		if query.AutoRenew != nil {
			filters = append(filters, map[string]any{"term": map[string]any{"auto_renew": *query.AutoRenew}})
		}
		if !query.StartTime.IsZero() || !query.EndTime.IsZero() {
			rangeQuery := map[string]any{}
			if !query.StartTime.IsZero() {
				rangeQuery["gte"] = query.StartTime
			}
			if !query.EndTime.IsZero() {
				rangeQuery["lte"] = query.EndTime
			}
			filters = append(filters, map[string]any{"range": map[string]any{"created_at": rangeQuery}})
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
				Source domain.Subscription `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, esQuery, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	items := make([]*domain.Subscription, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		sub := hit.Source
		items[i] = &sub
	}
	return items, searchRes.Hits.Total.Value, nil
}

func (r *subscriptionSearchRepository) documentID(id uint64) string {
	return strings.TrimSpace(fmt.Sprintf("%d", id))
}
