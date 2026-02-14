package elasticsearch

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/wyfcoding/ecommerce/internal/inventory/domain"
	"github.com/wyfcoding/pkg/search"
)

// inventorySearchRepository 基于 Elasticsearch 的库存搜索仓储。
type inventorySearchRepository struct {
	client *search.Client
	index  string
}

// NewInventorySearchRepository 创建库存搜索仓储实现。
func NewInventorySearchRepository(client *search.Client, index string) domain.InventorySearchRepository {
	if index == "" {
		index = "inventories"
	}
	return &inventorySearchRepository{client: client, index: index}
}

// Index 将库存写入搜索索引。
func (r *inventorySearchRepository) Index(ctx context.Context, inventory *domain.Inventory) error {
	if inventory == nil {
		return nil
	}
	docID := fmt.Sprintf("%d", inventory.SkuID)
	return r.client.Index(ctx, r.index, docID, inventory)
}

// Delete 从索引中删除库存。
func (r *inventorySearchRepository) Delete(ctx context.Context, skuID uint64) error {
	docID := fmt.Sprintf("%d", skuID)
	return r.client.Delete(ctx, r.index, docID)
}

// Search 按条件检索库存（支持 SKU / 商品 / 仓库 / 状态过滤、分页）。
func (r *inventorySearchRepository) Search(ctx context.Context, skuID, productID, warehouseID *uint64, status *domain.InventoryStatus, offset, limit int, startTime, endTime *time.Time, sortBy string) ([]*domain.Inventory, int64, error) {
	filters := make([]any, 0)
	if skuID != nil {
		filters = append(filters, map[string]any{"term": map[string]any{"sku_id": *skuID}})
	}
	if productID != nil {
		filters = append(filters, map[string]any{"term": map[string]any{"product_id": *productID}})
	}
	if warehouseID != nil {
		filters = append(filters, map[string]any{"term": map[string]any{"warehouse_id": *warehouseID}})
	}
	if status != nil {
		filters = append(filters, map[string]any{"term": map[string]any{"status": *status}})
	}
	if startTime != nil || endTime != nil {
		rangeFilter := map[string]any{}
		if startTime != nil {
			rangeFilter["gte"] = startTime.Format(time.RFC3339)
		}
		if endTime != nil {
			rangeFilter["lte"] = endTime.Format(time.RFC3339)
		}
		filters = append(filters, map[string]any{"range": map[string]any{"created_at": rangeFilter}})
	}

	sortField, desc := parseInventorySort(sortBy)
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
		query["query"] = map[string]any{"bool": map[string]any{"filter": filters}}
	}

	var searchRes struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source domain.Inventory `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, query, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	inventories := make([]*domain.Inventory, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		inv := hit.Source
		inventories[i] = &inv
	}

	return inventories, searchRes.Hits.Total.Value, nil
}

// FindBySkuID 通过 SKU ID 检索库存。
func (r *inventorySearchRepository) FindBySkuID(ctx context.Context, skuID uint64) (*domain.Inventory, error) {
	query := map[string]any{
		"query": map[string]any{
			"term": map[string]any{"sku_id": skuID},
		},
		"size": 1,
	}

	var searchRes struct {
		Hits struct {
			Hits []struct {
				Source domain.Inventory `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, query, &searchRes); err != nil {
		return nil, fmt.Errorf("es search failed: %w", err)
	}

	if len(searchRes.Hits.Hits) == 0 {
		return nil, nil
	}

	inv := searchRes.Hits.Hits[0].Source
	return &inv, nil
}

func parseInventorySort(sortBy string) (string, bool) {
	allowed := map[string]string{
		"created_at":      "created_at",
		"updated_at":      "updated_at",
		"available_stock": "available_stock",
		"total_stock":     "total_stock",
		"locked_stock":    "locked_stock",
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
