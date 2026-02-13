package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrStockTakeNotFound       = errors.New("stock take not found")
	ErrStockTakeAlreadyClosed  = errors.New("stock take already closed")
	ErrStockTakeNotInProgress  = errors.New("stock take not in progress")
	ErrStockTakeItemNotFound   = errors.New("stock take item not found")
	ErrStockTakeCannotModify   = errors.New("stock take cannot be modified")
	ErrInvalidStockTakeStatus  = errors.New("invalid stock take status")
)

type StockTakeStatus int8

const (
	StockTakeStatusPending    StockTakeStatus = 1
	StockTakeStatusInProgress StockTakeStatus = 2
	StockTakeStatusCompleted  StockTakeStatus = 3
	StockTakeStatusCancelled  StockTakeStatus = 4
	StockTakeStatusAdjusting  StockTakeStatus = 5
	StockTakeStatusAdjusted   StockTakeStatus = 6
)

func (s StockTakeStatus) String() string {
	switch s {
	case StockTakeStatusPending:
		return "PENDING"
	case StockTakeStatusInProgress:
		return "IN_PROGRESS"
	case StockTakeStatusCompleted:
		return "COMPLETED"
	case StockTakeStatusCancelled:
		return "CANCELLED"
	case StockTakeStatusAdjusting:
		return "ADJUSTING"
	case StockTakeStatusAdjusted:
		return "ADJUSTED"
	default:
		return "UNKNOWN"
	}
}

type StockTakeType int8

const (
	StockTakeTypeFull      StockTakeType = 1
	StockTakeTypePartial   StockTakeType = 2
	StockTakeTypeCycle     StockTakeType = 3
	StockTakeTypeSpotCheck StockTakeType = 4
)

func (t StockTakeType) String() string {
	switch t {
	case StockTakeTypeFull:
		return "FULL"
	case StockTakeTypePartial:
		return "PARTIAL"
	case StockTakeTypeCycle:
		return "CYCLE"
	case StockTakeTypeSpotCheck:
		return "SPOT_CHECK"
	default:
		return "UNKNOWN"
	}
}

type StockTakeItemStatus int8

const (
	StockTakeItemStatusPending    StockTakeItemStatus = 1
	StockTakeItemStatusCounted    StockTakeItemStatus = 2
	StockTakeItemStatusVerified   StockTakeItemStatus = 3
	StockTakeItemStatusAdjusted   StockTakeItemStatus = 4
	StockTakeItemStatusDiscrepancy StockTakeItemStatus = 5
)

func (s StockTakeItemStatus) String() string {
	switch s {
	case StockTakeItemStatusPending:
		return "PENDING"
	case StockTakeItemStatusCounted:
		return "COUNTED"
	case StockTakeItemStatusVerified:
		return "VERIFIED"
	case StockTakeItemStatusAdjusted:
		return "ADJUSTED"
	case StockTakeItemStatusDiscrepancy:
		return "DISCREPANCY"
	default:
		return "UNKNOWN"
	}
}

type StockTake struct {
	ID               uint              `json:"id"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
	StockTakeNo      string            `json:"stock_take_no"`
	WarehouseID      uint64            `json:"warehouse_id"`
	WarehouseName    string            `json:"warehouse_name"`
	Type             StockTakeType     `json:"type"`
	Status           StockTakeStatus   `json:"status"`
	ScheduledAt      *time.Time        `json:"scheduled_at"`
	StartedAt        *time.Time        `json:"started_at"`
	CompletedAt      *time.Time        `json:"completed_at"`
	CancelledAt      *time.Time        `json:"cancelled_at"`
	CancelReason     string            `json:"cancel_reason"`
	CreatorID        uint64            `json:"creator_id"`
	CreatorName      string            `json:"creator_name"`
	SupervisorID     uint64            `json:"supervisor_id"`
	SupervisorName   string            `json:"supervisor_name"`
	TotalItems       int               `json:"total_items"`
	CountedItems     int               `json:"counted_items"`
	DiscrepancyItems int               `json:"discrepancy_items"`
	AdjustedItems    int               `json:"adjusted_items"`
	TotalValue       int64             `json:"total_value"`
	DiscrepancyValue int64             `json:"discrepancy_value"`
	Notes            string            `json:"notes"`
	Items            []*StockTakeItem  `json:"items"`
	Histories        []*StockTakeHistory `json:"histories"`
	FreezeInventory  bool              `json:"freeze_inventory"`
}

type StockTakeItem struct {
	ID               uint                `json:"id"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
	StockTakeID      uint                `json:"stock_take_id"`
	SkuID            uint64              `json:"sku_id"`
	SkuCode          string              `json:"sku_code"`
	ProductName      string              `json:"product_name"`
	SkuName          string              `json:"sku_name"`
	Location         string              `json:"location"`
	BatchNo          string              `json:"batch_no"`
	SystemQuantity   int32               `json:"system_quantity"`
	CountedQuantity  int32               `json:"counted_quantity"`
	VarianceQuantity int32               `json:"variance_quantity"`
	UnitCost         int64               `json:"unit_cost"`
	VarianceValue    int64               `json:"variance_value"`
	Status           StockTakeItemStatus `json:"status"`
	CounterID        uint64              `json:"counter_id"`
	CounterName      string              `json:"counter_name"`
	CountedAt        *time.Time          `json:"counted_at"`
	VerifierID       uint64              `json:"verifier_id"`
	VerifierName     string              `json:"verifier_name"`
	VerifiedAt       *time.Time          `json:"verified_at"`
	AdjustmentReason string              `json:"adjustment_reason"`
	AdjustedAt       *time.Time          `json:"adjusted_at"`
	Notes            string              `json:"notes"`
	SecondCount      int32               `json:"second_count"`
	ThirdCount       int32               `json:"third_count"`
	IsRecounted      bool                `json:"is_recounted"`
}

type StockTakeHistory struct {
	ID           uint      `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	StockTakeID  uint      `json:"stock_take_id"`
	OperatorID   uint64    `json:"operator_id"`
	OperatorName string    `json:"operator_name"`
	Action       string    `json:"action"`
	OldStatus    string    `json:"old_status"`
	NewStatus    string    `json:"new_status"`
	Comment      string    `json:"comment"`
}

type StockTakeConfig struct {
	AutoFreezeInventory  bool          `json:"auto_freeze_inventory"`
	RequireSecondCount   bool          `json:"require_second_count"`
	DiscrepancyThreshold float64       `json:"discrepancy_threshold"`
	MaxDiscrepancyValue  int64         `json:"max_discrepancy_value"`
	RequireSupervisor    bool          `json:"require_supervisor"`
	TimeoutHours         int           `json:"timeout_hours"`
	AllowPartialComplete bool          `json:"allow_partial_complete"`
}

func DefaultStockTakeConfig() *StockTakeConfig {
	return &StockTakeConfig{
		AutoFreezeInventory:  true,
		RequireSecondCount:   true,
		DiscrepancyThreshold: 5.0,
		MaxDiscrepancyValue:  100000,
		RequireSupervisor:    true,
		TimeoutHours:         24,
		AllowPartialComplete: false,
	}
}

func NewStockTake(stockTakeNo string, warehouseID uint64, warehouseName string, takeType StockTakeType, creatorID uint64, creatorName string) *StockTake {
	return &StockTake{
		StockTakeNo:   stockTakeNo,
		WarehouseID:   warehouseID,
		WarehouseName: warehouseName,
		Type:          takeType,
		Status:        StockTakeStatusPending,
		CreatorID:     creatorID,
		CreatorName:   creatorName,
		TotalItems:    0,
		CountedItems:  0,
		Items:         make([]*StockTakeItem, 0),
		Histories:     make([]*StockTakeHistory, 0),
		FreezeInventory: false,
	}
}

func (s *StockTake) AddItem(skuID uint64, skuCode, productName, skuName, location, batchNo string, systemQuantity int32, unitCost int64) *StockTakeItem {
	item := &StockTakeItem{
		StockTakeID:     s.ID,
		SkuID:           skuID,
		SkuCode:         skuCode,
		ProductName:     productName,
		SkuName:         skuName,
		Location:        location,
		BatchNo:         batchNo,
		SystemQuantity:  systemQuantity,
		CountedQuantity: 0,
		VarianceQuantity: 0,
		UnitCost:        unitCost,
		VarianceValue:   0,
		Status:          StockTakeItemStatusPending,
		IsRecounted:     false,
	}
	s.Items = append(s.Items, item)
	s.TotalItems = len(s.Items)
	s.TotalValue += int64(systemQuantity) * unitCost
	return item
}

func (s *StockTake) Start(supervisorID uint64, supervisorName string, freezeInventory bool) error {
	if s.Status != StockTakeStatusPending {
		return ErrInvalidStockTakeStatus
	}

	oldStatus := s.Status.String()
	now := time.Now()

	s.Status = StockTakeStatusInProgress
	s.StartedAt = &now
	s.SupervisorID = supervisorID
	s.SupervisorName = supervisorName
	s.FreezeInventory = freezeInventory

	s.addHistory(0, "SYSTEM", "START", oldStatus, s.Status.String(), "盘点开始")

	return nil
}

func (s *StockTake) CountItem(skuID uint64, countedQuantity int32, counterID uint64, counterName, notes string) error {
	if s.Status != StockTakeStatusInProgress {
		return ErrStockTakeNotInProgress
	}

	item := s.findItemBySkuID(skuID)
	if item == nil {
		return ErrStockTakeItemNotFound
	}

	now := time.Now()
	oldStatus := item.Status.String()

	if item.IsRecounted {
		if item.SecondCount == 0 {
			item.SecondCount = countedQuantity
		} else if item.ThirdCount == 0 {
			item.ThirdCount = countedQuantity
		}
	} else {
		item.CountedQuantity = countedQuantity
	}

	item.VarianceQuantity = item.CountedQuantity - item.SystemQuantity
	item.VarianceValue = int64(item.VarianceQuantity) * item.UnitCost
	item.CounterID = counterID
	item.CounterName = counterName
	item.CountedAt = &now
	item.Notes = notes

	if item.VarianceQuantity != 0 {
		item.Status = StockTakeItemStatusDiscrepancy
	} else {
		item.Status = StockTakeItemStatusCounted
	}

	s.CountedItems = s.countItemsByStatus(StockTakeItemStatusCounted, StockTakeItemStatusDiscrepancy)
	s.DiscrepancyItems = s.countItemsByStatus(StockTakeItemStatusDiscrepancy)
	s.DiscrepancyValue = s.calculateDiscrepancyValue()

	s.addHistory(0, counterName, "COUNT_ITEM", oldStatus, item.Status.String(),
		fmt.Sprintf("SKU: %s, Counted: %d, Variance: %d", item.SkuCode, countedQuantity, item.VarianceQuantity))

	return nil
}

func (s *StockTake) VerifyItem(skuID uint64, verifierID uint64, verifierName string, approved bool) error {
	if s.Status != StockTakeStatusInProgress {
		return ErrStockTakeNotInProgress
	}

	item := s.findItemBySkuID(skuID)
	if item == nil {
		return ErrStockTakeItemNotFound
	}

	if item.Status != StockTakeItemStatusCounted && item.Status != StockTakeItemStatusDiscrepancy {
		return ErrInvalidStockTakeStatus
	}

	now := time.Now()
	oldStatus := item.Status.String()

	item.VerifierID = verifierID
	item.VerifierName = verifierName
	item.VerifiedAt = &now

	if approved {
		item.Status = StockTakeItemStatusVerified
	} else {
		item.IsRecounted = true
		item.Status = StockTakeItemStatusPending
	}

	s.addHistory(0, verifierName, "VERIFY_ITEM", oldStatus, item.Status.String(),
		fmt.Sprintf("SKU: %s, Approved: %v", item.SkuCode, approved))

	return nil
}

func (s *StockTake) Complete() error {
	if s.Status != StockTakeStatusInProgress {
		return ErrStockTakeNotInProgress
	}

	oldStatus := s.Status.String()
	now := time.Now()

	s.Status = StockTakeStatusCompleted
	s.CompletedAt = &now

	s.addHistory(0, "SYSTEM", "COMPLETE", oldStatus, s.Status.String(), "盘点完成")

	return nil
}

func (s *StockTake) StartAdjustment() error {
	if s.Status != StockTakeStatusCompleted {
		return ErrInvalidStockTakeStatus
	}

	oldStatus := s.Status.String()

	s.Status = StockTakeStatusAdjusting

	s.addHistory(0, "SYSTEM", "START_ADJUSTMENT", oldStatus, s.Status.String(), "开始调整库存")

	return nil
}

func (s *StockTake) AdjustItem(skuID uint64, reason string) error {
	if s.Status != StockTakeStatusAdjusting {
		return ErrInvalidStockTakeStatus
	}

	item := s.findItemBySkuID(skuID)
	if item == nil {
		return ErrStockTakeItemNotFound
	}

	if item.Status != StockTakeItemStatusVerified {
		return ErrInvalidStockTakeStatus
	}

	now := time.Now()
	oldStatus := item.Status.String()

	item.Status = StockTakeItemStatusAdjusted
	item.AdjustmentReason = reason
	item.AdjustedAt = &now

	s.AdjustedItems = s.countItemsByStatus(StockTakeItemStatusAdjusted)

	s.addHistory(0, "SYSTEM", "ADJUST_ITEM", oldStatus, item.Status.String(),
		fmt.Sprintf("SKU: %s, Reason: %s", item.SkuCode, reason))

	return nil
}

func (s *StockTake) CompleteAdjustment() error {
	if s.Status != StockTakeStatusAdjusting {
		return ErrInvalidStockTakeStatus
	}

	oldStatus := s.Status.String()

	s.Status = StockTakeStatusAdjusted

	s.addHistory(0, "SYSTEM", "COMPLETE_ADJUSTMENT", oldStatus, s.Status.String(), "库存调整完成")

	return nil
}

func (s *StockTake) Cancel(reason string, operatorID uint64, operatorName string) error {
	if s.Status == StockTakeStatusAdjusted || s.Status == StockTakeStatusCancelled {
		return ErrStockTakeAlreadyClosed
	}

	oldStatus := s.Status.String()
	now := time.Now()

	s.Status = StockTakeStatusCancelled
	s.CancelledAt = &now
	s.CancelReason = reason

	s.addHistory(operatorID, operatorName, "CANCEL", oldStatus, s.Status.String(), reason)

	return nil
}

func (s *StockTake) GetProgress() float64 {
	if s.TotalItems == 0 {
		return 0
	}
	return float64(s.CountedItems) / float64(s.TotalItems) * 100
}

func (s *StockTake) HasDiscrepancy() bool {
	return s.DiscrepancyItems > 0
}

func (s *StockTake) IsComplete() bool {
	return s.Status == StockTakeStatusCompleted || s.Status == StockTakeStatusAdjusted
}

func (s *StockTake) findItemBySkuID(skuID uint64) *StockTakeItem {
	for _, item := range s.Items {
		if item.SkuID == skuID {
			return item
		}
	}
	return nil
}

func (s *StockTake) countItemsByStatus(statuses ...StockTakeItemStatus) int {
	count := 0
	for _, item := range s.Items {
		for _, status := range statuses {
			if item.Status == status {
				count++
				break
			}
		}
	}
	return count
}

func (s *StockTake) calculateDiscrepancyValue() int64 {
	var total int64
	for _, item := range s.Items {
		if item.Status == StockTakeItemStatusDiscrepancy {
			total += item.VarianceValue
		}
	}
	return total
}

func (s *StockTake) addHistory(operatorID uint64, operatorName, action, oldStatus, newStatus, comment string) {
	history := &StockTakeHistory{
		StockTakeID:  s.ID,
		OperatorID:   operatorID,
		OperatorName: operatorName,
		Action:       action,
		OldStatus:    oldStatus,
		NewStatus:    newStatus,
		Comment:      comment,
	}
	s.Histories = append(s.Histories, history)
}

type StockTakeRepository interface {
	Save(ctx context.Context, stockTake *StockTake) error
	FindByID(ctx context.Context, id uint64) (*StockTake, error)
	FindByStockTakeNo(ctx context.Context, stockTakeNo string) (*StockTake, error)
	FindByWarehouseID(ctx context.Context, warehouseID uint64, limit, offset int) ([]*StockTake, error)
	FindPending(ctx context.Context, limit, offset int) ([]*StockTake, error)
	FindInProgress(ctx context.Context) ([]*StockTake, error)
	FindByStatus(ctx context.Context, status StockTakeStatus, limit, offset int) ([]*StockTake, error)
	Update(ctx context.Context, stockTake *StockTake) error
}

type StockTakeService interface {
	CreateStockTake(ctx context.Context, warehouseID uint64, takeType StockTakeType, creatorID uint64, creatorName string, items []*StockTakeItemInput) (*StockTake, error)
	StartStockTake(ctx context.Context, stockTakeID uint64, supervisorID uint64, supervisorName string, freezeInventory bool) error
	CountItem(ctx context.Context, stockTakeID uint64, skuID uint64, countedQuantity int32, counterID uint64, counterName, notes string) error
	VerifyItem(ctx context.Context, stockTakeID uint64, skuID uint64, verifierID uint64, verifierName string, approved bool) error
	CompleteStockTake(ctx context.Context, stockTakeID uint64) error
	ProcessAdjustments(ctx context.Context, stockTakeID uint64) error
	CancelStockTake(ctx context.Context, stockTakeID uint64, reason string, operatorID uint64, operatorName string) error
}

type StockTakeItemInput struct {
	SkuID          uint64
	SkuCode        string
	ProductName    string
	SkuName        string
	Location       string
	BatchNo        string
	SystemQuantity int32
	UnitCost       int64
}
