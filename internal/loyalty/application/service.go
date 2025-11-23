package application

import (
	"context"
	"ecommerce/internal/loyalty/domain/entity"
	"ecommerce/internal/loyalty/domain/repository"

	"log/slog"
)

type LoyaltyService struct {
	repo   repository.LoyaltyRepository
	logger *slog.Logger
}

func NewLoyaltyService(repo repository.LoyaltyRepository, logger *slog.Logger) *LoyaltyService {
	return &LoyaltyService{
		repo:   repo,
		logger: logger,
	}
}

// GetOrCreateAccount 获取或创建会员账户
func (s *LoyaltyService) GetOrCreateAccount(ctx context.Context, userID uint64) (*entity.MemberAccount, error) {
	account, err := s.repo.GetMemberAccount(ctx, userID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		account = entity.NewMemberAccount(userID)
		if err := s.repo.SaveMemberAccount(ctx, account); err != nil {
			s.logger.Error("failed to create member account", "error", err)
			return nil, err
		}
	}
	return account, nil
}

// AddPoints 增加积分
func (s *LoyaltyService) AddPoints(ctx context.Context, userID uint64, points int64, transactionType, description string, orderID uint64) error {
	account, err := s.GetOrCreateAccount(ctx, userID)
	if err != nil {
		return err
	}

	account.AddPoints(points)
	if err := s.repo.SaveMemberAccount(ctx, account); err != nil {
		return err
	}

	// Record transaction
	tx := entity.NewPointsTransaction(userID, transactionType, points, account.AvailablePoints, orderID, description, nil)
	return s.repo.SavePointsTransaction(ctx, tx)
}

// DeductPoints 扣减积分
func (s *LoyaltyService) DeductPoints(ctx context.Context, userID uint64, points int64, transactionType, description string, orderID uint64) error {
	account, err := s.GetOrCreateAccount(ctx, userID)
	if err != nil {
		return err
	}

	if err := account.DeductPoints(points); err != nil {
		return err
	}

	if err := s.repo.SaveMemberAccount(ctx, account); err != nil {
		return err
	}

	// Record transaction
	tx := entity.NewPointsTransaction(userID, transactionType, -points, account.AvailablePoints, orderID, description, nil)
	return s.repo.SavePointsTransaction(ctx, tx)
}

// AddSpent 增加消费金额
func (s *LoyaltyService) AddSpent(ctx context.Context, userID uint64, amount uint64) error {
	account, err := s.GetOrCreateAccount(ctx, userID)
	if err != nil {
		return err
	}

	account.AddSpent(amount)
	return s.repo.SaveMemberAccount(ctx, account)
}

// GetPointsTransactions 获取积分交易记录
func (s *LoyaltyService) GetPointsTransactions(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.PointsTransaction, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListPointsTransactions(ctx, userID, offset, pageSize)
}

// AddBenefit 添加会员权益
func (s *LoyaltyService) AddBenefit(ctx context.Context, level entity.MemberLevel, name, description string, discountRate, pointsRate float64) (*entity.MemberBenefit, error) {
	benefit := entity.NewMemberBenefit(level, name, description, discountRate, pointsRate)
	if err := s.repo.SaveMemberBenefit(ctx, benefit); err != nil {
		s.logger.Error("failed to save member benefit", "error", err)
		return nil, err
	}
	return benefit, nil
}

// ListBenefits 获取会员权益列表
func (s *LoyaltyService) ListBenefits(ctx context.Context, level entity.MemberLevel) ([]*entity.MemberBenefit, error) {
	return s.repo.ListMemberBenefits(ctx, level)
}

// DeleteBenefit 删除会员权益
func (s *LoyaltyService) DeleteBenefit(ctx context.Context, id uint64) error {
	return s.repo.DeleteMemberBenefit(ctx, id)
}
