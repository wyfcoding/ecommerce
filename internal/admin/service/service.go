package service

import (
	"context"
	"time"

	"ecommerce/internal/admin/model"
	v1 "ecommerce/api/admin/v1" // Assuming v1 is the gRPC API definition
)

// AuthService 定义了认证相关的业务逻辑接口。
type AuthService interface {
	// AdminLogin 负责管理员登录的业务逻辑。
	AdminLogin(ctx context.Context, username, password string) (string, error)
	// GetJwtSecret 返回 JWT 密钥。
	GetJwtSecret() string
	// GetAdminUserByID 负责根据ID获取管理员用户信息的业务逻辑。
	GetAdminUserByID(ctx context.Context, id uint32) (*model.AdminUser, error)
}

// AdminService 定义了管理后台服务的业务逻辑接口。
// 这是一个聚合服务，它会调用多个下游微服务来完成管理操作。
type AdminService interface {
	// GetDashboardStatistics 获取仪表盘的聚合统计数据。
	// 该方法会并发调用多个下游服务，聚合数据后返回。
	GetDashboardStatistics(ctx context.Context) (map[string]interface{}, error)

	// 用户管理
	// ListUsers 列出用户列表。
	ListUsers(ctx context.Context, req *v1.ListUsersRequest) (*v1.ListUsersResponse, error)
	// GetUserDetails 获取用户详情。
	GetUserDetails(ctx context.Context, req *v1.GetUserDetailsRequest) (*v1.GetUserDetailsResponse, error)
	// UpdateUserStatus 更新用户状态。
	UpdateUserStatus(ctx context.Context, req *v1.UpdateUserStatusRequest) (*v1.UpdateUserStatusResponse, error)

	// 商品管理
	// CreateProduct 创建商品。
	CreateProduct(ctx context.Context, req *v1.CreateProductRequest) (*v1.CreateProductResponse, error)
	// ListProducts 列出商品列表。
	ListProducts(ctx context.Context, req *v1.ListProductsRequest) (*v1.ListProductsResponse, error)
	// UpdateProduct 更新商品信息。
	UpdateProduct(ctx context.Context, req *v1.UpdateProductRequest) (*v1.UpdateProductResponse, error)
	// DeleteProduct 删除商品。
	DeleteProduct(ctx context.Context, req *v1.DeleteProductRequest) (*v1.DeleteProductResponse, error)

	// 订单管理
	// ListOrders 列出订单列表。
	ListOrders(ctx context.Context, req *v1.ListOrdersRequest) (*v1.ListOrdersResponse, error)
	// GetOrderDetail 获取订单详情。
	GetOrderDetail(ctx context.Context, req *v1.GetOrderDetailRequest) (*v1.GetOrderDetailResponse, error)
	// ShipOrder 发货订单。
	ShipOrder(ctx context.Context, req *v1.ShipOrderRequest) (*v1.ShipOrderResponse, error)
	// UpdateOrderStatus 更新订单状态。
	UpdateOrderStatus(ctx context.Context, req *v1.UpdateOrderStatusRequest) (*v1.UpdateOrderStatusResponse, error)

	// 评论管理
	// ListReviews 列出评论列表。
	ListReviews(ctx context.Context, req *v1.ListReviewsRequest) (*v1.ListReviewsResponse, error)
	// ModerateReview 审核评论。
	ModerateReview(ctx context.Context, req *v1.ModerateReviewRequest) (*v1.ModerateReviewResponse, error)

	// 优惠券管理
	// CreateCoupon 创建优惠券。
	CreateCoupon(ctx context.Context, req *v1.CreateCouponRequest) (*v1.CreateCouponResponse, error)
	// ListCoupons 列出优惠券列表。
	ListCoupons(ctx context.Context, req *v1.ListCouponsRequest) (*v1.ListCouponsResponse, error)
	// UpdateCoupon 更新优惠券信息。
	UpdateCoupon(ctx context.Context, req *v1.UpdateCouponRequest) (*v1.UpdateCouponResponse, error)
	// DeleteCoupon 删除优惠券。
	DeleteCoupon(ctx context.Context, req *v1.DeleteCouponRequest) (*v1.DeleteCouponResponse, error)

	// 管理员信息
	// GetAdminUserInfo 获取管理员自身信息。
	GetAdminUserInfo(ctx context.Context, req *v1.GetAdminUserInfoRequest) (*v1.AdminUserInfoResponse, error)
}