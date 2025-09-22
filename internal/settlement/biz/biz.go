package biz

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrOrderAlreadySettled = errors.New("order already settled")
	ErrRecordNotFound      = errors.New("settlement record not found")
)

// SettlementRecord represents a settlement record in the business logic layer.
type SettlementRecord struct {
	ID               uint
	RecordID         string
	OrderID          uint64
	MerchantID       uint64
	TotalAmount      uint64
	PlatformFee      uint64
	SettlementAmount uint64
	Status           string // PENDING, COMPLETED, FAILED
	CreatedAt        time.Time
	SettledAt        *time.Time
}

// SettlementRepo defines the interface for settlement data access.
type SettlementRepo interface {
	CreateSettlementRecord(ctx context.Context, record *SettlementRecord) (*SettlementRecord, error)
	GetSettlementRecordByID(ctx context.Context, recordID string) (*SettlementRecord, error)
	ListSettlementRecords(ctx context.Context, merchantID uint64, status string, pageSize, pageNum uint32) ([]*SettlementRecord, uint64, error)
	UpdateSettlementRecordStatus(ctx context.Context, recordID string, newStatus string, settledAt *time.Time) error
}

// OrderClient defines the interface to interact with the Order Service (to get order details).
type OrderClient interface {
	// GetOrderDetails(ctx context.Context, orderID uint64) (*OrderInfo, error) // Assuming an OrderInfo struct exists in Order Service
}

// SettlementUsecase is the business logic for settlement.
type SettlementUsecase struct {
	repo        SettlementRepo
	orderClient OrderClient // Optional, for getting order details
}

// NewSettlementUsecase creates a new SettlementUsecase.
func NewSettlementUsecase(repo SettlementRepo, orderClient OrderClient) *SettlementUsecase {
	return &SettlementUsecase{repo: repo, orderClient: orderClient}
}

// ProcessOrderSettlement processes the settlement for a given order.
func (uc *SettlementUsecase) ProcessOrderSettlement(ctx context.Context, orderID, merchantID, totalAmount uint64) (*SettlementRecord, error) {
	// 1. Check if order is already settled (simplified)
	// In a real system, you'd query by orderID and check status.
	// For now, assume it's not settled if no record with this orderID exists.

	// 2. Calculate platform fee and settlement amount
	// Simplified: assume a fixed commission rate
	platformCommissionRate := 0.05 // 5%
	platformFee := uint64(float64(totalAmount) * platformCommissionRate)
	settlementAmount := totalAmount - platformFee

	// 3. Create a new settlement record
	recordID := uuid.New().String()
	record := &SettlementRecord{
		RecordID:         recordID,
		OrderID:          orderID,
		MerchantID:       merchantID,
		TotalAmount:      totalAmount,
		PlatformFee:      platformFee,
		SettlementAmount: settlementAmount,
		Status:           "PENDING", // Initial status
	}

	createdRecord, err := uc.repo.CreateSettlementRecord(ctx, record)
	if err != nil {
		return nil, fmt.Errorf("failed to create settlement record: %w", err)
	}

	// TODO: In a real system, trigger actual payment to merchant (e.g., via a payment service)
	// For now, simulate immediate completion
	settledAt := time.Now()
	err = uc.repo.UpdateSettlementRecordStatus(ctx, createdRecord.RecordID, "COMPLETED", &settledAt)
	if err != nil {
		// Log error, but record is already created.
		fmt.Printf("failed to update settlement record status to COMPLETED: %v\n", err)
	}
	createdRecord.Status = "COMPLETED"
	createdRecord.SettledAt = &settledAt

	return createdRecord, nil
}

// GetSettlementRecord retrieves a settlement record by its ID.
func (uc *SettlementUsecase) GetSettlementRecord(ctx context.Context, recordID string) (*SettlementRecord, error) {
	record, err := uc.repo.GetSettlementRecordByID(ctx, recordID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, ErrRecordNotFound
	}
	return record, nil
}

// ListSettlementRecords lists settlement records based on filters.
func (uc *SettlementUsecase) ListSettlementRecords(ctx context.Context, merchantID uint64, status string, pageSize, pageNum uint32) ([]*SettlementRecord, uint64, error) {
	return uc.repo.ListSettlementRecords(ctx, merchantID, status, pageSize, pageNum)
}
