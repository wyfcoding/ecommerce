package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/goapi/dynamicpricing/v1"
	"github.com/wyfcoding/ecommerce/internal/dynamicpricing/application"
	"github.com/wyfcoding/ecommerce/internal/dynamicpricing/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server 结构体实现了 DynamicPricingService 的 gRPC 服务端接口。
type Server struct {
	pb.UnimplementedDynamicPricingServer
	app *application.DynamicPricingService
}

// NewServer 创建并返回一个新的 DynamicPricing gRPC 服务端实例。
func NewServer(app *application.DynamicPricingService) *Server {
	return &Server{app: app}
}

// CalculatePrice 处理计算动态价格的gRPC请求。
func (s *Server) CalculatePrice(ctx context.Context, req *pb.CalculatePriceRequest) (*pb.CalculatePriceResponse, error) {
	pricingReq := &domain.PricingRequest{
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
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to calculate dynamic price: %v", err))
	}

	return &pb.CalculatePriceResponse{
		Price: convertDynamicPriceToProto(price),
	}, nil
}

// GetLatestPrice 处理获取最新动态价格的gRPC请求。
func (s *Server) GetLatestPrice(ctx context.Context, req *pb.GetLatestPriceRequest) (*pb.GetLatestPriceResponse, error) {
	price, err := s.app.GetLatestPrice(ctx, req.SkuId)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("latest price not found for sku %d: %v", req.SkuId, err))
	}
	return &pb.GetLatestPriceResponse{
		Price: convertDynamicPriceToProto(price),
	}, nil
}

// SaveStrategy 处理保存（创建或更新）定价策略的gRPC请求。
func (s *Server) SaveStrategy(ctx context.Context, req *pb.SaveStrategyRequest) (*emptypb.Empty, error) {
	strategy := &domain.PricingStrategy{
		SKUID:                 req.Strategy.SkuId,
		StrategyType:          req.Strategy.StrategyType,
		MinPrice:              req.Strategy.MinPrice,
		MaxPrice:              req.Strategy.MaxPrice,
		InventoryThreshold:    req.Strategy.InventoryThreshold,
		DemandThreshold:       req.Strategy.DemandThreshold,
		CompetitorPriceOffset: req.Strategy.CompetitorPriceOffset,
		Enabled:               req.Strategy.Enabled,
	}
	if req.Strategy.Id > 0 {
		strategy.ID = uint(req.Strategy.Id)
	}

	if err := s.app.SaveStrategy(ctx, strategy); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to save pricing strategy: %v", err))
	}

	return &emptypb.Empty{}, nil
}

// ListStrategies 处理列出定价策略的gRPC请求。
func (s *Server) ListStrategies(ctx context.Context, req *pb.ListStrategiesRequest) (*pb.ListStrategiesResponse, error) {
	page := max(int(req.PageNum), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	strategies, total, err := s.app.ListStrategies(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list pricing strategies: %v", err))
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

func convertDynamicPriceToProto(p *domain.DynamicPrice) *pb.DynamicPrice {
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
