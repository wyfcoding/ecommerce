package elasticsearch

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/wyfcoding/ecommerce/internal/warehouse/domain"
	"github.com/wyfcoding/pkg/search"
)

// warehouseSearchRepository 基于 Elasticsearch 的仓库搜索仓储。
type warehouseSearchRepository struct {
	client         *search.Client
	warehouseIndex string
	transferIndex  string
}

// NewWarehouseSearchRepository 创建仓库搜索仓储实现。
func NewWarehouseSearchRepository(client *search.Client, warehouseIndex, transferIndex string) domain.WarehouseSearchRepository {
	if warehouseIndex == "" {
		warehouseIndex = "warehouses"
	}
	if transferIndex == "" {
		transferIndex = "warehouse_transfers"
	}
	return &warehouseSearchRepository{
		client:         client,
		warehouseIndex: warehouseIndex,
		transferIndex:  transferIndex,
	}
}

// IndexWarehouse 将仓库写入搜索索引。
func (r *warehouseSearchRepository) IndexWarehouse(ctx context.Context, warehouse *domain.Warehouse) error {
	if warehouse == nil {
		return nil
	}
	docID := fmt.Sprintf("%d", warehouse.ID)
	return r.client.Index(ctx, r.warehouseIndex, docID, warehouse)
}

// DeleteWarehouse 从索引中删除仓库。
func (r *warehouseSearchRepository) DeleteWarehouse(ctx context.Context, id uint64) error {
	docID := fmt.Sprintf("%d", id)
	return r.client.Delete(ctx, r.warehouseIndex, docID)
}

// SearchWarehouses 按条件检索仓库（支持编码/名称/省市/状态过滤、分页）。
func (r *warehouseSearchRepository) SearchWarehouses(ctx context.Context, code, name, province, city, status *string, offset, limit int, startTime, endTime *time.Time, sortBy string) ([]*domain.Warehouse, int64, error) {
	filters := make([]any, 0)
	if code != nil && *code != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"code": *code}})
	}
	if name != nil && *name != "" {
		filters = append(filters, map[string]any{"match": map[string]any{"name": *name}})
	}
	if province != nil && *province != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"province": *province}})
	}
	if city != nil && *city != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"city": *city}})
	}
	if status != nil && *status != "" {
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

	sortField, desc := parseWarehouseSort(sortBy)
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
				Source domain.Warehouse `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.warehouseIndex, query, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	warehouses := make([]*domain.Warehouse, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		w := hit.Source
		warehouses[i] = &w
	}
	return warehouses, searchRes.Hits.Total.Value, nil
}

// IndexTransfer 将调拨单写入搜索索引。
func (r *warehouseSearchRepository) IndexTransfer(ctx context.Context, transfer *domain.StockTransfer) error {
	if transfer == nil {
		return nil
	}
	docID := fmt.Sprintf("%d", transfer.ID)
	return r.client.Index(ctx, r.transferIndex, docID, transfer)
}

// DeleteTransfer 从索引中删除调拨单。
func (r *warehouseSearchRepository) DeleteTransfer(ctx context.Context, id uint64) error {
	docID := fmt.Sprintf("%d", id)
	return r.client.Delete(ctx, r.transferIndex, docID)
}

// SearchTransfers 按条件检索调拨单。
func (r *warehouseSearchRepository) SearchTransfers(ctx context.Context, fromID, toID *uint64, status *string, offset, limit int, startTime, endTime *time.Time, sortBy string) ([]*domain.StockTransfer, int64, error) {
	filters := make([]any, 0)
	if fromID != nil && *fromID > 0 {
		filters = append(filters, map[string]any{"term": map[string]any{"from_warehouse_id": *fromID}})
	}
	if toID != nil && *toID > 0 {
		filters = append(filters, map[string]any{"term": map[string]any{"to_warehouse_id": *toID}})
	}
	if status != nil && *status != "" {
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

	sortField, desc := parseTransferSort(sortBy)
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
				Source domain.StockTransfer `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.transferIndex, query, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	transfers := make([]*domain.StockTransfer, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		t := hit.Source
		transfers[i] = &t
	}
	return transfers, searchRes.Hits.Total.Value, nil
}

func parseWarehouseSort(sortBy string) (string, bool) {
	allowed := map[string]string{
		"created_at": "created_at",
		"updated_at": "updated_at",
		"priority":   "priority",
	}

	sortBy = strings.TrimSpace(strings.ToLower(sortBy))
	if sortBy == "" {
		return "priority", true
	}

	desc := true
	if strings.HasPrefix(sortBy, "-") {
		sortBy = strings.TrimPrefix(sortBy, "-")
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
	return "priority", true
}

func parseTransferSort(sortBy string) (string, bool) {
	allowed := map[string]string{
		"created_at": "created_at",
		"updated_at": "updated_at",
		"quantity":   "quantity",
	}

	sortBy = strings.TrimSpace(strings.ToLower(sortBy))
	if sortBy == "" {
		return "created_at", true
	}

	desc := true
	if strings.HasPrefix(sortBy, "-") {
		sortBy = strings.TrimPrefix(sortBy, "-")
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
