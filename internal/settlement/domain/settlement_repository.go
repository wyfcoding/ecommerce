package domain

import (
	"context"
)

// SettlementRepository 是结算模块的仓储接口。
type SettlementRepository interface {
	// 事务支持
	BeginTx(ctx context.Context) any
	CommitTx(tx any) error
	RollbackTx(tx any) error
	WithTx(ctx context.Context, fn func(tx any) error) error

	// --- 结算单管理 (Settlement methods) ---

	SaveSettlement(ctx context.Context, settlement *Settlement) error
	SaveSettlementInTx(ctx context.Context, tx any, settlement *Settlement) error
	GetSettlement(ctx context.Context, id uint64) (*Settlement, error)
	GetSettlementByNo(ctx context.Context, no string) (*Settlement, error)
	ListSettlements(ctx context.Context, merchantID uint64, status *SettlementStatus, offset, limit int) ([]*Settlement, int64, error)

	// --- 结算明细管理 (SettlementDetail methods) ---

	SaveSettlementDetail(ctx context.Context, detail *SettlementDetail) error
	SaveSettlementDetailInTx(ctx context.Context, tx any, detail *SettlementDetail) error
	ListSettlementDetails(ctx context.Context, settlementID uint64) ([]*SettlementDetail, error)

	// --- 商户账户管理 (MerchantAccount methods) ---

	GetMerchantAccount(ctx context.Context, merchantID uint64) (*MerchantAccount, error)
	SaveMerchantAccount(ctx context.Context, account *MerchantAccount) error
	SaveMerchantAccountInTx(ctx context.Context, tx any, account *MerchantAccount) error
}
