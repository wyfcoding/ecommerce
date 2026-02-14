package domain

import (
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInsufficientBalance  = errors.New("insufficient balance")
	ErrWalletFrozen         = errors.New("wallet is frozen")
	ErrWalletDisabled       = errors.New("wallet is disabled")
	ErrInvalidPassword      = errors.New("invalid payment password")
	ErrPasswordNotSet       = errors.New("payment password not set")
	ErrExceedDailyLimit     = errors.New("exceed daily limit")
	ErrExceedSingleLimit    = errors.New("exceed single transaction limit")
	ErrExceedMonthlyLimit   = errors.New("exceed monthly limit")
	ErrInvalidAmount        = errors.New("invalid amount")
	ErrTransactionNotFound  = errors.New("transaction not found")
	ErrDuplicateTransaction = errors.New("duplicate transaction")
)

type WalletStatus int

const (
	WalletStatusDisabled WalletStatus = 0
	WalletStatusNormal   WalletStatus = 1
	WalletStatusFrozen   WalletStatus = 2
)

func (s WalletStatus) String() string {
	switch s {
	case WalletStatusDisabled:
		return "DISABLED"
	case WalletStatusNormal:
		return "NORMAL"
	case WalletStatusFrozen:
		return "FROZEN"
	default:
		return "UNKNOWN"
	}
}

type TransactionType string

const (
	TransactionTypeDeposit    TransactionType = "DEPOSIT"
	TransactionTypeWithdraw   TransactionType = "WITHDRAW"
	TransactionTypeTransfer   TransactionType = "TRANSFER"
	TransactionTypeFreeze     TransactionType = "FREEZE"
	TransactionTypeUnfreeze   TransactionType = "UNFREEZE"
	TransactionTypePayment    TransactionType = "PAYMENT"
	TransactionTypeRefund     TransactionType = "REFUND"
	TransactionTypeCommission TransactionType = "COMMISSION"
	TransactionTypeDividend   TransactionType = "DIVIDEND"
)

type TransactionStatus int

const (
	TransactionStatusPending   TransactionStatus = 0
	TransactionStatusSuccess   TransactionStatus = 1
	TransactionStatusFailed    TransactionStatus = 2
	TransactionStatusCancelled TransactionStatus = 3
)

func (s TransactionStatus) String() string {
	switch s {
	case TransactionStatusPending:
		return "PENDING"
	case TransactionStatusSuccess:
		return "SUCCESS"
	case TransactionStatusFailed:
		return "FAILED"
	case TransactionStatusCancelled:
		return "CANCELLED"
	default:
		return "UNKNOWN"
	}
}

type Wallet struct {
	ID               uint64        `json:"id"`
	WalletID         uint64        `json:"wallet_id"`
	UserID           uint64        `json:"user_id"`
	AccountNo        string        `json:"account_no"`
	Currency         string        `json:"currency"`
	WalletType       string        `json:"wallet_type"`
	Balance          int64         `json:"balance"`
	FrozenBalance    int64         `json:"frozen_balance"`
	AvailableBalance int64         `json:"available_balance"`
	Status           WalletStatus  `json:"status"`
	PasswordHash     string        `json:"-"`
	PasswordSalt     string        `json:"-"`
	HasPassword      bool          `json:"has_password"`
	SecurityLevel    int           `json:"security_level"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
	Limits           *WalletLimits `json:"limits,omitempty"`
	DailyUsage       *DailyUsage   `json:"daily_usage,omitempty"`
}

type WalletLimits struct {
	ID                   uint64    `json:"id"`
	WalletID             uint64    `json:"wallet_id"`
	SingleDepositLimit   int64     `json:"single_deposit_limit"`
	SingleWithdrawLimit  int64     `json:"single_withdraw_limit"`
	DailyDepositLimit    int64     `json:"daily_deposit_limit"`
	DailyWithdrawLimit   int64     `json:"daily_withdraw_limit"`
	MonthlyWithdrawLimit int64     `json:"monthly_withdraw_limit"`
	DailyTransferLimit   int64     `json:"daily_transfer_limit"`
	RequirePasswordMin   int64     `json:"require_password_min"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type DailyUsage struct {
	ID               uint64    `json:"id"`
	WalletID         uint64    `json:"wallet_id"`
	UsageDate        string    `json:"usage_date"`
	DepositAmount    int64     `json:"deposit_amount"`
	WithdrawAmount   int64     `json:"withdraw_amount"`
	TransferAmount   int64     `json:"transfer_amount"`
	TransactionCount int       `json:"transaction_count"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Transaction struct {
	ID               uint64            `json:"id"`
	TransactionNo    string            `json:"transaction_no"`
	WalletID         uint64            `json:"wallet_id"`
	UserID           uint64            `json:"user_id"`
	Type             TransactionType   `json:"type"`
	Amount           int64             `json:"amount"`
	BalanceBefore    int64             `json:"balance_before"`
	BalanceAfter     int64             `json:"balance_after"`
	Fee              int64             `json:"fee"`
	Status           TransactionStatus `json:"status"`
	Remark           string            `json:"remark"`
	ChannelTradeNo   string            `json:"channel_trade_no"`
	CounterpartyID   uint64            `json:"counterparty_id,omitempty"`
	CounterpartyType string            `json:"counterparty_type,omitempty"`
	RiskCheckID      string            `json:"risk_check_id,omitempty"`
	IPAddress        string            `json:"ip_address,omitempty"`
	DeviceID         string            `json:"device_id,omitempty"`
	Location         string            `json:"location,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
	CompletedAt      *time.Time        `json:"completed_at,omitempty"`
}

type PasswordRecord struct {
	ID           uint64     `json:"id"`
	UserID       uint64     `json:"user_id"`
	WalletID     uint64     `json:"wallet_id"`
	PasswordHash string     `json:"-"`
	Salt         string     `json:"-"`
	Version      int        `json:"version"`
	FailedCount  int        `json:"failed_count"`
	LockedUntil  *time.Time `json:"locked_until,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type FreezeRecord struct {
	ID          uint64     `json:"id"`
	WalletID    uint64     `json:"wallet_id"`
	UserID      uint64     `json:"user_id"`
	Amount      int64      `json:"amount"`
	Reason      string     `json:"reason"`
	ReferenceNo string     `json:"reference_no"`
	Status      string     `json:"status"`
	FrozenAt    time.Time  `json:"frozen_at"`
	UnfrozenAt  *time.Time `json:"unfrozen_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

func NewWallet(userID uint64, accountNo, currency, walletType string) *Wallet {
	now := time.Now()
	return &Wallet{
		UserID:           userID,
		AccountNo:        accountNo,
		Currency:         currency,
		WalletType:       walletType,
		Balance:          0,
		FrozenBalance:    0,
		AvailableBalance: 0,
		Status:           WalletStatusNormal,
		HasPassword:      false,
		SecurityLevel:    1,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func (w *Wallet) SetPassword(password string) error {
	if len(password) < 6 {
		return errors.New("password must be at least 6 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	w.PasswordHash = string(hash)
	w.HasPassword = true
	w.UpdatedAt = time.Now()
	return nil
}

func (w *Wallet) VerifyPassword(password string) error {
	if !w.HasPassword || w.PasswordHash == "" {
		return ErrPasswordNotSet
	}

	err := bcrypt.CompareHashAndPassword([]byte(w.PasswordHash), []byte(password))
	if err != nil {
		return ErrInvalidPassword
	}
	return nil
}

func (w *Wallet) Deposit(amount int64) (*Transaction, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	if w.Status == WalletStatusDisabled {
		return nil, ErrWalletDisabled
	}

	balanceBefore := w.Balance
	w.Balance += amount
	w.AvailableBalance += amount
	w.UpdatedAt = time.Now()

	return &Transaction{
		WalletID:      w.WalletID,
		UserID:        w.UserID,
		Type:          TransactionTypeDeposit,
		Amount:        amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  w.Balance,
		Status:        TransactionStatusSuccess,
		CreatedAt:     time.Now(),
	}, nil
}

func (w *Wallet) Withdraw(amount int64, password string, limits *WalletLimits, dailyUsage *DailyUsage) (*Transaction, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	if w.Status == WalletStatusDisabled {
		return nil, ErrWalletDisabled
	}

	if w.Status == WalletStatusFrozen {
		return nil, ErrWalletFrozen
	}

	if w.HasPassword {
		if err := w.VerifyPassword(password); err != nil {
			return nil, err
		}
	}

	if w.AvailableBalance < amount {
		return nil, ErrInsufficientBalance
	}

	if limits != nil {
		if limits.SingleWithdrawLimit > 0 && amount > limits.SingleWithdrawLimit {
			return nil, ErrExceedSingleLimit
		}
		if limits.DailyWithdrawLimit > 0 && dailyUsage != nil {
			if dailyUsage.WithdrawAmount+amount > limits.DailyWithdrawLimit {
				return nil, ErrExceedDailyLimit
			}
		}
	}

	balanceBefore := w.Balance
	w.Balance -= amount
	w.AvailableBalance -= amount
	w.UpdatedAt = time.Now()

	return &Transaction{
		WalletID:      w.WalletID,
		UserID:        w.UserID,
		Type:          TransactionTypeWithdraw,
		Amount:        amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  w.Balance,
		Status:        TransactionStatusSuccess,
		CreatedAt:     time.Now(),
	}, nil
}

func (w *Wallet) Freeze(amount int64, reason, referenceNo string) (*FreezeRecord, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	if w.Status == WalletStatusDisabled {
		return nil, ErrWalletDisabled
	}

	if w.AvailableBalance < amount {
		return nil, ErrInsufficientBalance
	}

	w.AvailableBalance -= amount
	w.FrozenBalance += amount
	w.UpdatedAt = time.Now()

	return &FreezeRecord{
		WalletID:    w.WalletID,
		UserID:      w.UserID,
		Amount:      amount,
		Reason:      reason,
		ReferenceNo: referenceNo,
		Status:      "FROZEN",
		FrozenAt:    time.Now(),
	}, nil
}

func (w *Wallet) Unfreeze(amount int64, reason string) (*FreezeRecord, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	if w.FrozenBalance < amount {
		return nil, ErrInsufficientBalance
	}

	w.FrozenBalance -= amount
	w.AvailableBalance += amount
	w.UpdatedAt = time.Now()

	now := time.Now()
	return &FreezeRecord{
		WalletID:   w.WalletID,
		UserID:     w.UserID,
		Amount:     amount,
		Reason:     reason,
		Status:     "UNFROZEN",
		UnfrozenAt: &now,
	}, nil
}

func (w *Wallet) Transfer(toWallet *Wallet, amount int64, password string, limits *WalletLimits, dailyUsage *DailyUsage) ([]*Transaction, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	if w.Status == WalletStatusDisabled || toWallet.Status == WalletStatusDisabled {
		return nil, ErrWalletDisabled
	}

	if w.Status == WalletStatusFrozen || toWallet.Status == WalletStatusFrozen {
		return nil, ErrWalletFrozen
	}

	if w.HasPassword {
		if err := w.VerifyPassword(password); err != nil {
			return nil, err
		}
	}

	if w.AvailableBalance < amount {
		return nil, ErrInsufficientBalance
	}

	if limits != nil && limits.DailyTransferLimit > 0 && dailyUsage != nil {
		if dailyUsage.TransferAmount+amount > limits.DailyTransferLimit {
			return nil, ErrExceedDailyLimit
		}
	}

	fromBalanceBefore := w.Balance
	w.Balance -= amount
	w.AvailableBalance -= amount

	toBalanceBefore := toWallet.Balance
	toWallet.Balance += amount
	toWallet.AvailableBalance += amount

	now := time.Now()
	w.UpdatedAt = now
	toWallet.UpdatedAt = now

	return []*Transaction{
		{
			WalletID:         w.WalletID,
			UserID:           w.UserID,
			Type:             TransactionTypeTransfer,
			Amount:           amount,
			BalanceBefore:    fromBalanceBefore,
			BalanceAfter:     w.Balance,
			Status:           TransactionStatusSuccess,
			CounterpartyID:   toWallet.WalletID,
			CounterpartyType: "WALLET",
			CreatedAt:        now,
		},
		{
			WalletID:         toWallet.WalletID,
			UserID:           toWallet.UserID,
			Type:             TransactionTypeDeposit,
			Amount:           amount,
			BalanceBefore:    toBalanceBefore,
			BalanceAfter:     toWallet.Balance,
			Status:           TransactionStatusSuccess,
			CounterpartyID:   w.WalletID,
			CounterpartyType: "WALLET",
			CreatedAt:        now,
		},
	}, nil
}

func (w *Wallet) Payment(amount int64, password string, limits *WalletLimits) (*Transaction, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	if w.Status == WalletStatusDisabled {
		return nil, ErrWalletDisabled
	}

	if w.Status == WalletStatusFrozen {
		return nil, ErrWalletFrozen
	}

	if limits != nil && limits.RequirePasswordMin > 0 && amount >= limits.RequirePasswordMin {
		if w.HasPassword {
			if err := w.VerifyPassword(password); err != nil {
				return nil, err
			}
		}
	}

	if w.AvailableBalance < amount {
		return nil, ErrInsufficientBalance
	}

	balanceBefore := w.Balance
	w.Balance -= amount
	w.AvailableBalance -= amount
	w.UpdatedAt = time.Now()

	return &Transaction{
		WalletID:      w.WalletID,
		UserID:        w.UserID,
		Type:          TransactionTypePayment,
		Amount:        amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  w.Balance,
		Status:        TransactionStatusSuccess,
		CreatedAt:     time.Now(),
	}, nil
}

func (w *Wallet) Refund(amount int64, originalTransactionNo string) (*Transaction, error) {
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	if w.Status == WalletStatusDisabled {
		return nil, ErrWalletDisabled
	}

	balanceBefore := w.Balance
	w.Balance += amount
	w.AvailableBalance += amount
	w.UpdatedAt = time.Now()

	return &Transaction{
		WalletID:      w.WalletID,
		UserID:        w.UserID,
		Type:          TransactionTypeRefund,
		Amount:        amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  w.Balance,
		Status:        TransactionStatusSuccess,
		Remark:        "Refund for " + originalTransactionNo,
		CreatedAt:     time.Now(),
	}, nil
}

func (w *Wallet) FreezeWallet() {
	w.Status = WalletStatusFrozen
	w.UpdatedAt = time.Now()
}

func (w *Wallet) UnfreezeWallet() {
	w.Status = WalletStatusNormal
	w.UpdatedAt = time.Now()
}

func (w *Wallet) Disable() {
	w.Status = WalletStatusDisabled
	w.UpdatedAt = time.Now()
}

func (w *Wallet) Enable() {
	w.Status = WalletStatusNormal
	w.UpdatedAt = time.Now()
}

func (w *Wallet) IsOperational() bool {
	return w.Status == WalletStatusNormal
}

func (w *Wallet) GetTotalBalance() int64 {
	return w.Balance + w.FrozenBalance
}

func NewWalletLimits(walletID uint64) *WalletLimits {
	return &WalletLimits{
		WalletID:             walletID,
		SingleDepositLimit:   500000,
		SingleWithdrawLimit:  100000,
		DailyDepositLimit:    5000000,
		DailyWithdrawLimit:   1000000,
		MonthlyWithdrawLimit: 5000000,
		DailyTransferLimit:   2000000,
		RequirePasswordMin:   10000,
	}
}

func (l *WalletLimits) UpdateLimits(singleDeposit, singleWithdraw, dailyDeposit, dailyWithdraw, monthlyWithdraw, dailyTransfer, requirePasswordMin int64) {
	if singleDeposit > 0 {
		l.SingleDepositLimit = singleDeposit
	}
	if singleWithdraw > 0 {
		l.SingleWithdrawLimit = singleWithdraw
	}
	if dailyDeposit > 0 {
		l.DailyDepositLimit = dailyDeposit
	}
	if dailyWithdraw > 0 {
		l.DailyWithdrawLimit = dailyWithdraw
	}
	if monthlyWithdraw > 0 {
		l.MonthlyWithdrawLimit = monthlyWithdraw
	}
	if dailyTransfer > 0 {
		l.DailyTransferLimit = dailyTransfer
	}
	if requirePasswordMin >= 0 {
		l.RequirePasswordMin = requirePasswordMin
	}
	l.UpdatedAt = time.Now()
}

func NewDailyUsage(walletID uint64, date string) *DailyUsage {
	return &DailyUsage{
		WalletID:  walletID,
		UsageDate: date,
	}
}

func (d *DailyUsage) AddDeposit(amount int64) {
	d.DepositAmount += amount
	d.TransactionCount++
	d.UpdatedAt = time.Now()
}

func (d *DailyUsage) AddWithdraw(amount int64) {
	d.WithdrawAmount += amount
	d.TransactionCount++
	d.UpdatedAt = time.Now()
}

func (d *DailyUsage) AddTransfer(amount int64) {
	d.TransferAmount += amount
	d.TransactionCount++
	d.UpdatedAt = time.Now()
}

func GenerateTransactionNo() string {
	return time.Now().Format("20060102150405") + hex.EncodeToString([]byte{byte(time.Now().Nanosecond() % 256)})
}

func ComparePasswords(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

type WalletRepository interface {
	Create(wallet *Wallet) error
	GetByID(walletID uint64) (*Wallet, error)
	GetByUserID(userID uint64, currency string) (*Wallet, error)
	GetByAccountNo(accountNo string) (*Wallet, error)
	Update(wallet *Wallet) error
	UpdateBalance(walletID uint64, balance, frozenBalance, availableBalance int64) error
	GetLimits(walletID uint64) (*WalletLimits, error)
	SaveLimits(limits *WalletLimits) error
	GetDailyUsage(walletID uint64, date string) (*DailyUsage, error)
	SaveDailyUsage(usage *DailyUsage) error
	Transaction(fn func(tx any) error) error
}

type TransactionRepository interface {
	Create(tx *Transaction) error
	GetByID(id uint64) (*Transaction, error)
	GetByTransactionNo(transactionNo string) (*Transaction, error)
	ListByWalletID(walletID uint64, txType TransactionType, startTime, endTime *time.Time, page, pageSize int) ([]*Transaction, int64, error)
	ListByUserID(userID uint64, txType TransactionType, startTime, endTime *time.Time, page, pageSize int) ([]*Transaction, int64, error)
	Update(tx *Transaction) error
	GetByChannelTradeNo(channelTradeNo string) (*Transaction, error)
}

type FreezeRecordRepository interface {
	Create(record *FreezeRecord) error
	GetByReferenceNo(referenceNo string) (*FreezeRecord, error)
	Update(record *FreezeRecord) error
	ListByWalletID(walletID uint64, status string, page, pageSize int) ([]*FreezeRecord, int64, error)
}

type PasswordRepository interface {
	Create(record *PasswordRecord) error
	GetByWalletID(walletID uint64) (*PasswordRecord, error)
	Update(record *PasswordRecord) error
	IncrementFailedCount(walletID uint64) error
	ResetFailedCount(walletID uint64) error
}

type RiskCheckService interface {
	CheckTransaction(ctx any, tx *Transaction) (*RiskCheckResult, error)
}

type RiskCheckResult struct {
	Pass      bool   `json:"pass"`
	RiskLevel string `json:"risk_level"`
	Message   string `json:"message"`
	CheckID   string `json:"check_id"`
}
