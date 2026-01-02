package domain

import (
	"errors"
	"fmt"

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
