// 生成摘要：实现售后搜索仓储（Elasticsearch）。
package elasticsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/aftersales/domain"
	"github.com/wyfcoding/pkg/search"
)

type afterSalesSearchRepository struct {
	client *search.Client
	index  string
}

// NewAfterSalesSearchRepository 创建售后搜索仓储实现。
func NewAfterSalesSearchRepository(client *search.Client, index string) domain.AfterSalesSearchRepository {
	if index == "" {
		index = "after_sales"
	}
	return &afterSalesSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *afterSalesSearchRepository) Index(ctx context.Context, afterSales *domain.AfterSales) error {
	if afterSales == nil {
		return nil
	}
	return r.client.Index(ctx, r.index, r.documentID(afterSales.ID), afterSales)
}

func (r *afterSalesSearchRepository) Delete(ctx context.Context, afterSalesID uint64) error {
	if afterSalesID == 0 {
		return nil
	}
	return r.client.Delete(ctx, r.index, r.documentID(afterSalesID))
}

func (r *afterSalesSearchRepository) Search(ctx context.Context, query *domain.AfterSalesQuery) ([]*domain.AfterSales, int64, error) {
	if query == nil {
		query = &domain.AfterSalesQuery{Page: 1, PageSize: 10}
	}
	limit := query.PageSize
	if limit <= 0 {
		limit = 10
	}
	offset := max((query.Page-1)*limit, 0)

	esQuery := map[string]any{
		"from":             offset,
		"size":             limit,
		"track_total_hits": true,
		"sort": []any{
			map[string]any{"created_at": map[string]any{"order": "desc"}},
		},
	}

	filters := make([]any, 0, 5)
	if query.UserID > 0 {
		filters = append(filters, map[string]any{"term": map[string]any{"user_id": query.UserID}})
	}
	if query.OrderID > 0 {
		filters = append(filters, map[string]any{"term": map[string]any{"order_id": query.OrderID}})
	}
	if query.Type > 0 {
		filters = append(filters, map[string]any{"term": map[string]any{"type": query.Type}})
	}
	if query.Status > 0 {
		filters = append(filters, map[string]any{"term": map[string]any{"status": query.Status}})
	}
	if query.AfterSalesNo != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"after_sales_no": query.AfterSalesNo}})
	}
	if len(filters) > 0 {
		esQuery["query"] = map[string]any{
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
				Source domain.AfterSales `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, esQuery, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	items := make([]*domain.AfterSales, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		item := hit.Source
		items[i] = &item
	}
	return items, searchRes.Hits.Total.Value, nil
}

func (r *afterSalesSearchRepository) documentID(id uint64) string {
	return strings.TrimSpace(fmt.Sprintf("%d", id))
}
