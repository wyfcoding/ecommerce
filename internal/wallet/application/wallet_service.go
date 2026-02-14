package application

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/wallet/domain"
	"github.com/wyfcoding/pkg/idgen"
)

var (
	ErrWalletNotFound      = errors.New("wallet not found")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrWalletFrozen        = errors.New("wallet is frozen")
)

type WalletService struct {
	walletRepo      domain.WalletRepository
	transactionRepo domain.TransactionRepository
}

func NewWalletService(
	walletRepo domain.WalletRepository,
	transactionRepo domain.TransactionRepository,
) *WalletService {
	return &WalletService{
		walletRepo:      walletRepo,
		transactionRepo: transactionRepo,
	}
}

func (s *WalletService) CreateWallet(ctx context.Context, userID uint64, currency, walletType string) (*domain.Wallet, error) {
	existing, err := s.walletRepo.GetByUserID(userID, currency)
	if err == nil && existing != nil {
		return existing, nil
	}

	wallet := &domain.Wallet{
		WalletID:         idgen.GenID(),
		UserID:           userID,
		AccountNo:        fmt.Sprintf("W%d%s", userID, currency),
		Currency:         currency,
		WalletType:       walletType,
		Balance:          0,
		FrozenBalance:    0,
		AvailableBalance: 0,
		Status:           domain.WalletStatusNormal,
	}

	if err := s.walletRepo.Create(wallet); err != nil {
		return nil, err
	}

	return wallet, nil
}

func (s *WalletService) GetWallet(ctx context.Context, userID uint64, currency string) (*domain.Wallet, error) {
	wallet, err := s.walletRepo.GetByUserID(userID, currency)
	if err != nil {
		return nil, ErrWalletNotFound
	}
	return wallet, nil
}

func (s *WalletService) Deposit(ctx context.Context, userID uint64, currency, amountStr string, remark string) (*domain.Transaction, error) {
	wallet, err := s.walletRepo.GetByUserID(userID, currency)
	if err != nil {
		return nil, ErrWalletNotFound
	}

	amount, err := parseAmount(amountStr)
	if err != nil {
		return nil, ErrInvalidAmount
	}

	balanceBefore := wallet.Balance
	wallet.Balance += amount
	wallet.AvailableBalance += amount

	if err := s.walletRepo.UpdateBalance(wallet.ID, wallet.Balance, wallet.FrozenBalance, wallet.AvailableBalance); err != nil {
		return nil, err
	}

	tx := &domain.Transaction{
		TransactionNo: fmt.Sprintf("D%d%d", userID, idgen.GenID()),
		WalletID:      wallet.ID,
		UserID:        userID,
		Type:          domain.TransactionTypeDeposit,
		Amount:        amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  wallet.Balance,
		Status:        domain.TransactionStatusSuccess,
		Remark:        remark,
	}

	if err := s.transactionRepo.Create(tx); err != nil {
		return nil, err
	}

	return tx, nil
}

func (s *WalletService) Withdraw(ctx context.Context, userID uint64, currency, amountStr string, remark string) (*domain.Transaction, error) {
	wallet, err := s.walletRepo.GetByUserID(userID, currency)
	if err != nil {
		return nil, ErrWalletNotFound
	}

	amount, err := parseAmount(amountStr)
	if err != nil {
		return nil, ErrInvalidAmount
	}

	if wallet.AvailableBalance < amount {
		return nil, ErrInsufficientBalance
	}

	balanceBefore := wallet.Balance
	wallet.Balance -= amount
	wallet.AvailableBalance -= amount

	if err := s.walletRepo.UpdateBalance(wallet.ID, wallet.Balance, wallet.FrozenBalance, wallet.AvailableBalance); err != nil {
		return nil, err
	}

	tx := &domain.Transaction{
		TransactionNo: fmt.Sprintf("W%d%d", userID, idgen.GenID()),
		WalletID:      wallet.ID,
		UserID:        userID,
		Type:          domain.TransactionTypeWithdraw,
		Amount:        -amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  wallet.Balance,
		Status:        domain.TransactionStatusSuccess,
		Remark:        remark,
	}

	if err := s.transactionRepo.Create(tx); err != nil {
		return nil, err
	}

	return tx, nil
}

func (s *WalletService) Transfer(ctx context.Context, fromUserID, toUserID uint64, currency, amountStr, remark string) (*domain.Transaction, error) {
	amount, err := parseAmount(amountStr)
	if err != nil {
		return nil, ErrInvalidAmount
	}

	fromWallet, err := s.walletRepo.GetByUserID(fromUserID, currency)
	if err != nil {
		return nil, ErrWalletNotFound
	}

	if fromWallet.AvailableBalance < amount {
		return nil, ErrInsufficientBalance
	}

	toWallet, err := s.walletRepo.GetByUserID(toUserID, currency)
	if err != nil {
		return nil, ErrWalletNotFound
	}

	fromBalanceBefore := fromWallet.Balance
	fromWallet.Balance -= amount
	fromWallet.AvailableBalance -= amount

	toBalanceBefore := toWallet.Balance
	toWallet.Balance += amount
	toWallet.AvailableBalance += amount

	if err := s.walletRepo.UpdateBalance(fromWallet.ID, fromWallet.Balance, fromWallet.FrozenBalance, fromWallet.AvailableBalance); err != nil {
		return nil, err
	}

	if err := s.walletRepo.UpdateBalance(toWallet.ID, toWallet.Balance, toWallet.FrozenBalance, toWallet.AvailableBalance); err != nil {
		return nil, err
	}

	tx := &domain.Transaction{
		TransactionNo: fmt.Sprintf("T%d%d", fromUserID, idgen.GenID()),
		WalletID:      fromWallet.ID,
		UserID:        fromUserID,
		Type:          domain.TransactionTypeTransfer,
		Amount:        -amount,
		BalanceBefore: fromBalanceBefore,
		BalanceAfter:  fromWallet.Balance,
		Status:        domain.TransactionStatusSuccess,
		Remark:        fmt.Sprintf("%s -> to=%d,before_to=%d,after_to=%d", remark, toUserID, toBalanceBefore, toWallet.Balance),
	}

	if err := s.transactionRepo.Create(tx); err != nil {
		return nil, err
	}

	return tx, nil
}

func parseAmount(amountStr string) (int64, error) {
	amount, err := strconv.ParseInt(amountStr, 10, 64)
	if err != nil {
		return 0, err
	}
	if amount <= 0 {
		return 0, ErrInvalidAmount
	}
	return amount, nil
}
