package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/usertier/domain"
)

// UserTierService 结构体定义了用户等级与积分相关的应用服务。
type UserTierService struct {
	repo   domain.UserTierRepository
	logger *slog.Logger
}

// NewUserTierService 创建并返回一个新的 UserTierService 实例。
func NewUserTierService(repo domain.UserTierRepository, logger *slog.Logger) *UserTierService {
	return &UserTierService{
		repo:   repo,
		logger: logger,
	}
}

// --- 用户等级 (User Tier) ---

func (s *UserTierService) GetUserTier(ctx context.Context, userID uint64) (*domain.UserTier, error) {
	tier, err := s.repo.GetUserTier(ctx, userID)
	if err != nil {
		return nil, err
	}
	if tier == nil {
		// 默认初始化
		tier = &domain.UserTier{
			UserID:       userID,
			Level:        0,
			LevelName:    "普通会员",
			DiscountRate: 100,
		}
		if err := s.repo.SaveUserTier(ctx, tier); err != nil {
			return nil, err
		}
	}
	return tier, nil
}

func (s *UserTierService) AddScore(ctx context.Context, userID uint64, score int64) error {
	tier, err := s.GetUserTier(ctx, userID)
	if err != nil {
		return err
	}

	tier.Score += score

	// 升级检查
	configs, err := s.repo.ListTierConfigs(ctx)
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
		s.logger.InfoContext(ctx, "user tier upgraded", "user_id", userID, "new_level", targetLevel)
	}

	return s.repo.SaveUserTier(ctx, tier)
}

// --- 积分 (Points) ---

func (s *UserTierService) GetPoints(ctx context.Context, userID uint64) (int64, error) {
	acc, err := s.repo.GetPointsAccount(ctx, userID)
	if err != nil {
		return 0, err
	}
	if acc == nil {
		return 0, nil
	}
	return acc.Balance, nil
}

func (s *UserTierService) AddPoints(ctx context.Context, userID uint64, points int64, reason string) error {
	acc, err := s.repo.GetPointsAccount(ctx, userID)
	if err != nil {
		return err
	}
	if acc == nil {
		acc = &domain.PointsAccount{UserID: userID, Balance: 0}
	}

	acc.Balance += points
	if err := s.repo.SavePointsAccount(ctx, acc); err != nil {
		return err
	}

	log := &domain.PointsLog{
		UserID: userID,
		Points: points,
		Reason: reason,
		Type:   "add",
	}
	return s.repo.SavePointsLog(ctx, log)
}

func (s *UserTierService) DeductPoints(ctx context.Context, userID uint64, points int64, reason string) error {
	acc, err := s.repo.GetPointsAccount(ctx, userID)
	if err != nil {
		return err
	}
	if acc == nil || acc.Balance < points {
		return errors.New("insufficient points")
	}

	acc.Balance -= points
	if err := s.repo.SavePointsAccount(ctx, acc); err != nil {
		return err
	}

	log := &domain.PointsLog{
		UserID: userID,
		Points: -points,
		Reason: reason,
		Type:   "deduct",
	}
	return s.repo.SavePointsLog(ctx, log)
}

func (s *UserTierService) ListPointsLogs(ctx context.Context, userID uint64, page, pageSize int) ([]*domain.PointsLog, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListPointsLogs(ctx, userID, offset, pageSize)
}

// --- 兑换 (Exchange) ---

func (s *UserTierService) ListExchanges(ctx context.Context, page, pageSize int) ([]*domain.Exchange, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListExchanges(ctx, offset, pageSize)
}

func (s *UserTierService) Exchange(ctx context.Context, userID uint64, exchangeID uint64) error {
	item, err := s.repo.GetExchange(ctx, exchangeID)
	if err != nil {
		return err
	}
	if item == nil {
		return errors.New("exchange item not found")
	}
	if item.Stock <= 0 {
		return errors.New("out of stock")
	}

	// 扣除积分
	if err := s.DeductPoints(ctx, userID, item.RequiredPoints, fmt.Sprintf("Exchange: %s", item.Name)); err != nil {
		return err
	}

	// 更新库存
	item.Stock--
	if err := s.repo.SaveExchange(ctx, item); err != nil {
		return err
	}

	// 记录兑换
	record := &domain.ExchangeRecord{
		UserID:     userID,
		ExchangeID: exchangeID,
		Points:     item.RequiredPoints,
	}
	return s.repo.SaveExchangeRecord(ctx, record)
}
