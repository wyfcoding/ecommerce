// Package domain 钱包读模型仓储接口（CQRS 查询侧）
// 生成摘要：
// 1) 定义钱包服务的读模型仓储接口，用于 CQRS 查询侧
// 2) 读模型基于 Redis 缓存，提供高性能的钱包余额查询、交易流水查询
// 3) 与写模型（MySQL）解耦，通过事件投影保持最终一致性
package domain

import (
	"context"
	"time"
)

// WalletReadModel 钱包读模型（用于快速查询的扁平化视图）
type WalletReadModel struct {
	WalletID         uint64    `json:"wallet_id"`
	UserID           uint64    `json:"user_id"`
	AccountNo        string    `json:"account_no"`
	Currency         string    `json:"currency"`
	WalletType       string    `json:"wallet_type"`
	Balance          int64     `json:"balance"`
	FrozenBalance    int64     `json:"frozen_balance"`
	AvailableBalance int64     `json:"available_balance"`
	Status           string    `json:"status"`
	HasPassword      bool      `json:"has_password"`
	SecurityLevel    int       `json:"security_level"`
	TodayDeposit     int64     `json:"today_deposit"`
	TodayWithdraw    int64     `json:"today_withdraw"`
	TodayTransfer    int64     `json:"today_transfer"`
	TodayTxCount     int       `json:"today_tx_count"`
	LastTxTime       time.Time `json:"last_tx_time"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TransactionReadModel 交易记录读模型
type TransactionReadModel struct {
	ID               uint64    `json:"id"`
	TransactionNo    string    `json:"transaction_no"`
	WalletID         uint64    `json:"wallet_id"`
	UserID           uint64    `json:"user_id"`
	Type             string    `json:"type"`
	Amount           int64     `json:"amount"`
	BalanceBefore    int64     `json:"balance_before"`
	BalanceAfter     int64     `json:"balance_after"`
	Fee              int64     `json:"fee"`
	Status           string    `json:"status"`
	Remark           string    `json:"remark"`
	CounterpartyID   uint64    `json:"counterparty_id,omitempty"`
	CounterpartyType string    `json:"counterparty_type,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// BalanceSummaryReadModel 余额汇总读模型（多币种）
type BalanceSummaryReadModel struct {
	UserID   uint64                `json:"user_id"`
	Wallets  []*WalletReadModel    `json:"wallets"`
	TotalCNY int64                 `json:"total_cny"`
}

// WalletReadRepository 钱包读模型仓储接口
type WalletReadRepository interface {
	// GetByWalletID 根据钱包ID获取读模型。
	GetByWalletID(ctx context.Context, walletID uint64) (*WalletReadModel, error)
	// GetByUserID 根据用户ID和币种获取读模型。
	GetByUserID(ctx context.Context, userID uint64, currency string) (*WalletReadModel, error)
	// GetAllByUserID 获取用户所有钱包的读模型。
	GetAllByUserID(ctx context.Context, userID uint64) ([]*WalletReadModel, error)
	// Save 保存或更新读模型。
	Save(ctx context.Context, model *WalletReadModel) error
	// Delete 删除读模型。
	Delete(ctx context.Context, walletID uint64) error
}

// TransactionReadRepository 交易记录读模型仓储接口
type TransactionReadRepository interface {
	// GetRecent 获取最近的交易记录（用于首页展示）。
	GetRecent(ctx context.Context, walletID uint64, limit int) ([]*TransactionReadModel, error)
	// Save 保存交易记录到读模型。
	Save(ctx context.Context, model *TransactionReadModel) error
}
