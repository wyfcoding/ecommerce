// 生成摘要：实现订单搜索仓储（Elasticsearch），支持分页与条件过滤。
// 假设：索引字段与 domain.Order 的 JSON 映射一致，CreatedAt 字段可用于排序。
package elasticsearch

import (
	"context"
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/order/domain"
	"github.com/wyfcoding/pkg/search"
)

// orderSearchRepository 基于 Elasticsearch 的订单搜索仓储。
type orderSearchRepository struct {
	client *search.Client
	index  string
}

// NewOrderSearchRepository 创建订单搜索仓储实现。
func NewOrderSearchRepository(client *search.Client) domain.OrderSearchRepository {
	return &orderSearchRepository{
		client: client,
		index:  "orders",
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
func (r *orderSearchRepository) Search(ctx context.Context, userID *uint64, status *int, offset, limit int) ([]*domain.Order, int64, error) {
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

	query := map[string]any{
		"from":             offset,
		"size":             limit,
		"track_total_hits": true,
		"sort": []any{
			map[string]any{"CreatedAt": map[string]any{"order": "desc"}},
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
