package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/usertier/domain"
)

// UserTierManager 处理所有用户等级与积分相关的写入操作（Commands）。
type UserTierManager struct {
	repo   domain.UserTierRepository
	logger *slog.Logger
}

// NewUserTierManager 构造函数。
func NewUserTierManager(repo domain.UserTierRepository, logger *slog.Logger) *UserTierManager {
	return &UserTierManager{
		repo:   repo,
		logger: logger,
	}
}

// EnsureUserTier 确保用户等级实体存在。
func (m *UserTierManager) EnsureUserTier(ctx context.Context, userID uint64) (*domain.UserTier, error) {
	tier, err := m.repo.GetUserTier(ctx, userID)
	if err != nil {
		return nil, err
	}
	if tier == nil {
		tier = &domain.UserTier{
			UserID:       userID,
			Level:        0,
			LevelName:    "普通会员",
			DiscountRate: 100,
		}
		if err := m.repo.SaveUserTier(ctx, tier); err != nil {
			return nil, err
		}
	}
	return tier, nil
}

func (m *UserTierManager) AddScore(ctx context.Context, userID uint64, score int64) error {
	tier, err := m.EnsureUserTier(ctx, userID)
	if err != nil {
		return err
	}

	tier.Score += score

	// 升级检查
	configs, err := m.repo.ListTierConfigs(ctx)
	if err != nil {
		return err
	}

	var targetLevel domain.TierLevel = 0
	targetName := "普通会员"
	targetDiscount := 100.0

	for _, cfg := range configs {
		if tier.Score >= cfg.MinScore && cfg.Level > targetLevel {
			targetLevel = cfg.Level
			targetName = cfg.LevelName
			targetDiscount = cfg.DiscountRate
		}
	}

	if targetLevel > tier.Level {
		tier.Level = targetLevel
		tier.LevelName = targetName
		tier.DiscountRate = targetDiscount
		m.logger.InfoContext(ctx, "user tier upgraded", "user_id", userID, "new_level", targetLevel)
	}

	return m.repo.SaveUserTier(ctx, tier)
}

func (m *UserTierManager) AddPoints(ctx context.Context, userID uint64, points int64, reason string) error {
	acc, err := m.repo.GetPointsAccount(ctx, userID)
	if err != nil {
		return err
	}
	if acc == nil {
		acc = &domain.PointsAccount{UserID: userID, Balance: 0}
	}

	acc.Balance += points
	if err := m.repo.SavePointsAccount(ctx, acc); err != nil {
		return err
	}

	log := &domain.PointsLog{
		UserID: userID,
		Points: points,
		Reason: reason,
		Type:   "add",
	}
	return m.repo.SavePointsLog(ctx, log)
}

func (m *UserTierManager) DeductPoints(ctx context.Context, userID uint64, points int64, reason string) error {
	acc, err := m.repo.GetPointsAccount(ctx, userID)
	if err != nil {
		return err
	}
	if acc == nil || acc.Balance < points {
		return errors.New("insufficient points")
	}

	acc.Balance -= points
	if err := m.repo.SavePointsAccount(ctx, acc); err != nil {
		return err
	}

	log := &domain.PointsLog{
		UserID: userID,
		Points: -points,
		Reason: reason,
		Type:   "deduct",
	}
	return m.repo.SavePointsLog(ctx, log)
}

func (m *UserTierManager) Exchange(ctx context.Context, userID uint64, exchangeID uint64) error {
	item, err := m.repo.GetExchange(ctx, exchangeID)
	if err != nil {
		return err
	}
	if item == nil {
		return errors.New("exchange item not found")
	}
	if item.Stock <= 0 {
		return errors.New("out of stock")
	}

	if err := m.DeductPoints(ctx, userID, item.RequiredPoints, fmt.Sprintf("Exchange: %s", item.Name)); err != nil {
		return err
	}

	item.Stock--
	if err := m.repo.SaveExchange(ctx, item); err != nil {
		return err
	}

	record := &domain.ExchangeRecord{
		UserID:     userID,
		ExchangeID: exchangeID,
		Points:     item.RequiredPoints,
	}
	return m.repo.SaveExchangeRecord(ctx, record)
}
