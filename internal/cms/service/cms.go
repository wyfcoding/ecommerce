package service

import (
	"context"
	"errors"

	"ecommerce/internal/cms/model"
	"ecommerce/internal/cms/repository"
)

// ErrContentPageNotFound is a specific error for when a content page is not found.
var ErrContentPageNotFound = errors.New("content page not found")

// ErrContentBlockNotFound is a specific error for when a content block is not found.
var ErrContentBlockNotFound = errors.New("content block not found")

// CmsService is the use case for CMS operations.
// It orchestrates the business logic.
type CmsService struct {
	repo repository.CmsRepo
	// You can also inject other dependencies like a logger
}

// NewCmsService creates a new CmsService.
func NewCmsService(repo repository.CmsRepo) *CmsService {
	return &CmsService{repo: repo}
}

// CreateContentPage creates a new content page.
func (s *CmsService) CreateContentPage(ctx context.Context, title, slug, contentHTML, status string) (*model.ContentPage, error) {
	page := &model.ContentPage{
		Title:       title,
		Slug:        slug,
		ContentHTML: contentHTML,
		Status:      status,
	}
	return s.repo.CreateContentPage(ctx, page)
}

// GetContentPage retrieves a content page by ID or slug.
func (s *CmsService) GetContentPage(ctx context.Context, id uint, slug string) (*model.ContentPage, error) {
	if id != 0 {
		return s.repo.GetContentPageByID(ctx, id)
	} else if slug != "" {
		return s.repo.GetContentPageBySlug(ctx, slug)
	}
	return nil, errors.New("either ID or Slug must be provided")
}

// UpdateContentPage updates an existing content page.
func (s *CmsService) UpdateContentPage(ctx context.Context, page *model.ContentPage) (*model.ContentPage, error) {
	return s.repo.UpdateContentPage(ctx, page)
}

// DeleteContentPage deletes a content page.
func (s *CmsService) DeleteContentPage(ctx context.Context, id uint) error {
	return s.repo.DeleteContentPage(ctx, id)
}

// ListContentPages lists content pages with optional filters.
func (s *CmsService) ListContentPages(ctx context.Context, statusFilter string, pageSize, pageToken int32) ([]*model.ContentPage, int32, error) {
	return s.repo.ListContentPages(ctx, statusFilter, pageSize, pageToken)
}

// CreateContentBlock creates a new content block.
func (s *CmsService) CreateContentBlock(ctx context.Context, name, contentHTML, blockType string) (*model.ContentBlock, error) {
	block := &model.ContentBlock{
		Name:        name,
		ContentHTML: contentHTML,
		Type:        blockType,
	}
	return s.repo.CreateContentBlock(ctx, block)
}

// GetContentBlock retrieves a content block by ID or name.
func (s *CmsService) GetContentBlock(ctx context.Context, id uint, name string) (*model.ContentBlock, error) {
	if id != 0 {
		return s.repo.GetContentBlockByID(ctx, id)
	} else if name != "" {
		return s.repo.GetContentBlockByName(ctx, name)
	}
	return nil, errors.New("either ID or Name must be provided")
}

// UpdateContentBlock updates an existing content block.
func (s *CmsService) UpdateContentBlock(ctx context.Context, block *model.ContentBlock) (*model.ContentBlock, error) {
	return s.repo.UpdateContentBlock(ctx, block)
}

// DeleteContentBlock deletes a content block.
func (s *CmsService) DeleteContentBlock(ctx context.Context, id uint) error {
	return s.repo.DeleteContentBlock(ctx, id)
}
