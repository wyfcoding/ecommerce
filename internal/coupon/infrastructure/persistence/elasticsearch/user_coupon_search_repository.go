// 生成摘要：实现用户优惠券搜索仓储（Elasticsearch），支持分页与条件过滤。
// 假设：索引字段与 domain.UserCoupon 的 JSON 映射一致。
package elasticsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/coupon/domain"
	"github.com/wyfcoding/pkg/search"
)

type userCouponSearchRepository struct {
	client *search.Client
	index  string
}

// NewUserCouponSearchRepository 创建用户优惠券搜索仓储实现。
func NewUserCouponSearchRepository(client *search.Client, index string) domain.UserCouponSearchRepository {
	if index == "" {
		index = "user_coupons"
	}
	return &userCouponSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *userCouponSearchRepository) IndexUserCoupon(ctx context.Context, coupon *domain.UserCoupon) error {
	if coupon == nil {
		return nil
	}
	return r.client.Index(ctx, r.index, r.documentID(coupon.ID), coupon)
}

func (r *userCouponSearchRepository) DeleteUserCoupon(ctx context.Context, userCouponID uint64) error {
	if userCouponID == 0 {
		return nil
	}
	return r.client.Delete(ctx, r.index, r.documentID(userCouponID))
}

func (r *userCouponSearchRepository) SearchUserCoupons(ctx context.Context, userID uint64, status string, offset, limit int) ([]*domain.UserCoupon, int64, error) {
	if limit <= 0 {
		limit = 10
	}

	filters := make([]any, 0, 2)
	filters = append(filters, map[string]any{"term": map[string]any{"user_id": userID}})
	if status != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"status": status}})
	}

	query := map[string]any{
		"from":             offset,
		"size":             limit,
		"track_total_hits": true,
		"sort": []any{
			map[string]any{"created_at": map[string]any{"order": "desc"}},
		},
		"query": map[string]any{
			"bool": map[string]any{
				"filter": filters,
			},
		},
	}

	var searchRes struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source domain.UserCoupon `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, query, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}
	items := make([]*domain.UserCoupon, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		item := hit.Source
		items[i] = &item
	}
	return items, searchRes.Hits.Total.Value, nil
}

func (r *userCouponSearchRepository) documentID(userCouponID uint64) string {
	return strings.TrimSpace(fmt.Sprintf("%d", userCouponID))
}
