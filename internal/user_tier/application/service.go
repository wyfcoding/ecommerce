package application

import (
	"context"
	"errors"
	"github.com/wyfcoding/ecommerce/internal/user_tier/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/user_tier/domain/repository"

	"log/slog"
)

type UserTierService struct {
	repo   repository.UserTierRepository
	logger *slog.Logger
}

func NewUserTierService(repo repository.UserTierRepository, logger *slog.Logger) *UserTierService {
	return &UserTierService{
		repo:   repo,
		logger: logger,
	}
}

// GetUserTier 获取用户等级
func (s *UserTierService) GetUserTier(ctx context.Context, userID uint64) (*entity.UserTier, error) {
	tier, err := s.repo.GetUserTier(ctx, userID)
	if err != nil {
		return nil, err
	}
	if tier == nil {
		// Initialize default tier
		config, _ := s.repo.GetTierConfig(ctx, entity.TierLevelRegular)
		tier = &entity.UserTier{
			UserID:    userID,
			Level:     entity.TierLevelRegular,
			LevelName: "Regular",
			Score:     0,
		}
		if config != nil {
			tier.LevelName = config.LevelName
			tier.DiscountRate = config.DiscountRate
		}
		_ = s.repo.SaveUserTier(ctx, tier)
	}
	return tier, nil
}

// AddScore 增加成长值并尝试升级
func (s *UserTierService) AddScore(ctx context.Context, userID uint64, score int64) error {
	tier, err := s.GetUserTier(ctx, userID)
	if err != nil {
		return err
	}

	tier.Score += score

	// Check for upgrade
	configs, err := s.repo.ListTierConfigs(ctx)
	if err == nil {
		for _, config := range configs {
			if tier.Score >= config.MinScore && config.Level > tier.Level {
				tier.Level = config.Level
				tier.LevelName = config.LevelName
				tier.DiscountRate = config.DiscountRate
			}
		}
	}

	return s.repo.SaveUserTier(ctx, tier)
}

// GetPoints 获取积分
func (s *UserTierService) GetPoints(ctx context.Context, userID uint64) (int64, error) {
	account, err := s.repo.GetPointsAccount(ctx, userID)
	if err != nil {
		return 0, err
	}
	if account == nil {
		return 0, nil
	}
	return account.Balance, nil
}

// AddPoints 增加积分
func (s *UserTierService) AddPoints(ctx context.Context, userID uint64, points int64, reason string) error {
	account, err := s.repo.GetPointsAccount(ctx, userID)
	if err != nil {
		return err
	}
	if account == nil {
		account = &entity.PointsAccount{
			UserID:  userID,
			Balance: 0,
		}
	}

	account.Balance += points
	if err := s.repo.SavePointsAccount(ctx, account); err != nil {
		return err
	}

	return s.repo.SavePointsLog(ctx, &entity.PointsLog{
		UserID: userID,
		Points: points,
		Reason: reason,
		Type:   "add",
	})
}

// DeductPoints 扣除积分
func (s *UserTierService) DeductPoints(ctx context.Context, userID uint64, points int64, reason string) error {
	account, err := s.repo.GetPointsAccount(ctx, userID)
	if err != nil {
		return err
	}
	if account == nil || account.Balance < points {
		return errors.New("insufficient points")
	}

	account.Balance -= points
	if err := s.repo.SavePointsAccount(ctx, account); err != nil {
		return err
	}

	return s.repo.SavePointsLog(ctx, &entity.PointsLog{
		UserID: userID,
		Points: points,
		Reason: reason,
		Type:   "deduct",
	})
}

// Exchange 兑换商品
func (s *UserTierService) Exchange(ctx context.Context, userID, exchangeID uint64) error {
	exchange, err := s.repo.GetExchange(ctx, exchangeID)
	if err != nil {
		return err
	}
	if exchange == nil {
		return errors.New("exchange item not found")
	}
	if exchange.Stock <= 0 {
		return errors.New("out of stock")
	}

	if err := s.DeductPoints(ctx, userID, exchange.RequiredPoints, "Exchange: "+exchange.Name); err != nil {
		return err
	}

	exchange.Stock--
	if err := s.repo.SaveExchange(ctx, exchange); err != nil {
		return err
	}

	return s.repo.SaveExchangeRecord(ctx, &entity.ExchangeRecord{
		UserID:     userID,
		ExchangeID: exchangeID,
		Points:     exchange.RequiredPoints,
	})
}

// ListExchanges 兑换列表
func (s *UserTierService) ListExchanges(ctx context.Context, page, pageSize int) ([]*entity.Exchange, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListExchanges(ctx, offset, pageSize)
}

// ListPointsLogs 积分日志
func (s *UserTierService) ListPointsLogs(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.PointsLog, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListPointsLogs(ctx, userID, offset, pageSize)
}
