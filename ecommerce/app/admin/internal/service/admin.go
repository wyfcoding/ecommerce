package service

import (
	"context"

	v1 "ecommerce/ecommerce/api/admin/v1"
	"ecommerce/ecommerce/app/admin/internal/biz"
)

type AdminService struct {
	v1.UnimplementedAdminServer
	authUC    *biz.AuthUsecase
	productUC *biz.ProductUsecase
	orderUC   *biz.OrderUsecase // 新增 OrderUsecase 依赖
}

func NewAdminService(authUC *biz.AuthUsecase, productUC *biz.ProductUsecase, orderUC *biz.OrderUsecase) *AdminService {
	return &AdminService{authUC: authUC, productUC: productUC, orderUC: orderUC}
}

func (s *AdminService) AdminLogin(ctx context.Context, req *v1.AdminLoginRequest) (*v1.AdminLoginResponse, error) {
	token, err := s.authUC.Login(ctx, req.Username, req.Password)
	if err != nil {
		// ... 错误处理
	}
	return &v1.AdminLoginResponse{Token: token}, nil
}

func (s *AdminService) CreateProduct(ctx context.Context, req *v1.CreateProductRequest) (*v1.CreateProductResponse, error) {
	// 无需写权限校验，拦截器会自动处理
	// 直接调用 biz 层
	// 注意：这里的 req 和 res 都复用了 product.v1 的消息体
	res, err := s.productUC.CreateProduct(ctx, req)
	if err != nil {
		// ... 错误处理
	}
	return res, nil
}

// 假设 admin.proto 中也有一个 ListProducts，并复用了 product.v1 的消息体
func (s *AdminService) ListProducts(ctx context.Context, req *v1.ListProductsRequest) (*v1.ListProductsResponse, error) {
	res, err := s.productUC.ListProducts(ctx, req)
	if err != nil {
		// ... 错误处理
	}
	return res, nil
}

func (s *AdminService) ListOrders(ctx context.Context, req *v1.ListOrdersRequest) (*v1.ListOrdersResponse, error) {
	// 权限由拦截器自动处理
	return s.orderUC.ListOrders(ctx, req)
}

func (s *AdminService) GetOrderDetail(ctx context.Context, req *v1.GetOrderDetailRequest) (*v1.GetOrderDetailResponse, error) {
	// 权限由拦截器自动处理
	return s.orderUC.GetOrderDetail(ctx, req)
}

func (s *AdminService) ShipOrder(ctx context.Context, req *v1.ShipOrderRequest) (*v1.ShipOrderResponse, error) {
	// 权限由拦截器自动处理
	return s.orderUC.ShipOrder(ctx, req)
}
