// 生成摘要：定义会员权益读模型仓储接口（Redis）。
package domain

import "context"

// MemberBenefitReadRepository 定义会员权益读模型接口。
type MemberBenefitReadRepository interface {
	Save(ctx context.Context, benefit *MemberBenefit) error
	GetByLevel(ctx context.Context, level MemberLevel) (*MemberBenefit, error)
	DeleteByLevel(ctx context.Context, level MemberLevel) error
}
