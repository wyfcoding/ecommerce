package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/loyalty/domain"
)

// LoyaltyQueryService 负责处理 Loyalty 相关的读操作和查询逻辑。
type LoyaltyQueryService struct {
	repo              domain.LoyaltyRepository
	accountReadRepo   domain.MemberAccountReadRepository
	benefitReadRepo   domain.MemberBenefitReadRepository
	txSearchRepo      domain.PointsTransactionSearchRepository
	benefitSearchRepo domain.MemberBenefitSearchRepository
	logger            *slog.Logger
}

// NewLoyaltyQueryService 负责处理 Loyalty 相关的读操作和查询逻辑。
func NewLoyaltyQueryService(
	repo domain.LoyaltyRepository,
	accountReadRepo domain.MemberAccountReadRepository,
	benefitReadRepo domain.MemberBenefitReadRepository,
	txSearchRepo domain.PointsTransactionSearchRepository,
	benefitSearchRepo domain.MemberBenefitSearchRepository,
	logger *slog.Logger,
) *LoyaltyQueryService {
	return &LoyaltyQueryService{
		repo:              repo,
		accountReadRepo:   accountReadRepo,
		benefitReadRepo:   benefitReadRepo,
		txSearchRepo:      txSearchRepo,
		benefitSearchRepo: benefitSearchRepo,
		logger:            logger,
	}
}

func (q *LoyaltyQueryService) GetOrCreateAccount(ctx context.Context, userID uint64) (*domain.MemberAccount, error) {
	if q.accountReadRepo != nil {
		if account, err := q.accountReadRepo.GetByUserID(ctx, userID); err == nil && account != nil {
			return account, nil
		}
	}

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
	if q.accountReadRepo != nil {
		if err := q.accountReadRepo.Save(ctx, account); err != nil {
			q.logger.WarnContext(ctx, "failed to warm account cache", "user_id", userID, "error", err)
		}
	}
	return account, nil
}

func (q *LoyaltyQueryService) GetMemberAccount(ctx context.Context, userID uint64) (*domain.MemberAccount, error) {
	if q.accountReadRepo != nil {
		if account, err := q.accountReadRepo.GetByUserID(ctx, userID); err == nil && account != nil {
			return account, nil
		}
	}
	account, err := q.repo.GetMemberAccount(ctx, userID)
	if err != nil {
		return nil, err
	}
	if account != nil && q.accountReadRepo != nil {
		if err := q.accountReadRepo.Save(ctx, account); err != nil {
			q.logger.WarnContext(ctx, "failed to warm account cache", "user_id", userID, "error", err)
		}
	}
	return account, nil
}

func (q *LoyaltyQueryService) GetPointsTransactions(ctx context.Context, userID uint64, page, pageSize int) ([]*domain.PointsTransaction, int64, error) {
	offset := (page - 1) * pageSize
	if q.txSearchRepo != nil {
		list, total, err := q.txSearchRepo.SearchTransactions(ctx, userID, offset, pageSize)
		if err == nil {
			return list, total, nil
		}
		q.logger.WarnContext(ctx, "transaction search fallback to mysql", "user_id", userID, "error", err)
	}
	return q.repo.ListPointsTransactions(ctx, userID, offset, pageSize)
}

func (q *LoyaltyQueryService) ListBenefits(ctx context.Context, level domain.MemberLevel) ([]*domain.MemberBenefit, error) {
	if q.benefitSearchRepo != nil {
		list, _, err := q.benefitSearchRepo.SearchBenefits(ctx, level, 0, 1000)
		if err == nil {
			return list, nil
		}
		q.logger.WarnContext(ctx, "benefit search fallback to mysql", "level", level, "error", err)
	}
	return q.repo.ListMemberBenefits(ctx, level)
}

func (q *LoyaltyQueryService) GetMemberBenefitByLevel(ctx context.Context, level domain.MemberLevel) (*domain.MemberBenefit, error) {
	if q.benefitReadRepo != nil {
		if benefit, err := q.benefitReadRepo.GetByLevel(ctx, level); err == nil && benefit != nil {
			return benefit, nil
		}
	}
	benefit, err := q.repo.GetMemberBenefitByLevel(ctx, level)
	if err != nil {
		return nil, err
	}
	if benefit != nil && q.benefitReadRepo != nil {
		if err := q.benefitReadRepo.Save(ctx, benefit); err != nil {
			q.logger.WarnContext(ctx, "failed to warm benefit cache", "level", level, "error", err)
		}
	}
	return benefit, nil
}
