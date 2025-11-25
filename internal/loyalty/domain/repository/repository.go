package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/loyalty/domain/entity"
)

// LoyaltyRepository 会员仓储接口
type LoyaltyRepository interface {
	// 会员账户
	SaveMemberAccount(ctx context.Context, account *entity.MemberAccount) error
	GetMemberAccount(ctx context.Context, userID uint64) (*entity.MemberAccount, error)

	// 积分交易
	SavePointsTransaction(ctx context.Context, transaction *entity.PointsTransaction) error
	ListPointsTransactions(ctx context.Context, userID uint64, offset, limit int) ([]*entity.PointsTransaction, int64, error)

	// 会员权益
	SaveMemberBenefit(ctx context.Context, benefit *entity.MemberBenefit) error
	GetMemberBenefit(ctx context.Context, id uint64) (*entity.MemberBenefit, error)
	ListMemberBenefits(ctx context.Context, level entity.MemberLevel) ([]*entity.MemberBenefit, error)
	DeleteMemberBenefit(ctx context.Context, id uint64) error
}
