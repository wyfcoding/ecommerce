package grpc

import (
	"context"
	pb "ecommerce/api/dynamic_pricing/v1"
	"ecommerce/internal/dynamic_pricing/application"
	"ecommerce/internal/dynamic_pricing/domain/entity"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedDynamicPricingServiceServer
	app *application.DynamicPricingService
}

func NewServer(app *application.DynamicPricingService) *Server {
	return &Server{app: app}
}

func (s *Server) CalculatePrice(ctx context.Context, req *pb.CalculatePriceRequest) (*pb.CalculatePriceResponse, error) {
	pricingReq := &entity.PricingRequest{
		SKUID:              req.SkuId,
		BasePrice:          req.BasePrice,
		CurrentStock:       req.CurrentStock,
		TotalStock:         req.TotalStock,
		DailyDemand:        req.DailyDemand,
		AverageDailyDemand: req.AverageDailyDemand,
		CompetitorPrice:    req.CompetitorPrice,
	}

	price, err := s.app.CalculatePrice(ctx, pricingReq)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CalculatePriceResponse{
		Price: convertDynamicPriceToProto(price),
	}, nil
}

func (s *Server) GetLatestPrice(ctx context.Context, req *pb.GetLatestPriceRequest) (*pb.GetLatestPriceResponse, error) {
	price, err := s.app.GetLatestPrice(ctx, req.SkuId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &pb.GetLatestPriceResponse{
		Price: convertDynamicPriceToProto(price),
	}, nil
}

func (s *Server) SaveStrategy(ctx context.Context, req *pb.SaveStrategyRequest) (*emptypb.Empty, error) {
	strategy := &entity.PricingStrategy{
		SKUID:                 req.Strategy.SkuId,
		StrategyType:          req.Strategy.StrategyType,
		MinPrice:              req.Strategy.MinPrice,
		MaxPrice:              req.Strategy.MaxPrice,
		InventoryThreshold:    req.Strategy.InventoryThreshold,
		DemandThreshold:       req.Strategy.DemandThreshold,
		CompetitorPriceOffset: req.Strategy.CompetitorPriceOffset,
		Enabled:               req.Strategy.Enabled,
	}
	// Assuming ID is handled by DB or passed if updating
	if req.Strategy.Id > 0 {
		strategy.ID = uint(req.Strategy.Id)
	}

	if err := s.app.SaveStrategy(ctx, strategy); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) ListStrategies(ctx context.Context, req *pb.ListStrategiesRequest) (*pb.ListStrategiesResponse, error) {
	page := int(req.PageNum)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	strategies, total, err := s.app.ListStrategies(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbStrategies := make([]*pb.PricingStrategy, len(strategies))
	for i, st := range strategies {
		pbStrategies[i] = &pb.PricingStrategy{
			Id:                    uint64(st.ID),
			SkuId:                 st.SKUID,
			StrategyType:          st.StrategyType,
			MinPrice:              st.MinPrice,
			MaxPrice:              st.MaxPrice,
			InventoryThreshold:    st.InventoryThreshold,
			DemandThreshold:       st.DemandThreshold,
			CompetitorPriceOffset: st.CompetitorPriceOffset,
			Enabled:               st.Enabled,
		}
	}

	return &pb.ListStrategiesResponse{
		Strategies: pbStrategies,
		TotalCount: uint64(total),
	}, nil
}

func convertDynamicPriceToProto(p *entity.DynamicPrice) *pb.DynamicPrice {
	if p == nil {
		return nil
	}
	return &pb.DynamicPrice{
		SkuId:            p.SKUID,
		BasePrice:        p.BasePrice,
		FinalPrice:       p.FinalPrice,
		PriceAdjustment:  p.PriceAdjustment,
		InventoryFactor:  p.InventoryFactor,
		DemandFactor:     p.DemandFactor,
		CompetitorFactor: p.CompetitorFactor,
		EffectiveTime:    timestamppb.New(p.EffectiveTime),
		ExpiryTime:       timestamppb.New(p.ExpiryTime),
	}
}
