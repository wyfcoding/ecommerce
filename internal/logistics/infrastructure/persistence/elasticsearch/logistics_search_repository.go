package elasticsearch

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/wyfcoding/ecommerce/internal/logistics/domain"
	"github.com/wyfcoding/pkg/search"
)

// logisticsSearchRepository 基于 Elasticsearch 的物流搜索仓储。
type logisticsSearchRepository struct {
	client *search.Client
	index  string
}

// NewLogisticsSearchRepository 创建物流搜索仓储实现。
func NewLogisticsSearchRepository(client *search.Client, index string) domain.LogisticsSearchRepository {
	if index == "" {
		index = "logistics"
	}
	return &logisticsSearchRepository{client: client, index: index}
}

// Index 将物流单写入搜索索引。
func (r *logisticsSearchRepository) Index(ctx context.Context, logistics *domain.Logistics) error {
	if logistics == nil {
		return nil
	}
	docID := fmt.Sprintf("%d", logistics.ID)
	return r.client.Index(ctx, r.index, docID, logistics)
}

// Delete 从索引中删除物流单。
func (r *logisticsSearchRepository) Delete(ctx context.Context, id uint64) error {
	docID := fmt.Sprintf("%d", id)
	return r.client.Delete(ctx, r.index, docID)
}

// Search 按条件检索物流单（支持订单/运单/承运商/状态过滤、分页）。
func (r *logisticsSearchRepository) Search(ctx context.Context, orderID *uint64, trackingNo, carrier *string, status *domain.LogisticsStatus, offset, limit int, startTime, endTime *time.Time, sortBy string) ([]*domain.Logistics, int64, error) {
	filters := make([]any, 0)
	if orderID != nil {
		filters = append(filters, map[string]any{"term": map[string]any{"order_id": *orderID}})
	}
	if trackingNo != nil && *trackingNo != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"tracking_no": *trackingNo}})
	}
	if carrier != nil && *carrier != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"carrier": *carrier}})
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

	sortField, desc := parseLogisticsSort(sortBy)
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
				Source domain.Logistics `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, query, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	list := make([]*domain.Logistics, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		l := hit.Source
		list[i] = &l
	}
	return list, searchRes.Hits.Total.Value, nil
}

// FindByTrackingNo 通过运单号检索物流单。
func (r *logisticsSearchRepository) FindByTrackingNo(ctx context.Context, trackingNo string) (*domain.Logistics, error) {
	query := map[string]any{
		"query": map[string]any{
			"term": map[string]any{"tracking_no": trackingNo},
		},
		"size": 1,
	}

	var searchRes struct {
		Hits struct {
			Hits []struct {
				Source domain.Logistics `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, query, &searchRes); err != nil {
		return nil, fmt.Errorf("es search failed: %w", err)
	}
	if len(searchRes.Hits.Hits) == 0 {
		return nil, nil
	}

	result := searchRes.Hits.Hits[0].Source
	return &result, nil
}

func parseLogisticsSort(sortBy string) (string, bool) {
	allowed := map[string]string{
		"created_at": "created_at",
		"updated_at": "updated_at",
		"status":     "status",
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
