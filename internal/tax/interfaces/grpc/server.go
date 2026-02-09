package grpc

import (
	"context"

	v1 "github.com/wyfcoding/ecommerce/go-api/tax/v1"
	"github.com/wyfcoding/ecommerce/internal/tax/application"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	v1.UnimplementedTaxServiceServer
	app *application.TaxService
}

func NewServer(app *application.TaxService) *Server {
	return &Server{app: app}
}

func (s *Server) CalculateTax(ctx context.Context, req *v1.CalculateTaxRequest) (*v1.CalculateTaxResponse, error) {
	// Map request to domain params
	// Simplified: utilizing the first item's category for now or doing per-item calculation?
	// The application service currently supports single category calculation.
	// For production, we should loop items.

	var totalTax int64
	var details []*v1.TaxLineItem

	// Assuming customer_id can be parsed to uint64, or use 0
	// This is a simplification. Actual implementation needs robust string->uint64 parsing or ID mapping.
	var userID uint64 = 0

	for _, item := range req.Items {
		amount := int64(item.UnitPrice * float64(item.Quantity) * 100) // Convert to cents

		res, err := s.app.CalculateOrderTax(
			ctx,
			userID,
			req.ShippingAddress.CountryCode,
			req.ShippingAddress.StateCode,
			item.CategoryCode,
			amount,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to calculate tax: %v", err)
		}

		totalTax += res.TotalTaxAmount

		for _, detail := range res.Items {
			details = append(details, &v1.TaxLineItem{
				Title:  detail.RuleName,
				Rate:   detail.Rate,
				Amount: float64(detail.Amount) / 100.0,
				Type:   detail.TaxType.String(),
			})
		}
	}

	return &v1.CalculateTaxResponse{
		TotalTaxAmount: float64(totalTax) / 100.0,
		Details:        details,
		Inclusive:      false,
	}, nil
}

func (s *Server) GetTaxRates(ctx context.Context, req *v1.GetTaxRatesRequest) (*v1.GetTaxRatesResponse, error) {
	// Not implemented in app service yet, returning empty
	return &v1.GetTaxRatesResponse{}, nil
}
