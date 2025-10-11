package service

import (
	"context"
	v1 "ecommerce/api/product/v1"
	"ecommerce/internal/product/biz"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// bizCategoryToProto 将 biz.Category 领域模型转换为 v1.CategoryInfo API 模型。
func bizCategoryToProto(c *biz.Category) *v1.CategoryInfo {
	if c == nil {
		return nil
	}
	res := &v1.CategoryInfo{
		Id:       c.ID,
		ParentId: c.ParentID,
		Name:     c.Name,
		Level:    c.Level,
	}
	if c.Icon != nil {
		res.Icon = *c.Icon
	}
	if c.SortOrder != nil {
		res.SortOrder = *c.SortOrder
	}
	if c.IsVisible != nil {
		res.IsVisible = *c.IsVisible
	}
	return res
}

// CreateCategory 实现了创建商品分类的 RPC。
func (s *service) CreateCategory(ctx context.Context, req *v1.CreateCategoryRequest) (*v1.CategoryInfo, error) {
	bizCate := &biz.Category{
		ParentID: req.ParentId,
		Name:     req.Name,
	}
	if req.HasIcon() {
		icon := req.GetIcon()
		bizCate.Icon = &icon
	}
	if req.HasSortOrder() {
		sortOrder := req.GetSortOrder()
		bizCate.SortOrder = &sortOrder
	}
	if req.HasIsVisible() {
		isVisible := req.GetIsVisible()
		bizCate.IsVisible = &isVisible
	}

	created, err := s.categoryUsecase.CreateCategory(ctx, bizCate)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create category: %v", err)
	}
	return bizCategoryToProto(created), nil
}

// UpdateCategory 实现了更新商品分类的 RPC。
func (s *service) UpdateCategory(ctx context.Context, req *v1.UpdateCategoryRequest) (*v1.CategoryInfo, error) {
	bizCate := &biz.Category{
		ID: req.Id,
	}
	if req.HasParentId() {
		bizCate.ParentID = req.GetParentId()
	}
	if req.HasName() {
		bizCate.Name = req.GetName()
	}
	if req.HasIcon() {
		icon := req.GetIcon()
		bizCate.Icon = &icon
	}
	if req.HasSortOrder() {
		sortOrder := req.GetSortOrder()
		bizCate.SortOrder = &sortOrder
	}
	if req.HasIsVisible() {
		isVisible := req.GetIsVisible()
		bizCate.IsVisible = &isVisible
	}

	updated, err := s.categoryUsecase.UpdateCategory(ctx, bizCate)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update category: %v", err)
	}
	return bizCategoryToProto(updated), nil
}

// DeleteCategory 实现了删除商品分类的 RPC。
func (s *service) DeleteCategory(ctx context.Context, req *v1.DeleteCategoryRequest) (*emptypb.Empty, error) {
	if err := s.categoryUsecase.DeleteCategory(ctx, req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete category: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// ListCategories 实现了获取商品分类列表的 RPC。
func (s *service) ListCategories(ctx context.Context, req *v1.ListCategoriesRequest) (*v1.ListCategoriesResponse, error) {
	categories, err := s.categoryUsecase.ListCategories(ctx, req.ParentId)
	if err != nil {
		return nil, err
	}

	var categoryInfos []*v1.CategoryInfo
	for _, c := range categories {
		categoryInfos = append(categoryInfos, bizCategoryToProto(c))
	}

	return &v1.ListCategoriesResponse{
		Categories: categoryInfos,
	}, nil
}
