package interfaces

import (
	"context"

	"github.com/shopspring/decimal"
	pb "github.com/wyfcoding/ecommerce/go-api/fiat/v1"
	"github.com/wyfcoding/ecommerce/internal/fiat/application"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type FiatHandler struct {
	pb.UnimplementedFiatServiceServer
	appService *application.FiatApplicationService
}

func NewFiatHandler(appService *application.FiatApplicationService) *FiatHandler {
	return &FiatHandler{
		appService: appService,
	}
}

func (h *FiatHandler) Exchange(ctx context.Context, req *pb.ExchangeRequest) (*pb.ExchangeResponse, error) {
	cmd := &application.ExchangeCommand{
		FromCurrency: req.FromCurrency,
		ToCurrency:   req.ToCurrency,
		Amount:       decimal.NewFromInt(req.Amount),
	}

	result, err := h.appService.Exchange(ctx, cmd)
	if err != nil {
		return nil, err
	}

	rate, _ := result.ExchangeRate.Float64()
	return &pb.ExchangeResponse{
		ExchangedAmount: result.ExchangedAmount.IntPart(),
		Rate:            rate,
	}, nil
}

func (h *FiatHandler) GetRate(ctx context.Context, req *pb.GetRateRequest) (*pb.GetRateResponse, error) {
	rate, err := h.appService.GetRate(ctx, req.FromCurrency, req.ToCurrency)
	if err != nil {
		return nil, err
	}

	return &pb.GetRateResponse{
		Rate:      rate,
		UpdatedAt: timestamppb.Now(),
	}, nil
}
