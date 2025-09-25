package service

import (
	"context"
	v1 "ecommerce/api/admin/v1"
	couponv1 "ecommerce/api/coupon/v1"
	orderv1 "ecommerce/api/order/v1"
	productv1 "ecommerce/api/product/v1"
	reviewv1 "ecommerce/api/review/v1"
	userv1 "ecommerce/api/user/v1"
	"ecommerce/internal/admin/biz"
	couponbiz "ecommerce/internal/coupon/biz"
	orderbiz "ecommerce/internal/order/biz"
	productbiz "ecommerce/internal/product/biz"
	reviewbiz "ecommerce/internal/review/biz"
	userbiz "ecommerce/internal/user/biz"
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AdminService 是 gRPC 服务的实现。
type AdminService struct {
	v1.UnimplementedAdminServer

	authUsecase     *biz.AuthUsecase
	productUsecase  *productbiz.ProductUsecase
	userUsecase     *userbiz.UserUsecase
	orderUsecase    *orderbiz.OrderUsecase
	reviewUsecase   *reviewbiz.ReviewUsecase
	couponUsecase   *couponbiz.CouponUsecase
}

// NewAdminService 是 AdminService 的构造函数。
func NewAdminService(authUC *biz.AuthUsecase, productUC *productbiz.ProductUsecase, userUC *userbiz.UserUsecase, orderUC *orderbiz.OrderUsecase, reviewUC *reviewbiz.ReviewUsecase, couponUC *couponbiz.CouponUsecase) *AdminService {
	return &AdminService{
		authUsecase:    authUC,
		productUsecase: productUC,
		userUsecase:    userUC,
		orderUsecase:   orderUC,
		reviewUsecase:  reviewUC,
		couponUsecase:  couponUC,
	}
}

// AdminLogin 实现了管理员登录 RPC。
func (s *AdminService) AdminLogin(ctx context.Context, req *v1.AdminLoginRequest) (*v1.AdminLoginResponse, error) {
	token, err := s.authUsecase.AdminLogin(ctx, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, biz.ErrAdminUserNotFound) {
			return nil, status.Errorf(codes.NotFound, "管理员用户不存在: %v", err)
		}
		if errors.Is(err, biz.ErrAdminPasswordIncorrect) {
			return nil, status.Errorf(codes.Unauthenticated, "管理员密码不正确: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "管理员登录失败: %v", err)
	}
	return &v1.AdminLoginResponse{Token: token}, nil
}

// CreateProduct 实现了创建商品 RPC。
func (s *AdminService) CreateProduct(ctx context.Context, req *v1.CreateProductRequest) (*v1.CreateProductResponse, error) {
	spu := &productbiz.Spu{
		CategoryID:    req.Spu.CategoryId,
		BrandID:       req.Spu.BrandId,
		Title:         req.Spu.Title,
		SubTitle:      req.Spu.SubTitle,
		MainImage:     req.Spu.MainImage,
		GalleryImages: req.Spu.GalleryImages,
		DetailHTML:    req.Spu.DetailHtml,
		Status:        req.Spu.Status,
	}

	skus := make([]*productbiz.Sku, 0, len(req.Skus))
	for _, s := range req.Skus {
		skus = append(skus, &productbiz.Sku{
			SpuID:         s.SpuId,
			Title:         s.Title,
			Price:         s.Price,
			OriginalPrice: s.OriginalPrice,
			Stock:         s.Stock,
			Image:         s.Image,
			Specs:         s.Specs,
			Status:        s.Status,
		})
	}

	createdSpu, _, err := s.productUsecase.CreateProduct(ctx, spu, skus)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "创建商品失败: %v", err)
	}

	return &v1.CreateProductResponse{SpuId: createdSpu.ID}, nil
}

// ListUsers 实现了列出用户 RPC。
func (s *AdminService) ListUsers(ctx context.Context, req *v1.ListUsersRequest) (*v1.ListUsersResponse, error) {
	users, total, err := s.userUsecase.ListUsers(ctx, req.Page, req.PageSize, req.GetSearchQuery(), req.GetStatusFilter())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "列出用户失败: %v", err)
	}

	pbUsers := make([]*userv1.UserInfo, len(users))
	for i, user := range users {
		pbUsers[i] = &userv1.UserInfo{
			Id:       fmt.Sprintf("%d", user.ID),
			Username: user.Username,
			Email:    user.Email,
			Phone:    user.Phone,
			Status:   user.Status,
			// ... other fields
		}
	}

	return &v1.ListUsersResponse{Users: pbUsers, Total: total},
		nil
}

// GetUserDetails 实现了获取用户详情 RPC。
func (s *AdminService) GetUserDetails(ctx context.Context, req *v1.GetUserDetailsRequest) (*v1.GetUserDetailsResponse, error) {
	user, err := s.userUsecase.GetUserByID(ctx, req.UserId)
	if err != nil {
		if errors.Is(err, userbiz.ErrUserNotFound) {
			return nil, status.Errorf(codes.NotFound, "用户不存在: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "获取用户详情失败: %v", err)
	}

	return &v1.GetUserDetailsResponse{
		User: &userv1.UserInfo{
			Id:       fmt.Sprintf("%d", user.ID),
			Username: user.Username,
			Email:    user.Email,
			Phone:    user.Phone,
			Status:   user.Status,
			// ... other fields
		},
	},
		nil
}

// UpdateUserStatus 实现了更新用户状态 RPC。
func (s *AdminService) UpdateUserStatus(ctx context.Context, req *v1.UpdateUserStatusRequest) (*v1.UpdateUserStatusResponse, error) {
	user, err := s.userUsecase.UpdateUserStatus(ctx, req.UserId, req.Status)
	if err != nil {
		if errors.Is(err, userbiz.ErrUserNotFound) {
			return nil, status.Errorf(codes.NotFound, "用户不存在: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "更新用户状态失败: %v", err)
	}

	return &v1.UpdateUserStatusResponse{
		User: &userv1.UserInfo{
			Id:       fmt.Sprintf("%d", user.ID),
			Username: user.Username,
			Email:    user.Email,
			Phone:    user.Phone,
			Status:   user.Status,
			// ... other fields
		},
	},
		nil
}

// ListProducts 实现了列出商品 RPC。
func (s *AdminService) ListProducts(ctx context.Context, req *v1.ListProductsRequest) (*v1.ListProductsResponse, error) {
	spus, total, err := s.productUsecase.ListSpus(ctx, req.Page, req.PageSize, req.GetSearchQuery(), req.GetCategoryId(), req.GetStatusFilter())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "列出商品失败: %v", err)
	}

	pbProducts := make([]*productv1.SpuInfo, len(spus))
	for i, spu := range spus {
		pbProducts[i] = &productv1.SpuInfo{
			Id:         fmt.Sprintf("%d", spu.ID),
			Title:      spu.Title,
			MainImage:  spu.MainImage,
			Status:     spu.Status,
			CategoryId: spu.CategoryID,
			BrandId:    spu.BrandID,
			// ... other fields
		}
	}

	return &v1.ListProductsResponse{Products: pbProducts, Total: total},
		nil
}

// UpdateProduct 实现了更新商品 RPC。
func (s *AdminService) UpdateProduct(ctx context.Context, req *v1.UpdateProductRequest) (*v1.UpdateProductResponse, error) {
	spuID, err := strconv.ParseUint(req.ProductId, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的商品ID: %v", err)
	}

	spu := &productbiz.Spu{
		ID:            uint(spuID),
		CategoryID:    req.Spu.CategoryId,
		BrandID:       req.Spu.BrandId,
		Title:         req.Spu.Title,
		SubTitle:      req.Spu.SubTitle,
		MainImage:     req.Spu.MainImage,
		GalleryImages: req.Spu.GalleryImages,
		DetailHTML:    req.Spu.DetailHtml,
		Status:        req.Spu.Status,
	}

	skus := make([]*productbiz.Sku, 0, len(req.Skus))
	for _, s := range req.Skus {
		skus = append(skus, &productbiz.Sku{
			SpuID:         s.SpuId,
			Title:         s.Title,
			Price:         s.Price,
			OriginalPrice: s.OriginalPrice,
			Stock:         s.Stock,
			Image:         s.Image,
			Specs:         s.Specs,
			Status:        s.Status,
		})
	}

	updatedSpu, err := s.productUsecase.UpdateProduct(ctx, spu, skus)
	if err != nil {
		if errors.Is(err, productbiz.ErrSpuNotFound) {
			return nil, status.Errorf(codes.NotFound, "商品不存在: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "更新商品失败: %v", err)
	}

	return &v1.UpdateProductResponse{
		Spu: &productv1.SpuInfo{
			Id:         fmt.Sprintf("%d", updatedSpu.ID),
			Title:      updatedSpu.Title,
			MainImage:  updatedSpu.MainImage,
			Status:     updatedSpu.Status,
			CategoryId: updatedSpu.CategoryID,
			BrandId:    updatedSpu.BrandID,
			// ... other fields
		},
	},
		nil
}

// DeleteProduct 实现了删除商品 RPC。
func (s *AdminService) DeleteProduct(ctx context.Context, req *v1.DeleteProductRequest) (*v1.DeleteProductResponse, error) {
	spuID, err := strconv.ParseUint(req.ProductId, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的商品ID: %v", err)
	}

	if err := s.productUsecase.DeleteSpu(ctx, uint(spuID)); err != nil {
		if errors.Is(err, productbiz.ErrSpuNotFound) {
			return nil, status.Errorf(codes.NotFound, "商品不存在: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "删除商品失败: %v", err)
	}

	return &v1.DeleteProductResponse{Success: true, Message: "商品删除成功"}, nil
}

// UpdateOrderStatus 实现了更新订单状态 RPC。
func (s *AdminService) UpdateOrderStatus(ctx context.Context, req *v1.UpdateOrderStatusRequest) (*v1.UpdateOrderStatusResponse, error) {
	orderID, err := strconv.ParseUint(req.OrderId, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的订单ID: %v", err)
	}

	updatedOrder, err := s.orderUsecase.UpdateOrderStatus(ctx, uint(orderID), req.Status)
	if err != nil {
		if errors.Is(err, orderbiz.ErrOrderNotFound) {
			return nil, status.Errorf(codes.NotFound, "订单不存在: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "更新订单状态失败: %v", err)
	}

	return &v1.UpdateOrderStatusResponse{
		Order: &orderv1.OrderInfo{
			Id:     fmt.Sprintf("%d", updatedOrder.ID),
			Status: updatedOrder.Status,
			// ... other fields
		},
	},
		nil
}

// ListReviews 实现了列出评论 RPC。
func (s *AdminService) ListReviews(ctx context.Context, req *v1.ListReviewsRequest) (*v1.ListReviewsResponse, error) {
	reviews, total, err := s.reviewUsecase.ListReviews(ctx, req.Page, req.PageSize, req.GetProductId(), req.GetUserId(), req.GetStatusFilter())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "列出评论失败: %v", err)
	}

	pbReviews := make([]*reviewv1.ReviewInfo, len(reviews))
	for i, review := range reviews {
		pbReviews[i] = &reviewv1.ReviewInfo{
			Id:        fmt.Sprintf("%d", review.ID),
			ProductId: review.ProductID,
			UserId:    review.UserID,
			Rating:    review.Rating,
			Title:     review.Title,
			Content:   review.Content,
			CreatedAt: timestamppb.New(review.CreatedAt),
			UpdatedAt: timestamppb.New(review.UpdatedAt),
		}
	}

	return &v1.ListReviewsResponse{Reviews: pbReviews, Total: total},
		nil
}

// ModerateReview 实现了审核评论 RPC。
func (s *AdminService) ModerateReview(ctx context.Context, req *v1.ModerateReviewRequest) (*v1.ModerateReviewResponse, error) {
	reviewID, err := strconv.ParseUint(req.ReviewId, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的评论ID: %v", err)
	}

	updatedReview, err := s.reviewUsecase.ModerateReview(ctx, uint(reviewID), req.Status, req.GetModerationNotes())
	if err != nil {
		if errors.Is(err, reviewbiz.ErrReviewNotFound) {
			return nil, status.Errorf(codes.NotFound, "评论不存在: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "审核评论失败: %v", err)
	}

	return &v1.ModerateReviewResponse{
		Review: &reviewv1.ReviewInfo{
			Id:        fmt.Sprintf("%d", updatedReview.ID),
			ProductId: updatedReview.ProductID,
			UserId:    updatedReview.UserID,
			Rating:    updatedReview.Rating,
			Title:     updatedReview.Title,
			Content:   updatedReview.Content,
			CreatedAt: timestamppb.New(updatedReview.CreatedAt),
			UpdatedAt: timestamppb.New(updatedReview.UpdatedAt),
		},
	},
		nil
}

// CreateCoupon 实现了创建优惠券 RPC。
func (s *AdminService) CreateCoupon(ctx context.Context, req *v1.CreateCouponRequest) (*v1.CreateCouponResponse, error) {
	coupon := &couponbiz.Coupon{
		Code:          req.Code,
		Description:   req.Description,
		DiscountValue: req.DiscountValue,
		DiscountType:  req.Type,
		ValidFrom:     req.ValidFrom.AsTime(),
		ValidTo:       req.ValidTo.AsTime(),
		MaxUsage:      req.GetMaxUsage(),
		IsActive:      true,
	}

	createdCoupon, err := s.couponUsecase.CreateCoupon(ctx, coupon)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "创建优惠券失败: %v", err)
	}

	return &v1.CreateCouponResponse{
		Coupon: &couponv1.CouponInfo{
			Id:            fmt.Sprintf("%d", createdCoupon.ID),
			Code:          createdCoupon.Code,
			Description:   createdCoupon.Description,
			DiscountValue: createdCoupon.DiscountValue,
			Type:          createdCoupon.DiscountType,
			ValidFrom:     timestamppb.New(createdCoupon.ValidFrom),
			ValidTo:       timestamppb.New(createdCoupon.ValidTo),
			IsActive:      createdCoupon.IsActive,
		},
	},
		nil
}

// ListCoupons 实现了列出优惠券 RPC。
func (s *AdminService) ListCoupons(ctx context.Context, req *v1.ListCouponsRequest) (*v1.ListCouponsResponse, error) {
	coupons, total, err := s.couponUsecase.ListCoupons(ctx, req.Page, req.PageSize, req.GetActiveOnly(), req.GetSearchQuery())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "列出优惠券失败: %v", err)
	}

	pbCoupons := make([]*couponv1.CouponInfo, len(coupons))
	for i, coupon := range coupons {
		pbCoupons[i] = &couponv1.CouponInfo{
			Id:            fmt.Sprintf("%d", coupon.ID),
			Code:          coupon.Code,
			Description:   coupon.Description,
			DiscountValue: coupon.DiscountValue,
			Type:          coupon.DiscountType,
			ValidFrom:     timestamppb.New(coupon.ValidFrom),
			ValidTo:       timestamppb.New(coupon.ValidTo),
			IsActive:      coupon.IsActive,
		}
	}

	return &v1.ListCouponsResponse{Coupons: pbCoupons, Total: total},
		nil
}

// UpdateCoupon 实现了更新优惠券 RPC。
func (s *AdminService) UpdateCoupon(ctx context.Context, req *v1.UpdateCouponRequest) (*v1.UpdateCouponResponse, error) {
	couponID, err := strconv.ParseUint(req.CouponId, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的优惠券ID: %v", err)
	}

	coupon := &couponbiz.Coupon{
		ID:            uint(couponID),
		Code:          req.Code,
		Description:   req.Description,
		DiscountValue: req.DiscountValue,
		DiscountType:  req.Type,
		ValidFrom:     req.ValidFrom.AsTime(),
		ValidTo:       req.ValidTo.AsTime(),
		IsActive:      req.IsActive,
		MaxUsage:      req.GetMaxUsage(),
	}

	updatedCoupon, err := s.couponUsecase.UpdateCoupon(ctx, coupon)
	if err != nil {
		if errors.Is(err, couponbiz.ErrCouponNotFound) {
			return nil, status.Errorf(codes.NotFound, "优惠券不存在: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "更新优惠券失败: %v", err)
	}

	return &v1.UpdateCouponResponse{
		Coupon: &couponv1.CouponInfo{
			Id:            fmt.Sprintf("%d", updatedCoupon.ID),
			Code:          updatedCoupon.Code,
			Description:   updatedCoupon.Description,
			DiscountValue: updatedCoupon.DiscountValue,
			Type:          updatedCoupon.DiscountType,
			ValidFrom:     timestamppb.New(updatedCoupon.ValidFrom),
			ValidTo:       timestamppb.New(updatedCoupon.ValidTo),
			IsActive:      updatedCoupon.IsActive,
		},
	},
		nil
}

// DeleteCoupon 实现了删除优惠券 RPC。
func (s *AdminService) DeleteCoupon(ctx context.Context, req *v1.DeleteCouponRequest) (*v1.DeleteCouponResponse, error) {
	couponID, err := strconv.ParseUint(req.CouponId, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的优惠券ID: %v", err)
	}

	if err := s.couponUsecase.DeleteCoupon(ctx, uint(couponID)); err != nil {
		if errors.Is(err, couponbiz.ErrCouponNotFound) {
			return nil, status.Errorf(codes.NotFound, "优惠券不存在: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "删除优惠券失败: %v", err)
	}

	return &v1.DeleteCouponResponse{Success: true, Message: "优惠券删除成功"}, nil
}

// GetOrderDetail 实现了获取订单详情 RPC。
func (s *AdminService) GetOrderDetail(ctx context.Context, req *v1.GetOrderDetailRequest) (*v1.GetOrderDetailResponse, error) {
	orderID, err := strconv.ParseUint(req.OrderId, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的订单ID: %v", err)
	}

	order, items, err := s.orderUsecase.GetOrderDetail(ctx, uint(orderID))
	if err != nil {
		if errors.Is(err, orderbiz.ErrOrderNotFound) {
			return nil, status.Errorf(codes.NotFound, "订单不存在: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "获取订单详情失败: %v", err)
	}

	pbOrder := &orderv1.OrderInfo{
		Id:     fmt.Sprintf("%d", order.ID),
		UserId: order.UserID,
		Status: order.Status,
		// ... other fields
	}

	pbItems := make([]*orderv1.OrderItem, len(items))
	for i, item := range items {
		pbItems[i] = &orderv1.OrderItem{
			Id:        fmt.Sprintf("%d", item.ID),
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.Price,
			// ... other fields
		}
	}

	return &v1.GetOrderDetailResponse{Order: pbOrder, Items: pbItems}, nil
}

// ListOrders 实现了列出订单 RPC。
func (s *AdminService) ListOrders(ctx context.Context, req *v1.ListOrdersRequest) (*v1.ListOrdersResponse, error) {
	orders, total, err := s.orderUsecase.ListOrders(ctx, req.Page, req.PageSize, req.GetUserId(), req.GetStatus())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "列出订单失败: %v", err)
	}

	pbOrders := make([]*orderv1.OrderInfo, len(orders))
	for i, order := range orders {
		pbOrders[i] = &orderv1.OrderInfo{
			Id:     fmt.Sprintf("%d", order.ID),
			UserId: order.UserID,
			Status: order.Status,
			// ... other fields
		}
	}

	return &v1.ListOrdersResponse{Orders: pbOrders, Total: total},
		nil
}

// ShipOrder 实现了发货订单 RPC。
func (s *AdminService) ShipOrder(ctx context.Context, req *v1.ShipOrderRequest) (*v1.ShipOrderResponse, error) {
	orderID, err := strconv.ParseUint(req.OrderId, 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的订单ID: %v", err)
	}

	if err := s.orderUsecase.ShipOrder(ctx, uint(orderID), req.TrackingCompany, req.TrackingNumber); err != nil {
		if errors.Is(err, orderbiz.ErrOrderNotFound) {
			return nil, status.Errorf(codes.NotFound, "订单不存在: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "发货失败: %v", err)
	}

	return &v1.ShipOrderResponse{}, nil
}

// --- RBAC 管理 ---
// CreateRole 实现了创建角色 RPC。
func (s *AdminService) CreateRole(ctx context.Context, req *v1.CreateRoleRequest) (*v1.Role, error) {
	// TODO: Implement RBAC logic
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

// ListRoles 实现了列出角色 RPC。
func (s *AdminService) ListRoles(ctx context.Context, req *v1.ListRolesRequest) (*v1.ListRolesResponse, error) {
	// TODO: Implement RBAC logic
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

// UpdateRolePermissions 实现了更新角色权限 RPC。
func (s *AdminService) UpdateRolePermissions(ctx context.Context, req *v1.UpdateRolePermissionsRequest) (*v1.UpdateRolePermissionsResponse, error) {
	// TODO: Implement RBAC logic
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

// ListPermissions 实现了列出权限 RPC。
func (s *AdminService) ListPermissions(ctx context.Context, req *v1.ListPermissionsRequest) (*v1.ListPermissionsResponse, error) {
	// TODO: Implement RBAC logic
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

// CreateAdminUser 实现了创建管理员用户 RPC。
func (s *AdminService) CreateAdminUser(ctx context.Context, req *v1.CreateAdminUserRequest) (*v1.AdminUser, error) {
	// TODO: Implement RBAC logic
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

// ListAdminUsers 实现了列出管理员用户 RPC。
func (s *AdminService) ListAdminUsers(ctx context.Context, req *v1.ListAdminUsersRequest) (*v1.ListAdminUsersResponse, error) {
	// TODO: Implement RBAC logic
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}

// UpdateUserRoles 实现了更新用户角色 RPC。
func (s *AdminService) UpdateUserRoles(ctx context.Context, req *v1.UpdateUserRolesRequest) (*v1.UpdateUserRolesResponse, error) {
	// TODO: Implement RBAC logic
	return nil, status.Errorf(codes.Unimplemented, "方法未实现")
}
