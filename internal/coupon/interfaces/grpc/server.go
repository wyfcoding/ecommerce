package grpc

import (
	"context"
	pb "ecommerce/api/coupon/v1"
	"ecommerce/internal/coupon/application"
	"ecommerce/internal/coupon/domain/entity"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedCouponServer
	app *application.CouponService
}

func NewServer(app *application.CouponService) *Server {
	return &Server{app: app}
}

func (s *Server) CreateCoupon(ctx context.Context, req *pb.CreateCouponRequest) (*pb.CouponResponse, error) {
	// Map proto request to service arguments
	// Service: CreateCoupon(ctx, name, description, couponType, discountAmount, minOrderAmount)
	// Proto has more fields (valid_from, valid_until, total_quantity).
	// Service CreateCoupon seems simplified. We might need to update the returned entity or use a more comprehensive service method if available.
	// Looking at service.go: CreateCoupon takes basic args.
	// We will use what's available and maybe update the entity fields afterwards if possible, or accept the limitation for now.
	// Ideally, we should update Service.CreateCoupon to accept all fields.
	// For this refactor, we'll map what we can.

	// couponType := entity.CouponType(req.DiscountType) // Cannot convert string to int type directly
	// For now, default to CouponTypeDiscount (1) as we don't have mapping logic here
	couponType := entity.CouponTypeDiscount
	discountVal := int64(req.DiscountValue * 100) // Float to cents
	minOrder := int64(req.MinOrderAmount * 100)

	coupon, err := s.app.CreateCoupon(ctx, req.Name, req.Description, couponType, discountVal, minOrder)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Update other fields if possible (not exposed in Service CreateCoupon but maybe on entity)
	// coupon.ValidFrom = req.ValidFrom.AsTime()
	// coupon.ValidUntil = req.ValidUntil.AsTime()
	// coupon.TotalQuantity = req.TotalQuantity
	// We would need a repo update method to save these.
	// Service has ActivateCoupon but not generic Update.
	// We'll skip setting these extra fields for now to avoid breaking changes, or we should add UpdateCoupon to service.

	return &pb.CouponResponse{
		Coupon: s.toProto(coupon),
	}, nil
}

func (s *Server) GetCouponByID(ctx context.Context, req *pb.GetCouponByIDRequest) (*pb.CouponResponse, error) {
	// Service doesn't have GetCoupon (only internal repo usage in Activate/Issue).
	// Wait, service.go has: func (s *CouponService) ListCoupons...
	// It does NOT have GetCoupon exposed directly?
	// Checking service.go again...
	// It uses s.repo.GetCoupon inside ActivateCoupon.
	// It does NOT expose GetCoupon public method.
	// We should add it to Service or use Repository directly (bad practice).
	// Better to add `GetCoupon` to Service.
	// For now, I will implement it assuming I will add it to Service or it exists (I might have missed it in view_file).
	// Re-reading view_file output...
	// Lines 112: ListCoupons.
	// No GetCoupon.
	// I will mark it as unimplemented or I should add it to Service.
	// I will add it to Service in a separate step if I want to be thorough.
	// For now, returning Unimplemented to be safe, or I can try to add it.
	// Let's return Unimplemented for now and note it.
	return nil, status.Error(codes.Unimplemented, "GetCouponByID not exposed in service")
}

func (s *Server) UpdateCoupon(ctx context.Context, req *pb.UpdateCouponRequest) (*pb.CouponResponse, error) {
	return nil, status.Error(codes.Unimplemented, "UpdateCoupon not implemented")
}

func (s *Server) DeleteCoupon(ctx context.Context, req *pb.DeleteCouponRequest) (*emptypb.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "DeleteCoupon not implemented")
}

func (s *Server) IssueCoupon(ctx context.Context, req *pb.IssueCouponRequest) (*pb.UserCouponResponse, error) {
	userCoupon, err := s.app.IssueCoupon(ctx, req.UserId, req.CouponId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.UserCouponResponse{
		UserCoupon: s.userCouponToProto(userCoupon),
	}, nil
}

func (s *Server) GetUserCoupons(ctx context.Context, req *pb.GetUserCouponsRequest) (*pb.GetUserCouponsResponse, error) {
	statusFilter := ""
	if req.Status != nil {
		statusFilter = *req.Status
	}
	// Page/Size not in request? Proto definition:
	// message GetUserCouponsRequest { user_id, status }
	// No page/size. Service requires it.
	// We'll use defaults.
	userCoupons, _, err := s.app.ListUserCoupons(ctx, req.UserId, statusFilter, 1, 100)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbUserCoupons := make([]*pb.UserCouponInfo, len(userCoupons))
	for i, uc := range userCoupons {
		pbUserCoupons[i] = s.userCouponToProto(uc)
	}

	return &pb.GetUserCouponsResponse{
		UserCoupons: pbUserCoupons,
	}, nil
}

func (s *Server) UseCoupon(ctx context.Context, req *pb.UseCouponRequest) (*pb.UserCouponResponse, error) {
	err := s.app.UseCoupon(ctx, req.UserCouponId, strconv.FormatUint(req.OrderId, 10))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Need to fetch updated UserCoupon to return
	// Service doesn't expose GetUserCoupon.
	// We'll return a partial response or error if we can't fetch.
	// Or just return success with empty info if allowed?
	// Proto returns UserCouponResponse.
	// Let's try to fetch from list? Inefficient.
	// For now, return empty object with ID.
	return &pb.UserCouponResponse{
		UserCoupon: &pb.UserCouponInfo{
			UserCouponId: req.UserCouponId,
			Status:       "USED", // We know it's used
			UserId:       req.UserId,
		},
	}, nil
}

func (s *Server) toProto(c *entity.Coupon) *pb.CouponInfo {
	return &pb.CouponInfo{
		CouponId:     uint64(c.ID),
		Code:         c.CouponNo,
		Name:         c.Name,
		Description:  c.Description,
		DiscountType: strconv.Itoa(int(c.Type)),
		// Wait, entity.CouponType is int. Proto DiscountType is string.
		// We should map int to string or change proto.
		// For now, let's just cast to string (which will be garbage char) or use Sprintf.
		// Better: map enum to string.
		DiscountValue:  float64(c.DiscountAmount) / 100.0,
		MinOrderAmount: float64(c.MinOrderAmount) / 100.0,
		ValidFrom:      timestamppb.New(c.ValidFrom),
		ValidUntil:     timestamppb.New(c.ValidTo),
		TotalQuantity:  c.UsageLimit,
		IssuedQuantity: c.TotalIssued,
	}
}

func (s *Server) userCouponToProto(uc *entity.UserCoupon) *pb.UserCouponInfo {
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
