package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/go-api/pricing/v1"              // 导入定价模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/pricing/application"   // 导入定价模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/pricing/domain/entity" // 导入定价模块的领域实体。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/emptypb"     // 导入空消息类型。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 PricingService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedPricingServiceServer                             // 嵌入生成的UnimplementedPricingServiceServer，确保前向兼容性。
	app                                  *application.PricingService // 依赖Pricing应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 Pricing gRPC 服务端实例。
func NewServer(app *application.PricingService) *Server {
	return &Server{app: app}
}

// CreateRule 处理创建定价规则的gRPC请求。
// req: 包含规则名称、商品/SKU ID、策略、价格范围、调整率、启用状态和时间范围的请求体。
// 返回创建成功的规则响应和可能发生的gRPC错误。
func (s *Server) CreateRule(ctx context.Context, req *pb.CreateRuleRequest) (*pb.CreateRuleResponse, error) {
	// 将protobuf请求转换为领域实体所需的 PricingRule 实体。
	rule := &entity.PricingRule{
		Name:       req.Name,
		ProductID:  req.ProductId,
		SkuID:      req.SkuId,
		Strategy:   entity.PricingStrategy(req.Strategy), // 直接转换策略类型。
		BasePrice:  req.BasePrice,
		MinPrice:   req.MinPrice,
		MaxPrice:   req.MaxPrice,
		AdjustRate: req.AdjustRate,
		Enabled:    req.Enabled,
		StartTime:  req.StartTime.AsTime(),
		EndTime:    req.EndTime.AsTime(),
	}

	// 调用应用服务层创建规则。
	if err := s.app.CreateRule(ctx, rule); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create rule: %v", err))
	}

	// 将领域实体转换为protobuf响应格式。
	return &pb.CreateRuleResponse{
		Rule: convertRuleToProto(rule),
	}, nil
}

// ListRules 处理列出定价规则的gRPC请求。
// req: 包含商品ID过滤和分页参数的请求体。
// 返回定价规则列表响应和可能发生的gRPC错误。
func (s *Server) ListRules(ctx context.Context, req *pb.ListRulesRequest) (*pb.ListRulesResponse, error) {
	// 获取分页参数。
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 调用应用服务层获取规则列表。
	rules, total, err := s.app.ListRules(ctx, req.ProductId, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list rules: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbRules := make([]*pb.PricingRule, len(rules))
	for i, r := range rules {
		pbRules[i] = convertRuleToProto(r)
	}

	return &pb.ListRulesResponse{
		Rules:      pbRules,
		TotalCount: total, // 总记录数。
	}, nil
}

// CalculatePrice 处理计算价格的gRPC请求。
// req: 包含商品ID、SKU ID、需求和竞争参数的请求体。
// 返回计算后的价格响应和可能发生的gRPC错误。
func (s *Server) CalculatePrice(ctx context.Context, req *pb.CalculatePriceRequest) (*pb.CalculatePriceResponse, error) {
	price, err := s.app.CalculatePrice(ctx, req.ProductId, req.SkuId, req.Demand, req.Competition)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to calculate price: %v", err))
	}

	return &pb.CalculatePriceResponse{
		Price: price, // 返回计算后的价格。
	}, nil
}

// RecordHistory 处理记录价格历史的gRPC请求。
// req: 包含商品ID、SKU ID、新旧价格和原因的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) RecordHistory(ctx context.Context, req *pb.RecordHistoryRequest) (*emptypb.Empty, error) {
	if err := s.app.RecordHistory(ctx, req.ProductId, req.SkuId, req.Price, req.OldPrice, req.Reason); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to record price history: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListHistory 处理列出价格历史的gRPC请求。
// req: 包含商品ID、SKU ID过滤和分页参数的请求体。
// 返回价格历史列表响应和可能发生的gRPC错误。
func (s *Server) ListHistory(ctx context.Context, req *pb.ListHistoryRequest) (*pb.ListHistoryResponse, error) {
	// 获取分页参数。
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 调用应用服务层获取价格历史列表。
	history, total, err := s.app.ListHistory(ctx, req.ProductId, req.SkuId, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list history: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbHistory := make([]*pb.PriceHistory, len(history))
	for i, h := range history {
		pbHistory[i] = convertHistoryToProto(h)
	}

	return &pb.ListHistoryResponse{
		History:    pbHistory,
		TotalCount: total, // 总记录数。
	}, nil
}

// convertRuleToProto 是一个辅助函数，将领域层的 PricingRule 实体转换为 protobuf 的 PricingRule 消息。
func convertRuleToProto(r *entity.PricingRule) *pb.PricingRule {
	if r == nil {
		return nil
	}
	return &pb.PricingRule{
		Id:         uint64(r.ID),                 // 规则ID。
		Name:       r.Name,                       // 名称。
		ProductId:  r.ProductID,                  // 商品ID。
		SkuId:      r.SkuID,                      // SKU ID。
		Strategy:   string(r.Strategy),           // 策略。
		BasePrice:  r.BasePrice,                  // 基础价格。
		MinPrice:   r.MinPrice,                   // 最低价格。
		MaxPrice:   r.MaxPrice,                   // 最高价格。
		AdjustRate: r.AdjustRate,                 // 调整率。
		Enabled:    r.Enabled,                    // 是否启用。
		StartTime:  timestamppb.New(r.StartTime), // 开始时间。
		EndTime:    timestamppb.New(r.EndTime),   // 结束时间。
		CreatedAt:  timestamppb.New(r.CreatedAt), // 创建时间。
		UpdatedAt:  timestamppb.New(r.UpdatedAt), // 更新时间。
	}
}

// convertHistoryToProto 是一个辅助函数，将领域层的 PriceHistory 实体转换为 protobuf 的 PriceHistory 消息。
func convertHistoryToProto(h *entity.PriceHistory) *pb.PriceHistory {
	if h == nil {
		return nil
	}
	return &pb.PriceHistory{
		Id:         uint64(h.ID),                 // ID。
		ProductId:  h.ProductID,                  // 商品ID。
		SkuId:      h.SkuID,                      // SKU ID。
		Price:      h.Price,                      // 价格。
		OldPrice:   h.OldPrice,                   // 原价格。
		ChangeRate: h.ChangeRate,                 // 变动率。
		Reason:     h.Reason,                     // 原因。
		CreatedAt:  timestamppb.New(h.CreatedAt), // 创建时间。
		UpdatedAt:  timestamppb.New(h.UpdatedAt), // 更新时间。
	}
}
