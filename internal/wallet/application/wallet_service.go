package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/wallet/domain"
	"github.com/wyfcoding/pkg/idgen"
)

var (
	ErrWalletNotFound      = domain.ErrTransactionNotFound // Reuse or define specific
	ErrInsufficientBalance = domain.ErrInsufficientBalance
	ErrInvalidAmount       = domain.ErrInvalidAmount
	ErrWalletFrozen        = domain.ErrWalletFrozen
)

// WalletService 提供钱包业务逻辑封装
type WalletService struct {
	walletRepo      domain.WalletRepository
	transactionRepo domain.TransactionRepository
	logger          *slog.Logger
}

func NewWalletService(
	walletRepo domain.WalletRepository,
	transactionRepo domain.TransactionRepository,
	logger *slog.Logger,
) *WalletService {
	return &WalletService{
		walletRepo:      walletRepo,
		transactionRepo: transactionRepo,
		logger:          logger.With("service", "wallet_application"),
	}
}

// CreateWallet 为用户创建新钱包
func (s *WalletService) CreateWallet(ctx context.Context, userID uint64, currency, walletType string) (*domain.Wallet, error) {
	existing, err := s.walletRepo.GetByUserID(userID, currency)
	if err == nil && existing != nil {
		return existing, nil
	}

	wallet := domain.NewWallet(userID, fmt.Sprintf("W%d%s", userID, currency), currency, walletType)
	wallet.WalletID = idgen.GenID()

	if err := s.walletRepo.Create(wallet); err != nil {
		s.logger.Error("failed to create wallet", "user_id", userID, "error", err)
		return nil, err
	}

	return wallet, nil
}

// GetWallet 获取用户钱包信息
func (s *WalletService) GetWallet(ctx context.Context, userID uint64, currency string) (*domain.Wallet, error) {
	wallet, err := s.walletRepo.GetByUserID(userID, currency)
	if err != nil {
		return nil, err
	}
	if wallet == nil {
		return nil, ErrWalletNotFound
	}
	return wallet, nil
}

// Deposit 充值操作
func (s *WalletService) Deposit(ctx context.Context, userID uint64, currency string, amount int64, remark string) (*domain.Transaction, error) {
	var tx *domain.Transaction
	err := s.walletRepo.Transaction(func(txObj any) error {
		wallet, err := s.walletRepo.GetByUserID(userID, currency)
		if err != nil || wallet == nil {
			return ErrWalletNotFound
		}

		transaction, err := wallet.Deposit(amount)
		if err != nil {
			return err
		}
		transaction.TransactionNo = fmt.Sprintf("D%d", idgen.GenID())
		transaction.Remark = remark

		if err := s.walletRepo.UpdateBalance(wallet.ID, wallet.Balance, wallet.FrozenBalance, wallet.AvailableBalance); err != nil {
			return err
		}

		if err := s.transactionRepo.Create(transaction); err != nil {
			return err
		}
		tx = transaction
		return nil
	})

	return tx, err
}

// Withdraw 提现操作
func (s *WalletService) Withdraw(ctx context.Context, userID uint64, currency string, amount int64, remark string) (*domain.Transaction, error) {
	var tx *domain.Transaction
	err := s.walletRepo.Transaction(func(txObj any) error {
		wallet, err := s.walletRepo.GetByUserID(userID, currency)
		if err != nil || wallet == nil {
			return ErrWalletNotFound
		}

		// 这里暂不传递密码校验逻辑，待 interfaces 层传入
		transaction, err := wallet.Withdraw(amount, "", nil, nil)
		if err != nil {
			return err
		}
		transaction.TransactionNo = fmt.Sprintf("W%d", idgen.GenID())
		transaction.Remark = remark

		if err := s.walletRepo.UpdateBalance(wallet.ID, wallet.Balance, wallet.FrozenBalance, wallet.AvailableBalance); err != nil {
			return err
		}

		if err := s.transactionRepo.Create(transaction); err != nil {
			return err
		}
		tx = transaction
		return nil
	})

	return tx, err
}

// Transfer 转账操作
func (s *WalletService) Transfer(ctx context.Context, fromUserID, toUserID uint64, currency string, amount int64, remark string) (*domain.Transaction, error) {
	var outTx *domain.Transaction
	err := s.walletRepo.Transaction(func(txObj any) error {
		fromWallet, err := s.walletRepo.GetByUserID(fromUserID, currency)
		if err != nil || fromWallet == nil {
			return ErrWalletNotFound
		}

		toWallet, err := s.walletRepo.GetByUserID(toUserID, currency)
		if err != nil || toWallet == nil {
			return ErrWalletNotFound
		}

		txs, err := fromWallet.Transfer(toWallet, amount, "", nil, nil)
		if err != nil {
			return err
		}

		batchNo := fmt.Sprintf("T%d", idgen.GenID())
		for _, t := range txs {
			t.TransactionNo = batchNo
			t.Remark = remark
			if err := s.transactionRepo.Create(t); err != nil {
				return err
			}
		}

		if err := s.walletRepo.UpdateBalance(fromWallet.ID, fromWallet.Balance, fromWallet.FrozenBalance, fromWallet.AvailableBalance); err != nil {
			return err
		}
		if err := s.walletRepo.UpdateBalance(toWallet.ID, toWallet.Balance, toWallet.FrozenBalance, toWallet.AvailableBalance); err != nil {
			return err
		}

		outTx = txs[0]
		return nil
	})

	return outTx, err
}
