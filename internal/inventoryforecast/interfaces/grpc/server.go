package grpc

import (
	"context"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/goapi/inventoryforecast/v1"
	"github.com/wyfcoding/ecommerce/internal/inventoryforecast/application"
	"github.com/wyfcoding/ecommerce/internal/inventoryforecast/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server 结构体实现了 InventoryForecastService 的 gRPC 服务端接口。
type Server struct {
	pb.UnimplementedInventoryForecastServiceServer
	app *application.InventoryForecastService
}

// NewServer 创建并返回一个新的 InventoryForecast gRPC 服务端实例。
func NewServer(app *application.InventoryForecastService) *Server {
	return &Server{app: app}
}

// GenerateForecast 处理生成销售预测的gRPC请求。
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
func (s *Server) GetForecast(ctx context.Context, req *pb.GetForecastRequest) (*pb.GetForecastResponse, error) {
	forecast, err := s.app.GetForecast(ctx, req.SkuId)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("forecast not found for sku %d: %v", req.SkuId, err))
	}
	return &pb.GetForecastResponse{
		Forecast: convertForecastToProto(forecast),
	}, nil
}

// ListWarnings 处理列出库存预警的gRPC请求。
func (s *Server) ListWarnings(ctx context.Context, req *pb.ListWarningsRequest) (*pb.ListWarningsResponse, error) {
	page := max(int(req.PageNum), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	warnings, total, err := s.app.ListWarnings(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list warnings: %v", err))
	}

	pbWarnings := make([]*pb.InventoryWarning, len(warnings))
	for i, w := range warnings {
		pbWarnings[i] = convertWarningToProto(w)
	}

	return &pb.ListWarningsResponse{
		Warnings:   pbWarnings,
		TotalCount: uint64(total),
	}, nil
}

// ListSlowMovingItems 处理列出滞销品的gRPC请求。
func (s *Server) ListSlowMovingItems(ctx context.Context, req *pb.ListSlowMovingItemsRequest) (*pb.ListSlowMovingItemsResponse, error) {
	page := max(int(req.PageNum), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	items, total, err := s.app.ListSlowMovingItems(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list slow moving items: %v", err))
	}

	pbItems := make([]*pb.SlowMovingItem, len(items))
	for i, item := range items {
		pbItems[i] = convertSlowMovingItemToProto(item)
	}

	return &pb.ListSlowMovingItemsResponse{
		Items:      pbItems,
		TotalCount: uint64(total),
	}, nil
}

// ListReplenishmentSuggestions 处理列出补货建议的gRPC请求。
func (s *Server) ListReplenishmentSuggestions(ctx context.Context, req *pb.ListReplenishmentSuggestionsRequest) (*pb.ListReplenishmentSuggestionsResponse, error) {
	page := max(int(req.PageNum), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	suggestions, total, err := s.app.ListReplenishmentSuggestions(ctx, req.Priority, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list replenishment suggestions: %v", err))
	}

	pbSuggestions := make([]*pb.ReplenishmentSuggestion, len(suggestions))
	for i, sug := range suggestions {
		pbSuggestions[i] = convertSuggestionToProto(sug)
	}

	return &pb.ListReplenishmentSuggestionsResponse{
		Suggestions: pbSuggestions,
		TotalCount:  uint64(total),
	}, nil
}

// ListStockoutRisks 处理列出缺货风险的gRPC请求。
func (s *Server) ListStockoutRisks(ctx context.Context, req *pb.ListStockoutRisksRequest) (*pb.ListStockoutRisksResponse, error) {
	page := max(int(req.PageNum), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	risks, total, err := s.app.ListStockoutRisks(ctx, domain.StockoutRiskLevel(req.Level), page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list stockout risks: %v", err))
	}

	pbRisks := make([]*pb.StockoutRisk, len(risks))
	for i, risk := range risks {
		pbRisks[i] = convertRiskToProto(risk)
	}

	return &pb.ListStockoutRisksResponse{
		Risks:      pbRisks,
		TotalCount: uint64(total),
	}, nil
}

func convertForecastToProto(f *domain.SalesForecast) *pb.SalesForecast {
	if f == nil {
		return nil
	}
	predictions := make([]*pb.DailyForecast, len(f.Predictions))
	for i, p := range f.Predictions {
		predictions[i] = &pb.DailyForecast{
			Date:       timestamppb.New(p.Date),
			Quantity:   p.Quantity,
			Confidence: p.Confidence,
		}
	}
	return &pb.SalesForecast{
		Id:                uint64(f.ID),
		SkuId:             f.SKUID,
		AverageDailySales: f.AverageDailySales,
		TrendRate:         f.TrendRate,
		Predictions:       predictions,
		CreatedAt:         timestamppb.New(f.CreatedAt),
		UpdatedAt:         timestamppb.New(f.UpdatedAt),
	}
}

func convertWarningToProto(w *domain.InventoryWarning) *pb.InventoryWarning {
	if w == nil {
		return nil
	}
	return &pb.InventoryWarning{
		Id:                          uint64(w.ID),
		SkuId:                       w.SKUID,
		CurrentStock:                w.CurrentStock,
		WarningThreshold:            w.WarningThreshold,
		DaysUntilEmpty:              w.DaysUntilEmpty,
		EstimatedEmptyDate:          timestamppb.New(w.EstimatedEmptyDate),
		NeedReplenishment:           w.NeedReplenishment,
		RecommendedReplenishmentQty: w.RecommendedReplenishmentQty,
		CreatedAt:                   timestamppb.New(w.CreatedAt),
	}
}

func convertSlowMovingItemToProto(item *domain.SlowMovingItem) *pb.SlowMovingItem {
	if item == nil {
		return nil
	}
	return &pb.SlowMovingItem{
		Id:              uint64(item.ID),
		SkuId:           item.SKUID,
		ProductName:     item.ProductName,
		CurrentStock:    item.CurrentStock,
		DailySalesRate:  item.DailySalesRate,
		DaysInStock:     item.DaysInStock,
		TurnoverRate:    item.TurnoverRate,
		RecommendAction: item.RecommendAction,
		CreatedAt:       timestamppb.New(item.CreatedAt),
	}
}

func convertSuggestionToProto(s *domain.ReplenishmentSuggestion) *pb.ReplenishmentSuggestion {
	if s == nil {
		return nil
	}
	return &pb.ReplenishmentSuggestion{
		Id:            uint64(s.ID),
		SkuId:         s.SKUID,
		ProductName:   s.ProductName,
		CurrentStock:  s.CurrentStock,
		SuggestedQty:  s.SuggestedQty,
		Priority:      s.Priority,
		Reason:        s.Reason,
		EstimatedCost: s.EstimatedCost,
		CreatedAt:     timestamppb.New(s.CreatedAt),
	}
}

func convertRiskToProto(r *domain.StockoutRisk) *pb.StockoutRisk {
	if r == nil {
		return nil
	}
	return &pb.StockoutRisk{
		Id:                    uint64(r.ID),
		SkuId:                 r.SKUID,
		CurrentStock:          r.CurrentStock,
		DaysUntilStockout:     r.DaysUntilStockout,
		EstimatedStockoutDate: timestamppb.New(r.EstimatedStockoutDate),
		RiskLevel:             string(r.RiskLevel),
		CreatedAt:             timestamppb.New(r.CreatedAt),
	}
}
