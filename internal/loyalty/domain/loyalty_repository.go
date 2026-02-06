package domain

import (
	"context"
)

// LoyaltyRepository 定义了数据持久层接口。
type LoyaltyRepository interface {
	// 事务管理
	BeginTx(ctx context.Context) any
	CommitTx(tx any) error
	RollbackTx(tx any) error
	WithTx(ctx context.Context, fn func(tx any) error) error

	// 会员账户
	SaveMemberAccount(ctx context.Context, account *MemberAccount) error
	SaveMemberAccountInTx(ctx context.Context, tx any, account *MemberAccount) error
	GetMemberAccount(ctx context.Context, userID uint64) (*MemberAccount, error)

	// 积分交易
	SavePointsTransaction(ctx context.Context, transaction *PointsTransaction) error
	SavePointsTransactionInTx(ctx context.Context, tx any, transaction *PointsTransaction) error
	GetPointsTransaction(ctx context.Context, id uint64) (*PointsTransaction, error)
	ListPointsTransactions(ctx context.Context, userID uint64, offset, limit int) ([]*PointsTransaction, int64, error)

	// 会员权益
	SaveMemberBenefit(ctx context.Context, benefit *MemberBenefit) error
	SaveMemberBenefitInTx(ctx context.Context, tx any, benefit *MemberBenefit) error
	GetMemberBenefit(ctx context.Context, id uint64) (*MemberBenefit, error)
	GetMemberBenefitByLevel(ctx context.Context, level MemberLevel) (*MemberBenefit, error)
	ListMemberBenefits(ctx context.Context, level MemberLevel) ([]*MemberBenefit, error)
	DeleteMemberBenefit(ctx context.Context, id uint64) error
	DeleteMemberBenefitInTx(ctx context.Context, tx any, id uint64) error
}
