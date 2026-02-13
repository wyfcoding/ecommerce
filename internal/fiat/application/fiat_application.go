package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/wyfcoding/ecommerce/internal/fiat/domain"
)

type DepositCommand struct {
	UserID       uint64
	Amount       decimal.Decimal
	Currency     string
	Channel      domain.ChannelType
	BankAccountID uint64
}

type DepositResult struct {
	TransactionID string
	Status        domain.TransactionStatus
	FeeAmount     decimal.Decimal
}

type WithdrawCommand struct {
	UserID        uint64
	Amount        decimal.Decimal
	Currency      string
	Channel       domain.ChannelType
	BankAccountID uint64
}

type WithdrawResult struct {
	TransactionID string
	Status        domain.TransactionStatus
	FeeAmount     decimal.Decimal
}

type ExchangeCommand struct {
	UserID       uint64
	Amount       decimal.Decimal
	FromCurrency string
	ToCurrency   string
}

type ExchangeResult struct {
	TransactionID    string
	OriginalAmount   decimal.Decimal
	ExchangedAmount  decimal.Decimal
	ExchangeRate     decimal.Decimal
	FeeAmount        decimal.Decimal
}

type FiatApplicationService struct {
	txRepo         domain.FiatTransactionRepository
	rateRepo       domain.ExchangeRateRepository
	bankAccountRepo domain.BankAccountRepository
	channelRepo    domain.FiatChannelRepository
	fiatService    *domain.FiatService
	bankGateway    domain.BankGateway
	logger         *slog.Logger
}

func NewFiatApplicationService(
	txRepo domain.FiatTransactionRepository,
	rateRepo domain.ExchangeRateRepository,
	bankAccountRepo domain.BankAccountRepository,
	channelRepo domain.FiatChannelRepository,
	fiatService *domain.FiatService,
	bankGateway domain.BankGateway,
	logger *slog.Logger,
) *FiatApplicationService {
	return &FiatApplicationService{
		txRepo:         txRepo,
		rateRepo:       rateRepo,
		bankAccountRepo: bankAccountRepo,
		channelRepo:    channelRepo,
		fiatService:    fiatService,
		bankGateway:    bankGateway,
		logger:         logger,
	}
}

func (s *FiatApplicationService) Deposit(ctx context.Context, cmd *DepositCommand) (*DepositResult, error) {
	channel, err := s.channelRepo.GetByCode(ctx, string(cmd.Channel))
	if err != nil {
		return nil, err
	}
	if !channel.IsActive {
		return nil, domain.ErrChannelNotAvailable
	}
	if !channel.SupportsCurrency(cmd.Currency) {
		return nil, domain.ErrCurrencyNotSupported
	}
	if cmd.Amount.LessThan(channel.MinAmount) || cmd.Amount.GreaterThan(channel.MaxAmount) {
		return nil, domain.ErrAmountOutOfRange
	}

	bankAccount, err := s.bankAccountRepo.GetByID(ctx, cmd.BankAccountID)
	if err != nil {
		return nil, err
	}
	if bankAccount == nil || bankAccount.UserID != cmd.UserID {
		return nil, domain.ErrBankAccountNotFound
	}

	transactionID := fmt.Sprintf("FI%s", uuid.New().String()[:16])
	tx := domain.NewFiatTransaction(transactionID, cmd.UserID, domain.TxTypeDeposit, cmd.Amount, cmd.Currency, cmd.Channel)
	tx.SetBankInfo(bankAccount.BankCode, cmd.BankAccountID)

	fee := channel.CalculateFee(cmd.Amount)
	tx.SetFee(fee, cmd.Currency)

	if err := s.txRepo.Save(ctx, tx); err != nil {
		return nil, err
	}

	s.logger.Info("deposit transaction created", "transaction_id", transactionID, "user_id", cmd.UserID, "amount", cmd.Amount)

	return &DepositResult{
		TransactionID: transactionID,
		Status:        tx.Status,
		FeeAmount:     fee,
	}, nil
}

func (s *FiatApplicationService) Withdraw(ctx context.Context, cmd *WithdrawCommand) (*WithdrawResult, error) {
	channel, err := s.channelRepo.GetByCode(ctx, string(cmd.Channel))
	if err != nil {
		return nil, err
	}
	if !channel.IsActive {
		return nil, domain.ErrChannelNotAvailable
	}
	if !channel.SupportsCurrency(cmd.Currency) {
		return nil, domain.ErrCurrencyNotSupported
	}
	if cmd.Amount.LessThan(channel.MinAmount) || cmd.Amount.GreaterThan(channel.MaxAmount) {
		return nil, domain.ErrAmountOutOfRange
	}

	bankAccount, err := s.bankAccountRepo.GetByID(ctx, cmd.BankAccountID)
	if err != nil {
		return nil, err
	}
	if bankAccount == nil || bankAccount.UserID != cmd.UserID {
		return nil, domain.ErrBankAccountNotFound
	}
	if bankAccount.Status != domain.AccountStatusActive {
		return nil, fmt.Errorf("bank account is not active")
	}

	transactionID := fmt.Sprintf("FO%s", uuid.New().String()[:16])
	tx := domain.NewFiatTransaction(transactionID, cmd.UserID, domain.TxTypeWithdraw, cmd.Amount, cmd.Currency, cmd.Channel)
	tx.SetBankInfo(bankAccount.BankCode, cmd.BankAccountID)

	fee := channel.CalculateFee(cmd.Amount)
	tx.SetFee(fee, cmd.Currency)

	if err := s.txRepo.Save(ctx, tx); err != nil {
		return nil, err
	}

	if err := tx.StartProcessing(); err != nil {
		return nil, err
	}

	result, err := s.bankGateway.InitiateTransfer(ctx, &domain.BankTransferRequest{
		TransactionID: transactionID,
		BankCode:      bankAccount.BankCode,
		AccountNo:     bankAccount.AccountNo,
		AccountName:   bankAccount.AccountName,
		Amount:        cmd.Amount.Sub(fee),
		Currency:      cmd.Currency,
		Reference:     transactionID,
	})
	if err != nil {
		tx.Fail(err.Error())
		s.txRepo.Update(ctx, tx)
		return nil, err
	}

	if result.Status == domain.TxStatusSuccess {
		tx.Complete(result.ExternalTxID)
	} else {
		tx.Fail(result.FailReason)
	}

	if err := s.txRepo.Update(ctx, tx); err != nil {
		return nil, err
	}

	s.logger.Info("withdraw transaction processed", "transaction_id", transactionID, "user_id", cmd.UserID, "status", tx.Status)

	return &WithdrawResult{
		TransactionID: transactionID,
		Status:        tx.Status,
		FeeAmount:     fee,
	}, nil
}

func (s *FiatApplicationService) Exchange(ctx context.Context, cmd *ExchangeCommand) (*ExchangeResult, error) {
	if cmd.FromCurrency == cmd.ToCurrency {
		return nil, fmt.Errorf("same currency")
	}

	rate, err := s.rateRepo.Get(ctx, cmd.FromCurrency, cmd.ToCurrency)
	if err != nil {
		return nil, err
	}

	exchangedAmount := cmd.Amount.Mul(rate.Rate)
	fee := exchangedAmount.Mul(decimal.NewFromFloat(0.001))

	transactionID := fmt.Sprintf("FE%s", uuid.New().String()[:16])
	tx := domain.NewFiatTransaction(transactionID, cmd.UserID, domain.TxTypeExchange, cmd.Amount, cmd.FromCurrency, domain.ChannelShortcut)
	tx.SetExchangeRate(rate.Rate)
	tx.SetFee(fee, cmd.ToCurrency)

	if err := s.txRepo.Save(ctx, tx); err != nil {
		return nil, err
	}

	return &ExchangeResult{
		TransactionID:   transactionID,
		OriginalAmount:  cmd.Amount,
		ExchangedAmount: exchangedAmount,
		ExchangeRate:    rate.Rate,
		FeeAmount:       fee,
	}, nil
}

func (s *FiatApplicationService) GetRate(ctx context.Context, from, to string) (float64, error) {
	return s.fiatService.GetRate(ctx, from, to)
}

func (s *FiatApplicationService) ConvertCurrency(ctx context.Context, amount decimal.Decimal, from, to string) (decimal.Decimal, error) {
	return s.fiatService.ConvertCurrency(ctx, amount, from, to)
}

func (s *FiatApplicationService) GetTransaction(ctx context.Context, transactionID string) (*domain.FiatTransaction, error) {
	return s.txRepo.GetByID(ctx, transactionID)
}

func (s *FiatApplicationService) ListTransactions(ctx context.Context, userID uint64, txType domain.TransactionType, status domain.TransactionStatus, page, pageSize int) ([]*domain.FiatTransaction, int64, error) {
	return s.txRepo.ListByUserID(ctx, userID, txType, status, page, pageSize)
}

func (s *FiatApplicationService) AddBankAccount(ctx context.Context, userID uint64, bankName, bankCode, accountName, accountNo, currency string) (*domain.BankAccount, error) {
	account := domain.NewBankAccount(userID, bankName, bankCode, accountName, accountNo, currency)
	if err := s.bankAccountRepo.Save(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *FiatApplicationService) ListBankAccounts(ctx context.Context, userID uint64) ([]*domain.BankAccount, error) {
	return s.bankAccountRepo.GetByUserID(ctx, userID)
}

func (s *FiatApplicationService) VerifyBankAccount(ctx context.Context, accountID uint64) error {
	account, err := s.bankAccountRepo.GetByID(ctx, accountID)
	if err != nil {
		return err
	}
	if account == nil {
		return domain.ErrBankAccountNotFound
	}

	result, err := s.bankGateway.ValidateAccount(ctx, &domain.ValidateAccountRequest{
		BankCode:    account.BankCode,
		AccountNo:   account.AccountNo,
		AccountName: account.AccountName,
	})
	if err != nil {
		return err
	}

	if result.Valid {
		account.Verify()
		return s.bankAccountRepo.Update(ctx, account)
	}

	return fmt.Errorf("bank account validation failed")
}

func (s *FiatApplicationService) SetDefaultBankAccount(ctx context.Context, userID, accountID uint64) error {
	accounts, err := s.bankAccountRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	for _, a := range accounts {
		if a.ID == accountID {
			a.SetDefault(true)
		} else {
			a.SetDefault(false)
		}
		if err := s.bankAccountRepo.Update(ctx, a); err != nil {
			return err
		}
	}
	return nil
}

func (s *FiatApplicationService) GetAvailableChannels(ctx context.Context, currency string) ([]*domain.FiatChannel, error) {
	return s.fiatService.GetAvailableChannels(ctx, currency)
}

func (s *FiatApplicationService) UpdateExchangeRates(ctx context.Context) error {
	return s.fiatService.UpdateExchangeRates(ctx)
}

func (s *FiatApplicationService) ProcessCallback(ctx context.Context, transactionID string, success bool, externalTxID, failReason string) error {
	tx, err := s.txRepo.GetByID(ctx, transactionID)
	if err != nil {
		return err
	}
	if tx == nil {
		return domain.ErrTransactionNotFound
	}

	if success {
		if err := tx.Complete(externalTxID); err != nil {
			return err
		}
	} else {
		if err := tx.Fail(failReason); err != nil {
			return err
		}
	}

	return s.txRepo.Update(ctx, tx)
}

type TransactionDTO struct {
	ID             uint64           `json:"id"`
	TransactionID  string           `json:"transaction_id"`
	UserID         uint64           `json:"user_id"`
	Type           string           `json:"type"`
	Amount         decimal.Decimal  `json:"amount"`
	Currency       string           `json:"currency"`
	Channel        string           `json:"channel"`
	Status         string           `json:"status"`
	FeeAmount      decimal.Decimal  `json:"fee_amount"`
	FeeCurrency    string           `json:"fee_currency"`
	ExchangeRate   decimal.Decimal  `json:"exchange_rate"`
	ExternalTxID   string           `json:"external_tx_id"`
	FailReason     string           `json:"fail_reason"`
	CreatedAt      time.Time        `json:"created_at"`
	CompletedAt    *time.Time       `json:"completed_at"`
}

func toTransactionDTO(tx *domain.FiatTransaction) *TransactionDTO {
	return &TransactionDTO{
		ID:            tx.ID,
		TransactionID: tx.TransactionID,
		UserID:        tx.UserID,
		Type:          string(tx.Type),
		Amount:        tx.Amount,
		Currency:      tx.Currency,
		Channel:       string(tx.Channel),
		Status:        string(tx.Status),
		FeeAmount:     tx.FeeAmount,
		FeeCurrency:   tx.FeeCurrency,
		ExchangeRate:  tx.ExchangeRate,
		ExternalTxID:  tx.ExternalTxID,
		FailReason:    tx.FailReason,
		CreatedAt:     tx.CreatedAt,
		CompletedAt:   tx.CompletedAt,
	}
}
