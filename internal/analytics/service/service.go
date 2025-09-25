package service

import (
	"context"
	"strconv" // For getUserIDFromContext

	v1 "ecommerce/api/analytics/v1"
	"ecommerce/internal/analytics/biz"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AnalyticsService is the gRPC service implementation for analytics.
type AnalyticsService struct {
	v1.UnimplementedAnalyticsServiceServer
	uc *biz.AnalyticsUsecase
}

// NewAnalyticsService creates a new AnalyticsService.
func NewAnalyticsService(uc *biz.AnalyticsUsecase) *AnalyticsService {
	return &AnalyticsService{uc: uc}
}

// getUserIDFromContext 从 gRPC 上下文的 metadata 中提取用户ID。
func getUserIDFromContext(ctx context.Context) (uint64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Errorf(codes.Unauthenticated, "无法获取元数据")
	}
	// 兼容 gRPC-Gateway 在 HTTP 请求时注入的用户ID
	values := md.Get("x-md-global-user-id")
	if len(values) == 0 {
		// 兼容直接 gRPC 调用时注入的用户ID
		values = md.Get("x-user-id")
		if len(values) == 0 {
			return 0, status.Errorf(codes.Unauthenticated, "请求头中缺少 x-user-id 信息")
		}
	}
	userID, err := strconv.ParseUint(values[0], 10, 64)
	if err != nil {
		return 0, status.Errorf(codes.Unauthenticated, "x-user-id 格式无效")
	}
	return userID, nil
}

// RecordProductView implements the RecordProductView RPC。
func (s *AnalyticsService) RecordProductView(ctx context.Context, req *v1.RecordProductViewRequest) (*v1.RecordProductViewResponse, error) {
	// For simplicity, we'll use the user_id from the request directly.
	// In a real system, you might get it from context (JWT token) or validate it.
	userID := req.UserId

	if userID == 0 || req.ProductId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id and product_id are required")
	}

	err := s.uc.RecordProductView(ctx, userID, req.ProductId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to record product view: %v", err)
	}

	return &v1.RecordProductViewResponse{}, nil
}
