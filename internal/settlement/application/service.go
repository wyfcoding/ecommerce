package application

import (
	"context"
	"ecommerce/internal/settlement/domain/entity"
	"ecommerce/internal/settlement/domain/repository"
	"errors"
	"fmt"
	"time"

	"log/slog"
)

type SettlementService struct {
	repo   repository.SettlementRepository
	logger *slog.Logger
}

func NewSettlementService(repo repository.SettlementRepository, logger *slog.Logger) *SettlementService {
	return &SettlementService{
		repo:   repo,
		logger: logger,
	}
}

// CreateSettlement 创建结算单
func (s *SettlementService) CreateSettlement(ctx context.Context, merchantID uint64, cycle string, startDate, endDate time.Time) (*entity.Settlement, error) {
	settlementNo := fmt.Sprintf("S%d%d", merchantID, time.Now().UnixNano())

	settlement := &entity.Settlement{
		SettlementNo: settlementNo,
		MerchantID:   merchantID,
		Cycle:        entity.SettlementCycle(cycle),
		StartDate:    startDate,
		EndDate:      endDate,
		Status:       entity.SettlementStatusPending,
	}

	if err := s.repo.SaveSettlement(ctx, settlement); err != nil {
		return nil, err
	}
	return settlement, nil
}

// AddOrderToSettlement 添加订单到结算单
func (s *SettlementService) AddOrderToSettlement(ctx context.Context, settlementID uint64, orderID uint64, orderNo string, amount uint64) error {
	settlement, err := s.repo.GetSettlement(ctx, settlementID)
	if err != nil {
		return err
	}
	if settlement == nil {
		return errors.New("settlement not found")
	}

	if settlement.Status != entity.SettlementStatusPending {
		return errors.New("settlement is not pending")
	}

	// Get Merchant Account for fee rate
	account, err := s.repo.GetMerchantAccount(ctx, settlement.MerchantID)
	if err != nil {
		return err
	}
	feeRate := 0.0
	if account != nil {
		feeRate = account.FeeRate
	}

	platformFee := uint64(float64(amount) * feeRate / 100)
	settlementAmount := amount - platformFee

	detail := &entity.SettlementDetail{
		SettlementID:     settlementID,
		OrderID:          orderID,
		OrderNo:          orderNo,
		OrderAmount:      amount,
		PlatformFee:      platformFee,
		SettlementAmount: settlementAmount,
	}

	if err := s.repo.SaveSettlementDetail(ctx, detail); err != nil {
		return err
	}

	// Update Settlement
	settlement.OrderCount++
	settlement.TotalAmount += amount
	settlement.PlatformFee += platformFee
	settlement.SettlementAmount += settlementAmount

	return s.repo.SaveSettlement(ctx, settlement)
}

// ProcessSettlement 处理结算
func (s *SettlementService) ProcessSettlement(ctx context.Context, id uint64) error {
	settlement, err := s.repo.GetSettlement(ctx, id)
	if err != nil {
		return err
	}
	if settlement == nil {
		return errors.New("settlement not found")
	}

	if settlement.Status != entity.SettlementStatusPending {
		return errors.New("settlement is not pending")
	}

	settlement.Status = entity.SettlementStatusProcessing
	return s.repo.SaveSettlement(ctx, settlement)
}

// CompleteSettlement 完成结算
func (s *SettlementService) CompleteSettlement(ctx context.Context, id uint64) error {
	settlement, err := s.repo.GetSettlement(ctx, id)
	if err != nil {
		return err
	}
	if settlement == nil {
		return errors.New("settlement not found")
	}

	if settlement.Status != entity.SettlementStatusProcessing {
		return errors.New("settlement is not processing")
	}

	// Update Merchant Account
	account, err := s.repo.GetMerchantAccount(ctx, settlement.MerchantID)
	if err != nil {
		return err
	}
	if account == nil {
		account = &entity.MerchantAccount{
			MerchantID: settlement.MerchantID,
			FeeRate:    0.0, // Default
		}
	}

	account.Balance += settlement.SettlementAmount
	account.TotalIncome += settlement.SettlementAmount
	if err := s.repo.SaveMerchantAccount(ctx, account); err != nil {
		return err
	}

	now := time.Now()
	settlement.Status = entity.SettlementStatusCompleted
	settlement.SettledAt = &now
	return s.repo.SaveSettlement(ctx, settlement)
}

// GetMerchantAccount 获取商户账户
func (s *SettlementService) GetMerchantAccount(ctx context.Context, merchantID uint64) (*entity.MerchantAccount, error) {
	return s.repo.GetMerchantAccount(ctx, merchantID)
}

// ListSettlements 列表
func (s *SettlementService) ListSettlements(ctx context.Context, merchantID uint64, status *int, page, pageSize int) ([]*entity.Settlement, int64, error) {
	offset := (page - 1) * pageSize
	var st *entity.SettlementStatus
	if status != nil {
		s := entity.SettlementStatus(*status)
		st = &s
	}
	return s.repo.ListSettlements(ctx, merchantID, st, offset, pageSize)
}
