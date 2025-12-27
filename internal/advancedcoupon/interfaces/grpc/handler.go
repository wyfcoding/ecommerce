package grpc

import (
	"context"

	pb "github.com/wyfcoding/ecommerce/goapi/advancedcoupon/v1"
	"github.com/wyfcoding/ecommerce/internal/advancedcoupon/application"
	"github.com/wyfcoding/ecommerce/internal/advancedcoupon/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server 结构体定义。
type Server struct {
	pb.UnimplementedAdvancedCouponServer
	app *application.AdvancedCoupon
}

// NewServer 函数。
func NewServer(app *application.AdvancedCoupon) *Server {
	return &Server{app: app}
}

func (s *Server) CreateCoupon(ctx context.Context, req *pb.CreateCouponRequest) (*pb.CreateCouponResponse, error) {
	cType := domain.CouponType(req.Type)
	validFrom := req.ValidFrom.AsTime()
	validUntil := req.ValidUntil.AsTime()

	coupon, err := s.app.CreateCoupon(ctx, req.Code, cType, req.DiscountValue, validFrom, validUntil, req.TotalQuantity)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CreateCouponResponse{
		Coupon: convertCouponToProto(coupon),
	}, nil
}

func (s *Server) GetCoupon(ctx context.Context, req *pb.GetCouponRequest) (*pb.GetCouponResponse, error) {
	coupon, err := s.app.GetCoupon(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &pb.GetCouponResponse{
		Coupon: convertCouponToProto(coupon),
	}, nil
}

func (s *Server) ListCoupons(ctx context.Context, req *pb.ListCouponsRequest) (*pb.ListCouponsResponse, error) {
	statusVal := domain.CouponStatus(req.Status)
	page := max(int(req.PageNum), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	coupons, total, err := s.app.ListCoupons(ctx, statusVal, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbCoupons := make([]*pb.Coupon, len(coupons))
	for i, c := range coupons {
		pbCoupons[i] = convertCouponToProto(c)
	}

	return &pb.ListCouponsResponse{
		Coupons:    pbCoupons,
		TotalCount: uint64(total),
	}, nil
}

func (s *Server) UseCoupon(ctx context.Context, req *pb.UseCouponRequest) (*pb.UseCouponResponse, error) {
	if err := s.app.UseCoupon(ctx, req.UserId, req.Code, req.OrderId); err != nil {
		return &pb.UseCouponResponse{Success: false}, status.Error(codes.Internal, err.Error())
	}
	return &pb.UseCouponResponse{Success: true}, nil
}

func (s *Server) CalculateBestDiscount(ctx context.Context, req *pb.CalculateBestDiscountRequest) (*pb.CalculateBestDiscountResponse, error) {
	bestIds, finalPrice, discount, err := s.app.CalculateBestDiscount(ctx, req.OrderAmount, req.CouponIds)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CalculateBestDiscountResponse{
		BestCouponIds:  bestIds,
		FinalPrice:     finalPrice,
		DiscountAmount: discount,
	}, nil
}

func convertCouponToProto(c *domain.Coupon) *pb.Coupon {
	if c == nil {
		return nil
	}
	return &pb.Coupon{
		Id:                uint64(c.ID),
		Code:              c.Code,
		Type:              string(c.Type),
		DiscountValue:     c.DiscountValue,
		MinPurchaseAmount: c.MinPurchaseAmount,
		MaxDiscountAmount: c.MaxDiscountAmount,
		ValidFrom:         timestamppb.New(c.ValidFrom),
		ValidUntil:        timestamppb.New(c.ValidUntil),
		TotalQuantity:     c.TotalQuantity,
		UsedQuantity:      c.UsedQuantity,
		PerUserLimit:      c.PerUserLimit,
		Status:            string(c.Status),
	}
}
