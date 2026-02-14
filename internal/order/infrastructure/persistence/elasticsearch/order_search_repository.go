// 生成摘要：实现订单搜索仓储（Elasticsearch），支持分页与条件过滤。
// 假设：索引字段与 domain.Order 的 JSON 映射一致，CreatedAt 字段可用于排序。
package elasticsearch

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/wyfcoding/ecommerce/internal/order/domain"
	"github.com/wyfcoding/pkg/search"
)

// orderSearchRepository 基于 Elasticsearch 的订单搜索仓储。
type orderSearchRepository struct {
	client *search.Client
	index  string
}

// NewOrderSearchRepository 创建订单搜索仓储实现。
func NewOrderSearchRepository(client *search.Client, index string) domain.OrderSearchRepository {
	if index == "" {
		index = "orders"
	}
	return &orderSearchRepository{
		client: client,
		index:  index,
	}
}

// Index 将订单写入搜索索引。
func (r *orderSearchRepository) Index(ctx context.Context, order *domain.Order) error {
	if order == nil {
		return nil
	}
	docID := fmt.Sprintf("%d", order.ID)
	return r.client.Index(ctx, r.index, docID, order)
}

// Delete 从索引中删除订单。
func (r *orderSearchRepository) Delete(ctx context.Context, orderID uint64) error {
	docID := fmt.Sprintf("%d", orderID)
	return r.client.Delete(ctx, r.index, docID)
}

// Search 按条件检索订单（支持用户与状态过滤、分页）。
func (r *orderSearchRepository) Search(ctx context.Context, userID *uint64, status *int, offset, limit int, startTime, endTime *time.Time, sortBy string) ([]*domain.Order, int64, error) {
	filters := make([]any, 0)
	if userID != nil {
		filters = append(filters, map[string]any{
			"term": map[string]any{"user_id": *userID},
		})
	}
	if status != nil {
		filters = append(filters, map[string]any{
			"term": map[string]any{"status": *status},
		})
	}
	if startTime != nil || endTime != nil {
		rangeFilter := map[string]any{}
		if startTime != nil {
			rangeFilter["gte"] = startTime.Format(time.RFC3339)
		}
		if endTime != nil {
			rangeFilter["lte"] = endTime.Format(time.RFC3339)
		}
		filters = append(filters, map[string]any{
			"range": map[string]any{"created_at": rangeFilter},
		})
	}

	sortField, desc := parseOrderSort(sortBy)
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
				Source domain.Order `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, query, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	orders := make([]*domain.Order, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		order := hit.Source
		orders[i] = &order
	}

	return orders, searchRes.Hits.Total.Value, nil
}

// FindByOrderNo 通过订单号检索订单。
func (r *orderSearchRepository) FindByOrderNo(ctx context.Context, orderNo string) (*domain.Order, error) {
	query := map[string]any{
		"query": map[string]any{
			"term": map[string]any{"order_no": orderNo},
		},
		"size": 1,
	}

	var searchRes struct {
		Hits struct {
			Hits []struct {
				Source domain.Order `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, query, &searchRes); err != nil {
		return nil, fmt.Errorf("es search failed: %w", err)
	}

	if len(searchRes.Hits.Hits) == 0 {
		return nil, nil
	}

	order := searchRes.Hits.Hits[0].Source
	return &order, nil
}

func parseOrderSort(sortBy string) (string, bool) {
	allowed := map[string]string{
		"created_at":    "created_at",
		"updated_at":    "updated_at",
		"paid_at":       "paid_at",
		"shipped_at":    "shipped_at",
		"delivered_at":  "delivered_at",
		"completed_at":  "completed_at",
		"cancelled_at":  "cancelled_at",
		"total_amount":  "total_amount",
		"actual_amount": "actual_amount",
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
