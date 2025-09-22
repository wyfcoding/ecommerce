package service

import (
	"context"
	"errors"
	"time"

	v1 "ecommerce/api/sales_forecasting/v1"
	"ecommerce/internal/sales_forecasting/biz"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// SalesForecastingService is the gRPC service implementation for sales forecasting.
type SalesForecastingService struct {
	v1.UnimplementedSalesForecastingServiceServer
	uc *biz.SalesForecastingUsecase
}

// NewSalesForecastingService creates a new SalesForecastingService.
func NewSalesForecastingService(uc *biz.SalesForecastingUsecase) *SalesForecastingService {
	return &SalesForecastingService{uc: uc}
}

// bizForecastResultToProto converts biz.ForecastResult to v1.ForecastResult.
func bizForecastResultToProto(result *biz.ForecastResult) *v1.ForecastResult {
	if result == nil {
		return nil
	}
	protoResult := &v1.ForecastResult{
		ProductId:               result.ProductID,
		ForecastDate:            timestamppb.New(result.ForecastDate),
		PredictedSalesQuantity:  result.PredictedSalesQuantity,
		ConfidenceIntervalLower: result.ConfidenceIntervalLower,
		ConfidenceIntervalUpper: result.ConfidenceIntervalUpper,
	}
	return protoResult
}

// GetProductSalesForecast implements the GetProductSalesForecast RPC.
func (s *SalesForecastingService) GetProductSalesForecast(ctx context.Context, req *v1.GetProductSalesForecastRequest) (*v1.GetProductSalesForecastResponse, error) {
	if req.ProductId == 0 || req.ForecastDays == 0 {
		return nil, status.Error(codes.InvalidArgument, "product_id and forecast_days are required")
	}

	forecasts, err := s.uc.GetProductSalesForecast(ctx, req.ProductId, req.ForecastDays)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get product sales forecast: %v", err)
	}

	protoForecasts := make([]*v1.ForecastResult, len(forecasts))
	for i, f := range forecasts {
		protoForecasts[i] = bizForecastResultToProto(f)
	}

	return &v1.GetProductSalesForecastResponse{Forecasts: protoForecasts}, nil
}

// TrainSalesForecastModel implements the TrainSalesForecastModel RPC.
func (s *SalesForecastingService) TrainSalesForecastModel(ctx context.Context, req *v1.TrainSalesForecastModelRequest) (*v1.TrainSalesForecastModelResponse, error) {
	if req.ModelName == "" || req.DataSource == "" {
		return nil, status.Error(codes.InvalidArgument, "model_name and data_source are required")
	}

	parameters := make(map[string]string)
	for k, v := range req.Parameters {
		parameters[k] = v
	}

	modelID, statusStr, err := s.uc.TrainSalesForecastModel(ctx, req.ModelName, req.DataSource, parameters)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to train sales forecast model: %v", err)
	}

	return &v1.TrainSalesForecastModelResponse{
		ModelId: modelID,
		Status:  statusStr,
	}, nil
}
