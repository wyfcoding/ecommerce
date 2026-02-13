package domain

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrTransferNotFound       = errors.New("transfer record not found")
	ErrTransferAlreadyPending = errors.New("transfer already pending")
	ErrTransferCannotCancel   = errors.New("transfer cannot be cancelled in current status")
	ErrInsufficientStockForTransfer = errors.New("insufficient stock for transfer")
)

type StockTransferStatus int8

const (
	StockTransferStatusPending    StockTransferStatus = 1
	StockTransferStatusApproved   StockTransferStatus = 2
	StockTransferStatusInProgress StockTransferStatus = 3
	StockTransferStatusCompleted  StockTransferStatus = 4
	StockTransferStatusCancelled  StockTransferStatus = 5
	StockTransferStatusRejected   StockTransferStatus = 6
)

type TransferType int8

const (
	TransferTypeWarehouse TransferType = 1
	TransferTypePurchase  TransferType = 2
	TransferTypeReturn    TransferType = 3
	TransferTypeAdjust    TransferType = 4
)

type StockTransfer struct {
	ID               uint64             `json:"id"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
	TransferNo       string             `json:"transfer_no"`
	SkuID            uint64             `json:"sku_id"`
	ProductID        uint64             `json:"product_id"`
	FromWarehouseID  uint64             `json:"from_warehouse_id"`
	ToWarehouseID    uint64             `json:"to_warehouse_id"`
	Quantity         int32              `json:"quantity"`
	Status           StockTransferStatus `json:"status"`
	TransferType     TransferType       `json:"transfer_type"`
	Reason           string             `json:"reason"`
	RequestedBy      uint64             `json:"requested_by"`
	ApprovedBy       uint64             `json:"approved_by"`
	ApprovedAt       *time.Time         `json:"approved_at"`
	RejectedBy       uint64             `json:"rejected_by"`
	RejectedAt       *time.Time         `json:"rejected_at"`
	RejectionReason  string             `json:"rejection_reason"`
	ShippedAt        *time.Time         `json:"shipped_at"`
	ReceivedAt       *time.Time         `json:"received_at"`
	ReceivedQuantity int32              `json:"received_quantity"`
	Notes            string             `json:"notes"`
	Logs             []*TransferLog     `json:"logs"`
}

type TransferLog struct {
	ID         uint64    `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	TransferID uint64    `json:"transfer_id"`
	Action     string    `json:"action"`
	OldStatus  string    `json:"old_status"`
	NewStatus  string    `json:"new_status"`
	Operator   string    `json:"operator"`
	Remark     string    `json:"remark"`
}

type TransferConfig struct {
	RequireApproval      bool
	ApprovalThreshold    int32
	AutoApproveSameOwner bool
	MaxTransferQuantity  int32
}

func DefaultTransferConfig() *TransferConfig {
	return &TransferConfig{
		RequireApproval:      true,
		ApprovalThreshold:    100,
		AutoApproveSameOwner: true,
		MaxTransferQuantity:  10000,
	}
}

func NewStockTransfer(transferNo string, skuID, productID, fromWarehouseID, toWarehouseID uint64, quantity int32, transferType TransferType, reason string, requestedBy uint64) *StockTransfer {
	return &StockTransfer{
		TransferNo:      transferNo,
		SkuID:           skuID,
		ProductID:       productID,
		FromWarehouseID: fromWarehouseID,
		ToWarehouseID:   toWarehouseID,
		Quantity:        quantity,
		Status:          StockTransferStatusPending,
		TransferType:    transferType,
		Reason:          reason,
		RequestedBy:     requestedBy,
		Logs:            []*TransferLog{},
	}
}

func (t *StockTransfer) Approve(approverID uint64) error {
	if t.Status != StockTransferStatusPending {
		return ErrTransferAlreadyPending
	}

	t.Status = StockTransferStatusApproved
	t.ApprovedBy = approverID
	now := time.Now()
	t.ApprovedAt = &now
	t.addLog("Approve", "PENDING", "APPROVED", fmt.Sprintf("%d", approverID), "")
	return nil
}

func (t *StockTransfer) Reject(approverID uint64, reason string) error {
	if t.Status != StockTransferStatusPending {
		return ErrTransferAlreadyPending
	}

	t.Status = StockTransferStatusRejected
	t.RejectedBy = approverID
	t.RejectionReason = reason
	now := time.Now()
	t.RejectedAt = &now
	t.addLog("Reject", "PENDING", "REJECTED", fmt.Sprintf("%d", approverID), reason)
	return nil
}

func (t *StockTransfer) Ship() error {
	if t.Status != StockTransferStatusApproved {
		return fmt.Errorf("cannot ship in status %d", t.Status)
	}

	t.Status = StockTransferStatusInProgress
	now := time.Now()
	t.ShippedAt = &now
	t.addLog("Ship", "APPROVED", "IN_PROGRESS", "System", "Shipment started")
	return nil
}

func (t *StockTransfer) Receive(receivedQuantity int32, notes string) error {
	if t.Status != StockTransferStatusInProgress {
		return fmt.Errorf("cannot receive in status %d", t.Status)
	}

	t.Status = StockTransferStatusCompleted
	t.ReceivedQuantity = receivedQuantity
	t.Notes = notes
	now := time.Now()
	t.ReceivedAt = &now
	t.addLog("Receive", "IN_PROGRESS", "COMPLETED", "System", fmt.Sprintf("Received %d, notes: %s", receivedQuantity, notes))
	return nil
}

func (t *StockTransfer) Cancel(reason string) error {
	if t.Status == StockTransferStatusCompleted || t.Status == StockTransferStatusCancelled {
		return ErrTransferCannotCancel
	}

	t.Status = StockTransferStatusCancelled
	t.addLog("Cancel", t.Status.String(), "CANCELLED", "System", reason)
	return nil
}

func (t *StockTransfer) IsPending() bool {
	return t.Status == StockTransferStatusPending
}

func (t *StockTransfer) IsCompleted() bool {
	return t.Status == StockTransferStatusCompleted
}

func (t *StockTransfer) CanModify() bool {
	return t.Status == StockTransferStatusPending
}

func (t *StockTransfer) addLog(action, oldStatus, newStatus, operator, remark string) {
	t.Logs = append(t.Logs, &TransferLog{
		TransferID: t.ID,
		Action:     action,
		OldStatus:  oldStatus,
		NewStatus:  newStatus,
		Operator:   operator,
		Remark:     remark,
	})
}

func (s StockTransferStatus) String() string {
	switch s {
	case StockTransferStatusPending:
		return "PENDING"
	case StockTransferStatusApproved:
		return "APPROVED"
	case StockTransferStatusInProgress:
		return "IN_PROGRESS"
	case StockTransferStatusCompleted:
		return "COMPLETED"
	case StockTransferStatusCancelled:
		return "CANCELLED"
	case StockTransferStatusRejected:
		return "REJECTED"
	default:
		return "UNKNOWN"
	}
}

type TransferRepository interface {
	Save(ctx context.Context, transfer *StockTransfer) error
	FindByID(ctx context.Context, id uint64) (*StockTransfer, error)
	FindByTransferNo(ctx context.Context, transferNo string) (*StockTransfer, error)
	FindByWarehouseID(ctx context.Context, warehouseID uint64, limit, offset int) ([]*StockTransfer, error)
	FindBySkuID(ctx context.Context, skuID uint64, limit, offset int) ([]*StockTransfer, error)
	FindPending(ctx context.Context, limit, offset int) ([]*StockTransfer, error)
	FindInProgress(ctx context.Context) ([]*StockTransfer, error)
	Update(ctx context.Context, transfer *StockTransfer) error
}

type TransferService interface {
	CreateTransfer(ctx context.Context, skuID, productID, fromWarehouseID, toWarehouseID uint64, quantity int32, transferType TransferType, reason string, requestedBy uint64) (*StockTransfer, error)
	ApproveTransfer(ctx context.Context, transferID, approverID uint64) error
	RejectTransfer(ctx context.Context, transferID, approverID uint64, reason string) error
	ShipTransfer(ctx context.Context, transferID uint64) error
	ReceiveTransfer(ctx context.Context, transferID uint64, receivedQuantity int32, notes string) error
	CancelTransfer(ctx context.Context, transferID uint64, reason string) error
}
