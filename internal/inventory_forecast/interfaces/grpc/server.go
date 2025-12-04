package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/api/inventory_forecast/v1"              // 导入库存预测模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/inventory_forecast/application"   // 导入库存预测模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/inventory_forecast/domain/entity" // 导入库存预测模块的领域实体。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 InventoryForecastService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedInventoryForecastServiceServer                                       // 嵌入生成的UnimplementedInventoryForecastServiceServer，确保前向兼容性。
	app                                            *application.InventoryForecastService // 依赖InventoryForecast应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 InventoryForecast gRPC 服务端实例。
func NewServer(app *application.InventoryForecastService) *Server {
	return &Server{app: app}
}

// GenerateForecast 处理生成销售预测的gRPC请求。
// req: 包含SKU ID的请求体。
// 返回生成的销售预测响应和可能发生的gRPC错误。
func (s *Server) GenerateForecast(ctx context.Context, req *pb.GenerateForecastRequest) (*pb.GenerateForecastResponse, error) {
	forecast, err := s.app.GenerateForecast(ctx, req.SkuId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to generate forecast: %v", err))
	}

	return &pb.GenerateForecastResponse{
		Forecast: convertForecastToProto(forecast),
	}, nil
}

// GetForecast 处理获取销售预测的gRPC请求。
// req: 包含SKU ID的请求体。
// 返回销售预测响应和可能发生的gRPC错误。
func (s *Server) GetForecast(ctx context.Context, req *pb.GetForecastRequest) (*pb.GetForecastResponse, error) {
	forecast, err := s.app.GetForecast(ctx, req.SkuId)
	if err != nil {
		// 如果预测未找到，返回NotFound状态码。
		return nil, status.Error(codes.NotFound, fmt.Sprintf("forecast not found for sku %d: %v", req.SkuId, err))
	}
	return &pb.GetForecastResponse{
		Forecast: convertForecastToProto(forecast),
	}, nil
}

// ListWarnings 处理列出库存预警的gRPC请求。
// req: 包含分页参数的请求体。
// 返回库存预警列表响应和可能发生的gRPC错误。
func (s *Server) ListWarnings(ctx context.Context, req *pb.ListWarningsRequest) (*pb.ListWarningsResponse, error) {
	// 获取分页参数。
	page := int(req.PageNum)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 调用应用服务层获取库存预警列表。
	warnings, total, err := s.app.ListWarnings(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list warnings: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbWarnings := make([]*pb.InventoryWarning, len(warnings))
	for i, w := range warnings {
		pbWarnings[i] = convertWarningToProto(w)
	}

	return &pb.ListWarningsResponse{
		Warnings:   pbWarnings,
		TotalCount: uint64(total), // 总记录数。
	}, nil
}

// ListSlowMovingItems 处理列出滞销品的gRPC请求。
// req: 包含分页参数的请求体。
// 返回滞销品列表响应和可能发生的gRPC错误。
func (s *Server) ListSlowMovingItems(ctx context.Context, req *pb.ListSlowMovingItemsRequest) (*pb.ListSlowMovingItemsResponse, error) {
	// 获取分页参数。
	page := int(req.PageNum)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 调用应用服务层获取滞销品列表。
	items, total, err := s.app.ListSlowMovingItems(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list slow moving items: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbItems := make([]*pb.SlowMovingItem, len(items))
	for i, item := range items {
		pbItems[i] = convertSlowMovingItemToProto(item)
	}

	return &pb.ListSlowMovingItemsResponse{
		Items:      pbItems,
		TotalCount: uint64(total), // 总记录数。
	}, nil
}

// ListReplenishmentSuggestions 处理列出补货建议的gRPC请求。
// req: 包含优先级和分页参数的请求体。
// 返回补货建议列表响应和可能发生的gRPC错误。
func (s *Server) ListReplenishmentSuggestions(ctx context.Context, req *pb.ListReplenishmentSuggestionsRequest) (*pb.ListReplenishmentSuggestionsResponse, error) {
	// 获取分页参数。
	page := int(req.PageNum)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 调用应用服务层获取补货建议列表。
	suggestions, total, err := s.app.ListReplenishmentSuggestions(ctx, req.Priority, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list replenishment suggestions: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbSuggestions := make([]*pb.ReplenishmentSuggestion, len(suggestions))
	for i, sug := range suggestions {
		pbSuggestions[i] = convertSuggestionToProto(sug)
	}

	return &pb.ListReplenishmentSuggestionsResponse{
		Suggestions: pbSuggestions,
		TotalCount:  uint64(total), // 总记录数。
	}, nil
}

// ListStockoutRisks 处理列出缺货风险的gRPC请求。
// req: 包含风险等级和分页参数的请求体。
// 返回缺货风险列表响应和可能发生的gRPC错误。
func (s *Server) ListStockoutRisks(ctx context.Context, req *pb.ListStockoutRisksRequest) (*pb.ListStockoutRisksResponse, error) {
	// 获取分页参数。
	page := int(req.PageNum)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 调用应用服务层获取缺货风险列表。
	risks, total, err := s.app.ListStockoutRisks(ctx, entity.StockoutRiskLevel(req.Level), page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list stockout risks: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbRisks := make([]*pb.StockoutRisk, len(risks))
	for i, risk := range risks {
		pbRisks[i] = convertRiskToProto(risk)
	}

	return &pb.ListStockoutRisksResponse{
		Risks:      pbRisks,
		TotalCount: uint64(total), // 总记录数。
	}, nil
}

// convertForecastToProto 是一个辅助函数，将领域层的 SalesForecast 实体转换为 protobuf 的 SalesForecast 消息。
func convertForecastToProto(f *entity.SalesForecast) *pb.SalesForecast {
	if f == nil {
		return nil
	}
	predictions := make([]*pb.DailyForecast, len(f.Predictions))
	for i, p := range f.Predictions {
		predictions[i] = &pb.DailyForecast{
			Date:       timestamppb.New(p.Date), // 预测日期。
			Quantity:   p.Quantity,              // 预测销量。
			Confidence: p.Confidence,            // 置信度。
		}
	}
	return &pb.SalesForecast{
		Id:                uint64(f.ID),                 // 销售预测ID。
		SkuId:             f.SKUID,                      // SKU ID。
		AverageDailySales: f.AverageDailySales,          // 日均销量。
		TrendRate:         f.TrendRate,                  // 趋势率。
		Predictions:       predictions,                  // 每日预测详情。
		CreatedAt:         timestamppb.New(f.CreatedAt), // 创建时间。
		UpdatedAt:         timestamppb.New(f.UpdatedAt), // 更新时间。
	}
}

// convertWarningToProto 是一个辅助函数，将领域层的 InventoryWarning 实体转换为 protobuf 的 InventoryWarning 消息。
func convertWarningToProto(w *entity.InventoryWarning) *pb.InventoryWarning {
	if w == nil {
		return nil
	}
	return &pb.InventoryWarning{
		Id:                          uint64(w.ID),                          // 预警ID。
		SkuId:                       w.SKUID,                               // SKU ID。
		CurrentStock:                w.CurrentStock,                        // 当前库存。
		WarningThreshold:            w.WarningThreshold,                    // 预警阈值。
		DaysUntilEmpty:              w.DaysUntilEmpty,                      // 预计售罄天数。
		EstimatedEmptyDate:          timestamppb.New(w.EstimatedEmptyDate), // 预计售罄日期。
		NeedReplenishment:           w.NeedReplenishment,                   // 是否需要补货。
		RecommendedReplenishmentQty: w.RecommendedReplenishmentQty,         // 建议补货数量。
		CreatedAt:                   timestamppb.New(w.CreatedAt),          // 创建时间。
	}
}

// convertSlowMovingItemToProto 是一个辅助函数，将领域层的 SlowMovingItem 实体转换为 protobuf 的 SlowMovingItem 消息。
func convertSlowMovingItemToProto(item *entity.SlowMovingItem) *pb.SlowMovingItem {
	if item == nil {
		return nil
	}
	return &pb.SlowMovingItem{
		Id:              uint64(item.ID),                 // 滞销品ID。
		SkuId:           item.SKUID,                      // SKU ID。
		ProductName:     item.ProductName,                // 商品名称。
		CurrentStock:    item.CurrentStock,               // 当前库存。
		DailySalesRate:  item.DailySalesRate,             // 日均动销率。
		DaysInStock:     item.DaysInStock,                // 库龄。
		TurnoverRate:    item.TurnoverRate,               // 周转率。
		RecommendAction: item.RecommendAction,            // 建议措施。
		CreatedAt:       timestamppb.New(item.CreatedAt), // 创建时间。
	}
}

// convertSuggestionToProto 是一个辅助函数，将领域层的 ReplenishmentSuggestion 实体转换为 protobuf 的 ReplenishmentSuggestion 消息。
func convertSuggestionToProto(s *entity.ReplenishmentSuggestion) *pb.ReplenishmentSuggestion {
	if s == nil {
		return nil
	}
	return &pb.ReplenishmentSuggestion{
		Id:            uint64(s.ID),                 // 建议ID。
		SkuId:         s.SKUID,                      // SKU ID。
		ProductName:   s.ProductName,                // 商品名称。
		CurrentStock:  s.CurrentStock,               // 当前库存。
		SuggestedQty:  s.SuggestedQty,               // 建议补货数量。
		Priority:      s.Priority,                   // 优先级。
		Reason:        s.Reason,                     // 原因。
		EstimatedCost: s.EstimatedCost,              // 预计成本。
		CreatedAt:     timestamppb.New(s.CreatedAt), // 创建时间。
	}
}

// convertRiskToProto 是一个辅助函数，将领域层的 StockoutRisk 实体转换为 protobuf 的 StockoutRisk 消息。
func convertRiskToProto(r *entity.StockoutRisk) *pb.StockoutRisk {
	if r == nil {
		return nil
	}
	return &pb.StockoutRisk{
		Id:                    uint64(r.ID),                             // 风险ID。
		SkuId:                 r.SKUID,                                  // SKU ID。
		CurrentStock:          r.CurrentStock,                           // 当前库存。
		DaysUntilStockout:     r.DaysUntilStockout,                      // 预计缺货天数。
		EstimatedStockoutDate: timestamppb.New(r.EstimatedStockoutDate), // 预计缺货日期。
		RiskLevel:             string(r.RiskLevel),                      // 风险等级。
		CreatedAt:             timestamppb.New(r.CreatedAt),             // 创建时间。
	}
}
