package infrastructure

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wyfcoding/ecommerce/internal/fiat/domain"
	"gorm.io/gorm"
)

type FiatTransactionPO struct {
	ID             uint64           `gorm:"column:id;primaryKey;autoIncrement"`
	TransactionID  string           `gorm:"column:transaction_id;type:varchar(32);uniqueIndex;not null"`
	UserID         uint64           `gorm:"column:user_id;index;not null"`
	Type           string           `gorm:"column:type;type:varchar(20);not null"`
	Amount         decimal.Decimal  `gorm:"column:amount;type:decimal(20,4);not null"`
	Currency       string           `gorm:"column:currency;type:varchar(10);not null"`
	Channel        string           `gorm:"column:channel;type:varchar(20);not null"`
	BankCode       string           `gorm:"column:bank_code;type:varchar(20)"`
	BankAccountID  uint64           `gorm:"column:bank_account_id"`
	Status         string           `gorm:"column:status;type:varchar(20);not null;default:'PENDING'"`
	FeeAmount      decimal.Decimal  `gorm:"column:fee_amount;type:decimal(20,4);not null"`
	FeeCurrency    string           `gorm:"column:fee_currency;type:varchar(10)"`
	ExchangeRate   decimal.Decimal  `gorm:"column:exchange_rate;type:decimal(20,8)"`
	ReferenceNo    string           `gorm:"column:reference_no;type:varchar(64)"`
	ExternalTxID   string           `gorm:"column:external_tx_id;type:varchar(64)"`
	FailReason     string           `gorm:"column:fail_reason;type:varchar(255)"`
	ProcessedAt    *time.Time       `gorm:"column:processed_at"`
	CompletedAt    *time.Time       `gorm:"column:completed_at"`
	CreatedAt      time.Time        `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time        `gorm:"column:updated_at;autoUpdateTime"`
}

func (FiatTransactionPO) TableName() string { return "fiat_transactions" }

type ExchangeRatePO struct {
	ID           uint64          `gorm:"column:id;primaryKey;autoIncrement"`
	FromCurrency string          `gorm:"column:from_currency;type:varchar(10);not null"`
	ToCurrency   string          `gorm:"column:to_currency;type:varchar(10);not null"`
	Rate         decimal.Decimal `gorm:"column:rate;type:decimal(20,8);not null"`
	BuyRate      decimal.Decimal `gorm:"column:buy_rate;type:decimal(20,8);not null"`
	SellRate     decimal.Decimal `gorm:"column:sell_rate;type:decimal(20,8);not null"`
	Source       string          `gorm:"column:source;type:varchar(50)"`
	UpdatedAt    time.Time       `gorm:"column:updated_at;autoUpdateTime"`
}

func (ExchangeRatePO) TableName() string { return "exchange_rates" }

type BankAccountPO struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	UserID      uint64    `gorm:"column:user_id;index;not null"`
	BankName    string    `gorm:"column:bank_name;type:varchar(64);not null"`
	BankCode    string    `gorm:"column:bank_code;type:varchar(20);not null"`
	AccountName string    `gorm:"column:account_name;type:varchar(64);not null"`
	AccountNo   string    `gorm:"column:account_no;type:varchar(32);not null"`
	AccountType string    `gorm:"column:account_type;type:varchar(20)"`
	Currency    string    `gorm:"column:currency;type:varchar(10);not null"`
	Country     string    `gorm:"column:country;type:varchar(10)"`
	SwiftCode   string    `gorm:"column:swift_code;type:varchar(20)"`
	IBAN        string    `gorm:"column:iban;type:varchar(34)"`
	IsVerified  bool      `gorm:"column:is_verified;default:false"`
	IsDefault   bool      `gorm:"column:is_default;default:false"`
	Status      string    `gorm:"column:status;type:varchar(20);not null;default:'ACTIVE'"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (BankAccountPO) TableName() string { return "bank_accounts" }

type FiatChannelPO struct {
	ID          uint64          `gorm:"column:id;primaryKey;autoIncrement"`
	Code        string          `gorm:"column:code;type:varchar(20);uniqueIndex;not null"`
	Name        string          `gorm:"column:name;type:varchar(64);not null"`
	ChannelType string          `gorm:"column:channel_type;type:varchar(20);not null"`
	Currencies  string          `gorm:"column:currencies;type:json"`
	Countries   string          `gorm:"column:countries;type:json"`
	MinAmount   decimal.Decimal `gorm:"column:min_amount;type:decimal(20,4);not null"`
	MaxAmount   decimal.Decimal `gorm:"column:max_amount;type:decimal(20,4);not null"`
	FeeRate     decimal.Decimal `gorm:"column:fee_rate;type:decimal(10,4);not null"`
	FeeFixed    decimal.Decimal `gorm:"column:fee_fixed;type:decimal(20,4);not null"`
	IsActive    bool            `gorm:"column:is_active;default:true"`
	Priority    int             `gorm:"column:priority;default:0"`
	CreatedAt   time.Time       `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time       `gorm:"column:updated_at;autoUpdateTime"`
}

func (FiatChannelPO) TableName() string { return "fiat_channels" }

type GormFiatTransactionRepository struct {
	db *gorm.DB
}

func NewGormFiatTransactionRepository(db *gorm.DB) *GormFiatTransactionRepository {
	return &GormFiatTransactionRepository{db: db}
}

func (r *GormFiatTransactionRepository) Save(ctx context.Context, tx *domain.FiatTransaction) error {
	po := toFiatTransactionPO(tx)
	return r.db.WithContext(ctx).Create(po).Error
}

func (r *GormFiatTransactionRepository) Update(ctx context.Context, tx *domain.FiatTransaction) error {
	po := toFiatTransactionPO(tx)
	return r.db.WithContext(ctx).Save(po).Error
}

func (r *GormFiatTransactionRepository) GetByID(ctx context.Context, id string) (*domain.FiatTransaction, error) {
	var po FiatTransactionPO
	err := r.db.WithContext(ctx).Where("transaction_id = ?", id).First(&po).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toFiatTransaction(&po), nil
}

func (r *GormFiatTransactionRepository) GetByReferenceNo(ctx context.Context, refNo string) (*domain.FiatTransaction, error) {
	var po FiatTransactionPO
	err := r.db.WithContext(ctx).Where("reference_no = ?", refNo).First(&po).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toFiatTransaction(&po), nil
}

func (r *GormFiatTransactionRepository) GetByExternalTxID(ctx context.Context, externalTxID string) (*domain.FiatTransaction, error) {
	var po FiatTransactionPO
	err := r.db.WithContext(ctx).Where("external_tx_id = ?", externalTxID).First(&po).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toFiatTransaction(&po), nil
}

func (r *GormFiatTransactionRepository) ListByUserID(ctx context.Context, userID uint64, txType domain.TransactionType, status domain.TransactionStatus, page, pageSize int) ([]*domain.FiatTransaction, int64, error) {
	var pos []*FiatTransactionPO
	var total int64

	query := r.db.WithContext(ctx).Model(&FiatTransactionPO{}).Where("user_id = ?", userID)
	if txType != "" {
		query = query.Where("type = ?", txType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	txs := make([]*domain.FiatTransaction, len(pos))
	for i, po := range pos {
		txs[i] = toFiatTransaction(po)
	}

	return txs, total, nil
}

type GormExchangeRateRepository struct {
	db *gorm.DB
}

func NewGormExchangeRateRepository(db *gorm.DB) *GormExchangeRateRepository {
	return &GormExchangeRateRepository{db: db}
}

func (r *GormExchangeRateRepository) Save(ctx context.Context, rate *domain.ExchangeRate) error {
	po := toExchangeRatePO(rate)
	return r.db.WithContext(ctx).Create(po).Error
}

func (r *GormExchangeRateRepository) Get(ctx context.Context, from, to string) (*domain.ExchangeRate, error) {
	var po ExchangeRatePO
	err := r.db.WithContext(ctx).Where("from_currency = ? AND to_currency = ?", from, to).First(&po).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toExchangeRate(&po), nil
}

func (r *GormExchangeRateRepository) GetAll(ctx context.Context) ([]*domain.ExchangeRate, error) {
	var pos []*ExchangeRatePO
	err := r.db.WithContext(ctx).Find(&pos).Error
	if err != nil {
		return nil, err
	}

	rates := make([]*domain.ExchangeRate, len(pos))
	for i, po := range pos {
		rates[i] = toExchangeRate(po)
	}
	return rates, nil
}

func (r *GormExchangeRateRepository) Update(ctx context.Context, rate *domain.ExchangeRate) error {
	po := toExchangeRatePO(rate)
	return r.db.WithContext(ctx).Save(po).Error
}

type GormBankAccountRepository struct {
	db *gorm.DB
}

func NewGormBankAccountRepository(db *gorm.DB) *GormBankAccountRepository {
	return &GormBankAccountRepository{db: db}
}

func (r *GormBankAccountRepository) Save(ctx context.Context, account *domain.BankAccount) error {
	po := toBankAccountPO(account)
	return r.db.WithContext(ctx).Create(po).Error
}

func (r *GormBankAccountRepository) Update(ctx context.Context, account *domain.BankAccount) error {
	po := toBankAccountPO(account)
	return r.db.WithContext(ctx).Save(po).Error
}

func (r *GormBankAccountRepository) GetByID(ctx context.Context, id uint64) (*domain.BankAccount, error) {
	var po BankAccountPO
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&po).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toBankAccount(&po), nil
}

func (r *GormBankAccountRepository) GetByUserID(ctx context.Context, userID uint64) ([]*domain.BankAccount, error) {
	var pos []*BankAccountPO
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&pos).Error
	if err != nil {
		return nil, err
	}

	accounts := make([]*domain.BankAccount, len(pos))
	for i, po := range pos {
		accounts[i] = toBankAccount(po)
	}
	return accounts, nil
}

func (r *GormBankAccountRepository) GetDefaultByUserID(ctx context.Context, userID uint64, currency string) (*domain.BankAccount, error) {
	var po BankAccountPO
	query := r.db.WithContext(ctx).Where("user_id = ? AND is_default = ? AND status = ?", userID, true, "ACTIVE")
	if currency != "" {
		query = query.Where("currency = ?", currency)
	}
	err := query.First(&po).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toBankAccount(&po), nil
}

func (r *GormBankAccountRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&BankAccountPO{}, id).Error
}

type GormFiatChannelRepository struct {
	db *gorm.DB
}

func NewGormFiatChannelRepository(db *gorm.DB) *GormFiatChannelRepository {
	return &GormFiatChannelRepository{db: db}
}

func (r *GormFiatChannelRepository) GetByCode(ctx context.Context, code string) (*domain.FiatChannel, error) {
	var po FiatChannelPO
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&po).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toFiatChannel(&po), nil
}

func (r *GormFiatChannelRepository) GetActiveChannels(ctx context.Context) ([]*domain.FiatChannel, error) {
	var pos []*FiatChannelPO
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("priority DESC").Find(&pos).Error
	if err != nil {
		return nil, err
	}

	channels := make([]*domain.FiatChannel, len(pos))
	for i, po := range pos {
		channels[i] = toFiatChannel(po)
	}
	return channels, nil
}

func (r *GormFiatChannelRepository) GetChannelsByCurrency(ctx context.Context, currency string) ([]*domain.FiatChannel, error) {
	var pos []*FiatChannelPO
	err := r.db.WithContext(ctx).
		Where("is_active = ? AND JSON_CONTAINS(currencies, ?)", true, "\""+currency+"\"").
		Order("priority DESC").
		Find(&pos).Error
	if err != nil {
		return nil, err
	}

	channels := make([]*domain.FiatChannel, len(pos))
	for i, po := range pos {
		channels[i] = toFiatChannel(po)
	}
	return channels, nil
}

func toFiatTransactionPO(tx *domain.FiatTransaction) *FiatTransactionPO {
	return &FiatTransactionPO{
		TransactionID: tx.TransactionID,
		UserID:        tx.UserID,
		Type:          string(tx.Type),
		Amount:        tx.Amount,
		Currency:      tx.Currency,
		Channel:       string(tx.Channel),
		BankCode:      tx.BankCode,
		BankAccountID: tx.BankAccountID,
		Status:        string(tx.Status),
		FeeAmount:     tx.FeeAmount,
		FeeCurrency:   tx.FeeCurrency,
		ExchangeRate:  tx.ExchangeRate,
		ReferenceNo:   tx.ReferenceNo,
		ExternalTxID:  tx.ExternalTxID,
		FailReason:    tx.FailReason,
		ProcessedAt:   tx.ProcessedAt,
		CompletedAt:   tx.CompletedAt,
		CreatedAt:     tx.CreatedAt,
		UpdatedAt:     tx.UpdatedAt,
	}
}

func toFiatTransaction(po *FiatTransactionPO) *domain.FiatTransaction {
	return &domain.FiatTransaction{
		ID:            po.ID,
		TransactionID: po.TransactionID,
		UserID:        po.UserID,
		Type:          domain.TransactionType(po.Type),
		Amount:        po.Amount,
		Currency:      po.Currency,
		Channel:       domain.ChannelType(po.Channel),
		BankCode:      po.BankCode,
		BankAccountID: po.BankAccountID,
		Status:        domain.TransactionStatus(po.Status),
		FeeAmount:     po.FeeAmount,
		FeeCurrency:   po.FeeCurrency,
		ExchangeRate:  po.ExchangeRate,
		ReferenceNo:   po.ReferenceNo,
		ExternalTxID:  po.ExternalTxID,
		FailReason:    po.FailReason,
		ProcessedAt:   po.ProcessedAt,
		CompletedAt:   po.CompletedAt,
		CreatedAt:     po.CreatedAt,
		UpdatedAt:     po.UpdatedAt,
	}
}

func toExchangeRatePO(rate *domain.ExchangeRate) *ExchangeRatePO {
	return &ExchangeRatePO{
		ID:           rate.ID,
		FromCurrency: rate.FromCurrency,
		ToCurrency:   rate.ToCurrency,
		Rate:         rate.Rate,
		BuyRate:      rate.BuyRate,
		SellRate:     rate.SellRate,
		Source:       rate.Source,
		UpdatedAt:    rate.UpdatedAt,
	}
}

func toExchangeRate(po *ExchangeRatePO) *domain.ExchangeRate {
	return &domain.ExchangeRate{
		ID:           po.ID,
		FromCurrency: po.FromCurrency,
		ToCurrency:   po.ToCurrency,
		Rate:         po.Rate,
		BuyRate:      po.BuyRate,
		SellRate:     po.SellRate,
		Source:       po.Source,
		UpdatedAt:    po.UpdatedAt,
	}
}

func toBankAccountPO(account *domain.BankAccount) *BankAccountPO {
	return &BankAccountPO{
		ID:          account.ID,
		UserID:      account.UserID,
		BankName:    account.BankName,
		BankCode:    account.BankCode,
		AccountName: account.AccountName,
		AccountNo:   account.AccountNo,
		AccountType: account.AccountType,
		Currency:    account.Currency,
		Country:     account.Country,
		SwiftCode:   account.SwiftCode,
		IBAN:        account.IBAN,
		IsVerified:  account.IsVerified,
		IsDefault:   account.IsDefault,
		Status:      string(account.Status),
		CreatedAt:   account.CreatedAt,
		UpdatedAt:   account.UpdatedAt,
	}
}

func toBankAccount(po *BankAccountPO) *domain.BankAccount {
	return &domain.BankAccount{
		ID:          po.ID,
		UserID:      po.UserID,
		BankName:    po.BankName,
		BankCode:    po.BankCode,
		AccountName: po.AccountName,
		AccountNo:   po.AccountNo,
		AccountType: po.AccountType,
		Currency:    po.Currency,
		Country:     po.Country,
		SwiftCode:   po.SwiftCode,
		IBAN:        po.IBAN,
		IsVerified:  po.IsVerified,
		IsDefault:   po.IsDefault,
		Status:      domain.AccountStatus(po.Status),
		CreatedAt:   po.CreatedAt,
		UpdatedAt:   po.UpdatedAt,
	}
}

func toFiatChannel(po *FiatChannelPO) *domain.FiatChannel {
	return &domain.FiatChannel{
		ID:          po.ID,
		Code:        po.Code,
		Name:        po.Name,
		ChannelType: domain.ChannelType(po.ChannelType),
		Currencies:  []string{},
		Countries:   []string{},
		MinAmount:   po.MinAmount,
		MaxAmount:   po.MaxAmount,
		FeeRate:     po.FeeRate,
		FeeFixed:    po.FeeFixed,
		IsActive:    po.IsActive,
		Priority:    po.Priority,
		CreatedAt:   po.CreatedAt,
		UpdatedAt:   po.UpdatedAt,
	}
}
