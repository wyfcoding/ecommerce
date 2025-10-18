package repository

import (
	"context"

	"ecommerce/internal/marketing/model"
)

type CouponRepo interface {
	CreateTemplate(ctx context.Context, template *model.CouponTemplate) (*model.CouponTemplate, error)
	ClaimCoupon(ctx context.Context, userID, templateID uint64) (*model.UserCoupon, error)
	GetUserCouponByCode(ctx context.Context, userID uint64, code string) (*model.UserCoupon, error)
	GetTemplateByID(ctx context.Context, templateID uint64) (*model.CouponTemplate, error)
	ListUserCoupons(ctx context.Context, userID uint64, status int8) ([]*model.UserCoupon, error)
	UpdateUserCouponStatus(ctx context.Context, userCouponID uint64, newStatus int8, orderID *uint64) error
}