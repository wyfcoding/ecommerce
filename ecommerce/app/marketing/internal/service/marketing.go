package service

import (
	"context"

	v1 "ecommerce/ecommerce/api/marketing/v1"
	"ecommerce/ecommerce/app/marketing/internal/biz"
)

type MarketingService struct {
	v1.UnimplementedMarketingServer
	uc *biz.CouponUsecase
}

func NewMarketingService(uc *biz.CouponUsecase) *MarketingService {
	return &MarketingService{uc: uc}
}

func (s *MarketingService) CreateCouponTemplate(ctx context.Context, req *v1.CreateCouponTemplateRequest) (*v1.CreateCouponTemplateResponse, error) {
	// ... 模型转换: v1.CreateCouponTemplateRequest -> biz.CouponTemplate
	// 调用 Usecase
	createdTemplate, err := s.uc.CreateCouponTemplate(ctx, &bizTemplate)
	if err != nil {
		// ... 错误处理
	}
	return &v1.CreateCouponTemplateResponse{TemplateId: createdTemplate.ID}, nil
}

func (s *MarketingService) ClaimCoupon(ctx context.Context, req *v1.ClaimCouponRequest) (*v1.ClaimCouponResponse, error) {
	_, err := s.uc.ClaimCoupon(ctx, req.UserId, req.TemplateId)
	if err != nil {
		// ... 错误处理
	}
	return &v1.ClaimCouponResponse{}, nil
}

func (s *MarketingService) CalculateDiscount(ctx context.Context, req *v1.CalculateDiscountRequest) (*v1.CalculateDiscountResponse, error) {
	// ... 模型转换: v1.OrderItemInfo -> biz.OrderItemInfo
	discount, err := s.uc.CalculateDiscount(ctx, req.UserId, req.CouponCode, bizItems)
	if err != nil {
		// ... 错误处理
	}

	var totalAmount uint64
	for _, item := range req.Items {
		totalAmount += item.Price * uint64(item.Quantity)
	}

	return &v1.CalculateDiscountResponse{
		DiscountAmount: discount,
		FinalAmount:    totalAmount - discount,
	}, nil
}

// ... ListMyCoupons 等其他接口的实现 ...
