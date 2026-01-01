package grpc

import (
	"context" // 导入上下文。
	"fmt"     // 导入格式化库。
	"log/slog"
	"strconv" // 导入字符串转换工具。
	"time"

	pb "github.com/wyfcoding/ecommerce/goapi/review/v1"          // 导入评论模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/review/application" // 导入评论模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/review/domain"      // 导入评论模块的领域层。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 Review 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedReviewServiceServer                     // 嵌入生成的UnimplementedReviewServiceServer，确保前向兼容性。
	app                                 *application.Review // 依赖Review应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 Review gRPC 服务端实例。
func NewServer(app *application.Review) *Server {
	return &Server{app: app}
}

// CreateReview 处理创建评论的gRPC请求。
func (s *Server) CreateReview(ctx context.Context, req *pb.CreateReviewRequest) (*pb.CreateReviewResponse, error) {
	start := time.Now()
	slog.Info("gRPC CreateReview received", "user_id", req.UserId, "product_id", req.ProductId, "rating", req.Rating)

	// 将字符串类型的用户ID和商品ID转换为uint64。
	userID, err := strconv.ParseUint(req.UserId, 10, 64)
	if err != nil {
		slog.Warn("gRPC CreateReview invalid user_id", "user_id", req.UserId, "error", err)
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid user_id: %v", err))
	}
	productID, err := strconv.ParseUint(req.ProductId, 10, 64)
	if err != nil {
		slog.Warn("gRPC CreateReview invalid product_id", "product_id", req.ProductId, "error", err)
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid product_id: %v", err))
	}

	// 映射 Proto 字段到应用服务层.
	review, err := s.app.CreateReview(ctx, userID, productID, 0, 0, int(req.Rating), req.Content, nil)
	if err != nil {
		slog.Error("gRPC CreateReview failed", "user_id", req.UserId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create review: %v", err))
	}

	slog.Info("gRPC CreateReview successful", "review_id", review.ID, "duration", time.Since(start))
	// 将领域实体转换为protobuf响应格式。
	return &pb.CreateReviewResponse{
		Review: s.toProto(review),
	}, nil
}

// ListProductReviews 处理列出商品评论的gRPC请求。
func (s *Server) ListProductReviews(ctx context.Context, req *pb.ListProductReviewsRequest) (*pb.ListProductReviewsResponse, error) {
	start := time.Now()
	slog.Debug("gRPC ListProductReviews received", "product_id", req.ProductId)

	// 将字符串类型的商品ID转换为uint64。
	productID, err := strconv.ParseUint(req.ProductId, 10, 64)
	if err != nil {
		slog.Warn("gRPC ListProductReviews invalid product_id", "product_id", req.ProductId, "error", err)
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid product_id: %v", err))
	}

	// 将PageToken作为页码进行简单处理。
	page := max(int(req.PageToken), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 评论列表通常只显示已通过审核的评论给用户。
	approvedStatus := int(domain.ReviewStatusApproved)
	reviews, total, err := s.app.ListReviews(ctx, productID, &approvedStatus, page, pageSize)
	if err != nil {
		slog.Error("gRPC ListReviews failed", "product_id", productID, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list reviews: %v", err))
	}

	// 获取商品评分统计。
	stats, err := s.app.GetProductStats(ctx, productID)
	if err != nil {
		slog.Error("gRPC GetProductStats failed", "product_id", productID, "error", err)
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get product stats: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbReviews := make([]*pb.ReviewInfo, len(reviews))
	for i, r := range reviews {
		pbReviews[i] = s.toProto(r)
	}

	slog.Debug("gRPC ListProductReviews successful", "product_id", productID, "count", len(pbReviews), "duration", time.Since(start))
	return &pb.ListProductReviewsResponse{
		Reviews:       pbReviews,
		TotalCount:    int32(total),
		AverageRating: stats.AverageRating, // 包含平均评分。
	}, nil
}

// ListUserReviews 处理列出用户评论的gRPC请求。
func (s *Server) ListUserReviews(ctx context.Context, req *pb.ListUserReviewsRequest) (*pb.ListUserReviewsResponse, error) {
	start := time.Now()
	slog.Debug("gRPC ListUserReviews received", "user_id", req.UserId)

	// 将字符串类型的用户ID转换为uint64。
	userID, err := strconv.ParseUint(req.UserId, 10, 64)
	if err != nil {
		slog.Warn("gRPC ListUserReviews invalid user_id", "user_id", req.UserId, "error", err)
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid user_id: %v", err))
	}

	// 将PageToken作为页码进行简单处理。
	page := max(int(req.PageToken), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	reviews, total, err := s.app.ListUserReviews(ctx, userID, page, pageSize)
	if err != nil {
		slog.Error("gRPC ListUserReviews failed", "user_id", userID, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list user reviews: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbReviews := make([]*pb.ReviewInfo, len(reviews))
	for i, r := range reviews {
		pbReviews[i] = s.toProto(r)
	}

	slog.Debug("gRPC ListUserReviews successful", "user_id", userID, "count", len(pbReviews), "duration", time.Since(start))
	return &pb.ListUserReviewsResponse{
		Reviews:    pbReviews,
		TotalCount: int32(total),
	}, nil
}

// GetProductRating 处理获取商品评分的gRPC请求。
func (s *Server) GetProductRating(ctx context.Context, req *pb.GetProductRatingRequest) (*pb.GetProductRatingResponse, error) {
	start := time.Now()
	slog.Debug("gRPC GetProductRating received", "product_id", req.ProductId)

	// 将字符串类型的商品ID转换为uint64。
	productID, err := strconv.ParseUint(req.ProductId, 10, 64)
	if err != nil {
		slog.Warn("gRPC GetProductRating invalid product_id", "product_id", req.ProductId, "error", err)
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid product_id: %v", err))
	}

	// 调用应用服务层获取商品评分统计。
	stats, err := s.app.GetProductStats(ctx, productID)
	if err != nil {
		slog.Error("gRPC GetProductStats failed", "product_id", productID, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get product stats: %v", err))
	}

	slog.Debug("gRPC GetProductRating successful", "product_id", productID, "duration", time.Since(start))
	return &pb.GetProductRatingResponse{
		AverageRating: stats.AverageRating,
		TotalReviews:  int32(stats.TotalReviews),
	}, nil
}

// toProto 是一个辅助函数，将领域层的 Review 实体转换为 protobuf 的 ReviewInfo 消息。
func (s *Server) toProto(r *domain.Review) *pb.ReviewInfo {
	if r == nil {
		return nil
	}
	return &pb.ReviewInfo{
		Id:        strconv.FormatUint(uint64(r.ID), 10), // 评论ID。
		ProductId: strconv.FormatUint(r.ProductID, 10),  // 商品ID。
		UserId:    strconv.FormatUint(r.UserID, 10),     // 用户ID。
		Rating:    int32(r.Rating),                      // 评分。
		Title:     "",                                   // 注意：Review 实体中没有 Title 字段，此处留空。
		Content:   r.Content,                            // 内容。
		// Images 字段需要单独处理，如果Proto中有此字段。
		CreatedAt: timestamppb.New(r.CreatedAt), // 创建时间。
	}
}
