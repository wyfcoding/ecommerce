package service

import (
	"context"
	"time"

	v1 "ecommerce/api/product/v1"
	"ecommerce/internal/product/model"
	"ecommerce/internal/product/repository"
)

// ReviewService is a Review service.
type ReviewService struct {
	repo repository.ReviewRepo
}

// NewReviewService creates a new ReviewService.
func NewReviewService(repo repository.ReviewRepo) *ReviewService {
	return &ReviewService{repo: repo}
}

// CreateReview creates a Review.
func (s *ReviewService) CreateReview(ctx context.Context, req *v1.CreateReviewRequest) (*v1.ReviewInfo, error) {
	review := &model.Review{
		SpuID:   req.SpuId,
		UserID:  req.UserId,
		Rating:  req.Rating,
		Comment: req.Comment,
		Images:  req.Images,
	}
	createdReview, err := s.repo.CreateReview(ctx, review)
	if err != nil {
		return nil, err
	}
	return bizReviewToProto(createdReview), nil
}

// ListReviews lists Reviews.
func (s *ReviewService) ListReviews(ctx context.Context, req *v1.ListReviewsRequest) ([]*v1.ReviewInfo, uint64, error) {
	var minRating *uint32
	if req.HasMinRating() {
		mr := req.GetMinRating()
		minRating = &mr
	}
	reviews, total, err := s.repo.ListReviews(ctx, req.SpuId, req.PageSize, req.PageNum, minRating)
	if err != nil {
		return nil, 0, err
	}
	var reviewInfos []*v1.ReviewInfo
	for _, r := range reviews {
		reviewInfos = append(reviewInfos, bizReviewToProto(r))
	}
	return reviewInfos, total, nil
}

// DeleteReview deletes a Review.
func (s *ReviewService) DeleteReview(ctx context.Context, id uint64, userID uint64) error {
	return s.repo.DeleteReview(ctx, id, userID)
}

// bizReviewToProto converts biz.Review to v1.ReviewInfo
func bizReviewToProto(r *model.Review) *v1.ReviewInfo {
	if r == nil {
		return nil
	}
	return &v1.ReviewInfo{
		Id:        r.ID,
		SpuId:     r.SpuID,
		UserId:    r.UserID,
		Rating:    r.Rating,
		Comment:   r.Comment,
		Images:    r.Images,
		CreatedAt: r.CreatedAt.Format(time.RFC3339), // Format time to string
	}
}
