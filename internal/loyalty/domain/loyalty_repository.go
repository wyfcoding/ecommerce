package domain

import (
	"context"
)

type LoyaltyRepository interface {
	SaveMemberAccount(ctx context.Context, account *MemberAccount) error
	GetMemberAccount(ctx context.Context, userID uint64) (*MemberAccount, error)

	SavePointsTransaction(ctx context.Context, transaction *PointsTransaction) error
	ListPointsTransactions(ctx context.Context, userID uint64, offset, limit int) ([]*PointsTransaction, int64, error)

	SaveMemberBenefit(ctx context.Context, benefit *MemberBenefit) error
	GetMemberBenefit(ctx context.Context, id uint64) (*MemberBenefit, error)
	ListMemberBenefits(ctx context.Context, level MemberLevel) ([]*MemberBenefit, error)
	DeleteMemberBenefit(ctx context.Context, id uint64) error
}
