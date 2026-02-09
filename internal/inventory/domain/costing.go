// 变更说明：新增库存成本核算功能，支持FIFO、加权平均法、移动加权平均法。
// 假设：默认使用FIFO成本核算法。
package domain

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// --- 成本核算方法 ---

// CostMethod 成本核算方法
type CostMethod int

const (
	CostMethodFIFO            CostMethod = 1 // 先进先出法
	CostMethodWeightedAverage CostMethod = 2 // 加权平均法
	CostMethodMovingAverage   CostMethod = 3 // 移动加权平均法
	CostMethodSpecificID      CostMethod = 4 // 个别计价法
)

// --- 成本核算上下文 ---

// CostContext 成本核算上下文
type CostContext struct {
	SkuID       uint64     `json:"sku_id"`
	WarehouseID uint64     `json:"warehouse_id"`
	Method      CostMethod `json:"method"`
	Batches     []*Batch   `json:"batches"` // 当前批次列表
}

// --- 成本核算器接口 ---

// CostCalculator 成本核算器接口
type CostCalculator interface {
	// CalculateOutboundCost 计算出库成本
	CalculateOutboundCost(ctx context.Context, quantity int32) (*OutboundCostResult, error)
	// CalculateInboundCost 计算入库成本（更新加权平均）
	CalculateInboundCost(ctx context.Context, quantity int32, unitCost int64) (*InboundCostResult, error)
	// GetCurrentUnitCost 获取当前单位成本
	GetCurrentUnitCost(ctx context.Context) int64
	// GetInventoryValue 获取当前库存价值
	GetInventoryValue(ctx context.Context) int64
}

// OutboundCostResult 出库成本结果
type OutboundCostResult struct {
	TotalCost    int64              `json:"total_cost"`    // 总成本
	UnitCost     int64              `json:"unit_cost"`     // 单位成本
	BatchDetails []*BatchCostDetail `json:"batch_details"` // 批次明细（FIFO用）
}

// BatchCostDetail 批次成本明细
type BatchCostDetail struct {
	BatchID   uint64 `json:"batch_id"`
	BatchNo   string `json:"batch_no"`
	Quantity  int32  `json:"quantity"`
	UnitCost  int64  `json:"unit_cost"`
	TotalCost int64  `json:"total_cost"`
}

// InboundCostResult 入库成本结果
type InboundCostResult struct {
	NewUnitCost      int64 `json:"new_unit_cost"`      // 新的单位成本
	PreviousUnitCost int64 `json:"previous_unit_cost"` // 之前的单位成本
	TotalValue       int64 `json:"total_value"`        // 更新后总价值
	TotalQuantity    int32 `json:"total_quantity"`     // 更新后总数量
}

// --- FIFO成本核算器 ---

// FIFOCostCalculator FIFO成本核算器
type FIFOCostCalculator struct {
	Context *CostContext
}

// NewFIFOCostCalculator 创建FIFO成本核算器
func NewFIFOCostCalculator(ctx *CostContext) *FIFOCostCalculator {
	return &FIFOCostCalculator{Context: ctx}
}

// CalculateOutboundCost 计算FIFO出库成本
func (c *FIFOCostCalculator) CalculateOutboundCost(ctx context.Context, quantity int32) (*OutboundCostResult, error) {
	// 按生产日期排序（FIFO）
	batches := make([]*Batch, len(c.Context.Batches))
	copy(batches, c.Context.Batches)
	sort.Slice(batches, func(i, j int) bool {
		return batches[i].ProductionDate.Before(batches[j].ProductionDate)
	})

	var totalCost int64
	var details []*BatchCostDetail
	remaining := quantity

	for _, batch := range batches {
		if remaining <= 0 {
			break
		}
		if batch.AvailableQuantity() <= 0 {
			continue
		}

		allocQty := batch.AvailableQuantity()
		if allocQty > remaining {
			allocQty = remaining
		}

		batchCost := int64(allocQty) * batch.UnitCost
		totalCost += batchCost
		remaining -= allocQty

		details = append(details, &BatchCostDetail{
			BatchID:   batch.ID,
			BatchNo:   batch.BatchNo,
			Quantity:  allocQty,
			UnitCost:  batch.UnitCost,
			TotalCost: batchCost,
		})
	}

	if remaining > 0 {
		return nil, fmt.Errorf("insufficient stock for FIFO: required=%d, available=%d", quantity, quantity-remaining)
	}

	return &OutboundCostResult{
		TotalCost:    totalCost,
		UnitCost:     totalCost / int64(quantity),
		BatchDetails: details,
	}, nil
}

// CalculateInboundCost 计算FIFO入库成本（FIFO不需要更新加权成本）
func (c *FIFOCostCalculator) CalculateInboundCost(ctx context.Context, quantity int32, unitCost int64) (*InboundCostResult, error) {
	// FIFO法入库直接以批次成本记录
	totalQty := int32(0)
	totalValue := int64(0)
	for _, batch := range c.Context.Batches {
		totalQty += batch.CurrentQuantity
		totalValue += int64(batch.CurrentQuantity) * batch.UnitCost
	}

	// 加上新入库
	totalQty += quantity
	totalValue += int64(quantity) * unitCost

	return &InboundCostResult{
		NewUnitCost:      unitCost, // FIFO使用批次成本
		PreviousUnitCost: 0,
		TotalValue:       totalValue,
		TotalQuantity:    totalQty,
	}, nil
}

// GetCurrentUnitCost 获取FIFO当前单位成本（最早批次）
func (c *FIFOCostCalculator) GetCurrentUnitCost(ctx context.Context) int64 {
	batches := make([]*Batch, len(c.Context.Batches))
	copy(batches, c.Context.Batches)
	sort.Slice(batches, func(i, j int) bool {
		return batches[i].ProductionDate.Before(batches[j].ProductionDate)
	})

	for _, batch := range batches {
		if batch.AvailableQuantity() > 0 {
			return batch.UnitCost
		}
	}
	return 0
}

// GetInventoryValue 获取FIFO库存总价值
func (c *FIFOCostCalculator) GetInventoryValue(ctx context.Context) int64 {
	var totalValue int64
	for _, batch := range c.Context.Batches {
		totalValue += int64(batch.CurrentQuantity) * batch.UnitCost
	}
	return totalValue
}

// --- 加权平均成本核算器 ---

// WeightedAverageCostCalculator 加权平均成本核算器
type WeightedAverageCostCalculator struct {
	Context         *CostContext
	CurrentUnitCost int64 // 当前加权单位成本
	TotalQuantity   int32 // 当前总数量
	TotalValue      int64 // 当前总价值
}

// NewWeightedAverageCostCalculator 创建加权平均成本核算器
func NewWeightedAverageCostCalculator(ctx *CostContext) *WeightedAverageCostCalculator {
	calc := &WeightedAverageCostCalculator{Context: ctx}
	calc.calculateCurrentAverage()
	return calc
}

// calculateCurrentAverage 计算当前加权平均
func (c *WeightedAverageCostCalculator) calculateCurrentAverage() {
	var totalQty int32
	var totalValue int64
	for _, batch := range c.Context.Batches {
		totalQty += batch.CurrentQuantity
		totalValue += int64(batch.CurrentQuantity) * batch.UnitCost
	}
	c.TotalQuantity = totalQty
	c.TotalValue = totalValue
	if totalQty > 0 {
		c.CurrentUnitCost = totalValue / int64(totalQty)
	}
}

// CalculateOutboundCost 计算加权平均出库成本
func (c *WeightedAverageCostCalculator) CalculateOutboundCost(ctx context.Context, quantity int32) (*OutboundCostResult, error) {
	if c.TotalQuantity < quantity {
		return nil, fmt.Errorf("insufficient stock: available=%d, required=%d", c.TotalQuantity, quantity)
	}

	totalCost := int64(quantity) * c.CurrentUnitCost

	return &OutboundCostResult{
		TotalCost:    totalCost,
		UnitCost:     c.CurrentUnitCost,
		BatchDetails: nil, // 加权平均不需要批次明细
	}, nil
}

// CalculateInboundCost 计算加权平均入库成本
func (c *WeightedAverageCostCalculator) CalculateInboundCost(ctx context.Context, quantity int32, unitCost int64) (*InboundCostResult, error) {
	previousCost := c.CurrentUnitCost

	// 加权平均 = (现有价值 + 新入库价值) / (现有数量 + 新入库数量)
	newTotalValue := c.TotalValue + int64(quantity)*unitCost
	newTotalQty := c.TotalQuantity + quantity
	newUnitCost := newTotalValue / int64(newTotalQty)

	c.TotalValue = newTotalValue
	c.TotalQuantity = newTotalQty
	c.CurrentUnitCost = newUnitCost

	return &InboundCostResult{
		NewUnitCost:      newUnitCost,
		PreviousUnitCost: previousCost,
		TotalValue:       newTotalValue,
		TotalQuantity:    newTotalQty,
	}, nil
}

// GetCurrentUnitCost 获取加权平均单位成本
func (c *WeightedAverageCostCalculator) GetCurrentUnitCost(ctx context.Context) int64 {
	return c.CurrentUnitCost
}

// GetInventoryValue 获取加权平均库存总价值
func (c *WeightedAverageCostCalculator) GetInventoryValue(ctx context.Context) int64 {
	return c.TotalValue
}

// --- 移动加权平均成本核算器 ---

// MovingAverageCostCalculator 移动加权平均成本核算器
type MovingAverageCostCalculator struct {
	*WeightedAverageCostCalculator
	History []*CostChangeRecord // 成本变动历史
}

// CostChangeRecord 成本变动记录
type CostChangeRecord struct {
	ID             uint64    `json:"id"`
	Timestamp      time.Time `json:"timestamp"`
	Action         string    `json:"action"` // INBOUND/OUTBOUND
	Quantity       int32     `json:"quantity"`
	UnitCost       int64     `json:"unit_cost"`        // 本次单位成本
	BeforeUnitCost int64     `json:"before_unit_cost"` // 变动前单位成本
	AfterUnitCost  int64     `json:"after_unit_cost"`  // 变动后单位成本
	BeforeQty      int32     `json:"before_qty"`
	AfterQty       int32     `json:"after_qty"`
}

// NewMovingAverageCostCalculator 创建移动加权平均成本核算器
func NewMovingAverageCostCalculator(ctx *CostContext) *MovingAverageCostCalculator {
	return &MovingAverageCostCalculator{
		WeightedAverageCostCalculator: NewWeightedAverageCostCalculator(ctx),
		History:                       make([]*CostChangeRecord, 0),
	}
}

// CalculateInboundCost 计算移动加权平均入库成本（带历史记录）
func (c *MovingAverageCostCalculator) CalculateInboundCost(ctx context.Context, quantity int32, unitCost int64) (*InboundCostResult, error) {
	beforeCost := c.CurrentUnitCost
	beforeQty := c.TotalQuantity

	result, err := c.WeightedAverageCostCalculator.CalculateInboundCost(ctx, quantity, unitCost)
	if err != nil {
		return nil, err
	}

	// 记录历史
	c.History = append(c.History, &CostChangeRecord{
		Timestamp:      time.Now(),
		Action:         "INBOUND",
		Quantity:       quantity,
		UnitCost:       unitCost,
		BeforeUnitCost: beforeCost,
		AfterUnitCost:  result.NewUnitCost,
		BeforeQty:      beforeQty,
		AfterQty:       result.TotalQuantity,
	})

	return result, nil
}

// CalculateOutboundCost 计算移动加权平均出库成本（带历史记录）
func (c *MovingAverageCostCalculator) CalculateOutboundCost(ctx context.Context, quantity int32) (*OutboundCostResult, error) {
	beforeCost := c.CurrentUnitCost
	beforeQty := c.TotalQuantity

	result, err := c.WeightedAverageCostCalculator.CalculateOutboundCost(ctx, quantity)
	if err != nil {
		return nil, err
	}

	// 更新数量（出库后）
	c.TotalQuantity -= quantity
	c.TotalValue -= result.TotalCost

	// 记录历史
	c.History = append(c.History, &CostChangeRecord{
		Timestamp:      time.Now(),
		Action:         "OUTBOUND",
		Quantity:       quantity,
		UnitCost:       result.UnitCost,
		BeforeUnitCost: beforeCost,
		AfterUnitCost:  c.CurrentUnitCost,
		BeforeQty:      beforeQty,
		AfterQty:       c.TotalQuantity,
	})

	return result, nil
}

// --- 成本核算工厂 ---

// CostCalculatorFactory 成本核算器工厂
type CostCalculatorFactory struct{}

// CreateCalculator 创建成本核算器
func (f *CostCalculatorFactory) CreateCalculator(ctx *CostContext) CostCalculator {
	switch ctx.Method {
	case CostMethodFIFO:
		return NewFIFOCostCalculator(ctx)
	case CostMethodWeightedAverage:
		return NewWeightedAverageCostCalculator(ctx)
	case CostMethodMovingAverage:
		return NewMovingAverageCostCalculator(ctx)
	default:
		return NewFIFOCostCalculator(ctx) // 默认FIFO
	}
}

// --- 库存成本账 ---

// InventoryCostLedger 库存成本账
type InventoryCostLedger struct {
	ID          uint64     `json:"id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	SkuID       uint64     `json:"sku_id"`
	WarehouseID uint64     `json:"warehouse_id"`
	CostMethod  CostMethod `json:"cost_method"`
	CurrentQty  int32      `json:"current_qty"`  // 当前库存数量
	CurrentCost int64      `json:"current_cost"` // 当前单位成本
	TotalValue  int64      `json:"total_value"`  // 总价值
	LastUpdated time.Time  `json:"last_updated"`
}

// InventoryCostLedgerRepository 库存成本账仓储接口
type InventoryCostLedgerRepository interface {
	FindBySkuAndWarehouse(ctx context.Context, skuID, warehouseID uint64) (*InventoryCostLedger, error)
	Save(ctx context.Context, ledger *InventoryCostLedger) error
	Update(ctx context.Context, ledger *InventoryCostLedger) error
}

// --- 成本分析报告 ---

// CostAnalysisReport 成本分析报告
type CostAnalysisReport struct {
	SkuID           uint64              `json:"sku_id"`
	WarehouseID     uint64              `json:"warehouse_id"`
	ReportDate      time.Time           `json:"report_date"`
	CostMethod      CostMethod          `json:"cost_method"`
	TotalInbound    int32               `json:"total_inbound"`     // 总入库
	TotalOutbound   int32               `json:"total_outbound"`    // 总出库
	InboundCost     int64               `json:"inbound_cost"`      // 入库总成本
	OutboundCost    int64               `json:"outbound_cost"`     // 出库总成本（销货成本）
	ClosingQty      int32               `json:"closing_qty"`       // 期末库存
	ClosingValue    int64               `json:"closing_value"`     // 期末价值
	GrossProfit     int64               `json:"gross_profit"`      // 毛利（如果有销售价格）
	GrossProfitRate float64             `json:"gross_profit_rate"` // 毛利率
	BatchBreakdown  []*BatchCostSummary `json:"batch_breakdown"`   // 批次明细
}

// BatchCostSummary 批次成本汇总
type BatchCostSummary struct {
	BatchNo     string    `json:"batch_no"`
	InboundQty  int32     `json:"inbound_qty"`
	OutboundQty int32     `json:"outbound_qty"`
	RemainQty   int32     `json:"remain_qty"`
	UnitCost    int64     `json:"unit_cost"`
	TotalCost   int64     `json:"total_cost"`
	ExpiryDate  time.Time `json:"expiry_date"`
}

// GenerateCostReport 生成成本分析报告
func GenerateCostReport(ctx *CostContext, inboundRecords, outboundRecords []*CostChangeRecord) *CostAnalysisReport {
	report := &CostAnalysisReport{
		SkuID:       ctx.SkuID,
		WarehouseID: ctx.WarehouseID,
		ReportDate:  time.Now(),
		CostMethod:  ctx.Method,
	}

	// 汇总入库
	for _, record := range inboundRecords {
		report.TotalInbound += record.Quantity
		report.InboundCost += int64(record.Quantity) * record.UnitCost
	}

	// 汇总出库
	for _, record := range outboundRecords {
		report.TotalOutbound += record.Quantity
		report.OutboundCost += int64(record.Quantity) * record.UnitCost
	}

	// 计算期末
	report.ClosingQty = report.TotalInbound - report.TotalOutbound
	report.ClosingValue = report.InboundCost - report.OutboundCost

	// 批次明细
	for _, batch := range ctx.Batches {
		report.BatchBreakdown = append(report.BatchBreakdown, &BatchCostSummary{
			BatchNo:     batch.BatchNo,
			InboundQty:  batch.InitialQuantity,
			OutboundQty: batch.InitialQuantity - batch.CurrentQuantity,
			RemainQty:   batch.CurrentQuantity,
			UnitCost:    batch.UnitCost,
			TotalCost:   int64(batch.CurrentQuantity) * batch.UnitCost,
			ExpiryDate:  batch.ExpiryDate,
		})
	}

	return report
}
