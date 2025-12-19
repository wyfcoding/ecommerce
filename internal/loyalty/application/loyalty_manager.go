package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/loyalty/domain"

	"log/slog"
)

type LoyaltyManager struct {
	repo   domain.LoyaltyRepository
	logger *slog.Logger
}

func NewLoyaltyManager(repo domain.LoyaltyRepository, logger *slog.Logger) *LoyaltyManager {
	return &LoyaltyManager{
		repo:   repo,
		logger: logger,
	}
}

func (m *LoyaltyManager) AddPoints(ctx context.Context, userID uint64, points int64, transactionType, description string, orderID uint64) error {
	account, err := m.repo.GetMemberAccount(ctx, userID)
	if err != nil {
		return err
	}
	if account == nil {
		account = domain.NewMemberAccount(userID)
		if err := m.repo.SaveMemberAccount(ctx, account); err != nil {
			m.logger.ErrorContext(ctx, "failed to create member account", "user_id", userID, "error", err)
			return err
		}
	}

	account.AddPoints(points)
	if err := m.repo.SaveMemberAccount(ctx, account); err != nil {
		return err
	}

	tx := domain.NewPointsTransaction(userID, transactionType, points, account.AvailablePoints, orderID, description, nil)
	return m.repo.SavePointsTransaction(ctx, tx)
}

func (m *LoyaltyManager) DeductPoints(ctx context.Context, userID uint64, points int64, transactionType, description string, orderID uint64) error {
	account, err := m.repo.GetMemberAccount(ctx, userID)
	if err != nil {
		return err
	}
	if account == nil {
		return domain.ErrInsufficientPoints
	}

	if err := account.DeductPoints(points); err != nil {
		return err
	}

	if err := m.repo.SaveMemberAccount(ctx, account); err != nil {
		return err
	}

	tx := domain.NewPointsTransaction(userID, transactionType, -points, account.AvailablePoints, orderID, description, nil)
	return m.repo.SavePointsTransaction(ctx, tx)
}

func (m *LoyaltyManager) AddSpent(ctx context.Context, userID uint64, amount uint64) error {
	account, err := m.repo.GetMemberAccount(ctx, userID)
	if err != nil {
		return err
	}
	if account == nil {
		account = domain.NewMemberAccount(userID)
	}

	account.AddSpent(amount)
	return m.repo.SaveMemberAccount(ctx, account)
}

func (m *LoyaltyManager) AddBenefit(ctx context.Context, level domain.MemberLevel, name, description string, discountRate, pointsRate float64) (*domain.MemberBenefit, error) {
	benefit := domain.NewMemberBenefit(level, name, description, discountRate, pointsRate)
	if err := m.repo.SaveMemberBenefit(ctx, benefit); err != nil {
		m.logger.ErrorContext(ctx, "failed to save member benefit", "level", level, "error", err)
		return nil, err
	}
	return benefit, nil
}

func (m *LoyaltyManager) DeleteBenefit(ctx context.Context, id uint64) error {
	return m.repo.DeleteMemberBenefit(ctx, id)
}
