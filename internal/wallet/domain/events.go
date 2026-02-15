// Package domain 钱包领域事件定义
// 生成摘要：
// 1) 定义钱包服务所有领域事件，支持 Event Sourcing 与 CQRS 投影
// 2) 每个事件携带完整的业务上下文（用户ID、钱包ID、金额、交易流水号等）
// 3) 所有事件实现 DomainEvent 接口，可被 EventBus 发布与消费
package domain

import "time"

// DomainEvent 领域事件接口
type DomainEvent interface {
	// EventName 返回事件类型名称，用于路由和序列化。
	EventName() string
	// EventKey 返回事件分区键，用于 Kafka 分区保序。
	EventKey() string
	// OccurredAt 返回事件发生时间。
	OccurredAt() time.Time
}

// --- 钱包生命周期事件 ---

// WalletCreatedEvent 钱包创建事件
type WalletCreatedEvent struct {
	WalletID   uint64    `json:"wallet_id"`
	UserID     uint64    `json:"user_id"`
	AccountNo  string    `json:"account_no"`
	Currency   string    `json:"currency"`
	WalletType string    `json:"wallet_type"`
	Timestamp  time.Time `json:"timestamp"`
}

func (e *WalletCreatedEvent) EventName() string      { return "wallet.created" }
func (e *WalletCreatedEvent) EventKey() string        { return e.AccountNo }
func (e *WalletCreatedEvent) OccurredAt() time.Time   { return e.Timestamp }

// WalletFrozenEvent 钱包冻结事件
type WalletFrozenEvent struct {
	WalletID  uint64    `json:"wallet_id"`
	UserID    uint64    `json:"user_id"`
	Reason    string    `json:"reason"`
	Operator  string    `json:"operator"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *WalletFrozenEvent) EventName() string      { return "wallet.frozen" }
func (e *WalletFrozenEvent) EventKey() string        { return "" }
func (e *WalletFrozenEvent) OccurredAt() time.Time   { return e.Timestamp }

// WalletUnfrozenEvent 钱包解冻事件
type WalletUnfrozenEvent struct {
	WalletID  uint64    `json:"wallet_id"`
	UserID    uint64    `json:"user_id"`
	Operator  string    `json:"operator"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *WalletUnfrozenEvent) EventName() string      { return "wallet.unfrozen" }
func (e *WalletUnfrozenEvent) EventKey() string        { return "" }
func (e *WalletUnfrozenEvent) OccurredAt() time.Time   { return e.Timestamp }

// WalletDisabledEvent 钱包禁用事件
type WalletDisabledEvent struct {
	WalletID  uint64    `json:"wallet_id"`
	UserID    uint64    `json:"user_id"`
	Reason    string    `json:"reason"`
	Operator  string    `json:"operator"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *WalletDisabledEvent) EventName() string      { return "wallet.disabled" }
func (e *WalletDisabledEvent) EventKey() string        { return "" }
func (e *WalletDisabledEvent) OccurredAt() time.Time   { return e.Timestamp }

// --- 资金变动事件 ---

// DepositedEvent 充值成功事件
type DepositedEvent struct {
	TransactionNo string    `json:"transaction_no"`
	WalletID      uint64    `json:"wallet_id"`
	UserID        uint64    `json:"user_id"`
	Amount        int64     `json:"amount"`
	BalanceBefore int64     `json:"balance_before"`
	BalanceAfter  int64     `json:"balance_after"`
	Channel       string    `json:"channel"`
	Remark        string    `json:"remark"`
	Timestamp     time.Time `json:"timestamp"`
}

func (e *DepositedEvent) EventName() string      { return "wallet.deposited" }
func (e *DepositedEvent) EventKey() string        { return e.TransactionNo }
func (e *DepositedEvent) OccurredAt() time.Time   { return e.Timestamp }

// WithdrawnEvent 提现成功事件
type WithdrawnEvent struct {
	TransactionNo string    `json:"transaction_no"`
	WalletID      uint64    `json:"wallet_id"`
	UserID        uint64    `json:"user_id"`
	Amount        int64     `json:"amount"`
	Fee           int64     `json:"fee"`
	BalanceBefore int64     `json:"balance_before"`
	BalanceAfter  int64     `json:"balance_after"`
	Remark        string    `json:"remark"`
	Timestamp     time.Time `json:"timestamp"`
}

func (e *WithdrawnEvent) EventName() string      { return "wallet.withdrawn" }
func (e *WithdrawnEvent) EventKey() string        { return e.TransactionNo }
func (e *WithdrawnEvent) OccurredAt() time.Time   { return e.Timestamp }

// TransferredEvent 转账成功事件
type TransferredEvent struct {
	TransactionNo string    `json:"transaction_no"`
	FromWalletID  uint64    `json:"from_wallet_id"`
	FromUserID    uint64    `json:"from_user_id"`
	ToWalletID    uint64    `json:"to_wallet_id"`
	ToUserID      uint64    `json:"to_user_id"`
	Amount        int64     `json:"amount"`
	Remark        string    `json:"remark"`
	Timestamp     time.Time `json:"timestamp"`
}

func (e *TransferredEvent) EventName() string      { return "wallet.transferred" }
func (e *TransferredEvent) EventKey() string        { return e.TransactionNo }
func (e *TransferredEvent) OccurredAt() time.Time   { return e.Timestamp }

// PaymentDeductedEvent 支付扣款事件
type PaymentDeductedEvent struct {
	TransactionNo string    `json:"transaction_no"`
	WalletID      uint64    `json:"wallet_id"`
	UserID        uint64    `json:"user_id"`
	OrderNo       string    `json:"order_no"`
	Amount        int64     `json:"amount"`
	BalanceBefore int64     `json:"balance_before"`
	BalanceAfter  int64     `json:"balance_after"`
	Timestamp     time.Time `json:"timestamp"`
}

func (e *PaymentDeductedEvent) EventName() string      { return "wallet.payment_deducted" }
func (e *PaymentDeductedEvent) EventKey() string        { return e.OrderNo }
func (e *PaymentDeductedEvent) OccurredAt() time.Time   { return e.Timestamp }

// RefundCreditedEvent 退款到账事件
type RefundCreditedEvent struct {
	TransactionNo         string    `json:"transaction_no"`
	WalletID              uint64    `json:"wallet_id"`
	UserID                uint64    `json:"user_id"`
	OriginalTransactionNo string    `json:"original_transaction_no"`
	Amount                int64     `json:"amount"`
	BalanceBefore         int64     `json:"balance_before"`
	BalanceAfter          int64     `json:"balance_after"`
	Timestamp             time.Time `json:"timestamp"`
}

func (e *RefundCreditedEvent) EventName() string      { return "wallet.refund_credited" }
func (e *RefundCreditedEvent) EventKey() string        { return e.TransactionNo }
func (e *RefundCreditedEvent) OccurredAt() time.Time   { return e.Timestamp }

// --- 冻结/解冻资金事件 ---

// BalanceFrozenEvent 余额冻结事件
type BalanceFrozenEvent struct {
	WalletID    uint64    `json:"wallet_id"`
	UserID      uint64    `json:"user_id"`
	Amount      int64     `json:"amount"`
	Reason      string    `json:"reason"`
	ReferenceNo string    `json:"reference_no"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e *BalanceFrozenEvent) EventName() string      { return "wallet.balance_frozen" }
func (e *BalanceFrozenEvent) EventKey() string        { return e.ReferenceNo }
func (e *BalanceFrozenEvent) OccurredAt() time.Time   { return e.Timestamp }

// BalanceUnfrozenEvent 余额解冻事件
type BalanceUnfrozenEvent struct {
	WalletID    uint64    `json:"wallet_id"`
	UserID      uint64    `json:"user_id"`
	Amount      int64     `json:"amount"`
	Reason      string    `json:"reason"`
	ReferenceNo string    `json:"reference_no"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e *BalanceUnfrozenEvent) EventName() string      { return "wallet.balance_unfrozen" }
func (e *BalanceUnfrozenEvent) EventKey() string        { return e.ReferenceNo }
func (e *BalanceUnfrozenEvent) OccurredAt() time.Time   { return e.Timestamp }

// --- 安全事件 ---

// PasswordSetEvent 支付密码设置事件
type PasswordSetEvent struct {
	WalletID  uint64    `json:"wallet_id"`
	UserID    uint64    `json:"user_id"`
	Timestamp time.Time `json:"timestamp"`
}

func (e *PasswordSetEvent) EventName() string      { return "wallet.password_set" }
func (e *PasswordSetEvent) EventKey() string        { return "" }
func (e *PasswordSetEvent) OccurredAt() time.Time   { return e.Timestamp }

// PasswordVerifyFailedEvent 支付密码验证失败事件（用于风控）
type PasswordVerifyFailedEvent struct {
	WalletID    uint64    `json:"wallet_id"`
	UserID      uint64    `json:"user_id"`
	FailedCount int       `json:"failed_count"`
	IPAddress   string    `json:"ip_address"`
	DeviceID    string    `json:"device_id"`
	Timestamp   time.Time `json:"timestamp"`
}

func (e *PasswordVerifyFailedEvent) EventName() string      { return "wallet.password_verify_failed" }
func (e *PasswordVerifyFailedEvent) EventKey() string        { return "" }
func (e *PasswordVerifyFailedEvent) OccurredAt() time.Time   { return e.Timestamp }

// --- 限额事件 ---

// DailyLimitExceededEvent 日限额超限事件
type DailyLimitExceededEvent struct {
	WalletID      uint64    `json:"wallet_id"`
	UserID        uint64    `json:"user_id"`
	LimitType     string    `json:"limit_type"`
	CurrentUsage  int64     `json:"current_usage"`
	LimitAmount   int64     `json:"limit_amount"`
	AttemptAmount int64     `json:"attempt_amount"`
	Timestamp     time.Time `json:"timestamp"`
}

func (e *DailyLimitExceededEvent) EventName() string      { return "wallet.daily_limit_exceeded" }
func (e *DailyLimitExceededEvent) EventKey() string        { return "" }
func (e *DailyLimitExceededEvent) OccurredAt() time.Time   { return e.Timestamp }
