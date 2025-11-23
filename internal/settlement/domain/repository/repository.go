package repository

import (
	"context"
	"ecommerce/internal/settlement/domain/entity"
)

// SettlementRepository 结算仓储接口
type SettlementRepository interface {
	// 结算单
	SaveSettlement(ctx context.Context, settlement *entity.Settlement) error
	GetSettlement(ctx context.Context, id uint64) (*entity.Settlement, error)
	GetSettlementByNo(ctx context.Context, no string) (*entity.Settlement, error)
	ListSettlements(ctx context.Context, merchantID uint64, status *entity.SettlementStatus, offset, limit int) ([]*entity.Settlement, int64, error)

	// 结算明细
	SaveSettlementDetail(ctx context.Context, detail *entity.SettlementDetail) error
	ListSettlementDetails(ctx context.Context, settlementID uint64) ([]*entity.SettlementDetail, error)

	// 商户账户
	GetMerchantAccount(ctx context.Context, merchantID uint64) (*entity.MerchantAccount, error)
	SaveMerchantAccount(ctx context.Context, account *entity.MerchantAccount) error
}
