package service

import (
	"context"

	v1 "ecommerce/api/product/v1"
	"ecommerce/internal/product/model"
	"ecommerce/internal/product/repository"
)

// BrandService is a Brand service.
type BrandService struct {
	repo repository.BrandRepo
}

// NewBrandService creates a new BrandService.
func NewBrandService(repo repository.BrandRepo) *BrandService {
	return &BrandService{repo: repo}
}

// CreateBrand creates a Brand.
func (s *BrandService) CreateBrand(ctx context.Context, req *v1.CreateBrandRequest) (*v1.BrandInfo, error) {
	brand := &model.Brand{
		Name: req.Name,
	}
	if req.HasLogo() {
		logo := req.GetLogo()
		brand.Logo = &logo
	}
	if req.HasDescription() {
		desc := req.GetDescription()
		brand.Description = &desc
	}
	if req.HasWebsite() {
		website := req.GetWebsite()
		brand.Website = &website
	}
	if req.HasSortOrder() {
		sortOrder := req.GetSortOrder()
		brand.SortOrder = &sortOrder
	}
	if req.HasIsVisible() {
		isVisible := req.GetIsVisible()
		brand.IsVisible = &isVisible
	}

	createdBrand, err := s.repo.CreateBrand(ctx, brand)
	if err != nil {
		return nil, err
	}
	return bizBrandToProto(createdBrand), nil
}

// UpdateBrand updates a Brand.
func (s *BrandService) UpdateBrand(ctx context.Context, req *v1.UpdateBrandRequest) (*v1.BrandInfo, error) {
	brand := &model.Brand{
		ID: req.Id,
	}
	if req.HasName() {
		name := req.GetName()
		brand.Name = &name
	}
	if req.HasLogo() {
		logo := req.GetLogo()
		brand.Logo = &logo
	}
	if req.HasDescription() {
		desc := req.GetDescription()
		brand.Description = &desc
	}
	if req.HasWebsite() {
		website := req.GetWebsite()
		brand.Website = &website
	}
	if req.HasSortOrder() {
		sortOrder := req.GetSortOrder()
		brand.SortOrder = &sortOrder
	}
	if req.HasIsVisible() {
		isVisible := req.GetIsVisible()
		brand.IsVisible = &isVisible
	}

	updatedBrand, err := s.repo.UpdateBrand(ctx, brand)
	if err != nil {
		return nil, err
	}
	return bizBrandToProto(updatedBrand), nil
}

// DeleteBrand deletes a Brand.
func (s *BrandService) DeleteBrand(ctx context.Context, id uint64) error {
	return s.repo.DeleteBrand(ctx, id)
}

// ListBrands lists Brands.
func (s *BrandService) ListBrands(ctx context.Context, req *v1.ListBrandsRequest) ([]*v1.BrandInfo, uint64, error) {
	var name *string
	if req.HasName() {
		n := req.GetName()
		name = &n
	}
	var isVisible *bool
	if req.HasIsVisible() {
		v := req.GetIsVisible()
		isVisible = &v
	}

	brands, total, err := s.repo.ListBrands(ctx, req.PageSize, req.PageNum, name, isVisible)
	if err != nil {
		return nil, 0, err
	}

	var brandInfos []*v1.BrandInfo
	for _, b := range brands {
		brandInfos = append(brandInfos, bizBrandToProto(b))
	}
	return brandInfos, total, nil
}

// bizBrandToProto converts biz.Brand to v1.BrandInfo
func bizBrandToProto(b *model.Brand) *v1.BrandInfo {
	if b == nil {
		return nil
	}
	res := &v1.BrandInfo{
		Id:   b.ID,
		Name: b.Name,
	}
	if b.Logo != nil {
		res.Logo = *b.Logo
	}
	if b.Description != nil {
		res.Description = *b.Description
	}
	if b.Website != nil {
		res.Website = *b.Website
	}
	if b.SortOrder != nil {
		res.SortOrder = *b.SortOrder
	}
	if b.IsVisible != nil {
		res.IsVisible = *b.IsVisible
	}
	return res
}
