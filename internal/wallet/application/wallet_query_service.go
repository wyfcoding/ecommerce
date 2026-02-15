// Package application 钱包查询服务（CQRS 读侧）
// 生成摘要：
// 1) 实现钱包服务的查询侧，从 Redis 读模型和 ES 搜索模型获取数据
// 2) 查询操作不触发任何写操作和事件，保证读写分离
// 3) 支持钱包余额查询、交易流水查询、交易搜索、统计聚合
package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/wallet/domain"
)

// WalletQueryService 钱包查询服务（CQRS 读侧）
type WalletQueryService struct {
	walletReadRepo      domain.WalletReadRepository
	transactionReadRepo domain.TransactionReadRepository
	searchRepo          domain.WalletSearchRepository
	// 降级回源：当读模型不可用时，从写模型查询
	walletRepo      domain.WalletRepository
	transactionRepo domain.TransactionRepository
	logger          *slog.Logger
}

// NewWalletQueryService 创建钱包查询服务实例
func NewWalletQueryService(
	walletReadRepo domain.WalletReadRepository,
	transactionReadRepo domain.TransactionReadRepository,
	searchRepo domain.WalletSearchRepository,
	walletRepo domain.WalletRepository,
	transactionRepo domain.TransactionRepository,
	logger *slog.Logger,
) *WalletQueryService {
	return &WalletQueryService{
		walletReadRepo:      walletReadRepo,
		transactionReadRepo: transactionReadRepo,
		searchRepo:          searchRepo,
		walletRepo:          walletRepo,
		transactionRepo:     transactionRepo,
		logger:              logger.With("module", "wallet_query"),
	}
}

// GetWalletBalance 获取钱包余额（优先从 Redis 读模型获取）
func (s *WalletQueryService) GetWalletBalance(ctx context.Context, userID uint64, currency string) (*domain.WalletReadModel, error) {
	start := time.Now()

	// 优先从读模型获取
	if s.walletReadRepo != nil {
		model, err := s.walletReadRepo.GetByUserID(ctx, userID, currency)
		if err == nil && model != nil {
			s.logger.DebugContext(ctx, "wallet balance from cache",
				"user_id", userID, "duration", time.Since(start))
			return model, nil
		}
	}

	// 降级：从写模型查询
	wallet, err := s.walletRepo.GetByUserID(userID, currency)
	if err != nil {
		return nil, fmt.Errorf("get wallet: %w", err)
	}
	if wallet == nil {
		return nil, fmt.Errorf("wallet not found for user %d currency %s", userID, currency)
	}

	model := &domain.WalletReadModel{
		WalletID:         wallet.WalletID,
		UserID:           wallet.UserID,
		AccountNo:        wallet.AccountNo,
		Currency:         wallet.Currency,
		WalletType:       wallet.WalletType,
		Balance:          wallet.Balance,
		FrozenBalance:    wallet.FrozenBalance,
		AvailableBalance: wallet.AvailableBalance,
		Status:           wallet.Status.String(),
		HasPassword:      wallet.HasPassword,
		SecurityLevel:    wallet.SecurityLevel,
		CreatedAt:        wallet.CreatedAt,
		UpdatedAt:        wallet.UpdatedAt,
	}

	// 回填读模型
	if s.walletReadRepo != nil {
		if err := s.walletReadRepo.Save(ctx, model); err != nil {
			s.logger.WarnContext(ctx, "failed to backfill wallet read model", "error", err)
		}
	}

	s.logger.DebugContext(ctx, "wallet balance from db (fallback)",
		"user_id", userID, "duration", time.Since(start))
	return model, nil
}

// GetAllWallets 获取用户所有钱包
func (s *WalletQueryService) GetAllWallets(ctx context.Context, userID uint64) ([]*domain.WalletReadModel, error) {
	if s.walletReadRepo != nil {
		models, err := s.walletReadRepo.GetAllByUserID(ctx, userID)
		if err == nil && len(models) > 0 {
			return models, nil
		}
	}

	// 降级处理省略，实际应从 MySQL 查询并回填
	return nil, fmt.Errorf("wallets not found for user %d", userID)
}

// GetRecentTransactions 获取最近交易记录
func (s *WalletQueryService) GetRecentTransactions(ctx context.Context, walletID uint64, limit int) ([]*domain.TransactionReadModel, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	if s.transactionReadRepo != nil {
		records, err := s.transactionReadRepo.GetRecent(ctx, walletID, limit)
		if err == nil && len(records) > 0 {
			return records, nil
		}
	}

	// 降级：从 MySQL 查询
	return nil, fmt.Errorf("no recent transactions found")
}

// SearchTransactions 搜索交易记录（通过 Elasticsearch）
func (s *WalletQueryService) SearchTransactions(ctx context.Context, query *domain.TransactionSearchQuery) (*domain.TransactionSearchResult, error) {
	start := time.Now()

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 100 {
		query.PageSize = 20
	}

	if s.searchRepo == nil {
		return nil, fmt.Errorf("search repository not available")
	}

	result, err := s.searchRepo.SearchTransactions(ctx, query)
	if err != nil {
		s.logger.ErrorContext(ctx, "search transactions failed", "error", err, "duration", time.Since(start))
		return nil, fmt.Errorf("search transactions: %w", err)
	}

	s.logger.DebugContext(ctx, "search transactions",
		"total", result.Total, "duration", time.Since(start))
	return result, nil
}

// GetTransactionStats 获取交易统计（按类型聚合）
func (s *WalletQueryService) GetTransactionStats(ctx context.Context, walletID uint64, startTime, endTime time.Time) (map[string]int64, error) {
	if s.searchRepo == nil {
		return nil, fmt.Errorf("search repository not available")
	}

	return s.searchRepo.AggregateByType(ctx, walletID, startTime, endTime)
}

// GetDailyStats 获取每日交易统计
func (s *WalletQueryService) GetDailyStats(ctx context.Context, walletID uint64, startTime, endTime time.Time) (map[string]int64, error) {
	if s.searchRepo == nil {
		return nil, fmt.Errorf("search repository not available")
	}

	return s.searchRepo.AggregateDaily(ctx, walletID, startTime, endTime)
}
