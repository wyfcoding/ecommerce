package service

import (
	"context"

	"ecommerce/internal/review/biz"
	// Assuming the generated pb.go file will be in this path
	// v1 "ecommerce/api/review/v1"
)

// ReviewService is a gRPC service that implements the ReviewServer interface.
// It holds a reference to the business logic layer.
type ReviewService struct {
	// v1.UnimplementedReviewServer

	uc *biz.ReviewUsecase
}

// NewReviewService creates a new ReviewService.
func NewReviewService(uc *biz.ReviewUsecase) *ReviewService {
	return &ReviewService{uc: uc}
}

// Note: The actual RPC methods like CreateReview, ListProductReviews, etc., will be implemented here.
// These methods will call the corresponding business logic in the 'biz' layer.

/*
Example Implementation (once gRPC code is generated):

func (s *ReviewService) CreateReview(ctx context.Context, req *v1.CreateReviewRequest) (*v1.CreateReviewResponse, error) {
    // 1. Call business logic
    review, err := s.uc.CreateReview(ctx, req.ProductId, req.UserId, req.Rating, req.Title, req.Content)
    if err != nil {
        return nil, err
    }

    // 2. Convert biz model to API model and return
    return &v1.CreateReviewResponse{Review: review}, nil
}

*/
