// 生成摘要：实现支付搜索仓储（Elasticsearch），支持分页与条件过滤。
// 假设：索引字段与 domain.Payment 的 JSON 映射一致，CreatedAt 字段可用于排序。
package elasticsearch

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
	"github.com/wyfcoding/pkg/search"
)

// paymentSearchRepository 基于 Elasticsearch 的支付搜索仓储。
type paymentSearchRepository struct {
	client *search.Client
	index  string
}

// NewPaymentSearchRepository 创建支付搜索仓储实现。
func NewPaymentSearchRepository(client *search.Client, index string) domain.PaymentSearchRepository {
	if index == "" {
		index = "payments"
	}
	return &paymentSearchRepository{
		client: client,
		index:  index,
	}
}

// Index 将支付写入搜索索引。
func (r *paymentSearchRepository) Index(ctx context.Context, payment *domain.Payment) error {
	if payment == nil || payment.ID == 0 {
		return nil
	}
	docID := fmt.Sprintf("%d", payment.ID)
	return r.client.Index(ctx, r.index, docID, payment)
}

// Delete 从索引中删除支付。
func (r *paymentSearchRepository) Delete(ctx context.Context, paymentID uint64) error {
	docID := fmt.Sprintf("%d", paymentID)
	return r.client.Delete(ctx, r.index, docID)
}

// Search 按条件检索支付（支持用户与状态过滤、分页）。
func (r *paymentSearchRepository) Search(ctx context.Context, userID *uint64, status *domain.PaymentStatus, offset, limit int, startTime, endTime *time.Time, sortBy string) ([]*domain.Payment, int64, error) {
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

	sortField, desc := parsePaymentSort(sortBy)
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
				Source domain.Payment `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, query, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	payments := make([]*domain.Payment, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		p := hit.Source
		payments[i] = &p
	}

	return payments, searchRes.Hits.Total.Value, nil
}

// FindByPaymentNo 通过支付单号检索支付。
func (r *paymentSearchRepository) FindByPaymentNo(ctx context.Context, paymentNo string) (*domain.Payment, error) {
	query := map[string]any{
		"query": map[string]any{
			"term": map[string]any{"payment_no": paymentNo},
		},
		"size": 1,
	}

	var searchRes struct {
		Hits struct {
			Hits []struct {
				Source domain.Payment `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, query, &searchRes); err != nil {
		return nil, fmt.Errorf("es search failed: %w", err)
	}

	if len(searchRes.Hits.Hits) == 0 {
		return nil, nil
	}

	payment := searchRes.Hits.Hits[0].Source
	return &payment, nil
}

// FindByOrderID 通过订单ID检索支付。
func (r *paymentSearchRepository) FindByOrderID(ctx context.Context, orderID uint64) (*domain.Payment, error) {
	query := map[string]any{
		"query": map[string]any{
			"term": map[string]any{"order_id": orderID},
		},
		"size": 1,
	}

	var searchRes struct {
		Hits struct {
			Hits []struct {
				Source domain.Payment `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, query, &searchRes); err != nil {
		return nil, fmt.Errorf("es search failed: %w", err)
	}

	if len(searchRes.Hits.Hits) == 0 {
		return nil, nil
	}

	payment := searchRes.Hits.Hits[0].Source
	return &payment, nil
}

func parsePaymentSort(sortBy string) (string, bool) {
	allowed := map[string]string{
		"created_at":      "created_at",
		"updated_at":      "updated_at",
		"paid_at":         "paid_at",
		"amount":          "amount",
		"captured_amount": "captured_amount",
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
