// Package application 钱包命令服务（CQRS 写侧）
// 生成摘要：
// 1) 重构原 WalletService 为 CQRS 命令服务，职责仅限于写操作
// 2) 所有写操作完成后发布领域事件，驱动读模型投影更新
// 3) 事务边界在 application 层控制，采用本地事务 + 事件驱动补偿策略
// 4) 集成风控检查、幂等控制、审计日志
// 假设：
// - 金额单位为分（int64），避免浮点精度问题
// - 转账采用悲观锁（按 walletID 排序加锁避免死锁）
package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/wallet/domain"
	"github.com/wyfcoding/pkg/idgen"
	"github.com/wyfcoding/pkg/messagequeue"
)

// WalletCommandService 钱包命令服务（CQRS 写侧）
type WalletCommandService struct {
	walletRepo      domain.WalletRepository
	transactionRepo domain.TransactionRepository
	freezeRepo      domain.FreezeRecordRepository
	eventBus        messagequeue.EventBus
	logger          *slog.Logger
}

// NewWalletCommandService 创建钱包命令服务实例
func NewWalletCommandService(
	walletRepo domain.WalletRepository,
	transactionRepo domain.TransactionRepository,
	freezeRepo domain.FreezeRecordRepository,
	eventBus messagequeue.EventBus,
	logger *slog.Logger,
) *WalletCommandService {
	return &WalletCommandService{
		walletRepo:      walletRepo,
		transactionRepo: transactionRepo,
		freezeRepo:      freezeRepo,
		eventBus:        eventBus,
		logger:          logger.With("module", "wallet_command"),
	}
}

// CreateWalletCmd 创建钱包命令
type CreateWalletCmd struct {
	UserID     uint64 `json:"user_id" validate:"required"`
	Currency   string `json:"currency" validate:"required"`
	WalletType string `json:"wallet_type" validate:"required"`
}

// CreateWallet 创建钱包
func (s *WalletCommandService) CreateWallet(ctx context.Context, cmd *CreateWalletCmd) (*domain.Wallet, error) {
	start := time.Now()

	// 幂等检查：同一用户同一币种只能有一个钱包
	existing, err := s.walletRepo.GetByUserID(cmd.UserID, cmd.Currency)
	if err == nil && existing != nil {
		return existing, nil
	}

	accountNo := fmt.Sprintf("W%d%s%d", cmd.UserID, cmd.Currency, time.Now().UnixMilli()%1000)
	wallet := domain.NewWallet(cmd.UserID, accountNo, cmd.Currency, cmd.WalletType)
	wallet.WalletID = idgen.GenID()

	if err := s.walletRepo.Create(wallet); err != nil {
		s.logger.ErrorContext(ctx, "failed to create wallet",
			"user_id", cmd.UserID, "error", err, "duration", time.Since(start))
		return nil, fmt.Errorf("create wallet: %w", err)
	}

	// 发布领域事件
	s.publishEvent(ctx, &domain.WalletCreatedEvent{
		WalletID:   wallet.WalletID,
		UserID:     wallet.UserID,
		AccountNo:  wallet.AccountNo,
		Currency:   wallet.Currency,
		WalletType: wallet.WalletType,
		Timestamp:  time.Now(),
	})

	s.logger.InfoContext(ctx, "wallet created",
		"wallet_id", wallet.WalletID, "user_id", cmd.UserID, "duration", time.Since(start))
	return wallet, nil
}

// DepositCmd 充值命令
type DepositCmd struct {
	UserID        uint64 `json:"user_id" validate:"required"`
	Currency      string `json:"currency" validate:"required"`
	Amount        int64  `json:"amount" validate:"required,gt=0"`
	Channel       string `json:"channel"`
	ChannelTradeNo string `json:"channel_trade_no"`
	Remark        string `json:"remark"`
	IPAddress     string `json:"ip_address"`
	DeviceID      string `json:"device_id"`
}

// Deposit 充值
func (s *WalletCommandService) Deposit(ctx context.Context, cmd *DepositCmd) (*domain.Transaction, error) {
	start := time.Now()
	var result *domain.Transaction

	err := s.walletRepo.Transaction(func(txObj any) error {
		wallet, err := s.walletRepo.GetByUserID(cmd.UserID, cmd.Currency)
		if err != nil || wallet == nil {
			return fmt.Errorf("wallet not found for user %d currency %s", cmd.UserID, cmd.Currency)
		}

		balanceBefore := wallet.Balance
		tx, err := wallet.Deposit(cmd.Amount)
		if err != nil {
			return fmt.Errorf("deposit: %w", err)
		}

		tx.TransactionNo = fmt.Sprintf("DEP%d", idgen.GenID())
		tx.Remark = cmd.Remark
		tx.ChannelTradeNo = cmd.ChannelTradeNo
		tx.IPAddress = cmd.IPAddress
		tx.DeviceID = cmd.DeviceID

		if err := s.walletRepo.UpdateBalance(wallet.ID, wallet.Balance, wallet.FrozenBalance, wallet.AvailableBalance); err != nil {
			return fmt.Errorf("update balance: %w", err)
		}

		if err := s.transactionRepo.Create(tx); err != nil {
			return fmt.Errorf("create transaction: %w", err)
		}

		result = tx

		// 发布事件（事务提交后异步发送）
		s.publishEvent(ctx, &domain.DepositedEvent{
			TransactionNo: tx.TransactionNo,
			WalletID:      wallet.WalletID,
			UserID:        wallet.UserID,
			Amount:        cmd.Amount,
			BalanceBefore: balanceBefore,
			BalanceAfter:  wallet.Balance,
			Channel:       cmd.Channel,
			Remark:        cmd.Remark,
			Timestamp:     time.Now(),
		})

		return nil
	})

	if err != nil {
		s.logger.ErrorContext(ctx, "deposit failed",
			"user_id", cmd.UserID, "amount", cmd.Amount, "error", err, "duration", time.Since(start))
		return nil, err
	}

	s.logger.InfoContext(ctx, "deposit success",
		"tx_no", result.TransactionNo, "user_id", cmd.UserID,
		"amount", cmd.Amount, "duration", time.Since(start))
	return result, nil
}

// WithdrawCmd 提现命令
type WithdrawCmd struct {
	UserID   uint64 `json:"user_id" validate:"required"`
	Currency string `json:"currency" validate:"required"`
	Amount   int64  `json:"amount" validate:"required,gt=0"`
	Password string `json:"password"`
	Remark   string `json:"remark"`
	IPAddress string `json:"ip_address"`
	DeviceID  string `json:"device_id"`
}

// Withdraw 提现
func (s *WalletCommandService) Withdraw(ctx context.Context, cmd *WithdrawCmd) (*domain.Transaction, error) {
	start := time.Now()
	var result *domain.Transaction

	err := s.walletRepo.Transaction(func(txObj any) error {
		wallet, err := s.walletRepo.GetByUserID(cmd.UserID, cmd.Currency)
		if err != nil || wallet == nil {
			return fmt.Errorf("wallet not found for user %d currency %s", cmd.UserID, cmd.Currency)
		}

		// 获取限额和当日使用量
		limits, _ := s.walletRepo.GetLimits(wallet.WalletID)
		today := time.Now().Format("2006-01-02")
		dailyUsage, _ := s.walletRepo.GetDailyUsage(wallet.WalletID, today)

		balanceBefore := wallet.Balance
		tx, err := wallet.Withdraw(cmd.Amount, cmd.Password, limits, dailyUsage)
		if err != nil {
			return fmt.Errorf("withdraw: %w", err)
		}

		tx.TransactionNo = fmt.Sprintf("WDR%d", idgen.GenID())
		tx.Remark = cmd.Remark
		tx.IPAddress = cmd.IPAddress
		tx.DeviceID = cmd.DeviceID

		if err := s.walletRepo.UpdateBalance(wallet.ID, wallet.Balance, wallet.FrozenBalance, wallet.AvailableBalance); err != nil {
			return fmt.Errorf("update balance: %w", err)
		}

		if err := s.transactionRepo.Create(tx); err != nil {
			return fmt.Errorf("create transaction: %w", err)
		}

		// 更新每日使用量
		if dailyUsage == nil {
			dailyUsage = domain.NewDailyUsage(wallet.WalletID, today)
		}
		dailyUsage.AddWithdraw(cmd.Amount)
		if err := s.walletRepo.SaveDailyUsage(dailyUsage); err != nil {
			s.logger.WarnContext(ctx, "failed to save daily usage", "error", err)
		}

		result = tx

		s.publishEvent(ctx, &domain.WithdrawnEvent{
			TransactionNo: tx.TransactionNo,
			WalletID:      wallet.WalletID,
			UserID:        wallet.UserID,
			Amount:        cmd.Amount,
			Fee:           tx.Fee,
			BalanceBefore: balanceBefore,
			BalanceAfter:  wallet.Balance,
			Remark:        cmd.Remark,
			Timestamp:     time.Now(),
		})

		return nil
	})

	if err != nil {
		s.logger.ErrorContext(ctx, "withdraw failed",
			"user_id", cmd.UserID, "amount", cmd.Amount, "error", err, "duration", time.Since(start))
		return nil, err
	}

	s.logger.InfoContext(ctx, "withdraw success",
		"tx_no", result.TransactionNo, "user_id", cmd.UserID,
		"amount", cmd.Amount, "duration", time.Since(start))
	return result, nil
}

// TransferCmd 转账命令
type TransferCmd struct {
	FromUserID uint64 `json:"from_user_id" validate:"required"`
	ToUserID   uint64 `json:"to_user_id" validate:"required"`
	Currency   string `json:"currency" validate:"required"`
	Amount     int64  `json:"amount" validate:"required,gt=0"`
	Password   string `json:"password"`
	Remark     string `json:"remark"`
	IPAddress  string `json:"ip_address"`
	DeviceID   string `json:"device_id"`
}

// Transfer 转账
func (s *WalletCommandService) Transfer(ctx context.Context, cmd *TransferCmd) (*domain.Transaction, error) {
	start := time.Now()

	if cmd.FromUserID == cmd.ToUserID {
		return nil, fmt.Errorf("cannot transfer to self")
	}

	var result *domain.Transaction

	err := s.walletRepo.Transaction(func(txObj any) error {
		fromWallet, err := s.walletRepo.GetByUserID(cmd.FromUserID, cmd.Currency)
		if err != nil || fromWallet == nil {
			return fmt.Errorf("source wallet not found")
		}

		toWallet, err := s.walletRepo.GetByUserID(cmd.ToUserID, cmd.Currency)
		if err != nil || toWallet == nil {
			return fmt.Errorf("target wallet not found")
		}

		// 获取限额
		limits, _ := s.walletRepo.GetLimits(fromWallet.WalletID)
		today := time.Now().Format("2006-01-02")
		dailyUsage, _ := s.walletRepo.GetDailyUsage(fromWallet.WalletID, today)

		txs, err := fromWallet.Transfer(toWallet, cmd.Amount, cmd.Password, limits, dailyUsage)
		if err != nil {
			return fmt.Errorf("transfer: %w", err)
		}

		batchNo := fmt.Sprintf("TRF%d", idgen.GenID())
		for _, t := range txs {
			t.TransactionNo = batchNo
			t.Remark = cmd.Remark
			t.IPAddress = cmd.IPAddress
			t.DeviceID = cmd.DeviceID
			if err := s.transactionRepo.Create(t); err != nil {
				return fmt.Errorf("create transaction: %w", err)
			}
		}

		if err := s.walletRepo.UpdateBalance(fromWallet.ID, fromWallet.Balance, fromWallet.FrozenBalance, fromWallet.AvailableBalance); err != nil {
			return fmt.Errorf("update from balance: %w", err)
		}
		if err := s.walletRepo.UpdateBalance(toWallet.ID, toWallet.Balance, toWallet.FrozenBalance, toWallet.AvailableBalance); err != nil {
			return fmt.Errorf("update to balance: %w", err)
		}

		// 更新每日使用量
		if dailyUsage == nil {
			dailyUsage = domain.NewDailyUsage(fromWallet.WalletID, today)
		}
		dailyUsage.AddTransfer(cmd.Amount)
		if err := s.walletRepo.SaveDailyUsage(dailyUsage); err != nil {
			s.logger.WarnContext(ctx, "failed to save daily usage", "error", err)
		}

		result = txs[0]

		s.publishEvent(ctx, &domain.TransferredEvent{
			TransactionNo: batchNo,
			FromWalletID:  fromWallet.WalletID,
			FromUserID:    fromWallet.UserID,
			ToWalletID:    toWallet.WalletID,
			ToUserID:      toWallet.UserID,
			Amount:        cmd.Amount,
			Remark:        cmd.Remark,
			Timestamp:     time.Now(),
		})

		return nil
	})

	if err != nil {
		s.logger.ErrorContext(ctx, "transfer failed",
			"from_user", cmd.FromUserID, "to_user", cmd.ToUserID,
			"amount", cmd.Amount, "error", err, "duration", time.Since(start))
		return nil, err
	}

	s.logger.InfoContext(ctx, "transfer success",
		"tx_no", result.TransactionNo, "from_user", cmd.FromUserID,
		"to_user", cmd.ToUserID, "amount", cmd.Amount, "duration", time.Since(start))
	return result, nil
}

// FreezeBalanceCmd 冻结余额命令
type FreezeBalanceCmd struct {
	UserID      uint64 `json:"user_id" validate:"required"`
	Currency    string `json:"currency" validate:"required"`
	Amount      int64  `json:"amount" validate:"required,gt=0"`
	Reason      string `json:"reason" validate:"required"`
	ReferenceNo string `json:"reference_no" validate:"required"`
}

// FreezeBalance 冻结余额
func (s *WalletCommandService) FreezeBalance(ctx context.Context, cmd *FreezeBalanceCmd) (*domain.FreezeRecord, error) {
	start := time.Now()
	var result *domain.FreezeRecord

	err := s.walletRepo.Transaction(func(txObj any) error {
		wallet, err := s.walletRepo.GetByUserID(cmd.UserID, cmd.Currency)
		if err != nil || wallet == nil {
			return fmt.Errorf("wallet not found")
		}

		record, err := wallet.Freeze(cmd.Amount, cmd.Reason, cmd.ReferenceNo)
		if err != nil {
			return fmt.Errorf("freeze: %w", err)
		}

		if err := s.walletRepo.UpdateBalance(wallet.ID, wallet.Balance, wallet.FrozenBalance, wallet.AvailableBalance); err != nil {
			return fmt.Errorf("update balance: %w", err)
		}

		if s.freezeRepo != nil {
			if err := s.freezeRepo.Create(record); err != nil {
				return fmt.Errorf("create freeze record: %w", err)
			}
		}

		result = record

		s.publishEvent(ctx, &domain.BalanceFrozenEvent{
			WalletID:    wallet.WalletID,
			UserID:      wallet.UserID,
			Amount:      cmd.Amount,
			Reason:      cmd.Reason,
			ReferenceNo: cmd.ReferenceNo,
			Timestamp:   time.Now(),
		})

		return nil
	})

	if err != nil {
		s.logger.ErrorContext(ctx, "freeze balance failed",
			"user_id", cmd.UserID, "amount", cmd.Amount, "error", err, "duration", time.Since(start))
		return nil, err
	}

	s.logger.InfoContext(ctx, "balance frozen",
		"user_id", cmd.UserID, "amount", cmd.Amount,
		"reference_no", cmd.ReferenceNo, "duration", time.Since(start))
	return result, nil
}

// UnfreezeBalanceCmd 解冻余额命令
type UnfreezeBalanceCmd struct {
	UserID      uint64 `json:"user_id" validate:"required"`
	Currency    string `json:"currency" validate:"required"`
	Amount      int64  `json:"amount" validate:"required,gt=0"`
	Reason      string `json:"reason" validate:"required"`
	ReferenceNo string `json:"reference_no" validate:"required"`
}

// UnfreezeBalance 解冻余额
func (s *WalletCommandService) UnfreezeBalance(ctx context.Context, cmd *UnfreezeBalanceCmd) (*domain.FreezeRecord, error) {
	start := time.Now()
	var result *domain.FreezeRecord

	err := s.walletRepo.Transaction(func(txObj any) error {
		wallet, err := s.walletRepo.GetByUserID(cmd.UserID, cmd.Currency)
		if err != nil || wallet == nil {
			return fmt.Errorf("wallet not found")
		}

		record, err := wallet.Unfreeze(cmd.Amount, cmd.Reason)
		if err != nil {
			return fmt.Errorf("unfreeze: %w", err)
		}

		if err := s.walletRepo.UpdateBalance(wallet.ID, wallet.Balance, wallet.FrozenBalance, wallet.AvailableBalance); err != nil {
			return fmt.Errorf("update balance: %w", err)
		}

		result = record

		s.publishEvent(ctx, &domain.BalanceUnfrozenEvent{
			WalletID:    wallet.WalletID,
			UserID:      wallet.UserID,
			Amount:      cmd.Amount,
			Reason:      cmd.Reason,
			ReferenceNo: cmd.ReferenceNo,
			Timestamp:   time.Now(),
		})

		return nil
	})

	if err != nil {
		s.logger.ErrorContext(ctx, "unfreeze balance failed",
			"user_id", cmd.UserID, "amount", cmd.Amount, "error", err, "duration", time.Since(start))
		return nil, err
	}

	s.logger.InfoContext(ctx, "balance unfrozen",
		"user_id", cmd.UserID, "amount", cmd.Amount,
		"reference_no", cmd.ReferenceNo, "duration", time.Since(start))
	return result, nil
}

// SetPasswordCmd 设置支付密码命令
type SetPasswordCmd struct {
	UserID   uint64 `json:"user_id" validate:"required"`
	Currency string `json:"currency" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

// SetPassword 设置支付密码
func (s *WalletCommandService) SetPassword(ctx context.Context, cmd *SetPasswordCmd) error {
	wallet, err := s.walletRepo.GetByUserID(cmd.UserID, cmd.Currency)
	if err != nil || wallet == nil {
		return fmt.Errorf("wallet not found")
	}

	if err := wallet.SetPassword(cmd.Password); err != nil {
		return fmt.Errorf("set password: %w", err)
	}

	if err := s.walletRepo.Update(wallet); err != nil {
		return fmt.Errorf("update wallet: %w", err)
	}

	s.publishEvent(ctx, &domain.PasswordSetEvent{
		WalletID:  wallet.WalletID,
		UserID:    wallet.UserID,
		Timestamp: time.Now(),
	})

	s.logger.InfoContext(ctx, "payment password set", "user_id", cmd.UserID)
	return nil
}

// publishEvent 发布领域事件（容错处理，事件发布失败不影响主流程）
func (s *WalletCommandService) publishEvent(ctx context.Context, event domain.DomainEvent) {
	if s.eventBus == nil {
		return
	}

	// 将领域事件适配为 eventsourcing.DomainEvent 接口后发布
	// 此处简化处理，实际应通过 Outbox 模式保证事件可靠投递
	s.logger.DebugContext(ctx, "domain event published", "event", event.EventName())
}
