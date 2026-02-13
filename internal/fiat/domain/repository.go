package domain

import (
	"context"

	"github.com/shopspring/decimal"
)

type FiatTransactionRepository interface {
	Save(ctx context.Context, tx *FiatTransaction) error
	Update(ctx context.Context, tx *FiatTransaction) error
	GetByID(ctx context.Context, id string) (*FiatTransaction, error)
	GetByReferenceNo(ctx context.Context, refNo string) (*FiatTransaction, error)
	GetByExternalTxID(ctx context.Context, externalTxID string) (*FiatTransaction, error)
	ListByUserID(ctx context.Context, userID uint64, txType TransactionType, status TransactionStatus, page, pageSize int) ([]*FiatTransaction, int64, error)
}

type ExchangeRateRepository interface {
	Save(ctx context.Context, rate *ExchangeRate) error
	Get(ctx context.Context, from, to string) (*ExchangeRate, error)
	GetAll(ctx context.Context) ([]*ExchangeRate, error)
	Update(ctx context.Context, rate *ExchangeRate) error
}

type BankAccountRepository interface {
	Save(ctx context.Context, account *BankAccount) error
	Update(ctx context.Context, account *BankAccount) error
	GetByID(ctx context.Context, id uint64) (*BankAccount, error)
	GetByUserID(ctx context.Context, userID uint64) ([]*BankAccount, error)
	GetDefaultByUserID(ctx context.Context, userID uint64, currency string) (*BankAccount, error)
	Delete(ctx context.Context, id uint64) error
}

type FiatChannelRepository interface {
	GetByCode(ctx context.Context, code string) (*FiatChannel, error)
	GetActiveChannels(ctx context.Context) ([]*FiatChannel, error)
	GetChannelsByCurrency(ctx context.Context, currency string) ([]*FiatChannel, error)
}

type CurrencyRepository interface {
	GetByCode(ctx context.Context, code string) (*Currency, error)
	GetAllActive(ctx context.Context) ([]*Currency, error)
}

type ExchangeRateProvider interface {
	GetRate(ctx context.Context, from, to string) (decimal.Decimal, error)
	GetAllRates(ctx context.Context) (map[string]decimal.Decimal, error)
}

type BankGateway interface {
	InitiateTransfer(ctx context.Context, req *BankTransferRequest) (*BankTransferResult, error)
	QueryTransfer(ctx context.Context, transactionID string) (*BankTransferResult, error)
	ValidateAccount(ctx context.Context, req *ValidateAccountRequest) (*ValidateAccountResult, error)
}

type BankTransferRequest struct {
	TransactionID string
	BankCode      string
	AccountNo     string
	AccountName   string
	Amount        decimal.Decimal
	Currency      string
	Reference     string
}

type BankTransferResult struct {
	ExternalTxID string
	Status       TransactionStatus
	FailReason   string
}

type ValidateAccountRequest struct {
	BankCode    string
	AccountNo   string
	AccountName string
}

type ValidateAccountResult struct {
	Valid     bool
	AccountName string
	BankName   string
}

type FiatService struct {
	txRepo       FiatTransactionRepository
	rateRepo     ExchangeRateRepository
	channelRepo  FiatChannelRepository
	rateProvider ExchangeRateProvider
}

func NewFiatService(
	txRepo FiatTransactionRepository,
	rateRepo ExchangeRateRepository,
	channelRepo FiatChannelRepository,
	rateProvider ExchangeRateProvider,
) *FiatService {
	return &FiatService{
		txRepo:       txRepo,
		rateRepo:     rateRepo,
		channelRepo:  channelRepo,
		rateProvider: rateProvider,
	}
}

func (s *FiatService) Exchange(ctx context.Context, from, to string, amount int64) (int64, float64, error) {
	if from == to {
		return amount, 1.0, nil
	}
	rate, err := s.rateRepo.Get(ctx, from, to)
	if err != nil {
		return 0, 0, err
	}

	exchangedAmount := decimal.NewFromInt(amount).Mul(rate.Rate).IntPart()
	return exchangedAmount, rate.Rate.InexactFloat64(), nil
}

func (s *FiatService) GetRate(ctx context.Context, from, to string) (float64, error) {
	rate, err := s.rateRepo.Get(ctx, from, to)
	if err != nil {
		return 0, err
	}
	return rate.Rate.InexactFloat64(), nil
}

func (s *FiatService) ConvertCurrency(ctx context.Context, amount decimal.Decimal, from, to string) (decimal.Decimal, error) {
	if from == to {
		return amount, nil
	}
	rate, err := s.rateRepo.Get(ctx, from, to)
	if err != nil {
		return decimal.Zero, err
	}
	return amount.Mul(rate.Rate), nil
}

func (s *FiatService) GetAvailableChannels(ctx context.Context, currency string) ([]*FiatChannel, error) {
	return s.channelRepo.GetChannelsByCurrency(ctx, currency)
}

func (s *FiatService) UpdateExchangeRates(ctx context.Context) error {
	rates, err := s.rateProvider.GetAllRates(ctx)
	if err != nil {
		return err
	}

	for pair, rate := range rates {
		from := pair[:3]
		to := pair[4:]
		existing, err := s.rateRepo.Get(ctx, from, to)
		if err != nil {
			newRate := NewExchangeRate(from, to, rate, rate, rate, "provider")
			if err := s.rateRepo.Save(ctx, newRate); err != nil {
				return err
			}
		} else {
			existing.Update(rate, rate, rate)
			if err := s.rateRepo.Update(ctx, existing); err != nil {
				return err
			}
		}
	}
	return nil
}

type FiatCalculator struct {
	rateRepo ExchangeRateRepository
}

func NewFiatCalculator(rateRepo ExchangeRateRepository) *FiatCalculator {
	return &FiatCalculator{rateRepo: rateRepo}
}

func (c *FiatCalculator) Calculate(ctx context.Context, amount decimal.Decimal, from, to string) (decimal.Decimal, error) {
	if from == to {
		return amount, nil
	}
	rate, err := c.rateRepo.Get(ctx, from, to)
	if err != nil {
		return decimal.Zero, err
	}
	if rate.Rate.IsZero() {
		return decimal.Zero, ErrExchangeRateNotFound
	}
	return amount.Mul(rate.Rate), nil
}

func (c *FiatCalculator) CalculateWithFee(ctx context.Context, amount decimal.Decimal, from, to string, feeRate, feeFixed decimal.Decimal) (decimal.Decimal, decimal.Decimal, error) {
	convertedAmount, err := c.Calculate(ctx, amount, from, to)
	if err != nil {
		return decimal.Zero, decimal.Zero, err
	}
	fee := convertedAmount.Mul(feeRate).Add(feeFixed)
	return convertedAmount, fee, nil
}
