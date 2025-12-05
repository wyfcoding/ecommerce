package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/go-api/dynamic_pricing/v1"           // 导入动态定价模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/dynamic_pricing/application"   // 导入动态定价模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/dynamic_pricing/domain/entity" // 导入动态定价模块的领域实体。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/emptypb"     // 导入空消息类型。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 DynamicPricingService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedDynamicPricingServiceServer                                    // 嵌入生成的UnimplementedDynamicPricingServiceServer，确保前向兼容性。
	app                                         *application.DynamicPricingService // 依赖DynamicPricing应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 DynamicPricing gRPC 服务端实例。
func NewServer(app *application.DynamicPricingService) *Server {
	return &Server{app: app}
}

// CalculatePrice 处理计算动态价格的gRPC请求。
// req: 包含SKU ID、基础价格、库存、需求和竞品价格等信息的请求体。
// 返回计算出的动态价格响应和可能发生的gRPC错误。
func (s *Server) CalculatePrice(ctx context.Context, req *pb.CalculatePriceRequest) (*pb.CalculatePriceResponse, error) {
	// 将protobuf请求转换为应用服务层所需的 PricingRequest 实体。
	pricingReq := &entity.PricingRequest{
		SKUID:              req.SkuId,
		BasePrice:          req.BasePrice,
		CurrentStock:       req.CurrentStock,
		TotalStock:         req.TotalStock,
		DailyDemand:        req.DailyDemand,
		AverageDailyDemand: req.AverageDailyDemand,
		CompetitorPrice:    req.CompetitorPrice,
		// UserLevel 字段Proto未提供，应用服务层当前CalculatePrice也未用到。
	}

	// 调用应用服务层计算动态价格。
	price, err := s.app.CalculatePrice(ctx, pricingReq)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to calculate dynamic price: %v", err))
	}

	// 将领域实体转换为protobuf响应格式。
	return &pb.CalculatePriceResponse{
		Price: convertDynamicPriceToProto(price),
	}, nil
}

// GetLatestPrice 处理获取最新动态价格的gRPC请求。
// req: 包含SKU ID的请求体。
// 返回最新动态价格响应和可能发生的gRPC错误。
func (s *Server) GetLatestPrice(ctx context.Context, req *pb.GetLatestPriceRequest) (*pb.GetLatestPriceResponse, error) {
	price, err := s.app.GetLatestPrice(ctx, req.SkuId)
	if err != nil {
		// 如果未找到价格，返回NotFound状态码。
		return nil, status.Error(codes.NotFound, fmt.Sprintf("latest price not found for sku %d: %v", req.SkuId, err))
	}
	return &pb.GetLatestPriceResponse{
		Price: convertDynamicPriceToProto(price),
	}, nil
}

// SaveStrategy 处理保存（创建或更新）定价策略的gRPC请求。
// req: 包含定价策略详细信息的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) SaveStrategy(ctx context.Context, req *pb.SaveStrategyRequest) (*emptypb.Empty, error) {
	// 将protobuf请求转换为领域实体所需的 PricingStrategy 实体。
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
	// 假设ID由数据库处理，或者在更新现有策略时传入。
	if req.Strategy.Id > 0 {
		strategy.ID = uint(req.Strategy.Id)
	}

	// 调用应用服务层保存定价策略。
	if err := s.app.SaveStrategy(ctx, strategy); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to save pricing strategy: %v", err))
	}

	return &emptypb.Empty{}, nil
}

// ListStrategies 处理列出定价策略的gRPC请求。
// req: 包含分页参数的请求体。
// 返回定价策略列表响应和可能发生的gRPC错误。
func (s *Server) ListStrategies(ctx context.Context, req *pb.ListStrategiesRequest) (*pb.ListStrategiesResponse, error) {
	// 获取分页参数。
	page := int(req.PageNum)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 调用应用服务层获取定价策略列表。
	strategies, total, err := s.app.ListStrategies(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list pricing strategies: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
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
		TotalCount: uint64(total), // 总记录数。
	}, nil
}

// convertDynamicPriceToProto 是一个辅助函数，将领域层的 DynamicPrice 实体转换为 protobuf 的 DynamicPrice 消息。
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
		// TimeFactor, UserFactor 实体中有，但Proto中没有映射。
		EffectiveTime: timestamppb.New(p.EffectiveTime),
		ExpiryTime:    timestamppb.New(p.ExpiryTime),
	}
}
