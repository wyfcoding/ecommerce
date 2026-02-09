package application

import (
	"context"
	"log/slog"

	"github.com/wyfcoding/ecommerce/internal/fiat/domain"
)

type ExchangeCommand struct {
	FromCurrency string
	ToCurrency   string
	Amount       int64
}

type FiatApplicationService struct {
	fiatService *domain.FiatService
	logger      *slog.Logger
}

func NewFiatApplicationService(fiatService *domain.FiatService, logger *slog.Logger) *FiatApplicationService {
	return &FiatApplicationService{
		fiatService: fiatService,
		logger:      logger,
	}
}

func (s *FiatApplicationService) Exchange(ctx context.Context, cmd ExchangeCommand) (int64, float64, error) {
	s.logger.Info("processing fiat exchange", "from", cmd.FromCurrency, "to", cmd.ToCurrency, "amount", cmd.Amount)
	return s.fiatService.Exchange(ctx, cmd.FromCurrency, cmd.ToCurrency, cmd.Amount)
}

func (s *FiatApplicationService) GetRate(ctx context.Context, from, to string) (float64, error) {
	return s.fiatService.GetRate(ctx, from, to)
}
