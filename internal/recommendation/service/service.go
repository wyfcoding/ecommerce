package service

import (
	"context"
	v1 "ecommerce/api/recommendation/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecommendationService is the gRPC service implementation for recommendation.
type RecommendationService struct {
	v1.UnimplementedRecommendationServiceServer
	uc *biz.RecommendationUsecase
}

// NewRecommendationService creates a new RecommendationService.
func NewRecommendationService(uc *biz.RecommendationUsecase) *RecommendationService {
	return &RecommendationService{uc: uc}
}

// GetRecommendedProducts implements the GetRecommendedProducts RPC.
func (s *RecommendationService) GetRecommendedProducts(ctx context.Context, req *v1.GetRecommendedProductsRequest) (*v1.GetRecommendedProductsResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.Count <= 0 {
		req.Count = 10 // Default count
	}

	bizProducts, err := s.uc.GetRecommendedProducts(ctx, req.UserId, req.Count)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get recommended products: %v", err)
	}

	protoProducts := make([]*v1.Product, 0, len(bizProducts))
	for _, p := range bizProducts {
		protoProducts = append(protoProducts, &v1.Product{
			Id:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			ImageUrl:    p.ImageURL,
		})
	}

	return &v1.GetRecommendedProductsResponse{
		Products: protoProducts,
	}, nil
}

// IndexProductRelationship implements the IndexProductRelationship RPC.
func (s *RecommendationService) IndexProductRelationship(ctx context.Context, req *v1.IndexProductRelationshipRequest) (*v1.IndexProductRelationshipResponse, error) {
	if req.ProductId1 == "" || req.ProductId2 == "" || req.Type == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id_1, product_id_2, and type are required")
	}

	bizRel := &biz.ProductRelationship{
		ProductID1: req.ProductId1,
		ProductID2: req.ProductId2,
		Type:       req.Type,
		Weight:     req.Weight,
	}

	err := s.uc.IndexProductRelationship(ctx, bizRel)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to index product relationship: %v", err)
	}

	return &v1.IndexProductRelationshipResponse{}, nil
}

// GetGraphRecommendedProducts implements the GetGraphRecommendedProducts RPC.
func (s *RecommendationService) GetGraphRecommendedProducts(ctx context.Context, req *v1.GetGraphRecommendedProductsRequest) (*v1.GetGraphRecommendedProductsResponse, error) {
	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}
	if req.Count <= 0 {
		req.Count = 10 // Default count
	}

	bizProducts, err := s.uc.GetGraphRecommendedProducts(ctx, req.ProductId, req.Count)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get graph recommended products: %v", err)
	}

	protoProducts := make([]*v1.Product, 0, len(bizProducts))
	for _, p := range bizProducts {
		protoProducts = append(protoProducts, &v1.Product{
			Id:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			ImageUrl:    p.ImageURL,
		})
	}

	return &v1.GetGraphRecommendedProductsResponse{
		Products: protoProducts,
	}, nil
}

// GetAdvancedRecommendedProducts implements the GetAdvancedRecommendedProducts RPC.
func (s *RecommendationService) GetAdvancedRecommendedProducts(ctx context.Context, req *v1.GetAdvancedRecommendedProductsRequest) (*v1.GetAdvancedRecommendedProductsResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.Count <= 0 {
		req.Count = 10 // Default count
	}

	contextFeatures := make(map[string]string)
	for k, v := range req.ContextFeatures {
		contextFeatures[k] = v
	}

	bizProducts, explanation, err := s.uc.GetAdvancedRecommendedProducts(ctx, req.UserId, req.Count, contextFeatures)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get advanced recommended products: %v", err)
	}

	protoProducts := make([]*v1.Product, 0, len(bizProducts))
	for _, p := range bizProducts {
		protoProducts = append(protoProducts, &v1.Product{
			Id:          p.ID,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
			ImageUrl:    p.ImageURL,
		})
	}

	return &v1.GetAdvancedRecommendedProductsResponse{
		Products:    protoProducts,
		Explanation: explanation,
	}, nil
}
