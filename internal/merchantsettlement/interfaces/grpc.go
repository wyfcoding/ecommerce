package interfaces

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/merchantsettlement/application"
	"github.com/wyfcoding/ecommerce/internal/merchantsettlement/domain"
)

type MerchantSettlementGRPCServer struct {
	app *application.MerchantSettlementService
}

func NewMerchantSettlementGRPCServer(app *application.MerchantSettlementService) *MerchantSettlementGRPCServer {
	return &MerchantSettlementGRPCServer{app: app}
}

func (s *MerchantSettlementGRPCServer) CreateSettlement(ctx context.Context, merchantID uint64, cycle, periodStart, periodEnd string) (*application.SettlementDTO, error) {
	return s.app.CreateSettlement(ctx, merchantID, domain.SettlementCycle(cycle), periodStart, periodEnd)
}

func (s *MerchantSettlementGRPCServer) CalculateSettlement(ctx context.Context, settlementID string) error {
	return s.app.CalculateSettlement(ctx, settlementID)
}

func (s *MerchantSettlementGRPCServer) ApproveSettlement(ctx context.Context, settlementID string, approvedBy uint64) error {
	return s.app.ApproveSettlement(ctx, settlementID, approvedBy)
}

func (s *MerchantSettlementGRPCServer) RejectSettlement(ctx context.Context, settlementID, reason string) error {
	return s.app.RejectSettlement(ctx, settlementID, reason)
}

func (s *MerchantSettlementGRPCServer) PaySettlement(ctx context.Context, settlementID string, bankAccountID uint64) error {
	return s.app.PaySettlement(ctx, settlementID, bankAccountID)
}

func (s *MerchantSettlementGRPCServer) GetSettlement(ctx context.Context, settlementID string) (*application.SettlementDTO, error) {
	return s.app.GetSettlement(ctx, settlementID)
}

func (s *MerchantSettlementGRPCServer) ListSettlements(ctx context.Context, merchantID uint64, status string, page, pageSize int) (*application.ListSettlementsResult, error) {
	return s.app.ListSettlements(ctx, merchantID, domain.SettlementStatus(status), page, pageSize)
}

func (s *MerchantSettlementGRPCServer) GetSettlementDetails(ctx context.Context, settlementID string) ([]*application.SettlementDetailDTO, error) {
	return s.app.GetSettlementDetails(ctx, settlementID)
}
