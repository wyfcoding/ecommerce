package repository

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/settlement/domain/entity" // 导入结算领域的实体定义。
)

// SettlementRepository 是结算模块的仓储接口。
// 它定义了对结算单、结算明细和商户账户实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type SettlementRepository interface {
	// --- 结算单管理 (Settlement methods) ---

	// SaveSettlement 将结算单实体保存到数据存储中。
	// 如果结算单已存在，则更新；如果不存在，则创建。
	// ctx: 上下文。
	// settlement: 待保存的结算单实体。
	SaveSettlement(ctx context.Context, settlement *entity.Settlement) error
	// GetSettlement 根据ID获取结算单实体。
	GetSettlement(ctx context.Context, id uint64) (*entity.Settlement, error)
	// GetSettlementByNo 根据结算单号获取结算单实体。
	GetSettlementByNo(ctx context.Context, no string) (*entity.Settlement, error)
	// ListSettlements 列出指定商户ID的所有结算单实体，支持通过状态过滤和分页。
	ListSettlements(ctx context.Context, merchantID uint64, status *entity.SettlementStatus, offset, limit int) ([]*entity.Settlement, int64, error)

	// --- 结算明细管理 (SettlementDetail methods) ---

	// SaveSettlementDetail 将结算明细实体保存到数据存储中。
	SaveSettlementDetail(ctx context.Context, detail *entity.SettlementDetail) error
	// ListSettlementDetails 列出指定结算单ID的所有结算明细实体。
	ListSettlementDetails(ctx context.Context, settlementID uint64) ([]*entity.SettlementDetail, error)

	// --- 商户账户管理 (MerchantAccount methods) ---

	// GetMerchantAccount 根据商户ID获取商户账户实体。
	GetMerchantAccount(ctx context.Context, merchantID uint64) (*entity.MerchantAccount, error)
	// SaveMerchantAccount 将商户账户实体保存到数据存储中。
	SaveMerchantAccount(ctx context.Context, account *entity.MerchantAccount) error
}
