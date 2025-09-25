package service

import (
	"context"

	"ecommerce/internal/coupon/biz"
	// Assuming the generated pb.go file will be in this path
	// v1 "ecommerce/api/coupon/v1"
)

// CouponService is a gRPC service that implements the CouponServer interface.
// It holds a reference to the business logic layer.
type CouponService struct {
	// v1.UnimplementedCouponServer

	uc *biz.CouponUsecase
}

// NewCouponService creates a new CouponService.
func NewCouponService(uc *biz.CouponUsecase) *CouponService {
	return &CouponService{uc: uc}
}

// Note: The actual RPC methods like CreateCoupon, GetCoupon, etc., will be implemented here.
// These methods will call the corresponding business logic in the 'biz' layer.

/*
Example Implementation (once gRPC code is generated):

func (s *CouponService) GetCoupon(ctx context.Context, req *v1.GetCouponRequest) (*v1.GetCouponResponse, error) {
    // 1. Call business logic
    coupon, err := s.uc.GetByCode(ctx, req.Code)
    if err != nil {
        return nil, err
    }

    // 2. Convert biz model to API model and return
    return &v1.GetCouponResponse{Coupon: coupon}, nil
}

*/
