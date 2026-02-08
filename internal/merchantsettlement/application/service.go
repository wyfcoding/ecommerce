package application

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	pb "github.com/wyfcoding/ecommerce/goapi/merchantsettlement/v1"
	"github.com/wyfcoding/ecommerce/internal/merchantsettlement/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type MerchantSettlementService struct {
	repo domain.SettlementRepository
}

func NewMerchantSettlementService(repo domain.SettlementRepository) *MerchantSettlementService {
	return &MerchantSettlementService{repo: repo}
}

func (s *MerchantSettlementService) GenerateSettlement(ctx context.Context, req *pb.GenerateSettlementRequest) (*pb.GenerateSettlementResponse, error) {
	// 简单模拟生成逻辑：生成一个随机金额的结算单
	settlementID := fmt.Sprintf("set_%d", time.Now().UnixNano())
	amount := decimal.NewFromFloat(100.0) // 演示数据

	settlement := &domain.Settlement{
		SettlementID: settlementID,
		MerchantID:   req.MerchantId,
		Amount:       amount,
		Status:       domain.StatusUnpaid,
		PeriodStart:  req.StartDate.AsTime(),
		PeriodEnd:    req.EndDate.AsTime(),
	}

	if err := s.repo.Save(ctx, settlement); err != nil {
		return nil, err
	}

	return &pb.GenerateSettlementResponse{
		SettlementId: settlementID,
		Amount:       amount.String(),
	}, nil
}

func (s *MerchantSettlementService) GetSettlement(ctx context.Context, id string) (*pb.GetSettlementResponse, error) {
	settlement, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if settlement == nil {
		return nil, fmt.Errorf("settlement %s not found", id)
	}

	return &pb.GetSettlementResponse{
		Settlement: mapToPb(settlement),
	}, nil
}

func (s *MerchantSettlementService) ListSettlements(ctx context.Context, merchantID, status string) (*pb.ListSettlementsResponse, error) {
	settlements, err := s.repo.ListByMerchant(ctx, merchantID, status)
	if err != nil {
		return nil, err
	}

	var pbSettlements []*pb.SettlementDetail
	for _, st := range settlements {
		pbSettlements = append(pbSettlements, mapToPb(st))
	}

	return &pb.ListSettlementsResponse{Settlements: pbSettlements}, nil
}

func (s *MerchantSettlementService) MarkAsPaid(ctx context.Context, id, txRef string) (*pb.MarkAsPaidResponse, error) {
	settlement, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if settlement == nil {
		return nil, fmt.Errorf("settlement %s not found", id)
	}

	settlement.Status = domain.StatusPaid
	settlement.TransactionRef = txRef

	if err := s.repo.Save(ctx, settlement); err != nil {
		return nil, err
	}

	return &pb.MarkAsPaidResponse{Success: true}, nil
}

func mapToPb(s *domain.Settlement) *pb.SettlementDetail {
	return &pb.SettlementDetail{
		SettlementId: s.SettlementID,
		MerchantId:   s.MerchantID,
		Amount:       s.Amount.String(),
		Status:       string(s.Status),
		PeriodStart:  timestamppb.New(s.PeriodStart),
		PeriodEnd:    timestamppb.New(s.PeriodEnd),
		CreatedAt:    timestamppb.New(s.CreatedAt),
	}
}
