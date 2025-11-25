package grpc

import (
	"context"
	pb "github.com/wyfcoding/ecommerce/api/inventory_forecast/v1"
	"github.com/wyfcoding/ecommerce/internal/inventory_forecast/application"
	"github.com/wyfcoding/ecommerce/internal/inventory_forecast/domain/entity"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedInventoryForecastServiceServer
	app *application.InventoryForecastService
}

func NewServer(app *application.InventoryForecastService) *Server {
	return &Server{app: app}
}

func (s *Server) GenerateForecast(ctx context.Context, req *pb.GenerateForecastRequest) (*pb.GenerateForecastResponse, error) {
	forecast, err := s.app.GenerateForecast(ctx, req.SkuId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GenerateForecastResponse{
		Forecast: convertForecastToProto(forecast),
	}, nil
}

func (s *Server) GetForecast(ctx context.Context, req *pb.GetForecastRequest) (*pb.GetForecastResponse, error) {
	forecast, err := s.app.GetForecast(ctx, req.SkuId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &pb.GetForecastResponse{
		Forecast: convertForecastToProto(forecast),
	}, nil
}

func (s *Server) ListWarnings(ctx context.Context, req *pb.ListWarningsRequest) (*pb.ListWarningsResponse, error) {
	page := int(req.PageNum)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	warnings, total, err := s.app.ListWarnings(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
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

func (s *Server) ListSlowMovingItems(ctx context.Context, req *pb.ListSlowMovingItemsRequest) (*pb.ListSlowMovingItemsResponse, error) {
	page := int(req.PageNum)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	items, total, err := s.app.ListSlowMovingItems(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
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

func (s *Server) ListReplenishmentSuggestions(ctx context.Context, req *pb.ListReplenishmentSuggestionsRequest) (*pb.ListReplenishmentSuggestionsResponse, error) {
	page := int(req.PageNum)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	suggestions, total, err := s.app.ListReplenishmentSuggestions(ctx, req.Priority, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
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

func (s *Server) ListStockoutRisks(ctx context.Context, req *pb.ListStockoutRisksRequest) (*pb.ListStockoutRisksResponse, error) {
	page := int(req.PageNum)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	risks, total, err := s.app.ListStockoutRisks(ctx, entity.StockoutRiskLevel(req.Level), page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
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

func convertForecastToProto(f *entity.SalesForecast) *pb.SalesForecast {
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

func convertWarningToProto(w *entity.InventoryWarning) *pb.InventoryWarning {
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

func convertSlowMovingItemToProto(item *entity.SlowMovingItem) *pb.SlowMovingItem {
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

func convertSuggestionToProto(s *entity.ReplenishmentSuggestion) *pb.ReplenishmentSuggestion {
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

func convertRiskToProto(r *entity.StockoutRisk) *pb.StockoutRisk {
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
