package service

import (
	"context"
	v1 "ecommerce/api/pricing/v1"
	"ecommerce/internal/pricing/biz"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PricingService is the gRPC service implementation for pricing.
type PricingService struct {
	v1.UnimplementedPricingServiceServer
	uc *biz.PricingUsecase
}

// NewPricingService creates a new PricingService.
func NewPricingService(uc *biz.PricingUsecase) *PricingService {
	return &PricingService{uc: uc}
}

// CalculateFinalPrice implements the CalculateFinalPrice RPC.
func (s *PricingService) CalculateFinalPrice(ctx context.Context, req *v1.CalculateFinalPriceRequest) (*v1.CalculateFinalPriceResponse, error) {
	if req.UserId == 0 || len(req.Items) == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id and items are required")
	}

	bizItems := make([]*biz.SkuPriceInfo, len(req.Items))
	for i, item := range req.Items {
		bizItems[i] = &biz.SkuPriceInfo{
			SkuID:         item.SkuId,
			OriginalPrice: item.OriginalPrice,
			Quantity:      item.Quantity,
		}
	}

	totalOriginalPrice, totalDiscountAmount, finalPrice, err := s.uc.CalculateFinalPrice(ctx, req.UserId, bizItems, req.CouponCode)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to calculate final price: %v", err)
	}

	return &v1.CalculateFinalPriceResponse{
		TotalOriginalPrice:  totalOriginalPrice,
		TotalDiscountAmount: totalDiscountAmount,
		FinalPrice:          finalPrice,
	}, nil
}

// CalculateDynamicPrice implements the CalculateDynamicPrice RPC.
func (s *PricingService) CalculateDynamicPrice(ctx context.Context, req *v1.CalculateDynamicPriceRequest) (*v1.CalculateDynamicPriceResponse, error) {
	if req.ProductId == 0 || req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "product_id and user_id are required")
	}

	contextFeatures := make(map[string]string)
	for k, v := range req.ContextFeatures {
		contextFeatures[k] = v
	}

	dynamicPrice, explanation, err := s.uc.CalculateDynamicPrice(ctx, req.ProductId, req.UserId, contextFeatures)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to calculate dynamic price: %v", err)
	}

	return &v1.CalculateDynamicPriceResponse{
		DynamicPrice: dynamicPrice,
		Explanation:  explanation,
	}, nil
}
