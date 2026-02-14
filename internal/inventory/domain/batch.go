// 变更说明：新增批次管理和序列号追踪功能，支持效期管理、批次号、序列号追踪。
// 假设：批次按FIFO原则出库，序列号唯一不可重复。
package domain

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"
)

// --- 批次状态 ---

// BatchStatus 批次状态
type BatchStatus int

const (
	BatchStatusActive   BatchStatus = 1 // 有效
	BatchStatusExpiring BatchStatus = 2 // 即将过期（30天内）
	BatchStatusExpired  BatchStatus = 3 // 已过期
	BatchStatusDepleted BatchStatus = 4 // 已用完
	BatchStatusBlocked  BatchStatus = 5 // 锁定（质量问题等）
)

// --- 批次实体 ---

// Batch 批次实体
type Batch struct {
	ID              uint64      `json:"id"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	SkuID           uint64      `json:"sku_id"`
	ProductID       uint64      `json:"product_id"`
	WarehouseID     uint64      `json:"warehouse_id"`
	BatchNo         string      `json:"batch_no"`          // 批次号
	ProductionDate  time.Time   `json:"production_date"`   // 生产日期
	ExpiryDate      time.Time   `json:"expiry_date"`       // 保质期截止
	ShelfLifeDays   int         `json:"shelf_life_days"`   // 保质期天数
	InitialQuantity int32       `json:"initial_quantity"`  // 初始入库数量
	CurrentQuantity int32       `json:"current_quantity"`  // 当前库存数量
	LockedQuantity  int32       `json:"locked_quantity"`   // 锁定数量
	UnitCost        int64       `json:"unit_cost"`         // 批次成本单价（分）
	TotalCost       int64       `json:"total_cost"`        // 总成本（分）
	SupplierID      uint64      `json:"supplier_id"`       // 供应商ID
	SupplierBatchNo string      `json:"supplier_batch_no"` // 供应商批次号
	Status          BatchStatus `json:"status"`
	Remark          string      `json:"remark"`
}

// NewBatch 创建批次
func NewBatch(skuID, productID, warehouseID uint64, batchNo string, productionDate, expiryDate time.Time, quantity int32, unitCost int64, supplierID uint64) *Batch {
	return &Batch{
		SkuID:           skuID,
		ProductID:       productID,
		WarehouseID:     warehouseID,
		BatchNo:         batchNo,
		ProductionDate:  productionDate,
		ExpiryDate:      expiryDate,
		ShelfLifeDays:   int(expiryDate.Sub(productionDate).Hours() / 24),
		InitialQuantity: quantity,
		CurrentQuantity: quantity,
		LockedQuantity:  0,
		UnitCost:        unitCost,
		TotalCost:       int64(quantity) * unitCost,
		SupplierID:      supplierID,
		Status:          BatchStatusActive,
	}
}

// AvailableQuantity 获取可用数量
func (b *Batch) AvailableQuantity() int32 {
	return b.CurrentQuantity - b.LockedQuantity
}

// Lock 锁定批次库存
func (b *Batch) Lock(quantity int32) error {
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}
	if b.AvailableQuantity() < quantity {
		return fmt.Errorf("insufficient batch stock: available=%d, required=%d", b.AvailableQuantity(), quantity)
	}
	b.LockedQuantity += quantity
	return nil
}

// Unlock 解锁批次库存
func (b *Batch) Unlock(quantity int32) error {
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}
	if b.LockedQuantity < quantity {
		return fmt.Errorf("cannot unlock more than locked: locked=%d, requested=%d", b.LockedQuantity, quantity)
	}
	b.LockedQuantity -= quantity
	return nil
}

// Deduct 扣减批次库存
func (b *Batch) Deduct(quantity int32) error {
	if quantity <= 0 {
		return errors.New("quantity must be positive")
	}
	if b.CurrentQuantity < quantity {
		return fmt.Errorf("insufficient batch stock: current=%d, required=%d", b.CurrentQuantity, quantity)
	}
	b.CurrentQuantity -= quantity
	// 如果是从锁定中扣减，同时减少锁定数量
	if b.LockedQuantity >= quantity {
		b.LockedQuantity -= quantity
	}
	b.updateStatus()
	return nil
}

// Add 增加批次库存
func (b *Batch) Add(quantity int32) {
	b.CurrentQuantity += quantity
	b.updateStatus()
}

// updateStatus 更新批次状态
func (b *Batch) updateStatus() {
	now := time.Now()
	if b.CurrentQuantity == 0 {
		b.Status = BatchStatusDepleted
		return
	}
	if now.After(b.ExpiryDate) {
		b.Status = BatchStatusExpired
		return
	}
	// 30天内过期
	if b.ExpiryDate.Sub(now).Hours() < 30*24 {
		b.Status = BatchStatusExpiring
		return
	}
	b.Status = BatchStatusActive
}

// IsExpired 判断是否已过期
func (b *Batch) IsExpired() bool {
	return time.Now().After(b.ExpiryDate)
}

// DaysUntilExpiry 计算距离过期的天数
func (b *Batch) DaysUntilExpiry() int {
	days := int(time.Until(b.ExpiryDate).Hours() / 24)
	if days < 0 {
		return 0
	}
	return days
}

// Block 锁定批次（质量问题等）
func (b *Batch) Block(reason string) {
	b.Status = BatchStatusBlocked
	b.Remark = reason
}

// Unblock 解除锁定
func (b *Batch) Unblock() {
	b.updateStatus()
	b.Remark = ""
}

// --- 序列号状态 ---

// SerialNumberStatus 序列号状态
type SerialNumberStatus int

const (
	SNStatusAvailable SerialNumberStatus = 1 // 可用
	SNStatusReserved  SerialNumberStatus = 2 // 已预留
	SNStatusSold      SerialNumberStatus = 3 // 已售出
	SNStatusReturned  SerialNumberStatus = 4 // 已退货
	SNStatusDamaged   SerialNumberStatus = 5 // 损坏
	SNStatusLost      SerialNumberStatus = 6 // 丢失
)

// --- 序列号实体 ---

// SerialNumber 序列号实体
type SerialNumber struct {
	ID          uint64             `json:"id"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	SkuID       uint64             `json:"sku_id"`
	ProductID   uint64             `json:"product_id"`
	BatchID     uint64             `json:"batch_id"` // 所属批次
	WarehouseID uint64             `json:"warehouse_id"`
	SN          string             `json:"sn"` // 序列号
	Status      SerialNumberStatus `json:"status"`
	OrderID     uint64             `json:"order_id"` // 售出时关联订单
	OrderItemID uint64             `json:"order_item_id"`
	ReservedAt  *time.Time         `json:"reserved_at"`
	SoldAt      *time.Time         `json:"sold_at"`
	ReturnedAt  *time.Time         `json:"returned_at"`
	CustomerID  uint64             `json:"customer_id"`  // 购买客户
	WarrantyEnd *time.Time         `json:"warranty_end"` // 保修截止
	Remark      string             `json:"remark"`
}

// NewSerialNumber 创建序列号
func NewSerialNumber(skuID, productID, batchID, warehouseID uint64, sn string) *SerialNumber {
	return &SerialNumber{
		SkuID:       skuID,
		ProductID:   productID,
		BatchID:     batchID,
		WarehouseID: warehouseID,
		SN:          sn,
		Status:      SNStatusAvailable,
	}
}

// Reserve 预留序列号
func (s *SerialNumber) Reserve(orderID, orderItemID uint64) error {
	if s.Status != SNStatusAvailable {
		return fmt.Errorf("serial number %s is not available", s.SN)
	}
	s.Status = SNStatusReserved
	s.OrderID = orderID
	s.OrderItemID = orderItemID
	now := time.Now()
	s.ReservedAt = &now
	return nil
}

// Sell 售出序列号
func (s *SerialNumber) Sell(customerID uint64, warrantyDays int) error {
	if s.Status != SNStatusReserved && s.Status != SNStatusAvailable {
		return fmt.Errorf("serial number %s cannot be sold", s.SN)
	}
	s.Status = SNStatusSold
	s.CustomerID = customerID
	now := time.Now()
	s.SoldAt = &now
	if warrantyDays > 0 {
		warrantyEnd := now.AddDate(0, 0, warrantyDays)
		s.WarrantyEnd = &warrantyEnd
	}
	return nil
}

// Return 退货
func (s *SerialNumber) Return() error {
	if s.Status != SNStatusSold {
		return fmt.Errorf("serial number %s is not sold", s.SN)
	}
	s.Status = SNStatusReturned
	now := time.Now()
	s.ReturnedAt = &now
	return nil
}

// MakeAvailable 重新上架（退货后检查合格）
func (s *SerialNumber) MakeAvailable() {
	s.Status = SNStatusAvailable
	s.OrderID = 0
	s.OrderItemID = 0
	s.CustomerID = 0
	s.ReservedAt = nil
	s.SoldAt = nil
	s.ReturnedAt = nil
}

// MarkDamaged 标记损坏
func (s *SerialNumber) MarkDamaged(reason string) {
	s.Status = SNStatusDamaged
	s.Remark = reason
}

// MarkLost 标记丢失
func (s *SerialNumber) MarkLost() {
	s.Status = SNStatusLost
}

// IsUnderWarranty 是否在保修期内
func (s *SerialNumber) IsUnderWarranty() bool {
	if s.WarrantyEnd == nil {
		return false
	}
	return time.Now().Before(*s.WarrantyEnd)
}

// --- FIFO出库管理器 ---

// FIFOManager FIFO出库管理器
type FIFOManager struct {
	BatchRepository BatchRepository
}

// BatchRepository 批次仓储接口
type BatchRepository interface {
	FindBySkuAndWarehouse(ctx context.Context, skuID, warehouseID uint64) ([]*Batch, error)
	FindActiveBySkuAndWarehouse(ctx context.Context, skuID, warehouseID uint64) ([]*Batch, error)
	Save(ctx context.Context, batch *Batch) error
	Update(ctx context.Context, batch *Batch) error
	FindByID(ctx context.Context, id uint64) (*Batch, error)
	FindByBatchNo(ctx context.Context, batchNo string) (*Batch, error)
}

// AllocateFIFO 按FIFO原则分配批次
func (m *FIFOManager) AllocateFIFO(ctx context.Context, skuID, warehouseID uint64, quantity int32) ([]*BatchAllocation, error) {
	batches, err := m.BatchRepository.FindActiveBySkuAndWarehouse(ctx, skuID, warehouseID)
	if err != nil {
		return nil, err
	}

	// 按生产日期排序（FIFO）
	sort.Slice(batches, func(i, j int) bool {
		return batches[i].ProductionDate.Before(batches[j].ProductionDate)
	})

	var allocations []*BatchAllocation
	remaining := quantity

	for _, batch := range batches {
		if remaining <= 0 {
			break
		}
		if batch.IsExpired() || batch.Status == BatchStatusBlocked {
			continue
		}

		available := batch.AvailableQuantity()
		if available <= 0 {
			continue
		}

		allocQty := min(available, remaining)

		allocations = append(allocations, &BatchAllocation{
			Batch:    batch,
			Quantity: allocQty,
		})
		remaining -= allocQty
	}

	if remaining > 0 {
		return nil, fmt.Errorf("insufficient stock: required=%d, allocated=%d", quantity, quantity-remaining)
	}

	return allocations, nil
}

// BatchAllocation 批次分配结果
type BatchAllocation struct {
	Batch    *Batch `json:"batch"`
	Quantity int32  `json:"quantity"`
}

// --- 预售库存 ---

// PresaleInventory 预售库存（虚拟库存）
type PresaleInventory struct {
	ID              uint64    `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	SkuID           uint64    `json:"sku_id"`
	ProductID       uint64    `json:"product_id"`
	PresaleQuantity int32     `json:"presale_quantity"` // 预售总量
	SoldQuantity    int32     `json:"sold_quantity"`    // 已售数量
	AvailableQty    int32     `json:"available_qty"`    // 可售数量
	StartTime       time.Time `json:"start_time"`       // 预售开始时间
	EndTime         time.Time `json:"end_time"`         // 预售结束时间
	DeliveryDate    time.Time `json:"delivery_date"`    // 预计发货时间
	Status          string    `json:"status"`           // PENDING/ACTIVE/ENDED/CANCELLED
}

// NewPresaleInventory 创建预售库存
func NewPresaleInventory(skuID, productID uint64, quantity int32, startTime, endTime, deliveryDate time.Time) *PresaleInventory {
	return &PresaleInventory{
		SkuID:           skuID,
		ProductID:       productID,
		PresaleQuantity: quantity,
		SoldQuantity:    0,
		AvailableQty:    quantity,
		StartTime:       startTime,
		EndTime:         endTime,
		DeliveryDate:    deliveryDate,
		Status:          "PENDING",
	}
}

// Start 开始预售
func (p *PresaleInventory) Start() error {
	if p.Status != "PENDING" {
		return errors.New("presale already started or ended")
	}
	p.Status = "ACTIVE"
	return nil
}

// Sell 预售售出
func (p *PresaleInventory) Sell(quantity int32) error {
	if p.Status != "ACTIVE" {
		return errors.New("presale is not active")
	}
	if p.AvailableQty < quantity {
		return fmt.Errorf("insufficient presale stock: available=%d, required=%d", p.AvailableQty, quantity)
	}
	p.SoldQuantity += quantity
	p.AvailableQty -= quantity
	return nil
}

// Cancel 取消预售订单（恢复库存）
func (p *PresaleInventory) Cancel(quantity int32) {
	p.SoldQuantity -= quantity
	p.AvailableQty += quantity
}

// End 结束预售
func (p *PresaleInventory) End() {
	p.Status = "ENDED"
}

// IsActive 判断预售是否有效
func (p *PresaleInventory) IsActive() bool {
	now := time.Now()
	return p.Status == "ACTIVE" && now.After(p.StartTime) && now.Before(p.EndTime)
}

// --- 调拨单 ---

// TransferOrderStatus 调拨单状态
type TransferOrderStatus int

const (
	TransferStatusPending   TransferOrderStatus = 1 // 待审批
	TransferStatusApproved  TransferOrderStatus = 2 // 已审批
	TransferStatusShipping  TransferOrderStatus = 3 // 调拨中
	TransferStatusReceived  TransferOrderStatus = 4 // 已收货
	TransferStatusCompleted TransferOrderStatus = 5 // 已完成
	TransferStatusCancelled TransferOrderStatus = 6 // 已取消
)

// TransferOrder 调拨单
type TransferOrder struct {
	ID              uint64              `json:"id"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	TransferNo      string              `json:"transfer_no"`       // 调拨单号
	FromWarehouseID uint64              `json:"from_warehouse_id"` // 调出仓库
	ToWarehouseID   uint64              `json:"to_warehouse_id"`   // 调入仓库
	Status          TransferOrderStatus `json:"status"`
	Items           []*TransferItem     `json:"items"`          // 调拨商品
	TotalQuantity   int32               `json:"total_quantity"` // 总数量
	ShippedAt       *time.Time          `json:"shipped_at"`     // 发货时间
	ReceivedAt      *time.Time          `json:"received_at"`    // 收货时间
	ApprovedBy      string              `json:"approved_by"`    // 审批人
	ApprovedAt      *time.Time          `json:"approved_at"`    // 审批时间
	Remark          string              `json:"remark"`
}

// TransferItem 调拨商品项
type TransferItem struct {
	ID              uint64 `json:"id"`
	TransferOrderID uint64 `json:"transfer_order_id"`
	SkuID           uint64 `json:"sku_id"`
	ProductID       uint64 `json:"product_id"`
	BatchID         uint64 `json:"batch_id"`     // 调拨批次
	Quantity        int32  `json:"quantity"`     // 调拨数量
	ReceivedQty     int32  `json:"received_qty"` // 实收数量
}

// NewTransferOrder 创建调拨单
func NewTransferOrder(transferNo string, fromWarehouseID, toWarehouseID uint64) *TransferOrder {
	return &TransferOrder{
		TransferNo:      transferNo,
		FromWarehouseID: fromWarehouseID,
		ToWarehouseID:   toWarehouseID,
		Status:          TransferStatusPending,
		Items:           make([]*TransferItem, 0),
	}
}

// AddItem 添加调拨商品
func (t *TransferOrder) AddItem(skuID, productID, batchID uint64, quantity int32) {
	item := &TransferItem{
		SkuID:     skuID,
		ProductID: productID,
		BatchID:   batchID,
		Quantity:  quantity,
	}
	t.Items = append(t.Items, item)
	t.TotalQuantity += quantity
}

// Approve 审批通过
func (t *TransferOrder) Approve(approver string) error {
	if t.Status != TransferStatusPending {
		return errors.New("can only approve pending transfer")
	}
	t.Status = TransferStatusApproved
	t.ApprovedBy = approver
	now := time.Now()
	t.ApprovedAt = &now
	return nil
}

// Ship 发货
func (t *TransferOrder) Ship() error {
	if t.Status != TransferStatusApproved {
		return errors.New("can only ship approved transfer")
	}
	t.Status = TransferStatusShipping
	now := time.Now()
	t.ShippedAt = &now
	return nil
}

// Receive 收货
func (t *TransferOrder) Receive(receivedQuantities map[uint64]int32) error {
	if t.Status != TransferStatusShipping {
		return errors.New("can only receive shipping transfer")
	}
	for _, item := range t.Items {
		if qty, ok := receivedQuantities[item.SkuID]; ok {
			item.ReceivedQty = qty
		}
	}
	t.Status = TransferStatusReceived
	now := time.Now()
	t.ReceivedAt = &now
	return nil
}

// Complete 完成调拨
func (t *TransferOrder) Complete() error {
	if t.Status != TransferStatusReceived {
		return errors.New("can only complete received transfer")
	}
	t.Status = TransferStatusCompleted
	return nil
}

// Cancel 取消调拨
func (t *TransferOrder) Cancel(reason string) error {
	if t.Status == TransferStatusCompleted {
		return errors.New("cannot cancel completed transfer")
	}
	t.Status = TransferStatusCancelled
	t.Remark = reason
	return nil
}
