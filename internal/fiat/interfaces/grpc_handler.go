package interfaces

import (
	"context"

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
	cmd := application.ExchangeCommand{
		FromCurrency: req.FromCurrency,
		ToCurrency:   req.ToCurrency,
		Amount:       req.Amount,
	}

	amount, rate, err := h.appService.Exchange(ctx, cmd)
	if err != nil {
		return nil, err
	}

	return &pb.ExchangeResponse{
		ExchangedAmount: amount,
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
