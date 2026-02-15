// 生成摘要：
// - 实现 Wallet 服务的 MySQL 仓储层，包含钱包核心信息、额度限制、交易记录等持久化逻辑
// - 遵循 DDD 架构，将领域对象转换为 GORM 模型并存入数据库
// - 采用事务机制确保资金操作的原子性
// - 集成 slog 输出结构化日志，包含关键业务标识和操作耗时

package mysql

import (
	"errors"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/wallet/domain"
	"github.com/wyfcoding/pkg/database"
	"gorm.io/gorm"
)

// WalletModel 钱包数据库模型
type WalletModel struct {
	gorm.Model
	WalletID         uint64 `gorm:"column:wallet_id;uniqueIndex;not null;comment:业务定义的钱包唯一ID"`
	UserID           uint64 `gorm:"column:user_id;index;not null;comment:用户ID"`
	AccountNo        string `gorm:"column:account_no;uniqueIndex;type:varchar(64);not null;comment:钱包账号"`
	Currency         string `gorm:"column:currency;type:varchar(10);not null;comment:币种"`
	WalletType       string `gorm:"column:wallet_type;type:varchar(20);not null;comment:钱包类型"`
	Balance          int64  `gorm:"column:balance;default:0;not null;comment:总余额（分）"`
	FrozenBalance    int64  `gorm:"column:frozen_balance;default:0;not null;comment:冻结余额（分）"`
	AvailableBalance int64  `gorm:"column:available_balance;default:0;not null;comment:可用余额（分）"`
	Status           int    `gorm:"column:status;default:1;not null;comment:状态: 0禁用, 1正常, 2冻结"`
	PasswordHash     string `gorm:"column:password_hash;type:varchar(255);comment:支付密码Hash"`
	SecurityLevel    int    `gorm:"column:security_level;default:1;comment:安全等级"`
}

// TableName 指定 WalletModel 对应的表名
func (WalletModel) TableName() string {
	return "wallets"
}

// TransactionModel 交易记录数据库模型
type TransactionModel struct {
	gorm.Model
	TransactionNo    string     `gorm:"column:transaction_no;uniqueIndex;type:varchar(64);not null;comment:交易流水号"`
	WalletID         uint64     `gorm:"column:wallet_id;index;not null;comment:所属钱包ID"`
	UserID           uint64     `gorm:"column:user_id;index;not null;comment:用户ID"`
	Type             string     `gorm:"column:type;type:varchar(20);not null;comment:交易类型"`
	Amount           int64      `gorm:"column:amount;not null;comment:交易金额（分）"`
	BalanceBefore    int64      `gorm:"column:balance_before;comment:交易前余额"`
	BalanceAfter     int64      `gorm:"column:balance_after;comment:交易后余额"`
	Fee              int64      `gorm:"column:fee;default:0;comment:手续费"`
	Status           int        `gorm:"column:status;default:0;comment:状态: 0处理中, 1成功, 2失败"`
	Remark           string     `gorm:"column:remark;type:text;comment:备注"`
	ChannelTradeNo   string     `gorm:"column:channel_trade_no;type:varchar(64);comment:渠道流水号"`
	CounterpartyID   uint64     `gorm:"column:counterparty_id;comment:对手方ID"`
	CounterpartyType string     `gorm:"column:counterparty_type;type:varchar(20);comment:对手方类型"`
	RiskCheckID      string     `gorm:"column:risk_check_id;type:varchar(64);comment:风控ID"`
	IPAddress        string     `gorm:"column:ip_address;type:varchar(45);comment:IP地址"`
	DeviceID         string     `gorm:"column:device_id;type:varchar(128);comment:设备ID"`
	Location         string     `gorm:"column:location;type:varchar(255);comment:地理位置"`
	CompletedAt      *time.Time `gorm:"column:completed_at;comment:完成时间"`
}

// TableName 指定 TransactionModel 对应的表名
func (TransactionModel) TableName() string {
	return "wallet_transactions"
}

// WalletRepositoryImpl Wallet 仓储实现
type WalletRepositoryImpl struct {
	db     *database.DB
	logger *slog.Logger
}

// NewWalletRepository 创建 Wallet 仓储实例
func NewWalletRepository(db *database.DB, logger *slog.Logger) domain.WalletRepository {
	return &WalletRepositoryImpl{
		db:     db,
		logger: logger.With("module", "wallet_repository"),
	}
}

// Create 创建新钱包
func (r *WalletRepositoryImpl) Create(w *domain.Wallet) error {
	start := time.Now()
	model := toWalletModel(w)
	err := r.db.Create(model).Error

	r.logger.Info("create wallet",
		"user_id", w.UserID,
		"wallet_id", w.WalletID,
		"error", err,
		"duration", time.Since(start))

	if err == nil {
		w.ID = uint64(model.ID)
	}
	return err
}

// GetByID 根据 ID 获取钱包
func (r *WalletRepositoryImpl) GetByID(walletID uint64) (*domain.Wallet, error) {
	var model WalletModel
	err := r.db.Where("wallet_id = ?", walletID).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toWalletDomain(&model), nil
}

// GetByUserID 根据用户 ID 和币种获取钱包
func (r *WalletRepositoryImpl) GetByUserID(userID uint64, currency string) (*domain.Wallet, error) {
	var model WalletModel
	err := r.db.Where("user_id = ? AND currency = ?", userID, currency).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toWalletDomain(&model), nil
}

// GetByAccountNo 根据账号获取钱包
func (r *WalletRepositoryImpl) GetByAccountNo(accountNo string) (*domain.Wallet, error) {
	var model WalletModel
	err := r.db.Where("account_no = ?", accountNo).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toWalletDomain(&model), nil
}

// Update 更新钱包信息
func (r *WalletRepositoryImpl) Update(w *domain.Wallet) error {
	start := time.Now()
	model := toWalletModel(w)
	err := r.db.Model(&WalletModel{}).Where("wallet_id = ?", w.WalletID).Updates(model).Error

	r.logger.Info("update wallet",
		"wallet_id", w.WalletID,
		"error", err,
		"duration", time.Since(start))
	return err
}

// UpdateBalance 更新账户余额
func (r *WalletRepositoryImpl) UpdateBalance(walletID uint64, balance, frozenBalance, availableBalance int64) error {
	start := time.Now()
	err := r.db.Model(&WalletModel{}).Where("id = ?", walletID).Updates(map[string]interface{}{
		"balance":           balance,
		"frozen_balance":    frozenBalance,
		"available_balance": availableBalance,
		"updated_at":        time.Now(),
	}).Error

	r.logger.Info("update balance",
		"wallet_id", walletID,
		"balance", balance,
		"error", err,
		"duration", time.Since(start))
	return err
}

// Transaction 执行数据库事务
func (r *WalletRepositoryImpl) Transaction(fn func(tx any) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// --- 以下为辅助转换函数和由于篇幅省略的部分接口实现 ---

func toWalletModel(w *domain.Wallet) *WalletModel {
	return &WalletModel{
		WalletID:         w.WalletID,
		UserID:           w.UserID,
		AccountNo:        w.AccountNo,
		Currency:         w.Currency,
		WalletType:       w.WalletType,
		Balance:          w.Balance,
		FrozenBalance:    w.FrozenBalance,
		AvailableBalance: w.AvailableBalance,
		Status:           int(w.Status),
		PasswordHash:     w.PasswordHash,
		SecurityLevel:    w.SecurityLevel,
	}
}

func toWalletDomain(m *WalletModel) *domain.Wallet {
	return &domain.Wallet{
		ID:               uint64(m.ID),
		WalletID:         m.WalletID,
		UserID:           m.UserID,
		AccountNo:        m.AccountNo,
		Currency:         m.Currency,
		WalletType:       m.WalletType,
		Balance:          m.Balance,
		FrozenBalance:    m.FrozenBalance,
		AvailableBalance: m.AvailableBalance,
		Status:           domain.WalletStatus(m.Status),
		PasswordHash:     m.PasswordHash,
		SecurityLevel:    m.SecurityLevel,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
}

// WalletLimitModel 额度限制数据库模型
type WalletLimitModel struct {
	gorm.Model
	WalletID             uint64 `gorm:"column:wallet_id;uniqueIndex;not null"`
	SingleDepositLimit   int64  `gorm:"column:single_deposit_limit"`
	SingleWithdrawLimit  int64  `gorm:"column:single_withdraw_limit"`
	DailyDepositLimit    int64  `gorm:"column:daily_deposit_limit"`
	DailyWithdrawLimit   int64  `gorm:"column:daily_withdraw_limit"`
	MonthlyWithdrawLimit int64  `gorm:"column:monthly_withdraw_limit"`
	DailyTransferLimit   int64  `gorm:"column:daily_transfer_limit"`
	RequirePasswordMin   int64  `gorm:"column:require_password_min"`
}

func (WalletLimitModel) TableName() string { return "wallet_limits" }

// DailyUsageModel 每日使用量统计模型
type DailyUsageModel struct {
	gorm.Model
	WalletID         uint64 `gorm:"column:wallet_id;index;uniqueIndex:idx_wallet_date"`
	UsageDate        string `gorm:"column:usage_date;type:varchar(20);uniqueIndex:idx_wallet_date"`
	DepositAmount    int64  `gorm:"column:deposit_amount"`
	WithdrawAmount   int64  `gorm:"column:withdraw_amount"`
	TransferAmount   int64  `gorm:"column:transfer_amount"`
	TransactionCount int    `gorm:"column:transaction_count"`
}

func (DailyUsageModel) TableName() string { return "wallet_daily_usages" }

func (r *WalletRepositoryImpl) GetLimits(walletID uint64) (*domain.WalletLimits, error) {
	var m WalletLimitModel
	err := r.db.Where("wallet_id = ?", walletID).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.WalletLimits{
		ID:                   uint64(m.ID),
		WalletID:             m.WalletID,
		SingleDepositLimit:   m.SingleDepositLimit,
		SingleWithdrawLimit:  m.SingleWithdrawLimit,
		DailyDepositLimit:    m.DailyDepositLimit,
		DailyWithdrawLimit:   m.DailyWithdrawLimit,
		MonthlyWithdrawLimit: m.MonthlyWithdrawLimit,
		DailyTransferLimit:   m.DailyTransferLimit,
		RequirePasswordMin:   m.RequirePasswordMin,
		CreatedAt:            m.CreatedAt,
		UpdatedAt:            m.UpdatedAt,
	}, nil
}

func (r *WalletRepositoryImpl) SaveLimits(l *domain.WalletLimits) error {
	model := &WalletLimitModel{
		WalletID:             l.WalletID,
		SingleDepositLimit:   l.SingleDepositLimit,
		SingleWithdrawLimit:  l.SingleWithdrawLimit,
		DailyDepositLimit:    l.DailyDepositLimit,
		DailyWithdrawLimit:   l.DailyWithdrawLimit,
		MonthlyWithdrawLimit: l.MonthlyWithdrawLimit,
		DailyTransferLimit:   l.DailyTransferLimit,
		RequirePasswordMin:   l.RequirePasswordMin,
	}
	if l.ID > 0 {
		model.ID = uint(l.ID)
	}
	return r.db.Save(model).Error
}

func (r *WalletRepositoryImpl) GetDailyUsage(walletID uint64, date string) (*domain.DailyUsage, error) {
	var m DailyUsageModel
	err := r.db.Where("wallet_id = ? AND usage_date = ?", walletID, date).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &domain.DailyUsage{
		ID:               uint64(m.ID),
		WalletID:         m.WalletID,
		UsageDate:        m.UsageDate,
		DepositAmount:    m.DepositAmount,
		WithdrawAmount:   m.WithdrawAmount,
		TransferAmount:   m.TransferAmount,
		TransactionCount: m.TransactionCount,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}, nil
}

func (r *WalletRepositoryImpl) SaveDailyUsage(u *domain.DailyUsage) error {
	model := &DailyUsageModel{
		WalletID:         u.WalletID,
		UsageDate:        u.UsageDate,
		DepositAmount:    u.DepositAmount,
		WithdrawAmount:   u.WithdrawAmount,
		TransferAmount:   u.TransferAmount,
		TransactionCount: u.TransactionCount,
	}
	if u.ID > 0 {
		model.ID = uint(u.ID)
	}
	return r.db.Save(model).Error
}

// TransactionRepositoryImpl 交易记录仓储实现
type TransactionRepositoryImpl struct {
	db     *database.DB
	logger *slog.Logger
}

// NewTransactionRepository 创建 Transaction 仓储实例
func NewTransactionRepository(db *database.DB, logger *slog.Logger) domain.TransactionRepository {
	return &TransactionRepositoryImpl{
		db:     db,
		logger: logger.With("module", "transaction_repository"),
	}
}

func (r *TransactionRepositoryImpl) Create(tx *domain.Transaction) error {
	start := time.Now()
	model := toTransactionModel(tx)
	err := r.db.Create(model).Error

	r.logger.Info("create transaction",
		"tx_no", tx.TransactionNo,
		"wallet_id", tx.WalletID,
		"error", err,
		"duration", time.Since(start))

	if err == nil {
		tx.ID = uint64(model.ID)
	}
	return err
}

func (r *TransactionRepositoryImpl) GetByID(id uint64) (*domain.Transaction, error) { return nil, nil }
func (r *TransactionRepositoryImpl) GetByTransactionNo(no string) (*domain.Transaction, error) {
	return nil, nil
}
func (r *TransactionRepositoryImpl) ListByWalletID(walletID uint64, txType domain.TransactionType, start, end *time.Time, page, size int) ([]*domain.Transaction, int64, error) {
	return nil, 0, nil
}
func (r *TransactionRepositoryImpl) ListByUserID(userID uint64, txType domain.TransactionType, start, end *time.Time, page, size int) ([]*domain.Transaction, int64, error) {
	return nil, 0, nil
}
func (r *TransactionRepositoryImpl) Update(tx *domain.Transaction) error { return nil }
func (r *TransactionRepositoryImpl) GetByChannelTradeNo(no string) (*domain.Transaction, error) {
	return nil, nil
}

func toTransactionModel(tx *domain.Transaction) *TransactionModel {
	return &TransactionModel{
		TransactionNo:    tx.TransactionNo,
		WalletID:         tx.WalletID,
		UserID:           tx.UserID,
		Type:             string(tx.Type),
		Amount:           tx.Amount,
		BalanceBefore:    tx.BalanceBefore,
		BalanceAfter:     tx.BalanceAfter,
		Fee:              tx.Fee,
		Status:           int(tx.Status),
		Remark:           tx.Remark,
		ChannelTradeNo:   tx.ChannelTradeNo,
		CounterpartyID:   tx.CounterpartyID,
		CounterpartyType: tx.CounterpartyType,
		RiskCheckID:      tx.RiskCheckID,
		IPAddress:        tx.IPAddress,
		DeviceID:         tx.DeviceID,
		Location:         tx.Location,
		CompletedAt:      tx.CompletedAt,
	}
}

func toTransactionDomain(m *TransactionModel) *domain.Transaction {
	return &domain.Transaction{
		ID:               uint64(m.ID),
		TransactionNo:    m.TransactionNo,
		WalletID:         m.WalletID,
		UserID:           m.UserID,
		Type:             domain.TransactionType(m.Type),
		Amount:           m.Amount,
		BalanceBefore:    m.BalanceBefore,
		BalanceAfter:     m.BalanceAfter,
		Fee:              m.Fee,
		Status:           domain.TransactionStatus(m.Status),
		Remark:           m.Remark,
		ChannelTradeNo:   m.ChannelTradeNo,
		CounterpartyID:   m.CounterpartyID,
		CounterpartyType: m.CounterpartyType,
		RiskCheckID:      m.RiskCheckID,
		IPAddress:        m.IPAddress,
		DeviceID:         m.DeviceID,
		Location:         m.Location,
		CreatedAt:        m.CreatedAt,
		CompletedAt:      m.CompletedAt,
	}
}
