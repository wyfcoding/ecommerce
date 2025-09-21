package service

import (
	"context"
	v1 "ecommerce/api/admin/v1"
	"ecommerce/internal/admin/biz"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AdminService 是 gRPC 服务的实现。
type AdminService struct {
	v1.UnimplementedAdminServer

	authUsecase    *biz.AuthUsecase
	productUsecase *biz.ProductUsecase
	// TODO: Add other usecases as needed
}

// NewAdminService 是 AdminService 的构造函数。
func NewAdminService(authUC *biz.AuthUsecase, productUC *biz.ProductUsecase) *AdminService {
	return &AdminService{
		authUsecase:    authUC,
		productUsecase: productUC,
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
	spu := &biz.Spu{
		CategoryID:    req.Spu.CategoryId,
		BrandID:       req.Spu.BrandId,
		Title:         req.Spu.Title,
		SubTitle:      req.Spu.SubTitle,
		MainImage:     req.Spu.MainImage,
		GalleryImages: req.Spu.GalleryImages,
		DetailHTML:    req.Spu.DetailHtml,
		Status:        req.Spu.Status,
	}

	skus := make([]*biz.Sku, 0, len(req.Skus))
	for _, s := range req.Skus {
		skus = append(skus, &biz.Sku{
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
		return nil, err
	}

	return &v1.CreateProductResponse{SpuId: createdSpu.ID}, nil
}
