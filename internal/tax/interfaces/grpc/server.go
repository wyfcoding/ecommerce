package grpc

import (
	"context"
	"strconv"

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
	if req.Subtotal < 0 {
		return nil, status.Error(codes.InvalidArgument, "subtotal must be non-negative")
	}

	uid, _ := strconv.ParseUint(req.CustomerId, 10, 64)
	country := ""
	if req.ShippingAddress != nil {
		country = req.ShippingAddress.CountryCode
	}
	category := ""
	if len(req.Items) > 0 {
		category = req.Items[0].CategoryCode
	}

	result, err := s.app.CalculateOrderTax(
		ctx,
		uid,
		country,
		"",
		category,
		int64(req.Subtotal),
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to calculate tax: %v", err)
	}

	return &v1.CalculateTaxResponse{
		TotalTaxAmount: float64(result.TotalTaxAmount),
		Inclusive:      false,
	}, nil
}
