package grpc

import (
	"context"
	"fmt"     // 导入格式化包，用于错误信息。
	"strconv" // 导入字符串和数字转换工具。

	pb "github.com/wyfcoding/ecommerce/goapi/coupon/v1"          // 导入优惠券模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/coupon/application" // 导入优惠券模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/coupon/domain"      // 导入优惠券模块的领域层。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/emptypb"     // 导入空消息类型。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 Coupon 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedCouponServer                            // 嵌入生成的UnimplementedCouponServer，确保前向兼容性。
	app                          *application.Coupon // 依赖Coupon应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 Coupon gRPC 服务端实例。
func NewServer(app *application.Coupon) *Server {
	return &Server{app: app}
}

// CreateCoupon 处理创建优惠券的gRPC请求。
// req: 包含创建优惠券所需信息的请求体。
// 返回created successfully的优惠券响应和可能发生的gRPC错误。
func (s *Server) CreateCoupon(ctx context.Context, req *pb.CreateCouponRequest) (*pb.CouponResponse, error) {
	// 简化：默认使用 CouponTypeDiscount。
	couponType := int(domain.CouponTypeDiscount)
	// 将Proto中的浮点金额转换为整型（分）。
	discountVal := int64(req.DiscountValue * 100)
	minOrder := int64(req.MinOrderAmount * 100)

	// 调用应用服务层创建优惠券.
	coupon, err := s.app.CreateCoupon(ctx, req.Name, req.Description, couponType, discountVal, minOrder)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create coupon: %v", err))
	}

	return &pb.CouponResponse{
		Coupon: s.toProto(coupon),
	}, nil
}

// GetCouponByID 处理根据ID获取优惠券信息的gRPC请求。
// req: 包含优惠券ID的请求体。
// 返回优惠券响应和可能发生的gRPC错误。
func (s *Server) GetCouponByID(ctx context.Context, req *pb.GetCouponByIDRequest) (*pb.CouponResponse, error) {
	coupon, err := s.app.GetCoupon(ctx, req.CouponId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get coupon: %v", err))
	}
	return &pb.CouponResponse{
		Coupon: s.toProto(coupon),
	}, nil
}

// UpdateCoupon 处理更新优惠券信息的gRPC请求。
func (s *Server) UpdateCoupon(ctx context.Context, req *pb.UpdateCouponRequest) (*pb.CouponResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateCoupon not implemented")
}

// DeleteCoupon 处理删除优惠券的gRPC请求。
func (s *Server) DeleteCoupon(ctx context.Context, req *pb.DeleteCouponRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteCoupon not implemented")
}

// IssueCoupon 处理向用户发放优惠券的gRPC请求。
// req: 包含用户ID和优惠券ID的请求体。
// 返回用户优惠券响应和可能发生的gRPC错误。
func (s *Server) IssueCoupon(ctx context.Context, req *pb.IssueCouponRequest) (*pb.UserCouponResponse, error) {
	userCoupon, err := s.app.AcquireCoupon(ctx, req.UserId, req.CouponId)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to issue coupon: %v", err))
	}
	return &pb.UserCouponResponse{
		UserCoupon: s.userCouponToProto(userCoupon),
	}, nil
}

// GetUserCoupons 处理获取用户优惠券列表的gRPC请求。
// req: 包含用户ID和状态过滤的请求体。
// 返回用户优惠券列表响应和可能发生的gRPC错误。
func (s *Server) GetUserCoupons(ctx context.Context, req *pb.GetUserCouponsRequest) (*pb.GetUserCouponsResponse, error) {
	statusFilter := ""
	if req.Status != nil {
		statusFilter = *req.Status
	}
	// 注意：Proto请求中没有分页字段 (page/size)。
	// 应用服务层需要分页参数，此处使用默认值1页100条。
	userCoupons, _, err := s.app.ListUserCoupons(ctx, req.UserId, statusFilter, 1, 100)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list user coupons: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbUserCoupons := make([]*pb.UserCouponInfo, len(userCoupons))
	for i, uc := range userCoupons {
		pbUserCoupons[i] = s.userCouponToProto(uc)
	}

	return &pb.GetUserCouponsResponse{
		UserCoupons: pbUserCoupons,
	}, nil
}

// UseCoupon 处理使用优惠券的gRPC请求。
// req: 包含用户优惠券ID和订单ID的请求体。
// 返回更新后的用户优惠券响应和可能发生的gRPC错误。
func (s *Server) UseCoupon(ctx context.Context, req *pb.UseCouponRequest) (*pb.UserCouponResponse, error) {
	// 调用应用服务层使用优惠券。
	// 注意：Proto中的OrderId是uint64，这里将其转换为字符串传递给服务。
	err := s.app.UseCoupon(ctx, req.UserCouponId, req.UserId, strconv.FormatUint(req.OrderId, 10))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to use coupon: %v", err))
	}

	return &pb.UserCouponResponse{
		UserCoupon: &pb.UserCouponInfo{
			UserCouponId: req.UserCouponId,
			Status:       "used",
			UserId:       req.UserId,
		},
	}, nil
}

// toProto 转换 Coupon 实体为 protobuf 消息.
func (s *Server) toProto(c *domain.Coupon) *pb.CouponInfo {
	if c == nil {
		return nil
	}
	var discountType string
	switch c.Type {
	case domain.CouponTypeDiscount:
		discountType = "DISCOUNT"
	case domain.CouponTypeCash:
		discountType = "CASH"
	case domain.CouponTypeGift:
		discountType = "GIFT"
	case domain.CouponTypeExchange:
		discountType = "EXCHANGE"
	default:
		discountType = "UNKNOWN"
	}

	return &pb.CouponInfo{
		CouponId:       uint64(c.ID),
		Code:           c.CouponNo,
		Name:           c.Name,
		Description:    c.Description,
		DiscountType:   discountType,
		DiscountValue:  float64(c.DiscountAmount) / 100.0,
		MinOrderAmount: float64(c.MinOrderAmount) / 100.0,
		ValidFrom:      timestamppb.New(c.ValidFrom),
		ValidUntil:     timestamppb.New(c.ValidTo),
		TotalQuantity:  c.UsageLimit,
		IssuedQuantity: c.TotalIssued,
	}
}

// userCouponToProto 转换 UserCoupon 实体为 protobuf 消息.
func (s *Server) userCouponToProto(uc *domain.UserCoupon) *pb.UserCouponInfo {
	if uc == nil {
		return nil
	}
	var usedAt *timestamppb.Timestamp
	if uc.UsedAt != nil {
		usedAt = timestamppb.New(*uc.UsedAt)
	}
	return &pb.UserCouponInfo{
		UserCouponId: uint64(uc.ID),
		UserId:       uc.UserID,
		CouponId:     uc.CouponID,
		Code:         uc.CouponNo,
		Status:       uc.Status,
		IssuedAt:     timestamppb.New(uc.ReceivedAt),
		UsedAt:       usedAt,
	}
}
