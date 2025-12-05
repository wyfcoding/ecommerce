package application

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/loyalty/domain/entity"     // 导入忠诚度领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/loyalty/domain/repository" // 导入忠诚度领域的仓储接口。

	"log/slog" // 导入结构化日志库。
)

// LoyaltyService 结构体定义了忠诚度计划相关的应用服务。
// 它协调领域层和基础设施层，处理会员账户、积分管理、消费记录和会员权益等业务逻辑。
type LoyaltyService struct {
	repo   repository.LoyaltyRepository // 依赖LoyaltyRepository接口，用于数据持久化操作。
	logger *slog.Logger                 // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewLoyaltyService 创建并返回一个新的 LoyaltyService 实例。
func NewLoyaltyService(repo repository.LoyaltyRepository, logger *slog.Logger) *LoyaltyService {
	return &LoyaltyService{
		repo:   repo,
		logger: logger,
	}
}

// GetOrCreateAccount 获取或创建一个会员账户。
// 如果指定用户ID的会员账户不存在，则会自动创建一个新的账户。
// ctx: 上下文。
// userID: 用户的唯一标识符。
// 返回会员账户实体和可能发生的错误。
func (s *LoyaltyService) GetOrCreateAccount(ctx context.Context, userID uint64) (*entity.MemberAccount, error) {
	account, err := s.repo.GetMemberAccount(ctx, userID)
	if err != nil {
		return nil, err
	}
	// 如果账户不存在，则创建新账户。
	if account == nil {
		account = entity.NewMemberAccount(userID)
		if err := s.repo.SaveMemberAccount(ctx, account); err != nil {
			s.logger.Error("failed to create member account", "error", err)
			return nil, err
		}
	}
	return account, nil
}

// AddPoints 增加用户积分。
// ctx: 上下文。
// userID: 用户的唯一标识符。
// points: 待增加的积分数量。
// transactionType: 积分交易类型（例如，“购买”，“活动”）。
// description: 积分交易描述。
// orderID: 关联的订单ID（如果适用）。
// 返回可能发生的错误。
func (s *LoyaltyService) AddPoints(ctx context.Context, userID uint64, points int64, transactionType, description string, orderID uint64) error {
	// 获取或创建会员账户。
	account, err := s.GetOrCreateAccount(ctx, userID)
	if err != nil {
		return err
	}

	// 调用实体方法增加积分。
	account.AddPoints(points)
	// 保存更新后的会员账户。
	if err := s.repo.SaveMemberAccount(ctx, account); err != nil {
		return err
	}

	// 记录积分交易。
	tx := entity.NewPointsTransaction(userID, transactionType, points, account.AvailablePoints, orderID, description, nil)
	return s.repo.SavePointsTransaction(ctx, tx)
}

// DeductPoints 扣减用户积分。
// ctx: 上下文。
// userID: 用户的唯一标识符。
// points: 待扣减的积分数量。
// transactionType: 积分交易类型（例如，“兑换”，“退款”）。
// description: 积分交易描述。
// orderID: 关联的订单ID（如果适用）。
// 返回可能发生的错误。
func (s *LoyaltyService) DeductPoints(ctx context.Context, userID uint64, points int64, transactionType, description string, orderID uint64) error {
	// 获取或创建会员账户。
	account, err := s.GetOrCreateAccount(ctx, userID)
	if err != nil {
		return err
	}

	// 调用实体方法扣减积分。
	if err := account.DeductPoints(points); err != nil {
		return err
	}

	// 保存更新后的会员账户。
	if err := s.repo.SaveMemberAccount(ctx, account); err != nil {
		return err
	}

	// 记录积分交易。
	tx := entity.NewPointsTransaction(userID, transactionType, -points, account.AvailablePoints, orderID, description, nil)
	return s.repo.SavePointsTransaction(ctx, tx)
}

// AddSpent 增加用户总消费金额。
// ctx: 上下文。
// userID: 用户的唯一标识符。
// amount: 待增加的消费金额。
// 返回可能发生的错误。
func (s *LoyaltyService) AddSpent(ctx context.Context, userID uint64, amount uint64) error {
	// 获取或创建会员账户。
	account, err := s.GetOrCreateAccount(ctx, userID)
	if err != nil {
		return err
	}

	// 调用实体方法增加消费金额。
	account.AddSpent(amount)
	// 保存更新后的会员账户。
	return s.repo.SaveMemberAccount(ctx, account)
}

// GetPointsTransactions 获取指定用户的积分交易记录。
// ctx: 上下文。
// userID: 用户的唯一标识符。
// page, pageSize: 分页参数。
// 返回积分交易列表、总数和可能发生的错误。
func (s *LoyaltyService) GetPointsTransactions(ctx context.Context, userID uint64, page, pageSize int) ([]*entity.PointsTransaction, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListPointsTransactions(ctx, userID, offset, pageSize)
}

// AddBenefit 添加会员等级权益。
// ctx: 上下文。
// level: 会员等级。
// name: 权益名称。
// description: 权益描述。
// discountRate: 折扣率。
// pointsRate: 积分倍率。
// 返回created successfully的MemberBenefit实体和可能发生的错误。
func (s *LoyaltyService) AddBenefit(ctx context.Context, level entity.MemberLevel, name, description string, discountRate, pointsRate float64) (*entity.MemberBenefit, error) {
	benefit := entity.NewMemberBenefit(level, name, description, discountRate, pointsRate) // 创建MemberBenefit实体。
	if err := s.repo.SaveMemberBenefit(ctx, benefit); err != nil {
		s.logger.Error("failed to save member benefit", "error", err)
		return nil, err
	}
	return benefit, nil
}

// ListBenefits 获取会员权益列表。
// ctx: 上下文。
// level: 筛选会员等级的权益。
// 返回会员权益列表和可能发生的错误。
func (s *LoyaltyService) ListBenefits(ctx context.Context, level entity.MemberLevel) ([]*entity.MemberBenefit, error) {
	return s.repo.ListMemberBenefits(ctx, level)
}

// DeleteBenefit 删除指定ID的会员权益。
// ctx: 上下文。
// id: 会员权益ID。
// 返回可能发生的错误。
func (s *LoyaltyService) DeleteBenefit(ctx context.Context, id uint64) error {
	return s.repo.DeleteMemberBenefit(ctx, id)
}
