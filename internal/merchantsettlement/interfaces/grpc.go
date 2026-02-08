package interfaces

import (
	"context"

	pb "github.com/wyfcoding/ecommerce/goapi/merchantsettlement/v1"
	"github.com/wyfcoding/ecommerce/internal/merchantsettlement/application"
)

type MerchantSettlementHandler struct {
	pb.UnimplementedMerchantSettlementServiceServer
	app *application.MerchantSettlementService
}

func NewMerchantSettlementHandler(app *application.MerchantSettlementService) *MerchantSettlementHandler {
	return &MerchantSettlementHandler{app: app}
}

func (h *MerchantSettlementHandler) GenerateSettlement(ctx context.Context, req *pb.GenerateSettlementRequest) (*pb.GenerateSettlementResponse, error) {
	return h.app.GenerateSettlement(ctx, req)
}

func (h *MerchantSettlementHandler) GetSettlement(ctx context.Context, req *pb.GetSettlementRequest) (*pb.GetSettlementResponse, error) {
	return h.app.GetSettlement(ctx, req.SettlementId)
}

func (h *MerchantSettlementHandler) ListSettlements(ctx context.Context, req *pb.ListSettlementsRequest) (*pb.ListSettlementsResponse, error) {
	return h.app.ListSettlements(ctx, req.MerchantId, req.Status)
}

func (h *MerchantSettlementHandler) MarkAsPaid(ctx context.Context, req *pb.MarkAsPaidRequest) (*pb.MarkAsPaidResponse, error) {
	return h.app.MarkAsPaid(ctx, req.SettlementId, req.TransactionRef)
}
