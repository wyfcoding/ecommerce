package application

import (
	"context"
	"errors" // 导入标准错误处理库。

	// 导入格式化库。
	"github.com/wyfcoding/ecommerce/internal/user_tier/domain/entity"     // 导入用户等级领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/user_tier/domain/repository" // 导入用户等级领域的仓储接口。

	"log/slog" // 导入结构化日志库。
)

// UserTierService 结构体定义了用户等级和积分管理相关的应用服务。
// 它协调领域层和基础设施层，处理用户等级的获取、成长值的增加（可能触发等级升级）、
// 用户积分的管理（增加、扣除、兑换）以及相关的日志和兑换记录。
type UserTierService struct {
	repo   repository.UserTierRepository // 依赖UserTierRepository接口，用于数据持久化操作。
	logger *slog.Logger                  // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewUserTierService 创建并返回一个新的 UserTierService 实例。
func NewUserTierService(repo repository.UserTierRepository, logger *slog.Logger) *UserTierService {
	return &UserTierService{
		repo:   repo,
		logger: logger,
	}
}

// GetUserTier 获取指定用户的等级信息。
// 如果用户等级不存在，则会自动初始化一个默认的普通等级。
// ctx: 上下文。
// userID: 用户的唯一标识符。
// 返回UserTier实体和可能发生的错误。
func (s *UserTierService) GetUserTier(ctx context.Context, userID uint64) (*entity.UserTier, error) {
	tier, err := s.repo.GetUserTier(ctx, userID)
	if err != nil {
		return nil, err
	}
	if tier == nil {
		// 如果用户等级记录不存在，则初始化一个默认的“普通”等级。
		config, err := s.repo.GetTierConfig(ctx, entity.TierLevelRegular) // 获取普通等级的配置。
		if err != nil {
			s.logger.ErrorContext(ctx, "failed to get regular tier config", "error", err)
			return nil, err
		}
		// 创建一个新的用户等级实体。
		tier = &entity.UserTier{
			UserID:    userID,
			Level:     entity.TierLevelRegular, // 默认等级为普通。
			LevelName: "Regular",               // 默认等级名称。
			Score:     0,                       // 初始积分为0。
		}
		// 如果获取到配置，则更新等级名称和折扣率。
		if config != nil {
			tier.LevelName = config.LevelName
			tier.DiscountRate = config.DiscountRate
		}
		// 保存新创建的用户等级。
		_ = s.repo.SaveUserTier(ctx, tier) // 忽略保存错误，但实际生产应处理。
	}
	return tier, nil
}

// AddScore 增加用户的成长值，并根据成长值检查是否触发等级升级。
// ctx: 上下文。
// userID: 用户的唯一标识符。
// score: 待增加的成长值。
// 返回可能发生的错误。
func (s *UserTierService) AddScore(ctx context.Context, userID uint64, score int64) error {
	tier, err := s.GetUserTier(ctx, userID)
	if err != nil {
		return err
	}

	tier.Score += score // 增加成长值。

	// 检查是否满足等级升级条件。
	configs, err := s.repo.ListTierConfigs(ctx) // 获取所有等级配置。
	if err == nil {
		for _, config := range configs {
			// 如果用户当前分数达到更高等级的最低分数，且该等级高于当前等级，则进行升级。
			if tier.Score >= config.MinScore && config.Level > tier.Level {
				tier.Level = config.Level
				tier.LevelName = config.LevelName
				tier.DiscountRate = config.DiscountRate
			}
		}
	} else {
		s.logger.ErrorContext(ctx, "failed to list tier configs for upgrade check", "error", err)
		// 即使获取配置失败，也尝试保存当前分数。
	}

	// 保存更新后的用户等级。
	return s.repo.SaveUserTier(ctx, tier)
}

// GetPoints 获取指定用户的积分余额。
// ctx: 上下文。
// userID: 用户的唯一标识符。
// 返回积分余额和可能发生的错误。
func (s *UserTierService) GetPoints(ctx context.Context, userID uint64) (int64, error) {
	account, err := s.repo.GetPointsAccount(ctx, userID)
	if err != nil {
		return 0, err
	}
	if account == nil {
		return 0, nil // 如果积分账户不存在，则积分为0。
	}
	return account.Balance, nil
}

// AddPoints 增加指定用户的积分。
// ctx: 上下文。
// userID: 用户的唯一标识符。
// points: 待增加的积分数量。
// reason: 增加积分的原因。
// 返回可能发生的错误。
func (s *UserTierService) AddPoints(ctx context.Context, userID uint64, points int64, reason string) error {
	account, err := s.repo.GetPointsAccount(ctx, userID)
	if err != nil {
		return err
	}
	if account == nil {
		// 如果积分账户不存在，则创建一个新的账户。
		account = &entity.PointsAccount{
			UserID:  userID,
			Balance: 0,
		}
	}

	account.Balance += points // 增加积分余额。
	// 保存更新后的积分账户。
	if err := s.repo.SavePointsAccount(ctx, account); err != nil {
		return err
	}

	// 记录积分日志。
	return s.repo.SavePointsLog(ctx, &entity.PointsLog{
		UserID: userID,
		Points: points,
		Reason: reason,
		Type:   "add", // 记录类型为增加。
	})
}

// DeductPoints 扣除指定用户的积分。
// ctx: 上下文。
// userID: 用户的唯一标识符。
// points: 待扣除的积分数量。
// reason: 扣除积分的原因。
// 返回可能发生的错误。
func (s *UserTierService) DeductPoints(ctx context.Context, userID uint64, points int64, reason string) error {
	account, err := s.repo.GetPointsAccount(ctx, userID)
	if err != nil {
		return err
	}
	// 检查积分账户是否存在或积分是否充足。
	if account == nil || account.Balance < points {
		return errors.New("insufficient points")
	}

	account.Balance -= points // 扣除积分余额。
	// 保存更新后的积分账户。
	if err := s.repo.SavePointsAccount(ctx, account); err != nil {
		return err
	}

	// 记录积分日志。
	return s.repo.SavePointsLog(ctx, &entity.PointsLog{
		UserID: userID,
		Points: -points, // 扣除积分时，Points字段为负数。
		Reason: reason,
		Type:   "deduct", // 记录类型为扣除。
	})
}

// Exchange 兑换商品（使用积分）。
// ctx: 上下文。
// userID: 兑换用户ID。
// exchangeID: 兑换商品ID。
// 返回可能发生的错误。
func (s *UserTierService) Exchange(ctx context.Context, userID, exchangeID uint64) error {
	// 1. 获取兑换商品信息。
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

	// 2. 扣除用户积分。
	if err := s.DeductPoints(ctx, userID, exchange.RequiredPoints, "Exchange: "+exchange.Name); err != nil {
		return err
	}

	// 3. 扣减兑换商品库存。
	exchange.Stock--
	if err := s.repo.SaveExchange(ctx, exchange); err != nil {
		// TODO: 如果这里失败，需要回滚积分扣除操作。
		return err
	}

	// 4. 记录兑换记录。
	return s.repo.SaveExchangeRecord(ctx, &entity.ExchangeRecord{
		UserID:     userID,
		ExchangeID: exchangeID,
		Points:     exchange.RequiredPoints,
	})
}

// ListExchanges 获取可兑换商品列表。
// ctx: 上下文。
// page, pageSize: 分页参数。
// 返回兑换商品列表、总数和可能发生的错误。
func (s *UserTierService) ListExchanges(ctx context.Context, page, pageSize int) ([]*entity.Exchange, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListExchanges(ctx, offset, pageSize)
}

// ListPointsLogs 获取用户积分日志列表。
// ctx: 上下文。
// userID: 用户ID。
// page, pageSize: 分页参数。
// 返回积分日志列表、总数和可能发生的错误。
func (s *UserTierService) ListPointsLogs(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.PointsLog, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListPointsLogs(ctx, userID, offset, pageSize)
}
