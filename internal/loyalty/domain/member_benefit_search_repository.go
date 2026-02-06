// 生成摘要：定义会员权益搜索仓储接口（Elasticsearch）。
package domain

import "context"

// MemberBenefitSearchRepository 定义会员权益搜索的访问接口。
type MemberBenefitSearchRepository interface {
	IndexBenefit(ctx context.Context, benefit *MemberBenefit) error
	DeleteBenefit(ctx context.Context, benefitID uint64) error
	SearchBenefits(ctx context.Context, level MemberLevel, offset, limit int) ([]*MemberBenefit, int64, error)
}
