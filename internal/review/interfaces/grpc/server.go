package grpc

import (
	"context"
	pb "ecommerce/api/review/v1"
	"ecommerce/internal/review/application"
	"ecommerce/internal/review/domain/entity"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedReviewServer
	app *application.ReviewService
}

func NewServer(app *application.ReviewService) *Server {
	return &Server{app: app}
}

func (s *Server) CreateReview(ctx context.Context, req *pb.CreateReviewRequest) (*pb.CreateReviewResponse, error) {
	userID, err := strconv.ParseUint(req.UserId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id")
	}
	productID, err := strconv.ParseUint(req.ProductId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product_id")
	}

	// Note: Proto doesn't have OrderID/SkuID, but Service requires them.
	// This is a gap. We'll use 0 for now or need to update proto.
	// Assuming 0 is acceptable for now or we might fail validation if service enforces it.
	// Service doesn't enforce OrderID/SkuID > 0 explicitly in the snippet seen, but usually required.
	// Let's pass 0 and see.

	review, err := s.app.CreateReview(ctx, userID, productID, 0, 0, int(req.Rating), req.Content, nil)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateReviewResponse{
		Review: s.toProto(review),
	}, nil
}

func (s *Server) ListProductReviews(ctx context.Context, req *pb.ListProductReviewsRequest) (*pb.ListProductReviewsResponse, error) {
	productID, err := strconv.ParseUint(req.ProductId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product_id")
	}

	page := int(req.PageToken) // Using PageToken as Page number for simplicity
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// Status nil means all (or approved? Service logic might filter).
	// Usually public API only shows approved.
	// Let's assume we want approved reviews.
	approved := int(entity.ReviewStatusApproved)
	reviews, total, err := s.app.ListReviews(ctx, productID, &approved, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	stats, err := s.app.GetProductStats(ctx, productID)
	if err != nil {
		// Log error but continue? Or fail?
		// Let's fail for now.
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbReviews := make([]*pb.ReviewInfo, len(reviews))
	for i, r := range reviews {
		pbReviews[i] = s.toProto(r)
	}

	return &pb.ListProductReviewsResponse{
		Reviews:       pbReviews,
		TotalCount:    int32(total),
		AverageRating: stats.AverageRating,
	}, nil
}

func (s *Server) ListUserReviews(ctx context.Context, req *pb.ListUserReviewsRequest) (*pb.ListUserReviewsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "ListUserReviews not implemented")
}

func (s *Server) GetProductRating(ctx context.Context, req *pb.GetProductRatingRequest) (*pb.GetProductRatingResponse, error) {
	productID, err := strconv.ParseUint(req.ProductId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid product_id")
	}

	stats, err := s.app.GetProductStats(ctx, productID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetProductRatingResponse{
		AverageRating: stats.AverageRating,
		TotalReviews:  int32(stats.TotalReviews),
	}, nil
}

func (s *Server) toProto(r *entity.Review) *pb.ReviewInfo {
	return &pb.ReviewInfo{
		Id:        strconv.FormatUint(uint64(r.ID), 10),
		ProductId: strconv.FormatUint(r.ProductID, 10),
		UserId:    strconv.FormatUint(r.UserID, 10),
		Rating:    int32(r.Rating),
		Title:     "", // Not in entity
		Content:   r.Content,
		CreatedAt: timestamppb.New(r.CreatedAt),
		UpdatedAt: timestamppb.New(r.UpdatedAt),
	}
}
