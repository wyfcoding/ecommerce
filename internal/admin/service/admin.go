package service

import (
	"context"
	"errors"
	"strconv"

	v1 "ecommerce/api/admin/v1"
	"ecommerce/internal/admin/biz"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// --- SPU Converters ---

func protoToBizSpu(spu *v1.SpuInfo) *biz.Spu {
	if spu == nil {
		return nil
	}
	return &biz.Spu{
		ID:            spu.SpuId,
		CategoryID:    &spu.CategoryId,
		BrandID:       &spu.BrandId,
		Title:         &spu.Title,
		SubTitle:      &spu.SubTitle,
		MainImage:     &spu.MainImage,
		GalleryImages: spu.GalleryImages,
		DetailHTML:    &spu.DetailHtml,
		Status:        &spu.Status,
	}
}

// protoToBizSku 将 productV1.SkuInfo 转换为 biz.Sku。
func protoToBizSku(sku *v1.SkuInfo) *biz.Sku {
	if sku == nil {
		return nil
	}
	return &biz.Sku{
		ID:            sku.SkuId,
		SpuID:         sku.SpuId,
		Title:         &sku.Title,
		Price:         &sku.Price,
		OriginalPrice: &sku.OriginalPrice,
		Stock:         &sku.Stock,
		Image:         &sku.Image,
		Specs:         sku.Specs,
		Status:        &sku.Status,
	}
}

// AdminLogin 实现了管理员登录的 RPC。
func (s *AdminService) AdminLogin(ctx context.Context, req *v1.AdminLoginRequest) (*v1.AdminLoginResponse, error) {
	token, err := s.authUsecase.AdminLogin(ctx, req.Username, req.Password)
	if err != nil {
		// TODO: 更细致的错误处理，例如用户不存在、密码错误等
		return nil, status.Errorf(codes.Unauthenticated, "登录失败: %v", err)
	}
	return &v1.AdminLoginResponse{Token: token}, nil
}

// GetOrderDetail 实现了获取订单详情的 RPC。
func (s *AdminService) GetOrderDetail(ctx context.Context, req *v1.GetOrderDetailRequest) (*v1.GetOrderDetailResponse, error) {
	// TODO: 实现业务逻辑，调用 order service
	return nil, status.Errorf(codes.Unimplemented, "method GetOrderDetail not implemented")
}

// CreateProduct 实现了创建商品的 RPC。
func (s *AdminService) CreateProduct(ctx context.Context, req *v1.CreateProductRequest) (*v1.CreateProductResponse, error) {
	// TODO: 实现业务逻辑，调用 product usecase
	return nil, status.Errorf(codes.Unimplemented, "method CreateProduct not implemented")
}

// ListOrders 实现了获取订单列表的 RPC。
func (s *AdminService) ListOrders(ctx context.Context, req *v1.ListOrdersRequest) (*v1.ListOrdersResponse, error) {
	// TODO: 实现业务逻辑，调用 order service
	return nil, status.Errorf(codes.Unimplemented, "method ListOrders not implemented")
}

// ShipOrder 实现了发货的 RPC。
func (s *AdminService) ShipOrder(ctx context.Context, req *v1.ShipOrderRequest) (*v1.ShipOrderResponse, error) {
	// TODO: 实现业务逻辑，更新订单状态，调用物流服务
	return nil, status.Errorf(codes.Unimplemented, "method ShipOrder not implemented")
}

// CreateRole 实现了创建角色的 RPC。
func (s *AdminService) CreateRole(ctx context.Context, req *v1.CreateRoleRequest) (*v1.Role, error) {
	// TODO: 实现业务逻辑
	return nil, status.Errorf(codes.Unimplemented, "method CreateRole not implemented")
}

// ListRoles 实现了获取角色列表的 RPC。
func (s *AdminService) ListRoles(ctx context.Context, req *v1.ListRolesRequest) (*v1.ListRolesResponse, error) {
	// TODO: 实现业务逻辑
	return nil, status.Errorf(codes.Unimplemented, "method ListRoles not implemented")
}

// UpdateRolePermissions 实现了更新角色权限的 RPC。
func (s *AdminService) UpdateRolePermissions(ctx context.Context, req *v1.UpdateRolePermissionsRequest) (*v1.UpdateRolePermissionsResponse, error) {
	// TODO: 实现业务逻辑
	return nil, status.Errorf(codes.Unimplemented, "method UpdateRolePermissions not implemented")
}

// ListPermissions 实现了获取权限列表的 RPC。
func (s *AdminService) ListPermissions(ctx context.Context, req *v1.ListPermissionsRequest) (*v1.ListPermissionsResponse, error) {
	// TODO: 实现业务逻辑
	return nil, status.Errorf(codes.Unimplemented, "method ListPermissions not implemented")
}

// CreateAdminUser 实现了创建管理员用户的 RPC。
func (s *AdminService) CreateAdminUser(ctx context.Context, req *v1.CreateAdminUserRequest) (*v1.AdminUser, error) {
	// TODO: 实现业务逻辑
	return nil, status.Errorf(codes.Unimplemented, "method CreateAdminUser not implemented")
}

// ListAdminUsers 实现了获取管理员用户列表的 RPC。
func (s *AdminService) ListAdminUsers(ctx context.Context, req *v1.ListAdminUsersRequest) (*v1.ListAdminUsersResponse, error) {
	// TODO: 实现业务逻辑，获取管理员用户列表
	return nil, status.Errorf(codes.Unimplemented, "method ListAdminUsers not implemented")
}

// UpdateUserRoles 实现了更新用户角色的 RPC。
func (s *AdminService) UpdateUserRoles(ctx context.Context, req *v1.UpdateUserRolesRequest) (*v1.UpdateUserRolesResponse, error) {
	// TODO: 实现业务逻辑
	return nil, status.Errorf(codes.Unimplemented, "method UpdateUserRoles not implemented")
}

// GetAdminUserInfo 根据adminUserID获取管理员信息并返回。
func (s *AdminService) GetAdminUserInfo(ctx context.Context, req *v1.GetAdminUserInfoRequest) (*v1.AdminUserInfoResponse, error) {
	adminUserID, err := strconv.ParseInt(req.GetId(), 10, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "无效的管理员用户ID: %v", err)
	}

	adminUser, err := s.authUsecase.GetAdminUserByID(ctx, adminUserID)
	if err != nil {
		if errors.Is(err, biz.ErrAdminUserNotFound) {
			return nil, status.Errorf(codes.NotFound, "管理员用户不存在: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "获取管理员信息失败: %v", err)
	}

	return &v1.AdminUserInfoResponse{
		Id:       adminUser.ID,
		Username: adminUser.Username,
		Name:     adminUser.Name,
		Status:   adminUser.Status,
	}, nil
}
