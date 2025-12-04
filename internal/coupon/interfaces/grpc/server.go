package grpc

import (
	"context"
	"fmt"     // 导入格式化包，用于错误信息。
	"strconv" // 导入字符串和数字转换工具。

	pb "github.com/wyfcoding/ecommerce/api/coupon/v1"              // 导入优惠券模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/coupon/application"   // 导入优惠券模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/coupon/domain/entity" // 导入优惠券模块的领域实体。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/emptypb"     // 导入空消息类型。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 CouponService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedCouponServer                            // 嵌入生成的UnimplementedCouponServer，确保前向兼容性。
	app                          *application.CouponService // 依赖Coupon应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 Coupon gRPC 服务端实例。
func NewServer(app *application.CouponService) *Server {
	return &Server{app: app}
}

// CreateCoupon 处理创建优惠券的gRPC请求。
// req: 包含创建优惠券所需信息的请求体。
// 返回创建成功的优惠券响应和可能发生的gRPC错误。
func (s *Server) CreateCoupon(ctx context.Context, req *pb.CreateCouponRequest) (*pb.CouponResponse, error) {
	// 领域服务层的 CreateCoupon 方法参数相对简化。
	// Proto请求中包含更多字段 (valid_from, valid_until, total_quantity, etc.)，但服务方法未直接接收。
	// 当前实现仅使用服务方法支持的参数。理想情况下，服务方法应更全面或通过Update方法补充。

	// 将Proto的DiscountType映射到实体CouponType。
	// 注意：Proto中的DiscountType是字符串，而实体中的CouponType是int。
	// 这里简化为默认使用 CouponTypeDiscount。实际应根据req.DiscountType字符串进行映射。
	couponType := entity.CouponTypeDiscount
	// 将Proto中的浮点金额转换为整型（分）。
	discountVal := int64(req.DiscountValue * 100)
	minOrder := int64(req.MinOrderAmount * 100)

	// 调用应用服务层创建优惠券。
	coupon, err := s.app.CreateCoupon(ctx, req.Name, req.Description, couponType, discountVal, minOrder)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create coupon: %v", err))
	}

	// TODO: Proto中额外的字段 (valid_from, valid_until, total_quantity等) 未在此处设置。
	// 如果需要设置这些字段，需要扩展应用服务层的CreateCoupon方法，或者在Create后调用Update方法。

	return &pb.CouponResponse{
		Coupon: s.toProto(coupon),
	}, nil
}

// GetCouponByID 处理根据ID获取优惠券信息的gRPC请求。
// req: 包含优惠券ID的请求体。
// 返回优惠券响应和可能发生的gRPC错误。
func (s *Server) GetCouponByID(ctx context.Context, req *pb.GetCouponByIDRequest) (*pb.CouponResponse, error) {
	// 注意：应用服务层目前没有直接暴露根据ID获取Coupon的方法（只有仓储层有）。
	// 理想情况下，应用服务层应该提供一个公共方法来获取优惠券详情。
	// 为避免直接访问仓储（不符合DDD原则），此处暂时返回Unimplemented错误。
	return nil, status.Error(codes.Unimplemented, "GetCouponByID not exposed in application service")
}

// UpdateCoupon 处理更新优惠券信息的gRPC请求。
// 此方法尚未实现。
func (s *Server) UpdateCoupon(ctx context.Context, req *pb.UpdateCouponRequest) (*pb.CouponResponse, error) {
	// TODO: 实现UpdateCoupon逻辑。
	return nil, status.Error(codes.Unimplemented, "UpdateCoupon not implemented")
}

// DeleteCoupon 处理删除优惠券的gRPC请求。
// 此方法尚未实现。
func (s *Server) DeleteCoupon(ctx context.Context, req *pb.DeleteCouponRequest) (*emptypb.Empty, error) {
	// TODO: 实现DeleteCoupon逻辑。
	return nil, status.Error(codes.Unimplemented, "DeleteCoupon not implemented")
}

// IssueCoupon 处理向用户发放优惠券的gRPC请求。
// req: 包含用户ID和优惠券ID的请求体。
// 返回用户优惠券响应和可能发生的gRPC错误。
func (s *Server) IssueCoupon(ctx context.Context, req *pb.IssueCouponRequest) (*pb.UserCouponResponse, error) {
	userCoupon, err := s.app.IssueCoupon(ctx, req.UserId, req.CouponId)
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
		statusFilter = *req.Status // 获取状态过滤字符串。
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
	err := s.app.UseCoupon(ctx, req.UserCouponId, strconv.FormatUint(req.OrderId, 10))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to use coupon: %v", err))
	}

	// TODO: 应用服务层没有直接暴露GetUserCoupon方法来获取更新后的用户优惠券实体。
	// 此处返回一个包含已知状态的UserCouponInfo作为部分响应。
	// 理想情况下，应该能获取到完整的UserCoupon实体并进行转换。
	return &pb.UserCouponResponse{
		UserCoupon: &pb.UserCouponInfo{
			UserCouponId: req.UserCouponId,
			Status:       "USED", // 明确知道优惠券已使用。
			UserId:       req.UserId,
			// 其他字段缺失。
		},
	}, nil
}

// toProto 是一个辅助函数，将领域层的 Coupon 实体转换为 protobuf 的 CouponInfo 消息。
func (s *Server) toProto(c *entity.Coupon) *pb.CouponInfo {
	// 将实体CouponType（int）转换为protobuf所需的string类型。
	// 需要定义映射逻辑或在proto中定义enum。这里简单进行类型断言或转换。
	// strconv.Itoa(int(c.Type))
	var discountType string
	switch c.Type {
	case entity.CouponTypeDiscount:
		discountType = "DISCOUNT"
	case entity.CouponTypeCash:
		discountType = "CASH"
	case entity.CouponTypeGift:
		discountType = "GIFT"
	case entity.CouponTypeExchange:
		discountType = "EXCHANGE"
	default:
		discountType = "UNKNOWN"
	}

	return &pb.CouponInfo{
		CouponId:       uint64(c.ID),                      // 优惠券ID。
		Code:           c.CouponNo,                        // 优惠券编号。
		Name:           c.Name,                            // 名称。
		Description:    c.Description,                     // 描述。
		DiscountType:   discountType,                      // 优惠类型。
		DiscountValue:  float64(c.DiscountAmount) / 100.0, // 优惠金额（分转元）。
		MinOrderAmount: float64(c.MinOrderAmount) / 100.0, // 最低订单金额（分转元）。
		ValidFrom:      timestamppb.New(c.ValidFrom),      // 有效期开始时间。
		ValidUntil:     timestamppb.New(c.ValidTo),        // 有效期结束时间。
		TotalQuantity:  c.UsageLimit,                      // 总发行量。
		IssuedQuantity: c.TotalIssued,                     // 已发行量。
		// Proto中还包含其他字段如 MaxDiscount, UsagePerUser, ApplicableTo, Categories等，但实体中没有或未映射。
	}
}

// userCouponToProto 是一个辅助函数，将领域层的 UserCoupon 实体转换为 protobuf 的 UserCouponInfo 消息。
func (s *Server) userCouponToProto(uc *entity.UserCoupon) *pb.UserCouponInfo {
	var usedAt *timestamppb.Timestamp
	if uc.UsedAt != nil {
		usedAt = timestamppb.New(*uc.UsedAt)
	}
	return &pb.UserCouponInfo{
		UserCouponId: uint64(uc.ID),                  // 用户优惠券ID。
		UserId:       uc.UserID,                      // 用户ID。
		CouponId:     uc.CouponID,                    // 优惠券模板ID。
		Code:         uc.CouponNo,                    // 优惠券编号。
		Status:       uc.Status,                      // 优惠券状态。
		IssuedAt:     timestamppb.New(uc.ReceivedAt), // 领取时间。
		UsedAt:       usedAt,                         // 使用时间。
		// Proto中还包含其他字段如 OrderId 等，但实体中没有或未映射。
	}
}
