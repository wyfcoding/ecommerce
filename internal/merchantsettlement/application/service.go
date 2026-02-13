package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/wyfcoding/ecommerce/internal/merchantsettlement/domain"
)

type MerchantSettlementService struct {
	settlementRepo    domain.SettlementRepository
	bankAccountRepo   domain.MerchantBankAccountRepository
	configRepo        domain.MerchantSettlementConfigRepository
	calculator        domain.SettlementCalculatorService
	paymentGateway    domain.PaymentGateway
}

func NewMerchantSettlementService(
	settlementRepo domain.SettlementRepository,
	bankAccountRepo domain.MerchantBankAccountRepository,
	configRepo domain.MerchantSettlementConfigRepository,
	calculator domain.SettlementCalculatorService,
	paymentGateway domain.PaymentGateway,
) *MerchantSettlementService {
	return &MerchantSettlementService{
		settlementRepo:  settlementRepo,
		bankAccountRepo: bankAccountRepo,
		configRepo:      configRepo,
		calculator:      calculator,
		paymentGateway:  paymentGateway,
	}
}

func (s *MerchantSettlementService) CreateSettlement(ctx context.Context, merchantID uint64, cycle domain.SettlementCycle, periodStart, periodEnd string) (*SettlementDTO, error) {
	handler := NewCreateSettlementHandler(s.settlementRepo, s.configRepo)
	cmd := &CreateSettlementCommand{
		MerchantID: merchantID,
		Cycle:      cycle,
	}
	settlement, err := handler.Handle(ctx, cmd)
	if err != nil {
		return nil, err
	}
	return toSettlementDTO(settlement), nil
}

func (s *MerchantSettlementService) CalculateSettlement(ctx context.Context, settlementID string) error {
	handler := NewCalculateSettlementHandler(s.settlementRepo, s.calculator, s.configRepo)
	return handler.Handle(ctx, &CalculateSettlementCommand{SettlementID: settlementID})
}

func (s *MerchantSettlementService) ApproveSettlement(ctx context.Context, settlementID string, approvedBy uint64) error {
	handler := NewApproveSettlementHandler(s.settlementRepo)
	return handler.Handle(ctx, &ApproveSettlementCommand{
		SettlementID: settlementID,
		ApprovedBy:   approvedBy,
	})
}

func (s *MerchantSettlementService) RejectSettlement(ctx context.Context, settlementID, reason string) error {
	handler := NewRejectSettlementHandler(s.settlementRepo)
	return handler.Handle(ctx, &RejectSettlementCommand{
		SettlementID: settlementID,
		Reason:       reason,
	})
}

func (s *MerchantSettlementService) PaySettlement(ctx context.Context, settlementID string, bankAccountID uint64) error {
	handler := NewPaySettlementHandler(s.settlementRepo, s.bankAccountRepo, s.paymentGateway)
	return handler.Handle(ctx, &PaySettlementCommand{
		SettlementID:  settlementID,
		BankAccountID: bankAccountID,
	})
}

func (s *MerchantSettlementService) AdjustSettlement(ctx context.Context, settlementID string, amount decimal.Decimal, reason string) error {
	handler := NewAdjustSettlementHandler(s.settlementRepo)
	return handler.Handle(ctx, &AdjustSettlementCommand{
		SettlementID:     settlementID,
		AdjustmentAmount: amount,
		Reason:           reason,
	})
}

func (s *MerchantSettlementService) CancelSettlement(ctx context.Context, settlementID, reason string) error {
	handler := NewCancelSettlementHandler(s.settlementRepo)
	return handler.Handle(ctx, &CancelSettlementCommand{
		SettlementID: settlementID,
		Reason:       reason,
	})
}

func (s *MerchantSettlementService) GetSettlement(ctx context.Context, settlementID string) (*SettlementDTO, error) {
	handler := NewGetSettlementHandler(s.settlementRepo)
	return handler.Handle(ctx, &GetSettlementQuery{SettlementID: settlementID})
}

func (s *MerchantSettlementService) ListSettlements(ctx context.Context, merchantID uint64, status domain.SettlementStatus, page, pageSize int) (*ListSettlementsResult, error) {
	handler := NewListSettlementsHandler(s.settlementRepo)
	return handler.Handle(ctx, &ListSettlementsQuery{
		MerchantID: merchantID,
		Status:     status,
		Page:       page,
		PageSize:   pageSize,
	})
}

func (s *MerchantSettlementService) GetSettlementDetails(ctx context.Context, settlementID string) ([]*SettlementDetailDTO, error) {
	handler := NewGetSettlementDetailsHandler(s.settlementRepo)
	return handler.Handle(ctx, &GetSettlementDetailsQuery{SettlementID: settlementID})
}

func (s *MerchantSettlementService) AddBankAccount(ctx context.Context, merchantID uint64, bankName, bankCode, accountName, accountNo, branchName string, isDefault bool) (*BankAccountDTO, error) {
	account := &domain.MerchantBankAccount{
		ID:          uint64(uuid.New().ID()),
		MerchantID:  merchantID,
		BankName:    bankName,
		BankCode:    bankCode,
		AccountName: accountName,
		AccountNo:   accountNo,
		BranchName:  branchName,
		IsDefault:   isDefault,
		Status:      domain.AccountStatusActive,
	}
	if err := s.bankAccountRepo.Save(ctx, account); err != nil {
		return nil, err
	}
	return toBankAccountDTO(account), nil
}

func (s *MerchantSettlementService) ListBankAccounts(ctx context.Context, merchantID uint64) ([]*BankAccountDTO, error) {
	handler := NewListBankAccountsHandler(s.bankAccountRepo)
	return handler.Handle(ctx, &ListBankAccountsQuery{MerchantID: merchantID})
}

func (s *MerchantSettlementService) SetDefaultBankAccount(ctx context.Context, merchantID, accountID uint64) error {
	accounts, err := s.bankAccountRepo.GetByMerchantID(ctx, merchantID)
	if err != nil {
		return err
	}
	for _, a := range accounts {
		if a.ID == accountID {
			a.IsDefault = true
		} else {
			a.IsDefault = false
		}
		if err := s.bankAccountRepo.Update(ctx, a); err != nil {
			return err
		}
	}
	return nil
}

func (s *MerchantSettlementService) DeleteBankAccount(ctx context.Context, accountID uint64) error {
	return s.bankAccountRepo.Delete(ctx, accountID)
}

func (s *MerchantSettlementService) GetSettlementConfig(ctx context.Context, merchantID uint64) (*SettlementConfigDTO, error) {
	handler := NewGetSettlementConfigHandler(s.configRepo)
	return handler.Handle(ctx, &GetSettlementConfigQuery{MerchantID: merchantID})
}

func (s *MerchantSettlementService) UpdateSettlementConfig(ctx context.Context, config *domain.MerchantSettlementConfig) error {
	existing, err := s.configRepo.GetByMerchantID(ctx, config.MerchantID)
	if err != nil {
		return err
	}
	if existing == nil {
		return s.configRepo.Save(ctx, config)
	}
	return s.configRepo.Update(ctx, config)
}
