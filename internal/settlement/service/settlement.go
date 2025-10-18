package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ecommerce/internal/settlement/client"
	"ecommerce/internal/settlement/model"
	"ecommerce/internal/settlement/repository"
	"github.com/google/uuid"
)

var (
	ErrOrderAlreadySettled = errors.New("order already settled")
	ErrRecordNotFound      = errors.New("settlement record not found")
)

// SettlementService is the business logic for settlement.
type SettlementService struct {
	repo        repository.SettlementRepo
	orderClient client.OrderClient // Optional, for getting order details
}

// NewSettlementService creates a new SettlementService.
func NewSettlementService(repo repository.SettlementRepo, orderClient client.OrderClient) *SettlementService {
	return &SettlementService{repo: repo, orderClient: orderClient}
}

// ProcessOrderSettlement processes the settlement for a given order.
func (s *SettlementService) ProcessOrderSettlement(ctx context.Context, orderID, merchantID, totalAmount uint64) (*model.SettlementRecord, error) {
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
	record := &model.SettlementRecord{
		RecordID:         recordID,
		OrderID:          orderID,
		MerchantID:       merchantID,
		TotalAmount:      totalAmount,
		PlatformFee:      platformFee,
		SettlementAmount: settlementAmount,
		Status:           "PENDING", // Initial status
		CreatedAt:        time.Now(),
	}

	createdRecord, err := s.repo.CreateSettlementRecord(ctx, record)
	if err != nil {
		return nil, fmt.Errorf("failed to create settlement record: %w", err)
	}

	// TODO: In a real system, trigger actual payment to merchant (e.g., via a payment service)
	// For now, simulate immediate completion
	settledAt := time.Now()
	err = s.repo.UpdateSettlementRecordStatus(ctx, createdRecord.RecordID, "COMPLETED", &settledAt)
	if err != nil {
		// Log error, but record is already created.
		fmt.Printf("failed to update settlement record status to COMPLETED: %v\n", err)
	}
	createdRecord.Status = "COMPLETED"
	createdRecord.SettledAt = &settledAt

	return createdRecord, nil
}

// GetSettlementRecord retrieves a settlement record by its ID.
func (s *SettlementService) GetSettlementRecord(ctx context.Context, recordID string) (*model.SettlementRecord, error) {
	record, err := s.repo.GetSettlementRecordByID(ctx, recordID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, ErrRecordNotFound
	}
	return record, nil
}

// ListSettlementRecords lists settlement records based on filters.
func (s *SettlementService) ListSettlementRecords(ctx context.Context, merchantID uint64, status string, pageSize, pageNum uint32) ([]*model.SettlementRecord, uint64, error) {
	return s.repo.ListSettlementRecords(ctx, merchantID, status, pageSize, pageNum)
}
