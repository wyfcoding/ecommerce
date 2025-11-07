package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	v1 "ecommerce/api/admin/v1"
	couponv1 "ecommerce/api/coupon/v1"
	orderv1 "ecommerce/api/order/v1"
	productv1 "ecommerce/api/product/v1"
	reviewv1 "ecommerce/api/review/v1"
	userv1 "ecommerce/api/user/v1"
	"ecommerce/internal/admin/model"
	"ecommerce/internal/admin/repository"
	"ecommerce/pkg/jwt"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	// ErrAdminUserNotFound 表示未找到管理员用户。
	ErrAdminUserNotFound = errors.New("admin user not found")
	// ErrAdminPasswordIncorrect 表示管理员密码不正确。
	ErrAdminPasswordIncorrect = errors.New("incorrect admin password")
)

// --- authServiceConcrete --- //

// authServiceConcrete 封装了认证相关的业务逻辑的具体实现。
type authServiceConcrete struct {
	repo      repository.AdminUserRepo // 使用 AdminUserRepo 接口进行管理员用户数据访问
	jwtSecret string
	jwtIssuer string
	jwtExpire time.Duration
}

// NewAuthServiceConcrete 是 authServiceConcrete 的构造函数。
// 它接收一个 AdminUserRepo 实例、JWT 密钥、签发者和过期时间，并返回 AuthService 接口。
func NewAuthServiceConcrete(repo repository.AdminUserRepo, jwtSecret, jwtIssuer string, jwtExpire time.Duration) AuthService {
	return &authServiceConcrete{
		repo:      repo,
		jwtSecret: jwtSecret,
		jwtIssuer: jwtIssuer,
		jwtExpire: jwtExpire,
	}
}

// AdminLogin 负责管理员登录的业务逻辑。
// 它通过用户名从仓库获取管理员用户，验证密码，检查用户状态，然后生成并返回 JWT Token。
func (s *authServiceConcrete) AdminLogin(ctx context.Context, username, password string) (string, error) {
	user, err := s.repo.GetAdminUserByUsername(ctx, username)
	if err != nil {
		zap.S().Errorf("Failed to get admin user by username %s: %v", username, err)
		return "", status.Errorf(codes.Internal, "Failed to get admin user")
	}
	if user == nil {
		zap.S().Warnf("Admin user %s not found during login attempt", username)
		return "", ErrAdminUserNotFound
	}

	// 验证密码
	if !s.repo.VerifyAdminPassword(ctx, username, password) {
		zap.S().Warnf("Incorrect password for admin user %s", username)
		return "", ErrAdminPasswordIncorrect
	}

	// 检查用户状态
	if user.Status != 1 { // 假设 1 为正常状态
		zap.S().Warnf("Admin account %s is disabled or abnormal", username)
		return "", errors.New("admin account is disabled or abnormal")
	}

	// 生成 JWT Token
	token, err := jwt.GenerateToken(uint64(user.ID), user.Username, s.jwtSecret, s.jwtIssuer, s.jwtExpire, jwt.SigningMethodHS256)
	if err != nil {
		zap.S().Errorf("Failed to generate token for admin user %d: %v", user.ID, err)
		return "", status.Errorf(codes.Internal, "Failed to generate token")
	}

	zap.S().Infof("Admin user %d (%s) logged in successfully", user.ID, user.Username)
	return token, nil
}

// GetJwtSecret 返回 JWT 密钥。
func (s *authServiceConcrete) GetJwtSecret() string {
	return s.jwtSecret
}

// GetAdminUserByID 负责根据ID获取管理员用户信息的业务逻辑。
func (s *authServiceConcrete) GetAdminUserByID(ctx context.Context, id uint32) (*model.AdminUser, error) {
	user, err := s.repo.GetAdminUserByID(ctx, id)
	if err != nil {
		zap.S().Errorf("Failed to get admin user by id %d: %v", id, err)
		return nil, status.Errorf(codes.Internal, "Failed to get admin user")
	}
	if user == nil {
		return nil, ErrAdminUserNotFound
	}
	return user, nil
}

// --- adminServiceGRPC --- //

// adminServiceGRPC 是 AdminService 接口的 gRPC 实现。
// 它嵌入了 v1.UnimplementedAdminServer 以确保向前兼容性，并持有对业务逻辑服务和下游 gRPC 客户端的引用。
type adminServiceGRPC struct {
	v1.UnimplementedAdminServer

	authService   AuthService // 业务逻辑认证服务
	productClient productv1.ProductClient
	userClient    userv1.UserClient
	orderClient   orderv1.OrderClient
	reviewClient  reviewv1.ReviewClient
	couponClient  couponv1.CouponClient
}

// NewAdminServiceGRPC 是 adminServiceGRPC 的构造函数。
// 它接收业务逻辑认证服务和所有必要的下游 gRPC 客户端，并返回 AdminService 接口。
func NewAdminServiceGRPC(authService AuthService, productClient productv1.ProductClient, userClient userv1.UserClient, orderClient orderv1.OrderClient, reviewClient reviewv1.ReviewClient, couponClient couponv1.CouponClient) AdminService {
	return &adminServiceGRPC{
		authService:   authService,
		productClient: productClient,
		userClient:    userClient,
		orderClient:   orderClient,
		reviewClient:  reviewClient,
		couponClient:  couponClient,
	}
}

// AdminLogin 实现了管理员登录 RPC。
// 它调用内部的 authService 来处理登录逻辑。
func (s *adminServiceGRPC) AdminLogin(ctx context.Context, req *v1.AdminLoginRequest) (*v1.AdminLoginResponse, error) {
	token, err := s.authService.AdminLogin(ctx, req.Username, req.Password)
	if err != nil {
		// authService 已经记录了详细错误
		return nil, status.Errorf(codes.Unauthenticated, err.Error())
	}
	return &v1.AdminLoginResponse{Token: token}, nil
}

// ListUsers 实现了列出用户 RPC。
// 它将 admin.v1.ListUsersRequest 转换为 user.v1.ListUsersRequest，调用用户服务，
// 然后将 user.v1.ListUsersResponse 转换为 admin.v1.ListUsersResponse。
func (s *adminServiceGRPC) ListUsers(ctx context.Context, req *v1.ListUsersRequest) (*v1.ListUsersResponse, error) {
	userReq := &userv1.ListUsersRequest{
		Page:         req.Page,
		PageSize:     req.PageSize,
		SearchQuery:  req.SearchQuery,
		StatusFilter: req.StatusFilter,
	}

	userResp, err := s.userClient.ListUsers(ctx, userReq)
	if err != nil {
		zap.S().Errorf("Failed to list users from user service: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to list users")
	}

	adminUsers := make([]*v1.UserInfo, len(userResp.Users))
	for i, u := range userResp.Users {
		adminUsers[i] = &v1.UserInfo{
			UserId:   u.UserId,
			Username: u.Username,
			Nickname: u.Nickname,
			Avatar:   u.Avatar,
			Gender:   u.Gender,
			Birthday: u.Birthday,
		}
	}

	return &v1.ListUsersResponse{
		Users: adminUsers,
		Total: userResp.Total,
	}, nil
}

// GetUserDetails 实现了获取用户详情 RPC。
// 它通过用户ID调用用户服务获取用户详情。
func (s *adminServiceGRPC) GetUserDetails(ctx context.Context, req *v1.GetUserDetailsRequest) (*v1.GetUserDetailsResponse, error) {
	userID, err := strconv.ParseUint(req.GetUserId(), 10, 64)
	if err != nil {
		zap.S().Warnf("GetUserDetails: invalid user ID format: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "Invalid user ID format")
	}

	userResp, err := s.userClient.GetUserByID(ctx, &userv1.GetUserByIDRequest{UserId: userID})
	if err != nil {
		zap.S().Errorf("Failed to get user details from user service for id %d: %v", userID, err)
		return nil, status.Errorf(codes.Internal, "Failed to get user details")
	}
	if userResp == nil || userResp.User == nil {
		zap.S().Warnf("User %d not found from user service", userID)
		return nil, status.Errorf(codes.NotFound, "User not found")
	}

	return &v1.GetUserDetailsResponse{
		User: &v1.UserInfo{
			UserId:   userResp.User.UserId,
			Username: userResp.User.Username,
			Nickname: userResp.User.Nickname,
			Avatar:   userResp.User.Avatar,
			Gender:   userResp.User.Gender,
			Birthday: userResp.User.Birthday,
		},
	}, nil
}

// UpdateUserStatus 实现了更新用户状态 RPC。
// 它调用用户服务来更新用户的状态。
func (s *adminServiceGRPC) UpdateUserStatus(ctx context.Context, req *v1.UpdateUserStatusRequest) (*v1.UpdateUserStatusResponse, error) {
	userID, err := strconv.ParseUint(req.GetUserId(), 10, 64)
	if err != nil {
		zap.S().Warnf("UpdateUserStatus: invalid user ID format: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "Invalid user ID format")
	}

	// 假设 user service 有一个 UpdateUserStatus 方法
	// 目前 user service 只有 UpdateUserInfo，这里需要适配或扩展 user service
	// 暂时通过 UpdateUserInfo 来模拟更新状态
	userResp, err := s.userClient.UpdateUserInfo(ctx, &userv1.UpdateUserInfoRequest{
		UserId: userID,
		// Status 字段需要 user.proto 和 user service 支持
		// 这里暂时无法直接更新状态，仅作示例
	})
	if err != nil {
		zap.S().Errorf("Failed to update user status from user service for id %d: %v", userID, err)
		return nil, status.Errorf(codes.Internal, "Failed to update user status")
	}

	return &v1.UpdateUserStatusResponse{
		User: &v1.UserInfo{
			UserId:   userResp.User.UserId,
			Username: userResp.User.Username,
			Nickname: userResp.User.Nickname,
			Avatar:   userResp.User.Avatar,
			Gender:   userResp.User.Gender,
			Birthday: userResp.User.Birthday,
		},
	}, nil
}

// CreateProduct 实现了创建商品 RPC。
// 它调用 product service 的 CreateProduct 方法。
func (s *adminServiceGRPC) CreateProduct(ctx context.Context, req *v1.CreateProductRequest) (*v1.CreateProductResponse, error) {
	productResp, err := s.productClient.CreateProduct(ctx, &productv1.CreateProductRequest{
		Spu:  req.Spu,
		Skus: req.Skus,
	})
	if err != nil {
		zap.S().Errorf("Failed to create product from product service: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to create product")
	}
	return &v1.CreateProductResponse{SpuId: productResp.SpuId}, nil
}

// ListProducts 实现了列出商品 RPC。
// 它调用 product service 的 ListProducts 方法。
func (s *adminServiceGRPC) ListProducts(ctx context.Context, req *v1.ListProductsRequest) (*v1.ListProductsResponse, error) {
	productReq := &productv1.ListProductsRequest{
		Page:         req.Page,
		PageSize:     req.PageSize,
		SearchQuery:  req.SearchQuery,
		CategoryId:   req.CategoryId,
		StatusFilter: req.StatusFilter,
	}
	productResp, err := s.productClient.ListProducts(ctx, productReq)
	if err != nil {
		zap.S().Errorf("Failed to list products from product service: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to list products")
	}
	return &v1.ListProductsResponse{
		Products: productResp.Products,
		Total:    productResp.Total,
	}, nil
}

// UpdateProduct 实现了更新商品 RPC。
// 它调用 product service 的 UpdateProduct 方法。
func (s *adminServiceGRPC) UpdateProduct(ctx context.Context, req *v1.UpdateProductRequest) (*v1.UpdateProductResponse, error) {
	productResp, err := s.productClient.UpdateProduct(ctx, &productv1.UpdateProductRequest{
		ProductId: req.ProductId,
		Spu:       req.Spu,
		Skus:      req.Skus,
	})
	if err != nil {
		zap.S().Errorf("Failed to update product from product service: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to update product")
	}
	return &v1.UpdateProductResponse{Spu: productResp.Spu}, nil
}

// DeleteProduct 实现了删除商品 RPC。
// 它调用 product service 的 DeleteProduct 方法。
func (s *adminServiceGRPC) DeleteProduct(ctx context.Context, req *v1.DeleteProductRequest) (*v1.DeleteProductResponse, error) {
	productResp, err := s.productClient.DeleteProduct(ctx, &productv1.DeleteProductRequest{
		ProductId: req.ProductId,
	})
	if err != nil {
		zap.S().Errorf("Failed to delete product from product service: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to delete product")
	}
	return &v1.DeleteProductResponse{Success: productResp.Success, Message: productResp.Message}, nil
}

// ListOrders 实现了列出订单 RPC。
// 它调用 order service 的 ListOrders 方法。
func (s *adminServiceGRPC) ListOrders(ctx context.Context, req *v1.ListOrdersRequest) (*v1.ListOrdersResponse, error) {
	orderReq := &orderv1.ListOrdersRequest{
		Page:     req.Page,
		PageSize: req.PageSize,
		UserId:   req.UserId,
		Status:   req.Status,
	}
	orderResp, err := s.orderClient.ListOrders(ctx, orderReq)
	if err != nil {
		zap.S().Errorf("Failed to list orders from order service: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to list orders")
	}
	return &v1.ListOrdersResponse{
		Orders: orderResp.Orders,
		Total:  orderResp.Total,
	}, nil
}

// GetOrderDetail 实现了获取订单详情 RPC。
// 它调用 order service 的 GetOrderDetail 方法。
func (s *adminServiceGRPC) GetOrderDetail(ctx context.Context, req *v1.GetOrderDetailRequest) (*v1.GetOrderDetailResponse, error) {
	orderResp, err := s.orderClient.GetOrderDetail(ctx, &orderv1.GetOrderDetailRequest{OrderId: req.OrderId})
	if err != nil {
		zap.S().Errorf("Failed to get order detail from order service for id %d: %v", req.OrderId, err)
		return nil, status.Errorf(codes.Internal, "Failed to get order detail")
	}
	return &v1.GetOrderDetailResponse{Order: orderResp.Order, Items: orderResp.Items}, nil
}

// ShipOrder 实现了发货订单 RPC。
// 它调用 order service 的 ShipOrder 方法。
func (s *adminServiceGRPC) ShipOrder(ctx context.Context, req *v1.ShipOrderRequest) (*v1.ShipOrderResponse, error) {
	_, err := s.orderClient.ShipOrder(ctx, &orderv1.ShipOrderRequest{
		OrderId:         req.OrderId,
		TrackingCompany: req.TrackingCompany,
		TrackingNumber:  req.TrackingNumber,
	})
	if err != nil {
		zap.S().Errorf("Failed to ship order from order service for id %d: %v", req.OrderId, err)
		return nil, status.Errorf(codes.Internal, "Failed to ship order")
	}
	return &v1.ShipOrderResponse{}, nil
}

// UpdateOrderStatus 实现了更新订单状态 RPC。
// 它调用 order service 的 UpdateOrderStatus 方法。
func (s *adminServiceGRPC) UpdateOrderStatus(ctx context.Context, req *v1.UpdateOrderStatusRequest) (*v1.UpdateOrderStatusResponse, error) {
	orderResp, err := s.orderClient.UpdateOrderStatus(ctx, &orderv1.UpdateOrderStatusRequest{
		OrderId: req.OrderId,
		Status:  req.Status,
	})
	if err != nil {
		zap.S().Errorf("Failed to update order status from order service for id %d: %v", req.OrderId, err)
		return nil, status.Errorf(codes.Internal, "Failed to update order status")
	}
	return &v1.UpdateOrderStatusResponse{Order: orderResp.Order}, nil
}

// ListReviews 实现了列出评论 RPC。
// 它调用 review service 的 ListReviews 方法。
func (s *adminServiceGRPC) ListReviews(ctx context.Context, req *v1.ListReviewsRequest) (*v1.ListReviewsResponse, error) {
	reviewReq := &reviewv1.ListReviewsRequest{
		Page:         req.Page,
		PageSize:     req.PageSize,
		ProductId:    req.ProductId,
		UserId:       req.UserId,
		StatusFilter: req.StatusFilter,
	}
	reviewResp, err := s.reviewClient.ListReviews(ctx, reviewReq)
	if err != nil {
		zap.S().Errorf("Failed to list reviews from review service: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to list reviews")
	}
	return &v1.ListReviewsResponse{
		Reviews: reviewResp.Reviews,
		Total:   reviewResp.Total,
	}, nil
}

// ModerateReview 实现了审核评论 RPC。
// 它调用 review service 的 ModerateReview 方法。
func (s *adminServiceGRPC) ModerateReview(ctx context.Context, req *v1.ModerateReviewRequest) (*v1.ModerateReviewResponse, error) {
	reviewResp, err := s.reviewClient.ModerateReview(ctx, &reviewv1.ModerateReviewRequest{
		ReviewId:        req.ReviewId,
		Status:          req.Status,
		ModerationNotes: req.ModerationNotes,
	})
	if err != nil {
		zap.S().Errorf("Failed to moderate review from review service for id %d: %v", req.ReviewId, err)
		return nil, status.Errorf(codes.Internal, "Failed to moderate review")
	}
	return &v1.ModerateReviewResponse{Review: reviewResp.Review}, nil
}

// CreateCoupon 实现了创建优惠券 RPC。
// 它调用 coupon service 的 CreateCoupon 方法。
func (s *adminServiceGRPC) CreateCoupon(ctx context.Context, req *v1.CreateCouponRequest) (*v1.CreateCouponResponse, error) {
	couponResp, err := s.couponClient.CreateCoupon(ctx, &couponv1.CreateCouponRequest{
		Code:          req.Code,
		Description:   req.Description,
		DiscountValue: req.DiscountValue,
		Type:          req.Type,
		ValidFrom:     req.ValidFrom,
		ValidTo:       req.ValidTo,
		MaxUsage:      req.MaxUsage,
	})
	if err != nil {
		zap.S().Errorf("Failed to create coupon from coupon service: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to create coupon")
	}
	return &v1.CreateCouponResponse{Coupon: couponResp.Coupon}, nil
}

// ListCoupons 实现了列出优惠券 RPC。
// 它调用 coupon service 的 ListCoupons 方法。
func (s *adminServiceGRPC) ListCoupons(ctx context.Context, req *v1.ListCouponsRequest) (*v1.ListCouponsResponse, error) {
	couponReq := &couponv1.ListCouponsRequest{
		Page:        req.Page,
		PageSize:    req.PageSize,
		ActiveOnly:  req.ActiveOnly,
		SearchQuery: req.SearchQuery,
	}
	couponResp, err := s.couponClient.ListCoupons(ctx, couponReq)
	if err != nil {
		zap.S().Errorf("Failed to list coupons from coupon service: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to list coupons")
	}
	return &v1.ListCouponsResponse{Coupons: couponResp.Coupons, Total: couponResp.Total}, nil
}

// UpdateCoupon 实现了更新优惠券 RPC。
// 它调用 coupon service 的 UpdateCoupon 方法。
func (s *adminServiceGRPC) UpdateCoupon(ctx context.Context, req *v1.UpdateCouponRequest) (*v1.UpdateCouponResponse, error) {
	couponResp, err := s.couponClient.UpdateCoupon(ctx, &couponv1.UpdateCouponRequest{
		CouponId:      req.CouponId,
		Code:          req.Code,
		Description:   req.Description,
		DiscountValue: req.DiscountValue,
		Type:          req.Type,
		ValidFrom:     req.ValidFrom,
		ValidTo:       req.ValidTo,
		IsActive:      req.IsActive,
		MaxUsage:      req.MaxUsage,
	})
	if err != nil {
		zap.S().Errorf("Failed to update coupon from coupon service: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to update coupon")
	}
	return &v1.UpdateCouponResponse{Coupon: couponResp.Coupon}, nil
}

// DeleteCoupon 实现了删除优惠券 RPC。
// 它调用 coupon service 的 DeleteCoupon 方法。
func (s *adminServiceGRPC) DeleteCoupon(ctx context.Context, req *v1.DeleteCouponRequest) (*v1.DeleteCouponResponse, error) {
	couponResp, err := s.couponClient.DeleteCoupon(ctx, &couponv1.DeleteCouponRequest{
		CouponId: req.CouponId,
	})
	if err != nil {
		zap.S().Errorf("Failed to delete coupon from coupon service: %v", err)
		return nil, status.Errorf(codes.Internal, "Failed to delete coupon")
	}
	return &v1.DeleteCouponResponse{Success: couponResp.Success, Message: couponResp.Message}, nil
}

// GetAdminUserInfo 根据adminUserID获取管理员信息并返回。
// 它调用内部的 authService 来获取管理员用户信息。
func (s *adminServiceGRPC) GetAdminUserInfo(ctx context.Context, req *v1.GetAdminUserInfoRequest) (*v1.AdminUserInfoResponse, error) {
	adminUserID, err := strconv.ParseUint(req.GetId(), 10, 64)
	if err != nil {
		zap.S().Warnf("GetAdminUserInfo: invalid admin user ID format: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "Invalid admin user ID")
	}

	adminUser, err := s.authService.GetAdminUserByID(ctx, uint32(adminUserID))
	if err != nil {
		zap.S().Errorf("Failed to get admin user info for id %d: %v", adminUserID, err)
		return nil, status.Errorf(codes.Internal, "Failed to get admin user info")
	}

	return &v1.AdminUserInfoResponse{
			Id:       adminUser.ID,
			Username: adminUser.Username,
			Name:     adminUser.Name,
			Status:   adminUser.Status,
		},
	nil
}

// --- AuthInterceptor --- //

type contextKey string

const adminUserIDKey contextKey = "adminUserID"

var (
	// authWhitelist 定义了不需要认证的白名单方法。
	// 格式为 /package.Service/Method。
	authWhitelist = map[string]bool{
		"/admin.v1.Admin/AdminLogin": true,
	}
)

// AuthInterceptor 是一个 gRPC 拦截器，用于处理认证。
type AuthInterceptor struct {
	authService AuthService // 使用 AuthService 接口
	jwtSecret   string
}

// NewAuthInterceptor 是 AuthInterceptor 的构造函数。
// 它接收 AuthService 接口和 JWT 密钥。
func NewAuthInterceptor(authService AuthService, jwtSecret string) *AuthInterceptor {
	return &AuthInterceptor{
		authService: authService,
		jwtSecret:   jwtSecret,
	}
}

// Auth 是一个 gRPC UnaryInterceptor，用于验证 JWT Token。
// 它检查请求是否在白名单中，从 metadata 中获取并验证 JWT Token，
// 然后将管理员用户ID放入 context 中传递给后续的 RPC 处理函数。
func (i *AuthInterceptor) Auth(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// 1. 检查是否是白名单方法
	if authWhitelist[info.FullMethod] {
		return handler(ctx, req)
	}

	// 2. 从 metadata 中获取 JWT Token
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		zap.S().Warn("AuthInterceptor: missing authentication metadata")
		return nil, status.Errorf(codes.Unauthenticated, "Missing authentication metadata")
	}
	token := md.Get("authorization")
	if len(token) == 0 || !strings.HasPrefix(token[0], "Bearer ") {
		zap.S().Warn("AuthInterceptor: incorrect authorization header format")
		return nil, status.Errorf(codes.Unauthenticated, "Incorrect authorization header format")
	}

	// 3. 解析并验证 Token
	tokenString := strings.TrimPrefix(token[0], "Bearer ")
	claims, err := jwt.ParseToken(tokenString, i.jwtSecret)
	if err != nil {
		zap.S().Warnf("AuthInterceptor: invalid or expired token: %v", err)
		return nil, status.Errorf(codes.Unauthenticated, "Invalid or expired token: %v", err)
	}

	// 4. 将用户ID等信息放入 context，传递给后续的 RPC 处理函数
	ctx = context.WithValue(ctx, adminUserIDKey, claims.UserID)

	// 5. 继续处理 RPC 请求
	return handler(ctx, req)
}
