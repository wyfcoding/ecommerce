package grpc

import (
	"context" // 导入上下文。
	"fmt"     // 导入格式化库。
	"strconv" // 导入字符串转换工具。

	pb "github.com/wyfcoding/ecommerce/go-api/review/v1"           // 导入评论模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/review/application"   // 导入评论模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/review/domain/entity" // 导入评论模块的领域实体。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 ReviewService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedReviewServer                            // 嵌入生成的UnimplementedReviewServer，确保前向兼容性。
	app                          *application.ReviewService // 依赖Review应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 Review gRPC 服务端实例。
func NewServer(app *application.ReviewService) *Server {
	return &Server{app: app}
}

// CreateReview 处理创建评论的gRPC请求。
// req: 包含用户ID、商品ID、评分、内容和图片URL的请求体。
// 返回created successfully的评论信息响应和可能发生的gRPC错误。
func (s *Server) CreateReview(ctx context.Context, req *pb.CreateReviewRequest) (*pb.CreateReviewResponse, error) {
	// 将字符串类型的用户ID和商品ID转换为uint64。
	userID, err := strconv.ParseUint(req.UserId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid user_id: %v", err))
	}
	productID, err := strconv.ParseUint(req.ProductId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid product_id: %v", err))
	}

	// 注意：Proto请求（pb.CreateReviewRequest）中没有包含 OrderID 和 SkuID，但应用服务层期望这些字段。
	// 这里暂时传递0。如果应用服务层严格验证这些ID的有效性，则可能导致错误。
	// 生产环境中，通常会从请求上下文或通过其他服务获取这些ID。
	review, err := s.app.CreateReview(ctx, userID, productID, 0, 0, int(req.Rating), req.Content, nil) // 0 for OrderID, 0 for SkuID, nil for images.
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create review: %v", err))
	}

	// 将领域实体转换为protobuf响应格式。
	return &pb.CreateReviewResponse{
		Review: s.toProto(review),
	}, nil
}

// ListProductReviews 处理列出商品评论的gRPC请求。
// req: 包含商品ID、分页参数的请求体。
// 返回评论列表响应和可能发生的gRPC错误。
func (s *Server) ListProductReviews(ctx context.Context, req *pb.ListProductReviewsRequest) (*pb.ListProductReviewsResponse, error) {
	// 将字符串类型的商品ID转换为uint64。
	productID, err := strconv.ParseUint(req.ProductId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid product_id: %v", err))
	}

	// 将PageToken作为页码进行简单处理。
	page := int(req.PageToken)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 评论列表通常只显示已通过审核的评论给用户。
	approvedStatus := int(entity.ReviewStatusApproved)
	reviews, total, err := s.app.ListReviews(ctx, productID, &approvedStatus, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list reviews: %v", err))
	}

	// 获取商品评分统计。
	stats, err := s.app.GetProductStats(ctx, productID)
	if err != nil {
		// 如果获取统计数据失败，暂时选择失败整个请求。
		// 在实际系统中，可能选择记录错误并继续返回评论列表。
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get product stats: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbReviews := make([]*pb.ReviewInfo, len(reviews))
	for i, r := range reviews {
		pbReviews[i] = s.toProto(r)
	}

	return &pb.ListProductReviewsResponse{
		Reviews:       pbReviews,
		TotalCount:    int32(total),
		AverageRating: stats.AverageRating, // 包含平均评分。
	}, nil
}

// ListUserReviews 处理列出用户评论的gRPC请求。
func (s *Server) ListUserReviews(ctx context.Context, req *pb.ListUserReviewsRequest) (*pb.ListUserReviewsResponse, error) {
	// 将字符串类型的用户ID转换为uint64。
	userID, err := strconv.ParseUint(req.UserId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid user_id: %v", err))
	}

	// 将PageToken作为页码进行简单处理。
	page := int(req.PageToken)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	reviews, total, err := s.app.ListUserReviews(ctx, userID, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list user reviews: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbReviews := make([]*pb.ReviewInfo, len(reviews))
	for i, r := range reviews {
		pbReviews[i] = s.toProto(r)
	}

	return &pb.ListUserReviewsResponse{
		Reviews:    pbReviews,
		TotalCount: int32(total),
	}, nil
}

// GetProductRating 处理获取商品评分的gRPC请求。
// req: 包含商品ID的请求体。
// 返回商品评分响应和可能发生的gRPC错误。
func (s *Server) GetProductRating(ctx context.Context, req *pb.GetProductRatingRequest) (*pb.GetProductRatingResponse, error) {
	// 将字符串类型的商品ID转换为uint64。
	productID, err := strconv.ParseUint(req.ProductId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid product_id: %v", err))
	}

	// 调用应用服务层获取商品评分统计。
	stats, err := s.app.GetProductStats(ctx, productID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get product stats: %v", err))
	}

	return &pb.GetProductRatingResponse{
		AverageRating: stats.AverageRating,
		TotalReviews:  int32(stats.TotalReviews),
	}, nil
}

// toProto 是一个辅助函数，将领域层的 Review 实体转换为 protobuf 的 ReviewInfo 消息。
func (s *Server) toProto(r *entity.Review) *pb.ReviewInfo {
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
		UpdatedAt: timestamppb.New(r.UpdatedAt), // 更新时间。
	}
}
