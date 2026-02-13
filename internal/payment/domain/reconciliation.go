package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// ReconcileStatus 对账结果状态
type ReconcileStatus string

const (
	ReconcileMatch          ReconcileStatus = "MATCH"
	ReconcileMismatchAmount ReconcileStatus = "MISMATCH_AMOUNT"
	ReconcileMismatchStatus ReconcileStatus = "MISMATCH_STATUS"
	ReconcileMissingSystem  ReconcileStatus = "MISSING_SYSTEM"  // 长款 (渠道有，系统无)
	ReconcileMissingChannel ReconcileStatus = "MISSING_CHANNEL" // 短款 (系统有，渠道无)
)

// SystemTransaction 系统侧交易记录 (待对账数据)
type SystemTransaction struct {
	PaymentID     string
	OrderNo       string
	Amount        decimal.Decimal
	Status        PaymentStatus
	TransactionID string // 渠道流水号
}

// ChannelTransaction 渠道侧交易记录 (对账单数据)
type ChannelTransaction struct {
	TransactionID string
	PaymentNo     string // 商户订单号
	Amount        decimal.Decimal
	Status        string // 渠道原生状态
	Fee           decimal.Decimal
}

// ReconcileResult 单笔对账结果
type ReconcileResult struct {
	PaymentID     string
	TransactionID string
	Status        ReconcileStatus
	SystemAmount  decimal.Decimal
	ChannelAmount decimal.Decimal
	AmountDiff    decimal.Decimal
	Remark        string
}

// ReconciliationEngine 对账核心引擎
type ReconciliationEngine struct {
	// 可配置项：允许的金额误差 (例如 0.01 元)
	AmountTolerance decimal.Decimal
}

func NewReconciliationEngine() *ReconciliationEngine {
	return &ReconciliationEngine{
		AmountTolerance: decimal.NewFromFloat(0.01),
	}
}

// ReconcileBatch 执行批次对账
// systemTxns: 系统流水，Key 为 TransactionID (或 PaymentNo，需统一)
// channelTxns: 渠道流水，Key 为 TransactionID (或 PaymentNo)
// 假设以 TransactionID 为主要匹配键，PaymentNo 为辅
func (e *ReconciliationEngine) ReconcileBatch(systemTxns map[string]*SystemTransaction, channelTxns map[string]*ChannelTransaction) []*ReconcileResult {
	results := make([]*ReconcileResult, 0, len(systemTxns))
	processedChannelTxns := make(map[string]bool)

	// 1. 遍历系统流水，去匹配渠道流水
	for sysKey, sysTx := range systemTxns {
		// 尝试匹配
		chanTx, exists := channelTxns[sysKey]
		if !exists {
			// 如果 TransactionID 没匹配上，尝试用 PaymentNo 匹配 (假设 sysKey 就是 TransactionID)
			// 这里简化逻辑，假设 key 统一。

			// 短款：系统有，渠道无
			results = append(results, &ReconcileResult{
				PaymentID:     sysTx.PaymentID,
				TransactionID: sysTx.TransactionID,
				Status:        ReconcileMissingChannel,
				SystemAmount:  sysTx.Amount,
				Remark:        "Transaction found in system but missing in channel statement",
			})
			continue
		}

		processedChannelTxns[sysKey] = true

		// 匹配上了，开始核对细节
		res := e.compareTransaction(sysTx, chanTx)
		results = append(results, res)
	}

	// 2. 遍历渠道流水，找出长款 (系统没有的)
	for chanKey, chanTx := range channelTxns {
		if processedChannelTxns[chanKey] {
			continue
		}

		// 长款：渠道有，系统无
		results = append(results, &ReconcileResult{
			TransactionID: chanTx.TransactionID,
			Status:        ReconcileMissingSystem,
			ChannelAmount: chanTx.Amount,
			Remark:        fmt.Sprintf("Transaction %s found in channel but missing in system", chanKey),
		})
	}

	return results
}

// compareTransaction 核对单笔明细
func (e *ReconciliationEngine) compareTransaction(sys *SystemTransaction, ch *ChannelTransaction) *ReconcileResult {
	res := &ReconcileResult{
		PaymentID:     sys.PaymentID,
		TransactionID: sys.TransactionID,
		SystemAmount:  sys.Amount,
		ChannelAmount: ch.Amount,
	}

	// 1. 校验金额
	diff := sys.Amount.Sub(ch.Amount).Abs()
	res.AmountDiff = diff

	if diff.GreaterThan(e.AmountTolerance) {
		res.Status = ReconcileMismatchAmount
		res.Remark = fmt.Sprintf("Amount mismatch: System=%s, Channel=%s", sys.Amount, ch.Amount)
		return res
	}

	// 2. 校验状态
	// 这里需要一个状态映射逻辑，将 Channel 状态映射为 System 状态
	// 简化假设：只要渠道是 Success，系统也应该是 Success
	if sys.Status != PaymentSuccess {
		// 这是一个复杂点：如果渠道成功扣款，但系统状态不是 Success (可能是 Pending 或 Failed)
		// 这属于 "掉单" (Status Mismatch)
		res.Status = ReconcileMismatchStatus
		res.Remark = fmt.Sprintf("Status mismatch: System=%s, Channel=%s (Success)", sys.Status, ch.Status)
		return res
	}

	res.Status = ReconcileMatch
	return res
}

// AutoResolve 尝试自动解决差异 (Strategy Pattern)
// 针对简单的状态不一致，可以自动修正
func (e *ReconciliationEngine) AutoResolve(res *ReconcileResult, repo PaymentRepository) error {
	if res.Status == ReconcileMismatchStatus {
		// 场景：渠道成功，系统失败/处理中 -> 自动修补为成功
		// 这是一个高风险操作，通常需要记录详细日志
		// repo.UpdateStatus(...)
		return nil
	}
	return errors.New("manual intervention required")
}

// ========== 以下是新增的对账批次管理功能 ==========

var (
	ErrReconciliationNotFound    = errors.New("reconciliation record not found")
	ErrReconciliationInProgress  = errors.New("reconciliation already in progress")
	ErrReconciliationFailed      = errors.New("reconciliation failed")
	ErrInvalidReconciliationData = errors.New("invalid reconciliation data")
	ErrDiscrepancyNotFound       = errors.New("discrepancy not found")
)

type ReconciliationBatchStatus string

const (
	ReconciliationBatchStatusPending    ReconciliationBatchStatus = "PENDING"
	ReconciliationBatchStatusProcessing ReconciliationBatchStatus = "PROCESSING"
	ReconciliationBatchStatusCompleted  ReconciliationBatchStatus = "COMPLETED"
	ReconciliationBatchStatusFailed     ReconciliationBatchStatus = "FAILED"
	ReconciliationBatchStatusPartial    ReconciliationBatchStatus = "PARTIAL"
)

type DiscrepancyType string

const (
	DiscrepancyTypeAmountMismatch   DiscrepancyType = "AMOUNT_MISMATCH"
	DiscrepancyTypeMissingInSystem  DiscrepancyType = "MISSING_IN_SYSTEM"
	DiscrepancyTypeMissingInGateway DiscrepancyType = "MISSING_IN_GATEWAY"
	DiscrepancyTypeStatusMismatch   DiscrepancyType = "STATUS_MISMATCH"
	DiscrepancyTypeDuplicate        DiscrepancyType = "DUPLICATE"
	DiscrepancyTypeTimeMismatch     DiscrepancyType = "TIME_MISMATCH"
)

type DiscrepancyStatus string

const (
	DiscrepancyStatusOpen          DiscrepancyStatus = "OPEN"
	DiscrepancyStatusInvestigating DiscrepancyStatus = "INVESTIGATING"
	DiscrepancyStatusResolved      DiscrepancyStatus = "RESOLVED"
	DiscrepancyStatusIgnored       DiscrepancyStatus = "IGNORED"
)

// ReconciliationBatch 对账批次
type ReconciliationBatch struct {
	ID                uint                      `json:"id"`
	CreatedAt         time.Time                 `json:"created_at"`
	UpdatedAt         time.Time                 `json:"updated_at"`
	BatchNo           string                    `json:"batch_no"`
	GatewayType       GatewayType               `json:"gateway_type"`
	ReconcileDate     time.Time                 `json:"reconcile_date"`
	Status            ReconciliationBatchStatus `json:"status"`
	TotalCount        int                       `json:"total_count"`
	MatchedCount      int                       `json:"matched_count"`
	DiscrepancyCount  int                       `json:"discrepancy_count"`
	TotalAmount       int64                     `json:"total_amount"`
	MatchedAmount     int64                     `json:"matched_amount"`
	DiscrepancyAmount int64                     `json:"discrepancy_amount"`
	StartedAt         *time.Time                `json:"started_at"`
	CompletedAt       *time.Time                `json:"completed_at"`
	ErrorMessage      string                    `json:"error_message"`
	FilePath          string                    `json:"file_path"`
	Checksum          string                    `json:"checksum"`
}

// ReconciliationRecord 对账明细记录
type ReconciliationRecord struct {
	ID                   uint              `json:"id"`
	CreatedAt            time.Time         `json:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"`
	BatchID              uint              `json:"batch_id"`
	PaymentNo            string            `json:"payment_no"`
	GatewayTransactionID string            `json:"gateway_transaction_id"`
	OrderNo              string            `json:"order_no"`
	SystemAmount         int64             `json:"system_amount"`
	GatewayAmount        int64             `json:"gateway_amount"`
	AmountDiff           int64             `json:"amount_diff"`
	SystemStatus         PaymentStatus     `json:"system_status"`
	GatewayStatus        string            `json:"gateway_status"`
	StatusMatch          bool              `json:"status_match"`
	SystemPaidAt         *time.Time        `json:"system_paid_at"`
	GatewayPaidAt        *time.Time        `json:"gateway_paid_at"`
	IsMatched            bool              `json:"is_matched"`
	DiscrepancyType      DiscrepancyType   `json:"discrepancy_type"`
	DiscrepancyStatus    DiscrepancyStatus `json:"discrepancy_status"`
	Resolution           string            `json:"resolution"`
	ResolvedBy           string            `json:"resolved_by"`
	ResolvedAt           *time.Time        `json:"resolved_at"`
	Remarks              string            `json:"remarks"`
}

// GatewayBillItem 网关账单条目
type GatewayBillItem struct {
	TransactionID string    `json:"transaction_id"`
	OutTradeNo    string    `json:"out_trade_no"`
	Amount        int64     `json:"amount"`
	Status        string    `json:"status"`
	PaidAt        time.Time `json:"paid_at"`
	Fee           int64     `json:"fee"`
	RefundAmount  int64     `json:"refund_amount"`
	Currency      string    `json:"currency"`
	BuyerAccount  string    `json:"buyer_account"`
	RawData       string    `json:"raw_data"`
}

// ReconciliationRule 对账规则
type ReconciliationRule struct {
	ID                   uint        `json:"id"`
	CreatedAt            time.Time   `json:"created_at"`
	UpdatedAt            time.Time   `json:"updated_at"`
	Name                 string      `json:"name"`
	GatewayType          GatewayType `json:"gateway_type"`
	Enabled              bool        `json:"enabled"`
	AutoReconcile        bool        `json:"auto_reconcile"`
	ReconcileTime        string      `json:"reconcile_time"`
	ToleranceAmount      int64       `json:"tolerance_amount"`
	TolerancePercent     float64     `json:"tolerance_percent"`
	AutoResolveThreshold int64       `json:"auto_resolve_threshold"`
	NotifyOnDiscrepancy  bool        `json:"notify_on_discrepancy"`
	NotifyEmails         []string    `json:"notify_emails"`
	MaxRetries           int         `json:"max_retries"`
	RetryInterval        int         `json:"retry_interval"`
}

// ReconciliationSummary 对账汇总
type ReconciliationSummary struct {
	BatchID             uint        `json:"batch_id"`
	ReconcileDate       time.Time   `json:"reconcile_date"`
	GatewayType         GatewayType `json:"gateway_type"`
	TotalTransactions   int         `json:"total_transactions"`
	MatchedTransactions int         `json:"matched_transactions"`
	DiscrepancyCount    int         `json:"discrepancy_count"`
	TotalSystemAmount   int64       `json:"total_system_amount"`
	TotalGatewayAmount  int64       `json:"total_gateway_amount"`
	AmountDifference    int64       `json:"amount_difference"`
	MissingInSystem     int         `json:"missing_in_system"`
	MissingInGateway    int         `json:"missing_in_gateway"`
	StatusMismatches    int         `json:"status_mismatches"`
	AmountMismatches    int         `json:"amount_mismatches"`
	ResolvedCount       int         `json:"resolved_count"`
	PendingCount        int         `json:"pending_count"`
	SuccessRate         float64     `json:"success_rate"`
}

func NewReconciliationBatch(batchNo string, gatewayType GatewayType, reconcileDate time.Time) *ReconciliationBatch {
	return &ReconciliationBatch{
		BatchNo:       batchNo,
		GatewayType:   gatewayType,
		ReconcileDate: reconcileDate,
		Status:        ReconciliationBatchStatusPending,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func (b *ReconciliationBatch) Start() error {
	if b.Status != ReconciliationBatchStatusPending {
		return ErrReconciliationInProgress
	}
	b.Status = ReconciliationBatchStatusProcessing
	now := time.Now()
	b.StartedAt = &now
	b.UpdatedAt = now
	return nil
}

func (b *ReconciliationBatch) Complete() {
	b.Status = ReconciliationBatchStatusCompleted
	now := time.Now()
	b.CompletedAt = &now
	b.UpdatedAt = now
}

func (b *ReconciliationBatch) Fail(errMsg string) {
	b.Status = ReconciliationBatchStatusFailed
	b.ErrorMessage = errMsg
	now := time.Now()
	b.CompletedAt = &now
	b.UpdatedAt = now
}

func (b *ReconciliationBatch) UpdateStats(matchedCount, discrepancyCount int, matchedAmount, discrepancyAmount int64) {
	b.MatchedCount = matchedCount
	b.DiscrepancyCount = discrepancyCount
	b.MatchedAmount = matchedAmount
	b.DiscrepancyAmount = discrepancyAmount
	b.UpdatedAt = time.Now()
}

func (b *ReconciliationBatch) IsCompleted() bool {
	return b.Status == ReconciliationBatchStatusCompleted || b.Status == ReconciliationBatchStatusFailed
}

func NewReconciliationRecord(batchID uint, paymentNo, gatewayTxID, orderNo string) *ReconciliationRecord {
	return &ReconciliationRecord{
		BatchID:              batchID,
		PaymentNo:            paymentNo,
		GatewayTransactionID: gatewayTxID,
		OrderNo:              orderNo,
		DiscrepancyStatus:    DiscrepancyStatusOpen,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
}

func (r *ReconciliationRecord) SetSystemData(amount int64, status PaymentStatus, paidAt *time.Time) {
	r.SystemAmount = amount
	r.SystemStatus = status
	r.SystemPaidAt = paidAt
	r.UpdatedAt = time.Now()
}

func (r *ReconciliationRecord) SetGatewayData(amount int64, status string, paidAt *time.Time) {
	r.GatewayAmount = amount
	r.GatewayStatus = status
	r.GatewayPaidAt = paidAt
	r.UpdatedAt = time.Now()
}

func (r *ReconciliationRecord) CheckMatch(toleranceAmount int64, tolerancePercent float64) {
	r.AmountDiff = r.SystemAmount - r.GatewayAmount
	r.StatusMatch = r.matchStatus()

	amountTolerance := toleranceAmount
	if tolerancePercent > 0 {
		percentTolerance := int64(float64(r.SystemAmount) * tolerancePercent / 100)
		if percentTolerance > amountTolerance {
			amountTolerance = percentTolerance
		}
	}

	amountMatch := r.AmountDiff <= amountTolerance && r.AmountDiff >= -amountTolerance

	r.IsMatched = r.StatusMatch && amountMatch

	if !r.IsMatched {
		r.determineDiscrepancyType()
	}

	r.UpdatedAt = time.Now()
}

func (r *ReconciliationRecord) matchStatus() bool {
	if r.SystemStatus == PaymentSuccess && r.GatewayStatus == "SUCCESS" {
		return true
	}
	if r.SystemStatus == PaymentStatus_REFUNDED && r.GatewayStatus == "REFUND" {
		return true
	}
	if r.SystemStatus == PaymentStatus_CLOSED && r.GatewayStatus == "CLOSED" {
		return true
	}
	return false
}

func (r *ReconciliationRecord) determineDiscrepancyType() {
	if r.SystemAmount == 0 && r.GatewayAmount > 0 {
		r.DiscrepancyType = DiscrepancyTypeMissingInSystem
	} else if r.SystemAmount > 0 && r.GatewayAmount == 0 {
		r.DiscrepancyType = DiscrepancyTypeMissingInGateway
	} else if !r.StatusMatch {
		r.DiscrepancyType = DiscrepancyTypeStatusMismatch
	} else if r.AmountDiff != 0 {
		r.DiscrepancyType = DiscrepancyTypeAmountMismatch
	}
}

func (r *ReconciliationRecord) Resolve(resolvedBy, resolution string) {
	r.DiscrepancyStatus = DiscrepancyStatusResolved
	r.ResolvedBy = resolvedBy
	r.Resolution = resolution
	now := time.Now()
	r.ResolvedAt = &now
	r.UpdatedAt = now
}

func (r *ReconciliationRecord) Ignore(reason string) {
	r.DiscrepancyStatus = DiscrepancyStatusIgnored
	r.Resolution = "Ignored: " + reason
	now := time.Now()
	r.ResolvedAt = &now
	r.UpdatedAt = now
}

func (r *ReconciliationRecord) StartInvestigation() {
	r.DiscrepancyStatus = DiscrepancyStatusInvestigating
	r.UpdatedAt = time.Now()
}

func (r *ReconciliationRecord) IsResolved() bool {
	return r.DiscrepancyStatus == DiscrepancyStatusResolved || r.DiscrepancyStatus == DiscrepancyStatusIgnored
}

func NewReconciliationRule(name string, gatewayType GatewayType) *ReconciliationRule {
	return &ReconciliationRule{
		Name:                 name,
		GatewayType:          gatewayType,
		Enabled:              true,
		AutoReconcile:        true,
		ReconcileTime:        "02:00",
		ToleranceAmount:      1,
		TolerancePercent:     0.01,
		AutoResolveThreshold: 100,
		NotifyOnDiscrepancy:  true,
		NotifyEmails:         make([]string, 0),
		MaxRetries:           3,
		RetryInterval:        300,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
}

func (r *ReconciliationRule) ShouldAutoResolve(amountDiff int64) bool {
	absDiff := amountDiff
	if absDiff < 0 {
		absDiff = -absDiff
	}
	return absDiff <= r.AutoResolveThreshold
}

type ReconciliationBatchRepository interface {
	SaveBatch(ctx interface{}, batch *ReconciliationBatch) error
	UpdateBatch(ctx interface{}, batch *ReconciliationBatch) error
	FindBatchByID(ctx interface{}, id uint) (*ReconciliationBatch, error)
	FindBatchByNo(ctx interface{}, batchNo string) (*ReconciliationBatch, error)
	FindBatchesByDate(ctx interface{}, date time.Time) ([]*ReconciliationBatch, error)
	FindPendingBatches(ctx interface{}) ([]*ReconciliationBatch, error)

	SaveRecord(ctx interface{}, record *ReconciliationRecord) error
	UpdateRecord(ctx interface{}, record *ReconciliationRecord) error
	FindRecordByID(ctx interface{}, id uint) (*ReconciliationRecord, error)
	FindRecordsByBatchID(ctx interface{}, batchID uint) ([]*ReconciliationRecord, error)
	FindDiscrepancies(ctx interface{}, batchID uint) ([]*ReconciliationRecord, error)
	FindOpenDiscrepancies(ctx interface{}, limit int) ([]*ReconciliationRecord, error)

	SaveRule(ctx interface{}, rule *ReconciliationRule) error
	FindRuleByID(ctx interface{}, id uint) (*ReconciliationRule, error)
	FindRuleByGateway(ctx interface{}, gatewayType GatewayType) (*ReconciliationRule, error)
	FindEnabledRules(ctx interface{}) ([]*ReconciliationRule, error)
}

type ReconciliationBatchService interface {
	CreateBatch(gatewayType GatewayType, date time.Time) (*ReconciliationBatch, error)
	DownloadBill(ctx interface{}, batch *ReconciliationBatch) ([]*GatewayBillItem, error)
	Reconcile(ctx interface{}, batch *ReconciliationBatch) error
	ResolveDiscrepancy(ctx interface{}, recordID uint, resolvedBy, resolution string) error
	GetSummary(ctx interface{}, batchID uint) (*ReconciliationSummary, error)
	AutoReconcile(ctx interface{}) error
}
