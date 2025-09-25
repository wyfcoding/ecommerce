package biz

import (
	"context"
	"errors"
	couponbiz "ecommerce/internal/coupon/biz"
	orderbiz "ecommerce/internal/order/biz"
	productbiz "ecommerce/internal/product/biz"
	reviewbiz "ecommerce/internal/review/biz"
	userbiz "ecommerce/internal/user/biz"
	"strconv"
)

// Transaction 定义了事务管理器接口。
type Transaction interface {
	InTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// AdminUser 是管理员用户的业务领域模型。
type AdminUser struct {
	ID       uint32
	Username string
	Password string
	Name     string
	Status   int32
}

// AdminRepo 定义了管理员数据仓库需要实现的接口。
type AdminRepo interface {
	CreateAdminUser(ctx context.Context, user *AdminUser) (*AdminUser, error)
	GetAdminUserByUsername(ctx context.Context, username string) (*AdminUser, error)
}

// AdminUsecase 是管理员的业务用例。
type AdminUsecase struct {
	repo AdminRepo
	authUsecase    *AuthUsecase // Assuming AuthUsecase is in auth.go
	productUsecase *productbiz.ProductUsecase
	userUsecase    *userbiz.UserUsecase
	orderUsecase   *orderbiz.OrderUsecase
	reviewUsecase  *reviewbiz.ReviewUsecase
	couponUsecase  *couponbiz.CouponUsecase
	// TODO: Add password hasher interface
}

// NewAdminUsecase 创建一个新的 AdminUsecase。
func NewAdminUsecase(repo AdminRepo, authUC *AuthUsecase, productUC *productbiz.ProductUsecase, userUC *userbiz.UserUsecase, orderUC *orderbiz.OrderUsecase, reviewUC *reviewbiz.ReviewUsecase, couponUC *couponbiz.CouponUsecase) *AdminUsecase {
	return &AdminUsecase{
		repo:           repo,
		authUsecase:    authUC,
		productUsecase: productUC,
		userUsecase:    userUC,
		orderUsecase:   orderUC,
		reviewUsecase:  reviewUC,
		couponUsecase:  couponUC,
	}
}

// CreateAdminUser 注册一个新管理员用户。
func (uc *AdminUsecase) CreateAdminUser(ctx context.Context, username, password, name string) (*AdminUser, error) {
	// 1. 检查用户名是否已存在
	existingUser, err := uc.repo.GetAdminUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// 2. 密码哈希 (这里简化处理，实际应用中应使用 bcrypt 等安全哈希算法)
	hashedPassword := password // TODO: Implement actual password hashing

	// 3. 创建管理员用户
	user := &AdminUser{
		Username: username,
		Password: hashedPassword,
		Name:     name,
		Status:   1, // Default status to active
	}
	createdUser, err := uc.repo.CreateAdminUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return createdUser, nil
}

// ListUsers 列出用户。
func (uc *AdminUsecase) ListUsers(ctx context.Context, page, pageSize uint32, searchQuery, statusFilter string) ([]*userbiz.User, uint64, error) {
	return uc.userUsecase.ListUsers(ctx, page, pageSize, searchQuery, statusFilter)
}

// GetUserByID 获取用户详情。
func (uc *AdminUsecase) GetUserByID(ctx context.Context, userID string) (*userbiz.User, error) {
	return uc.userUsecase.GetUserByID(ctx, userID)
}

// UpdateUserStatus 更新用户状态。
func (uc *AdminUsecase) UpdateUserStatus(ctx context.Context, userID, status string) (*userbiz.User, error) {
	return uc.userUsecase.UpdateUserStatus(ctx, userID, status)
}

// ListSpus 列出商品。
func (uc *AdminUsecase) ListSpus(ctx context.Context, page, pageSize uint32, searchQuery, categoryID, statusFilter string) ([]*productbiz.Spu, uint64, error) {
	return uc.productUsecase.ListSpus(ctx, page, pageSize, searchQuery, categoryID, statusFilter)
}

// UpdateProduct 更新商品。
func (uc *AdminUsecase) UpdateProduct(ctx context.Context, spu *productbiz.Spu, skus []*productbiz.Sku) (*productbiz.Spu, error) {
	return uc.productUsecase.UpdateProduct(ctx, spu, skus)
}

// DeleteSpu 删除商品。
func (uc *AdminUsecase) DeleteSpu(ctx context.Context, spuID uint) error {
	return uc.productUsecase.DeleteSpu(ctx, spuID)
}

// UpdateOrderStatus 更新订单状态。
func (uc *AdminUsecase) UpdateOrderStatus(ctx context.Context, orderID uint, status string) (*orderbiz.Order, error) {
	return uc.orderUsecase.UpdateOrderStatus(ctx, orderID, status)
}

// ListReviews 列出评论。
func (uc *AdminUsecase) ListReviews(ctx context.Context, page, pageSize uint32, productID, userID, statusFilter string) ([]*reviewbiz.Review, uint64, error) {
	return uc.reviewUsecase.ListReviews(ctx, page, pageSize, productID, userID, statusFilter)
}

// ModerateReview 审核评论。
func (uc *AdminUsecase) ModerateReview(ctx context.Context, reviewID uint, status, moderationNotes string) (*reviewbiz.Review, error) {
	return uc.reviewUsecase.ModerateReview(ctx, reviewID, status, moderationNotes)
}

// CreateCoupon 创建优惠券。
func (uc *AdminUsecase) CreateCoupon(ctx context.Context, coupon *couponbiz.Coupon) (*couponbiz.Coupon, error) {
	return uc.couponUsecase.CreateCoupon(ctx, coupon)
}

// ListCoupons 列出优惠券。
func (uc *AdminUsecase) ListCoupons(ctx context.Context, page, pageSize uint32, activeOnly bool, searchQuery string) ([]*couponbiz.Coupon, uint64, error) {
	return uc.couponUsecase.ListCoupons(ctx, page, pageSize, activeOnly, searchQuery)
}

// UpdateCoupon 更新优惠券。
func (uc *AdminUsecase) UpdateCoupon(ctx context.Context, coupon *couponbiz.Coupon) (*couponbiz.Coupon, error) {
	return uc.couponUsecase.UpdateCoupon(ctx, coupon)
}

// DeleteCoupon 删除优惠券。
func (uc *AdminUsecase) DeleteCoupon(ctx context.Context, couponID uint) error {
	return uc.couponUsecase.DeleteCoupon(ctx, couponID)
}
