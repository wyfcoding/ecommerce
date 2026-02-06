// 生成摘要：实现优惠券搜索仓储（Elasticsearch），支持分页与条件过滤。
// 假设：索引字段与 domain.Coupon 的 JSON 映射一致。
package elasticsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/coupon/domain"
	"github.com/wyfcoding/pkg/search"
)

type couponSearchRepository struct {
	client *search.Client
	index  string
}

// NewCouponSearchRepository 创建优惠券搜索仓储实现。
func NewCouponSearchRepository(client *search.Client, index string) domain.CouponSearchRepository {
	if index == "" {
		index = "coupons"
	}
	return &couponSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *couponSearchRepository) IndexCoupon(ctx context.Context, coupon *domain.Coupon) error {
	if coupon == nil {
		return nil
	}
	return r.client.Index(ctx, r.index, r.documentID(coupon.ID), coupon)
}

func (r *couponSearchRepository) DeleteCoupon(ctx context.Context, couponID uint64) error {
	if couponID == 0 {
		return nil
	}
	return r.client.Delete(ctx, r.index, r.documentID(couponID))
}

func (r *couponSearchRepository) SearchCoupons(ctx context.Context, status *domain.CouponStatus, offset, limit int) ([]*domain.Coupon, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	query := map[string]any{
		"from":             offset,
		"size":             limit,
		"track_total_hits": true,
		"sort": []any{
			map[string]any{"created_at": map[string]any{"order": "desc"}},
		},
	}
	if status != nil {
		query["query"] = map[string]any{
			"term": map[string]any{"status": *status},
		}
	}

	var searchRes struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source domain.Coupon `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, query, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}
	items := make([]*domain.Coupon, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		item := hit.Source
		items[i] = &item
	}
	return items, searchRes.Hits.Total.Value, nil
}

func (r *couponSearchRepository) documentID(couponID uint64) string {
	return strings.TrimSpace(fmt.Sprintf("%d", couponID))
}
