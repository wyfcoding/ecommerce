package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInventoryCheckNotFound   = errors.New("inventory check not found")
	ErrCheckAlreadyCompleted    = errors.New("inventory check already completed")
	ErrCheckCannotCancel        = errors.New("inventory check cannot be cancelled")
	ErrCheckItemNotFound        = errors.New("check item not found")
)

type CheckStatus int8

const (
	CheckStatusPending    CheckStatus = 1
	CheckStatusInProgress CheckStatus = 2
	CheckStatusCompleted  CheckStatus = 3
	CheckStatusCancelled  CheckStatus = 4
)

type CheckType int8

const (
	CheckTypeFull       CheckType = 1
	CheckTypePartial    CheckType = 2
	CheckTypeCycle      CheckType = 3
	CheckTypeSpot       CheckType = 4
)

type InventoryCheck struct {
	ID               uint64            `json:"id"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
	CheckNo          string            `json:"check_no"`
	WarehouseID      uint64            `json:"warehouse_id"`
	CheckType        CheckType         `json:"check_type"`
	Status           CheckStatus       `json:"status"`
	ScheduledDate    time.Time         `json:"scheduled_date"`
	StartedAt        *time.Time        `json:"started_at"`
	CompletedAt      *time.Time        `json:"completed_at"`
	CreatedBy        uint64            `json:"created_by"`
	AssignedTo       []uint64          `json:"assigned_to"`
	TotalItems       int               `json:"total_items"`
	CheckedItems     int               `json:"checked_items"`
	MatchedItems     int               `json:"matched_items"`
	MismatchItems    int               `json:"mismatch_items"`
	Notes            string            `json:"notes"`
	Items            []*InventoryCheckItem `json:"items"`
}

type InventoryCheckItem struct {
	ID              uint64      `json:"id"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
	CheckID         uint64      `json:"check_id"`
	SkuID           uint64      `json:"sku_id"`
	ProductID       uint64      `json:"product_id"`
	ProductName     string      `json:"product_name"`
	SkuName         string      `json:"sku_name"`
	SystemQuantity  int32       `json:"system_quantity"`
	ActualQuantity  int32       `json:"actual_quantity"`
	Difference      int32       `json:"difference"`
	DifferenceReason string     `json:"difference_reason"`
	Status          CheckItemStatus `json:"status"`
	CheckedBy       uint64      `json:"checked_by"`
	CheckedAt       *time.Time  `json:"checked_at"`
	ApprovedBy      uint64      `json:"approved_by"`
	ApprovedAt      *time.Time  `json:"approved_at"`
	Location        string      `json:"location"`
	BatchNo         string      `json:"batch_no"`
	Images          []string    `json:"images"`
	Notes           string      `json:"notes"`
}

type CheckItemStatus int8

const (
	CheckItemStatusPending   CheckItemStatus = 1
	CheckItemStatusChecked   CheckItemStatus = 2
	CheckItemStatusConfirmed CheckItemStatus = 3
	CheckItemStatusAdjusted  CheckItemStatus = 4
)

type CheckDifferenceType int8

const (
	DifferenceTypeSurplus  CheckDifferenceType = 1
	DifferenceTypeShortage CheckDifferenceType = 2
	DifferenceTypeMatch    CheckDifferenceType = 3
)

type InventoryAdjustment struct {
	ID             uint64              `json:"id"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
	AdjustmentNo   string              `json:"adjustment_no"`
	CheckID        uint64              `json:"check_id"`
	CheckItemID    uint64              `json:"check_item_id"`
	SkuID          uint64              `json:"sku_id"`
	WarehouseID    uint64              `json:"warehouse_id"`
	AdjustmentType CheckDifferenceType `json:"adjustment_type"`
	BeforeQuantity int32               `json:"before_quantity"`
	AdjustQuantity int32               `json:"adjust_quantity"`
	AfterQuantity  int32               `json:"after_quantity"`
	Reason         string              `json:"reason"`
	ApprovedBy     uint64              `json:"approved_by"`
	ApprovedAt     *time.Time          `json:"approved_at"`
	Status         AdjustmentStatus    `json:"status"`
}

type AdjustmentStatus int8

const (
	AdjustmentStatusPending  AdjustmentStatus = 1
	AdjustmentStatusApproved AdjustmentStatus = 2
	AdjustmentStatusRejected AdjustmentStatus = 3
	AdjustmentStatusExecuted AdjustmentStatus = 4
)

type CheckConfig struct {
	RequireApproval        bool
	AutoAdjustThreshold    int32
	AllowPartialCheck      bool
	RequirePhotoEvidence   bool
	MaxDifferencePercent   float64
}

func DefaultCheckConfig() *CheckConfig {
	return &CheckConfig{
		RequireApproval:        true,
		AutoAdjustThreshold:    10,
		AllowPartialCheck:      true,
		RequirePhotoEvidence:   false,
		MaxDifferencePercent:   50.0,
	}
}

func NewInventoryCheck(checkNo string, warehouseID uint64, checkType CheckType, scheduledDate time.Time, createdBy uint64) *InventoryCheck {
	return &InventoryCheck{
		CheckNo:       checkNo,
		WarehouseID:   warehouseID,
		CheckType:     checkType,
		Status:        CheckStatusPending,
		ScheduledDate: scheduledDate,
		CreatedBy:     createdBy,
		AssignedTo:    []uint64{},
		Items:         []*InventoryCheckItem{},
	}
}

func (c *InventoryCheck) AddItem(skuID, productID uint64, productName, skuName string, systemQuantity int32, location, batchNo string) {
	item := &InventoryCheckItem{
		CheckID:        c.ID,
		SkuID:          skuID,
		ProductID:      productID,
		ProductName:    productName,
		SkuName:        skuName,
		SystemQuantity: systemQuantity,
		Status:         CheckItemStatusPending,
		Location:       location,
		BatchNo:        batchNo,
		Images:         []string{},
	}
	c.Items = append(c.Items, item)
	c.TotalItems++
}

func (c *InventoryCheck) Start() error {
	if c.Status != CheckStatusPending {
		return ErrCheckAlreadyCompleted
	}
	c.Status = CheckStatusInProgress
	now := time.Now()
	c.StartedAt = &now
	return nil
}

func (c *InventoryCheck) CheckItem(skuID uint64, actualQuantity int32, checkedBy uint64, images []string, notes string) error {
	if c.Status != CheckStatusInProgress {
		return ErrCheckAlreadyCompleted
	}

	var item *InventoryCheckItem
	for _, i := range c.Items {
		if i.SkuID == skuID && i.Status == CheckItemStatusPending {
			item = i
			break
		}
	}

	if item == nil {
		return ErrCheckItemNotFound
	}

	item.ActualQuantity = actualQuantity
	item.Difference = actualQuantity - item.SystemQuantity
	item.CheckedBy = checkedBy
	item.Status = CheckItemStatusChecked
	item.Images = images
	item.Notes = notes
	now := time.Now()
	item.CheckedAt = &now

	c.CheckedItems++
	if item.Difference == 0 {
		c.MatchedItems++
	} else {
		c.MismatchItems++
	}

	return nil
}

func (c *InventoryCheck) Complete() error {
	if c.Status != CheckStatusInProgress {
		return ErrCheckAlreadyCompleted
	}

	c.Status = CheckStatusCompleted
	now := time.Now()
	c.CompletedAt = &now
	return nil
}

func (c *InventoryCheck) Cancel(reason string) error {
	if c.Status == CheckStatusCompleted {
		return ErrCheckCannotCancel
	}

	c.Status = CheckStatusCancelled
	c.Notes = reason
	return nil
}

func (c *InventoryCheck) GetProgress() float64 {
	if c.TotalItems == 0 {
		return 0
	}
	return float64(c.CheckedItems) / float64(c.TotalItems) * 100
}

func (c *InventoryCheck) GetMismatchRate() float64 {
	if c.CheckedItems == 0 {
		return 0
	}
	return float64(c.MismatchItems) / float64(c.CheckedItems) * 100
}

func (c *InventoryCheck) GetItemsByDifference(diffType CheckDifferenceType) []*InventoryCheckItem {
	var items []*InventoryCheckItem
	for _, item := range c.Items {
		switch diffType {
		case DifferenceTypeSurplus:
			if item.Difference > 0 {
				items = append(items, item)
			}
		case DifferenceTypeShortage:
			if item.Difference < 0 {
				items = append(items, item)
			}
		case DifferenceTypeMatch:
			if item.Difference == 0 {
				items = append(items, item)
			}
		}
	}
	return items
}

func (i *InventoryCheckItem) GetDifferenceType() CheckDifferenceType {
	if i.Difference > 0 {
		return DifferenceTypeSurplus
	} else if i.Difference < 0 {
		return DifferenceTypeShortage
	}
	return DifferenceTypeMatch
}

func (i *InventoryCheckItem) Confirm(approvedBy uint64) {
	i.Status = CheckItemStatusConfirmed
	i.ApprovedBy = approvedBy
	now := time.Now()
	i.ApprovedAt = &now
}

func (i *InventoryCheckItem) SetDifferenceReason(reason string) {
	i.DifferenceReason = reason
}

func NewInventoryAdjustment(adjustmentNo string, checkID, checkItemID, skuID, warehouseID uint64, adjustmentType CheckDifferenceType, beforeQuantity, adjustQuantity int32, reason string) *InventoryAdjustment {
	return &InventoryAdjustment{
		AdjustmentNo:   adjustmentNo,
		CheckID:        checkID,
		CheckItemID:    checkItemID,
		SkuID:          skuID,
		WarehouseID:    warehouseID,
		AdjustmentType: adjustmentType,
		BeforeQuantity: beforeQuantity,
		AdjustQuantity: adjustQuantity,
		AfterQuantity:  beforeQuantity + adjustQuantity,
		Reason:         reason,
		Status:         AdjustmentStatusPending,
	}
}

func (a *InventoryAdjustment) Approve(approvedBy uint64) {
	a.Status = AdjustmentStatusApproved
	a.ApprovedBy = approvedBy
	now := time.Now()
	a.ApprovedAt = &now
}

func (a *InventoryAdjustment) Reject(reason string) {
	a.Status = AdjustmentStatusRejected
	a.Reason = reason
}

func (a *InventoryAdjustment) Execute() {
	a.Status = AdjustmentStatusExecuted
}

func (s CheckStatus) String() string {
	switch s {
	case CheckStatusPending:
		return "PENDING"
	case CheckStatusInProgress:
		return "IN_PROGRESS"
	case CheckStatusCompleted:
		return "COMPLETED"
	case CheckStatusCancelled:
		return "CANCELLED"
	default:
		return "UNKNOWN"
	}
}

type InventoryCheckRepository interface {
	Save(ctx context.Context, check *InventoryCheck) error
	FindByID(ctx context.Context, id uint64) (*InventoryCheck, error)
	FindByCheckNo(ctx context.Context, checkNo string) (*InventoryCheck, error)
	FindByWarehouseID(ctx context.Context, warehouseID uint64, limit, offset int) ([]*InventoryCheck, error)
	FindPending(ctx context.Context, limit, offset int) ([]*InventoryCheck, error)
	FindInProgress(ctx context.Context) ([]*InventoryCheck, error)
	Update(ctx context.Context, check *InventoryCheck) error
}

type InventoryAdjustmentRepository interface {
	Save(ctx context.Context, adjustment *InventoryAdjustment) error
	FindByID(ctx context.Context, id uint64) (*InventoryAdjustment, error)
	FindByCheckID(ctx context.Context, checkID uint64) ([]*InventoryAdjustment, error)
	FindPending(ctx context.Context) ([]*InventoryAdjustment, error)
	Update(ctx context.Context, adjustment *InventoryAdjustment) error
}
