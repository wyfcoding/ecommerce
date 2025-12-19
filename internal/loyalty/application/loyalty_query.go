package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/loyalty/domain"
)

type LoyaltyQuery struct {
	repo domain.LoyaltyRepository
}

func NewLoyaltyQuery(repo domain.LoyaltyRepository) *LoyaltyQuery {
	return &LoyaltyQuery{
		repo: repo,
	}
}

func (q *LoyaltyQuery) GetOrCreateAccount(ctx context.Context, userID uint64) (*domain.MemberAccount, error) {
	account, err := q.repo.GetMemberAccount(ctx, userID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		account = domain.NewMemberAccount(userID)
		if err := q.repo.SaveMemberAccount(ctx, account); err != nil {
			return nil, err
		}
	}
	return account, nil
}

func (q *LoyaltyQuery) GetPointsTransactions(ctx context.Context, userID uint64, page, pageSize int) ([]*domain.PointsTransaction, int64, error) {
	offset := (page - 1) * pageSize
	return q.repo.ListPointsTransactions(ctx, userID, offset, pageSize)
}

func (q *LoyaltyQuery) ListBenefits(ctx context.Context, level domain.MemberLevel) ([]*domain.MemberBenefit, error) {
	return q.repo.ListMemberBenefits(ctx, level)
}

func (q *LoyaltyQuery) GetMemberAccount(ctx context.Context, userID uint64) (*domain.MemberAccount, error) {
	return q.repo.GetMemberAccount(ctx, userID)
}
