package service

import (
	"context"
	"errors"
	"time"

	v1 "ecommerce/api/smart_product_selection/v1"
	"ecommerce/internal/smart_product_selection/biz"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SmartProductSelectionService is the gRPC service implementation for smart product selection.
type SmartProductSelectionService struct {
	v1.UnimplementedSmartProductSelectionServiceServer
	uc *biz.SmartProductSelectionUsecase
}

// NewSmartProductSelectionService creates a new SmartProductSelectionService.
func NewSmartProductSelectionService(uc *biz.SmartProductSelectionUsecase) *SmartProductSelectionService {
	return &SmartProductSelectionService{uc: uc}
}

// bizProductRecommendationToProto converts biz.ProductRecommendation to v1.ProductRecommendation.
func bizProductRecommendationToProto(rec *biz.ProductRecommendation) *v1.ProductRecommendation {
	if rec == nil {
		return nil
	}
	return &v1.ProductRecommendation{
		ProductId:   rec.ProductID,
		ProductName: rec.ProductName,
		Score:       rec.Score,
		Reason:      rec.Reason,
	}
}

// GetSmartProductRecommendations implements the GetSmartProductRecommendations RPC.
func (s *SmartProductSelectionService) GetSmartProductRecommendations(ctx context.Context, req *v1.GetSmartProductRecommendationsRequest) (*v1.GetSmartProductRecommendationsResponse, error) {
	if req.MerchantId == "" {
		return nil, status.Error(codes.InvalidArgument, "merchant_id is required")
	}

	contextFeatures := make(map[string]string)
	for k, v := range req.ContextFeatures {
		contextFeatures[k] = v
	}

	recommendations, err := s.uc.GetSmartProductRecommendations(ctx, req.MerchantId, contextFeatures)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get smart product recommendations: %v", err)
	}

	protoRecommendations := make([]*v1.ProductRecommendation, len(recommendations))
	for i, rec := range recommendations {
		protoRecommendations[i] = bizProductRecommendationToProto(rec)
	}

	return &v1.GetSmartProductRecommendationsResponse{Recommendations: protoRecommendations}, nil
}
