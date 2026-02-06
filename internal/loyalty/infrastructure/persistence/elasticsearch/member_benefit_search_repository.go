// 生成摘要：实现会员权益搜索仓储（Elasticsearch），支持分页与条件过滤。
// 假设：索引字段与 domain.MemberBenefit 的 JSON 映射一致。
package elasticsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/loyalty/domain"
	"github.com/wyfcoding/pkg/search"
)

type memberBenefitSearchRepository struct {
	client *search.Client
	index  string
}

// NewMemberBenefitSearchRepository 创建会员权益搜索仓储实现。
func NewMemberBenefitSearchRepository(client *search.Client, index string) domain.MemberBenefitSearchRepository {
	if index == "" {
		index = "member_benefits"
	}
	return &memberBenefitSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *memberBenefitSearchRepository) IndexBenefit(ctx context.Context, benefit *domain.MemberBenefit) error {
	if benefit == nil {
		return nil
	}
	return r.client.Index(ctx, r.index, r.documentID(benefit.ID), benefit)
}

func (r *memberBenefitSearchRepository) DeleteBenefit(ctx context.Context, benefitID uint64) error {
	if benefitID == 0 {
		return nil
	}
	return r.client.Delete(ctx, r.index, r.documentID(benefitID))
}

func (r *memberBenefitSearchRepository) SearchBenefits(ctx context.Context, level domain.MemberLevel, offset, limit int) ([]*domain.MemberBenefit, int64, error) {
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
	if level != "" {
		query["query"] = map[string]any{
			"term": map[string]any{"level": level},
		}
	}

	var searchRes struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source domain.MemberBenefit `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, query, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	items := make([]*domain.MemberBenefit, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		item := hit.Source
		items[i] = &item
	}
	return items, searchRes.Hits.Total.Value, nil
}

func (r *memberBenefitSearchRepository) documentID(benefitID uint64) string {
	return strings.TrimSpace(fmt.Sprintf("%d", benefitID))
}
