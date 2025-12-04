package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/loyalty/domain/entity" // 导入忠诚度领域的实体定义。
)

// LoyaltyRepository 是忠诚度模块的仓储接口。
// 它定义了对会员账户、积分交易和会员权益实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type LoyaltyRepository interface {
	// --- 会员账户 (MemberAccount methods) ---

	// SaveMemberAccount 将会员账户实体保存到数据存储中。
	// 如果账户已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// account: 待保存的会员账户实体。
	SaveMemberAccount(ctx context.Context, account *entity.MemberAccount) error
	// GetMemberAccount 根据用户ID获取会员账户实体。
	GetMemberAccount(ctx context.Context, userID uint64) (*entity.MemberAccount, error)

	// --- 积分交易 (PointsTransaction methods) ---

	// SavePointsTransaction 将积分交易实体保存到数据存储中。
	SavePointsTransaction(ctx context.Context, transaction *entity.PointsTransaction) error
	// ListPointsTransactions 列出指定用户ID的所有积分交易实体，支持分页。
	ListPointsTransactions(ctx context.Context, userID uint64, offset, limit int) ([]*entity.PointsTransaction, int64, error)

	// --- 会员权益 (MemberBenefit methods) ---

	// SaveMemberBenefit 将会员权益实体保存到数据存储中。
	SaveMemberBenefit(ctx context.Context, benefit *entity.MemberBenefit) error
	// GetMemberBenefit 根据ID获取会员权益实体。
	GetMemberBenefit(ctx context.Context, id uint64) (*entity.MemberBenefit, error)
	// ListMemberBenefits 列出指定会员等级的所有权益实体。
	ListMemberBenefits(ctx context.Context, level entity.MemberLevel) ([]*entity.MemberBenefit, error)
	// DeleteMemberBenefit 根据ID删除会员权益实体。
	DeleteMemberBenefit(ctx context.Context, id uint64) error
}
