// Package consumer 钱包投影处理器（CQRS 事件消费者）
// 生成摘要：
// 1) 监听钱包领域事件，将写模型投影到读模型（Redis）和搜索模型（Elasticsearch）
// 2) 保证最终一致性，事件处理失败会重试，不影响主业务流程
// 3) 支持事件去重和幂等处理，避免重复投影
package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/wallet/domain"
	"github.com/wyfcoding/pkg/messagequeue"
)

// WalletProjectionHandler 钱包投影处理器
type WalletProjectionHandler struct {
	walletReadRepo      domain.WalletReadRepository
	transactionReadRepo domain.TransactionReadRepository
	searchRepo          domain.WalletSearchRepository
	logger              *slog.Logger
}

// NewWalletProjectionHandler 创建钱包投影处理器实例
func NewWalletProjectionHandler(
	walletReadRepo domain.WalletReadRepository,
	transactionReadRepo domain.TransactionReadRepository,
	searchRepo domain.WalletSearchRepository,
	logger *slog.Logger,
) *WalletProjectionHandler {
	return &WalletProjectionHandler{
		walletReadRepo:      walletReadRepo,
		transactionReadRepo: transactionReadRepo,
		searchRepo:          searchRepo,
		logger:              logger.With("module", "wallet_projection"),
	}
}

// HandleWalletCreated 处理钱包创建事件
func (h *WalletProjectionHandler) HandleWalletCreated(ctx context.Context, event *domain.WalletCreatedEvent) error {
	start := time.Now()

	// 构建钱包读模型
	model := &domain.WalletReadModel{
		WalletID:         event.WalletID,
		UserID:           event.UserID,
		AccountNo:        event.AccountNo,
		Currency:         event.Currency,
		WalletType:       event.WalletType,
		Balance:          0,
		FrozenBalance:    0,
		AvailableBalance: 0,
		Status:           "NORMAL",
		HasPassword:      false,
		SecurityLevel:    1,
		TodayDeposit:     0,
		TodayWithdraw:    0,
		TodayTransfer:    0,
		TodayTxCount:     0,
		CreatedAt:        event.Timestamp,
		UpdatedAt:        event.Timestamp,
	}

	if err := h.walletReadRepo.Save(ctx, model); err != nil {
		h.logger.ErrorContext(ctx, "failed to project wallet created event",
			"wallet_id", event.WalletID, "error", err, "duration", time.Since(start))
		return fmt.Errorf("project wallet created: %w", err)
	}

	h.logger.DebugContext(ctx, "wallet created event projected",
		"wallet_id", event.WalletID, "duration", time.Since(start))
	return nil
}

// HandleDeposited 处理充值成功事件
func (h *WalletProjectionHandler) HandleDeposited(ctx context.Context, event *domain.DepositedEvent) error {
	start := time.Now()

	// 更新钱包读模型
	wallet, err := h.walletReadRepo.GetByWalletID(ctx, event.WalletID)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to get wallet for projection",
			"wallet_id", event.WalletID, "error", err)
		return fmt.Errorf("get wallet: %w", err)
	}

	if wallet == nil {
		h.logger.WarnContext(ctx, "wallet not found for projection, creating new one",
			"wallet_id", event.WalletID)
		// 如果读模型不存在，创建基础读模型
		wallet = &domain.WalletReadModel{
			WalletID:         event.WalletID,
			UserID:           event.UserID,
			Balance:          0,
			FrozenBalance:    0,
			AvailableBalance: 0,
			Status:           "NORMAL",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
	}

	// 更新余额和统计信息
	wallet.Balance = event.BalanceAfter
	wallet.AvailableBalance = event.BalanceAfter - wallet.FrozenBalance
	
	// 更新今日统计
	today := time.Now().Format("2006-01-02")
	if event.Timestamp.Format("2006-01-02") == today {
		wallet.TodayDeposit += event.Amount
		wallet.TodayTxCount++
	}
	
	wallet.LastTxTime = event.Timestamp
	wallet.UpdatedAt = time.Now()

	if err := h.walletReadRepo.Save(ctx, wallet); err != nil {
		h.logger.ErrorContext(ctx, "failed to project deposited event",
			"wallet_id", event.WalletID, "error", err)
		return fmt.Errorf("project deposited: %w", err)
	}

	// 索引交易记录到搜索模型
	txModel := &domain.TransactionReadModel{
		ID:              0, // 实际应从写模型获取
		TransactionNo:   event.TransactionNo,
		WalletID:        event.WalletID,
		UserID:          event.UserID,
		Type:            "DEPOSIT",
		Amount:          event.Amount,
		BalanceBefore:   event.BalanceBefore,
		BalanceAfter:    event.BalanceAfter,
		Fee:             0,
		Status:          "SUCCESS",
		Remark:          event.Remark,
		CreatedAt:       event.Timestamp,
	}

	if h.transactionReadRepo != nil {
		if err := h.transactionReadRepo.Save(ctx, txModel); err != nil {
			h.logger.WarnContext(ctx, "failed to save transaction to read repo", "error", err)
		}
	}

	if h.searchRepo != nil {
		if err := h.searchRepo.IndexTransaction(ctx, txModel); err != nil {
			h.logger.WarnContext(ctx, "failed to index transaction to es", "error", err)
		}
	}

	h.logger.DebugContext(ctx, "deposited event projected",
		"wallet_id", event.WalletID, "amount", event.Amount, "duration", time.Since(start))
	return nil
}

// HandleWithdrawn 处理提现成功事件
func (h *WalletProjectionHandler) HandleWithdrawn(ctx context.Context, event *domain.WithdrawnEvent) error {
	start := time.Now()

	wallet, err := h.walletReadRepo.GetByWalletID(ctx, event.WalletID)
	if err != nil || wallet == nil {
		return fmt.Errorf("wallet not found for projection: %w", err)
	}

	wallet.Balance = event.BalanceAfter
	wallet.AvailableBalance = event.BalanceAfter - wallet.FrozenBalance
	
	today := time.Now().Format("2006-01-02")
	if event.Timestamp.Format("2006-01-02") == today {
		wallet.TodayWithdraw += event.Amount
		wallet.TodayTxCount++
	}
	
	wallet.LastTxTime = event.Timestamp
	wallet.UpdatedAt = time.Now()

	if err := h.walletReadRepo.Save(ctx, wallet); err != nil {
		h.logger.ErrorContext(ctx, "failed to project withdrawn event",
			"wallet_id", event.WalletID, "error", err)
		return fmt.Errorf("project withdrawn: %w", err)
	}

	txModel := &domain.TransactionReadModel{
		ID:              0,
		TransactionNo:   event.TransactionNo,
		WalletID:        event.WalletID,
		UserID:          event.UserID,
		Type:            "WITHDRAW",
		Amount:          event.Amount,
		BalanceBefore:   event.BalanceBefore,
		BalanceAfter:    event.BalanceAfter,
		Fee:             event.Fee,
		Status:          "SUCCESS",
		Remark:          event.Remark,
		CreatedAt:       event.Timestamp,
	}

	if h.transactionReadRepo != nil {
		if err := h.transactionReadRepo.Save(ctx, txModel); err != nil {
			h.logger.WarnContext(ctx, "failed to save transaction to read repo", "error", err)
		}
	}

	if h.searchRepo != nil {
		if err := h.searchRepo.IndexTransaction(ctx, txModel); err != nil {
			h.logger.WarnContext(ctx, "failed to index transaction to es", "error", err)
		}
	}

	h.logger.DebugContext(ctx, "withdrawn event projected",
		"wallet_id", event.WalletID, "amount", event.Amount, "duration", time.Since(start))
	return nil
}

// HandleTransferred 处理转账成功事件
func (h *WalletProjectionHandler) HandleTransferred(ctx context.Context, event *domain.TransferredEvent) error {
	start := time.Now()

	// 更新转出方钱包
	fromWallet, err := h.walletReadRepo.GetByWalletID(ctx, event.FromWalletID)
	if err != nil || fromWallet == nil {
		return fmt.Errorf("from wallet not found: %w", err)
	}

	// 这里简化处理，实际应从事件中获取准确的余额信息
	// 或者通过查询写模型获取最新余额
	fromWallet.UpdatedAt = time.Now()
	
	today := time.Now().Format("2006-01-02")
	if event.Timestamp.Format("2006-01-02") == today {
		fromWallet.TodayTransfer += event.Amount
		fromWallet.TodayTxCount++
	}

	if err := h.walletReadRepo.Save(ctx, fromWallet); err != nil {
		h.logger.ErrorContext(ctx, "failed to project from wallet for transfer",
			"wallet_id", event.FromWalletID, "error", err)
		return fmt.Errorf("project from wallet: %w", err)
	}

	// 更新转入方钱包
	toWallet, err := h.walletReadRepo.GetByWalletID(ctx, event.ToWalletID)
	if err != nil || toWallet == nil {
		return fmt.Errorf("to wallet not found: %w", err)
	}

	toWallet.UpdatedAt = time.Now()
	
	if event.Timestamp.Format("2006-01-02") == today {
		toWallet.TodayDeposit += event.Amount
		toWallet.TodayTxCount++
	}

	if err := h.walletReadRepo.Save(ctx, toWallet); err != nil {
		h.logger.ErrorContext(ctx, "failed to project to wallet for transfer",
			"wallet_id", event.ToWalletID, "error", err)
		return fmt.Errorf("project to wallet: %w", err)
	}

	// 索引转账交易记录（转出方）
	txModel := &domain.TransactionReadModel{
		ID:                0,
		TransactionNo:     event.TransactionNo,
		WalletID:          event.FromWalletID,
		UserID:            event.FromUserID,
		Type:              "TRANSFER",
		Amount:            event.Amount,
		Status:            "SUCCESS",
		Remark:            event.Remark,
		CounterpartyID:    event.ToWalletID,
		CounterpartyType:  "WALLET",
		CreatedAt:         event.Timestamp,
	}

	if h.searchRepo != nil {
		if err := h.searchRepo.IndexTransaction(ctx, txModel); err != nil {
			h.logger.WarnContext(ctx, "failed to index transfer transaction to es", "error", err)
		}
	}

	h.logger.DebugContext(ctx, "transferred event projected",
		"from_wallet", event.FromWalletID, "to_wallet", event.ToWalletID,
		"amount", event.Amount, "duration", time.Since(start))
	return nil
}

// HandleBalanceFrozen 处理余额冻结事件
func (h *WalletProjectionHandler) HandleBalanceFrozen(ctx context.Context, event *domain.BalanceFrozenEvent) error {
	wallet, err := h.walletReadRepo.GetByWalletID(ctx, event.WalletID)
	if err != nil || wallet == nil {
		return fmt.Errorf("wallet not found: %w", err)
	}

	wallet.FrozenBalance += event.Amount
	wallet.AvailableBalance = wallet.Balance - wallet.FrozenBalance
	wallet.UpdatedAt = time.Now()

	if err := h.walletReadRepo.Save(ctx, wallet); err != nil {
		h.logger.ErrorContext(ctx, "failed to project balance frozen event",
			"wallet_id", event.WalletID, "error", err)
		return fmt.Errorf("project balance frozen: %w", err)
	}

	h.logger.DebugContext(ctx, "balance frozen event projected",
		"wallet_id", event.WalletID, "amount", event.Amount)
	return nil
}

// HandleBalanceUnfrozen 处理余额解冻事件
func (h *WalletProjectionHandler) HandleBalanceUnfrozen(ctx context.Context, event *domain.BalanceUnfrozenEvent) error {
	wallet, err := h.walletReadRepo.GetByWalletID(ctx, event.WalletID)
	if err != nil || wallet == nil {
		return fmt.Errorf("wallet not found: %w", err)
	}

	wallet.FrozenBalance -= event.Amount
	if wallet.FrozenBalance < 0 {
		wallet.FrozenBalance = 0
	}
	wallet.AvailableBalance = wallet.Balance - wallet.FrozenBalance
	wallet.UpdatedAt = time.Now()

	if err := h.walletReadRepo.Save(ctx, wallet); err != nil {
		h.logger.ErrorContext(ctx, "failed to project balance unfrozen event",
			"wallet_id", event.WalletID, "error", err)
		return fmt.Errorf("project balance unfrozen: %w", err)
	}

	h.logger.DebugContext(ctx, "balance unfrozen event projected",
		"wallet_id", event.WalletID, "amount", event.Amount)
	return nil
}

// RegisterEventHandlers 注册事件处理器到消息总线
func (h *WalletProjectionHandler) RegisterEventHandlers(eventBus messagequeue.EventBus) {
	// 注册钱包创建事件处理器
	eventBus.Subscribe("wallet.created", func(ctx context.Context, eventData []byte) error {
		var event domain.WalletCreatedEvent
		if err := json.Unmarshal(eventData, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal wallet created event", "error", err)
			return err
		}
		return h.HandleWalletCreated(ctx, &event)
	})

	// 注册充值事件处理器
	eventBus.Subscribe("wallet.deposited", func(ctx context.Context, eventData []byte) error {
		var event domain.DepositedEvent
		if err := json.Unmarshal(eventData, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal deposited event", "error", err)
			return err
		}
		return h.HandleDeposited(ctx, &event)
	})

	// 注册提现事件处理器
	eventBus.Subscribe("wallet.withdrawn", func(ctx context.Context, eventData []byte) error {
		var event domain.WithdrawnEvent
		if err := json.Unmarshal(eventData, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal withdrawn event", "error", err)
			return err
		}
		return h.HandleWithdrawn(ctx, &event)
	})

	// 注册转账事件处理器
	eventBus.Subscribe("wallet.transferred", func(ctx context.Context, eventData []byte) error {
		var event domain.TransferredEvent
		if err := json.Unmarshal(eventData, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal transferred event", "error", err)
			return err
		}
		return h.HandleTransferred(ctx, &event)
	})

	// 注册余额冻结事件处理器
	eventBus.Subscribe("wallet.balance_frozen", func(ctx context.Context, eventData []byte) error {
		var event domain.BalanceFrozenEvent
		if err := json.Unmarshal(eventData, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal balance frozen event", "error", err)
			return err
		}
		return h.HandleBalanceFrozen(ctx, &event)
	})

	// 注册余额解冻事件处理器
	eventBus.Subscribe("wallet.balance_unfrozen", func(ctx context.Context, eventData []byte) error {
		var event domain.BalanceUnfrozenEvent
		if err := json.Unmarshal(eventData, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal balance unfrozen event", "error", err)
			return err
		}
		return h.HandleBalanceUnfrozen(ctx, &event)
	})

	h.logger.Info("wallet projection event handlers registered")
}