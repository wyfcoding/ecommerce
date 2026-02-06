// 生成摘要：实现结算搜索仓储（Elasticsearch），支持分页与条件过滤。
// 假设：索引字段与 domain.Settlement 的 JSON 映射一致，CreatedAt 字段可用于排序。
package elasticsearch

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/wyfcoding/ecommerce/internal/settlement/domain"
	"github.com/wyfcoding/pkg/search"
)

type settlementSearchRepository struct {
	client *search.Client
	index  string
}

// NewSettlementSearchRepository 创建结算搜索仓储实现。
func NewSettlementSearchRepository(client *search.Client, index string) domain.SettlementSearchRepository {
	if index == "" {
		index = "settlements"
	}
	return &settlementSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *settlementSearchRepository) Index(ctx context.Context, settlement *domain.Settlement) error {
	if settlement == nil {
		return nil
	}
	docID := fmt.Sprintf("%d", settlement.ID)
	return r.client.Index(ctx, r.index, docID, settlement)
}

func (r *settlementSearchRepository) Delete(ctx context.Context, settlementID uint64) error {
	docID := fmt.Sprintf("%d", settlementID)
	return r.client.Delete(ctx, r.index, docID)
}

func (r *settlementSearchRepository) Search(ctx context.Context, merchantID *uint64, status *domain.SettlementStatus, offset, limit int, startTime, endTime *time.Time, sortBy string) ([]*domain.Settlement, int64, error) {
	filters := make([]any, 0)
	if merchantID != nil {
		filters = append(filters, map[string]any{
			"term": map[string]any{"merchant_id": *merchantID},
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

	sortField, desc := parseSettlementSort(sortBy)
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
				Source domain.Settlement `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, query, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	results := make([]*domain.Settlement, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		settlement := hit.Source
		results[i] = &settlement
	}

	return results, searchRes.Hits.Total.Value, nil
}

func (r *settlementSearchRepository) FindByNo(ctx context.Context, settlementNo string) (*domain.Settlement, error) {
	query := map[string]any{
		"query": map[string]any{
			"term": map[string]any{"settlement_no": settlementNo},
		},
		"size": 1,
	}

	var searchRes struct {
		Hits struct {
			Hits []struct {
				Source domain.Settlement `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, query, &searchRes); err != nil {
		return nil, fmt.Errorf("es search failed: %w", err)
	}

	if len(searchRes.Hits.Hits) == 0 {
		return nil, nil
	}

	settlement := searchRes.Hits.Hits[0].Source
	return &settlement, nil
}

func parseSettlementSort(sortBy string) (string, bool) {
	allowed := map[string]string{
		"created_at":        "created_at",
		"updated_at":        "updated_at",
		"settled_at":        "settled_at",
		"total_amount":      "total_amount",
		"settlement_amount": "settlement_amount",
		"platform_fee":      "platform_fee",
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
