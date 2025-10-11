package service

import (
	v1 "ecommerce/api/product/v1"
	"ecommerce/internal/product/biz"
)

// service 结构体同时实现了 CategoryService 和 ProductService
// 这样可以简化注册过程
type service struct {
	v1.UnimplementedCategoryServiceServer
	v1.UnimplementedProductServiceServer

	categoryUsecase *biz.CategoryUsecase
	productUsecase  *biz.ProductUsecase
	brandUsecase    *biz.BrandUsecase  // 新增：品牌用例
	reviewUsecase   *biz.ReviewUsecase // 新增：评论用例
}

// NewService 是 service 的构造函数
func NewService(cc *biz.CategoryUsecase, pc *biz.ProductUsecase, bc *biz.BrandUsecase, rc *biz.ReviewUsecase) *service {
	return &service{
		categoryUsecase: cc,
		productUsecase:  pc,
		brandUsecase:    bc,
		reviewUsecase:   rc,
	}
}
