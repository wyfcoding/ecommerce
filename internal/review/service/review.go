package service

import (
	"context"
	"time"

	v1 "ecommerce/api/review/v1"
	"ecommerce/internal/review/model"
	"ecommerce/internal/review/repository"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ReviewService struct {
	v1.UnimplementedReviewServer
	reviewRepo repository.ReviewRepository
	logger     *zap.Logger
}

func NewReviewService(
	reviewRepo repository.ReviewRepository,
	logger *zap.Logger,
) *ReviewService {
	return &ReviewService{
		reviewRepo: reviewRepo,
		logger:     logger,
	}
}

// CreateReview 创建评论
func (s *ReviewService) CreateReview(ctx context.Context, req *v1.CreateReviewRequest) (*v1.ReviewResponse, error) {
	s.logger.Info("CreateReview", zap.Uint64("user_id", req.UserId), zap.Uint64("product_id", req.ProductId))

	// 参数校验
	if req.UserId == 0 || req.ProductId == 0 || req.Rating < 1 || req.Rating > 5 {
		return nil, status.Error(codes.InvalidArgument, "invalid parameters")
	}

	review := &model.Review{
		UserID:    req.UserId,
		ProductID: req.ProductId,
		OrderID:   req.OrderId,
		SkuID:     req.SkuId,
		Rating:    int(req.Rating),
		Content:   req.Content,
		Images:    req.Images,
		Status:    model.ReviewStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.reviewRepo.CreateReview(ctx, review); err != nil {
		s.logger.Error("failed to create review", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to create review")
	}

	return &v1.ReviewResponse{
		Review: s.toProtoReview(review),
	}, nil
}

// GetReview 获取评论详情
func (s *ReviewService) GetReview(ctx context.Context, req *v1.GetReviewRequest) (*v1.ReviewResponse, error) {
	s.logger.Info("GetReview", zap.Uint64("review_id", req.Id))

	review, err := s.reviewRepo.GetReviewByID(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get review")
	}
	if review == nil {
		return nil, status.Error(codes.NotFound, "review not found")
	}

	return &v1.ReviewResponse{
		Review: s.toProtoReview(review),
	}, nil
}

// ListReviews 获取评论列表
func (s *ReviewService) ListReviews(ctx context.Context, req *v1.ListReviewsRequest) (*v1.ListReviewsResponse, error) {
	s.logger.Info("ListReviews", zap.Uint64("product_id", req.ProductId))

	page := int(req.Page)
	pageSize := int(req.PageSize)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	reviews, total, err := s.reviewRepo.ListReviewsByProductID(ctx, req.ProductId, offset, pageSize)
	if err != nil {
		s.logger.Error("failed to list reviews", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list reviews")
	}

	protoReviews := make([]*v1.ReviewInfo, len(reviews))
	for i, r := range reviews {
		protoReviews[i] = s.toProtoReview(r)
	}

	return &v1.ListReviewsResponse{
		Reviews:  protoReviews,
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

// UpdateReview 更新评论
func (s *ReviewService) UpdateReview(ctx context.Context, req *v1.UpdateReviewRequest) (*v1.ReviewResponse, error) {
	s.logger.Info("UpdateReview", zap.Uint64("review_id", req.Id))

	review, err := s.reviewRepo.GetReviewByID(ctx, req.Id)
	if err != nil || review == nil {
		return nil, status.Error(codes.NotFound, "review not found")
	}

	// 只允许用户修改自己的评论
	if review.UserID != req.UserId {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	if req.Content != nil {
		review.Content = *req.Content
	}
	if req.Images != nil {
		review.Images = req.Images
	}
	if req.Rating != nil {
		if *req.Rating < 1 || *req.Rating > 5 {
			return nil, status.Error(codes.InvalidArgument, "rating must be between 1 and 5")
		}
		review.Rating = int(*req.Rating)
	}

	review.UpdatedAt = time.Now()

	if err := s.reviewRepo.UpdateReview(ctx, review); err != nil {
		s.logger.Error("failed to update review", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update review")
	}

	return &v1.ReviewResponse{
		Review: s.toProtoReview(review),
	}, nil
}

// DeleteReview 删除评论
func (s *ReviewService) DeleteReview(ctx context.Context, req *v1.DeleteReviewRequest) (*emptypb.Empty, error) {
	s.logger.Info("DeleteReview", zap.Uint64("review_id", req.Id))

	review, err := s.reviewRepo.GetReviewByID(ctx, req.Id)
	if err != nil || review == nil {
		return nil, status.Error(codes.NotFound, "review not found")
	}

	// 只允许用户删除自己的评论
	if review.UserID != req.UserId {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	if err := s.reviewRepo.DeleteReview(ctx, req.Id); err != nil {
		s.logger.Error("failed to delete review", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to delete review")
	}

	return &emptypb.Empty{}, nil
}

// ApproveReview 审核通过评论
func (s *ReviewService) ApproveReview(ctx context.Context, req *v1.ApproveReviewRequest) (*v1.ReviewResponse, error) {
	s.logger.Info("ApproveReview", zap.Uint64("review_id", req.Id))

	review, err := s.reviewRepo.GetReviewByID(ctx, req.Id)
	if err != nil || review == nil {
		return nil, status.Error(codes.NotFound, "review not found")
	}

	review.Status = model.ReviewStatusApproved
	review.UpdatedAt = time.Now()

	if err := s.reviewRepo.UpdateReview(ctx, review); err != nil {
		s.logger.Error("failed to approve review", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to approve review")
	}

	return &v1.ReviewResponse{
		Review: s.toProtoReview(review),
	}, nil
}

// RejectReview 审核拒绝评论
func (s *ReviewService) RejectReview(ctx context.Context, req *v1.RejectReviewRequest) (*v1.ReviewResponse, error) {
	s.logger.Info("RejectReview", zap.Uint64("review_id", req.Id))

	review, err := s.reviewRepo.GetReviewByID(ctx, req.Id)
	if err != nil || review == nil {
		return nil, status.Error(codes.NotFound, "review not found")
	}

	review.Status = model.ReviewStatusRejected
	review.UpdatedAt = time.Now()

	if err := s.reviewRepo.UpdateReview(ctx, review); err != nil {
		s.logger.Error("failed to reject review", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to reject review")
	}

	return &v1.ReviewResponse{
		Review: s.toProtoReview(review),
	}, nil
}

// GetProductRating 获取商品评分统计
func (s *ReviewService) GetProductRating(ctx context.Context, req *v1.GetProductRatingRequest) (*v1.ProductRatingResponse, error) {
	s.logger.Info("GetProductRating", zap.Uint64("product_id", req.ProductId))

	stats, err := s.reviewRepo.GetProductRatingStats(ctx, req.ProductId)
	if err != nil {
		s.logger.Error("failed to get product rating", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get product rating")
	}

	return &v1.ProductRatingResponse{
		ProductId:    req.ProductId,
		AverageRating: stats.AverageRating,
		TotalReviews: int32(stats.TotalReviews),
		Rating5Count: int32(stats.Rating5Count),
		Rating4Count: int32(stats.Rating4Count),
		Rating3Count: int32(stats.Rating3Count),
		Rating2Count: int32(stats.Rating2Count),
		Rating1Count: int32(stats.Rating1Count),
	}, nil
}

func (s *ReviewService) toProtoReview(r *model.Review) *v1.ReviewInfo {
	return &v1.ReviewInfo{
		Id:        r.ID,
		UserId:    r.UserID,
		ProductId: r.ProductID,
		OrderId:   r.OrderID,
		SkuId:     r.SkuID,
		Rating:    int32(r.Rating),
		Content:   r.Content,
		Images:    r.Images,
		Status:    int32(r.Status),
		CreatedAt: timestamppb.New(r.CreatedAt),
		UpdatedAt: timestamppb.New(r.UpdatedAt),
	}
}
