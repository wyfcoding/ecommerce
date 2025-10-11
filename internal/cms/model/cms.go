package biz

import (
	"context"
	"errors"
	"time"
)

// ErrContentPageNotFound is a specific error for when a content page is not found.
var ErrContentPageNotFound = errors.New("content page not found")

// ErrContentBlockNotFound is a specific error for when a content block is not found.
var ErrContentBlockNotFound = errors.New("content block not found")

// ContentPage represents a content page in the business layer.
type ContentPage struct {
	ID          uint
	Title       string
	Slug        string // URL-friendly identifier
	ContentHTML string
	Status      string // e.g., DRAFT, PUBLISHED, ARCHIVED
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ContentBlock represents a reusable content block in the business layer.
type ContentBlock struct {
	ID          uint
	Name        string // Internal name for the block
	ContentHTML string
	Type        string // e.g., HTML, MARKDOWN, JSON
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CmsRepo defines the data storage interface for CMS data.
// The business layer depends on this interface, not on a concrete data implementation.
type CmsRepo interface {
	CreateContentPage(ctx context.Context, page *ContentPage) (*ContentPage, error)
	GetContentPageByID(ctx context.Context, id uint) (*ContentPage, error)
	GetContentPageBySlug(ctx context.Context, slug string) (*ContentPage, error)
	UpdateContentPage(ctx context.Context, page *ContentPage) (*ContentPage, error)
	DeleteContentPage(ctx context.Context, id uint) error
	ListContentPages(ctx context.Context, statusFilter string, pageSize, pageToken int32) ([]*ContentPage, int32, error)

	CreateContentBlock(ctx context.Context, block *ContentBlock) (*ContentBlock, error)
	GetContentBlockByID(ctx context.Context, id uint) (*ContentBlock, error)
	GetContentBlockByName(ctx context.Context, name string) (*ContentBlock, error)
	UpdateContentBlock(ctx context.Context, block *ContentBlock) (*ContentBlock, error)
	DeleteContentBlock(ctx context.Context, id uint) error
}

// CmsUsecase is the use case for CMS operations.
// It orchestrates the business logic.
type CmsUsecase struct {
	repo CmsRepo
	// You can also inject other dependencies like a logger
}

// NewCmsUsecase creates a new CmsUsecase.
func NewCmsUsecase(repo CmsRepo) *CmsUsecase {
	return &CmsUsecase{repo: repo}
}

// CreateContentPage creates a new content page.
func (uc *CmsUsecase) CreateContentPage(ctx context.Context, title, slug, contentHTML, status string) (*ContentPage, error) {
	page := &ContentPage{
		Title:       title,
		Slug:        slug,
		ContentHTML: contentHTML,
		Status:      status,
	}
	return uc.repo.CreateContentPage(ctx, page)
}

// GetContentPage retrieves a content page by ID or slug.
func (uc *CmsUsecase) GetContentPage(ctx context.Context, id uint, slug string) (*ContentPage, error) {
	if id != 0 {
		return uc.repo.GetContentPageByID(ctx, id)
	} else if slug != "" {
		return uc.repo.GetContentPageBySlug(ctx, slug)
	}
	return nil, errors.New("either ID or Slug must be provided")
}

// UpdateContentPage updates an existing content page.
func (uc *CmsUsecase) UpdateContentPage(ctx context.Context, page *ContentPage) (*ContentPage, error) {
	return uc.repo.UpdateContentPage(ctx, page)
}

// DeleteContentPage deletes a content page.
func (uc *CmsUsecase) DeleteContentPage(ctx context.Context, id uint) error {
	return uc.repo.DeleteContentPage(ctx, id)
}

// ListContentPages lists content pages with optional filters.
func (uc *CmsUsecase) ListContentPages(ctx context.Context, statusFilter string, pageSize, pageToken int32) ([]*ContentPage, int32, error) {
	return uc.repo.ListContentPages(ctx, statusFilter, pageSize, pageToken)
}

// CreateContentBlock creates a new content block.
func (uc *CmsUsecase) CreateContentBlock(ctx context.Context, name, contentHTML, blockType string) (*ContentBlock, error) {
	block := &ContentBlock{
		Name:        name,
		ContentHTML: contentHTML,
		Type:        blockType,
	}
	return uc.repo.CreateContentBlock(ctx, block)
}

// GetContentBlock retrieves a content block by ID or name.
func (uc *CmsUsecase) GetContentBlock(ctx context.Context, id uint, name string) (*ContentBlock, error) {
	if id != 0 {
		return uc.repo.GetContentBlockByID(ctx, id)
	} else if name != "" {
		return uc.repo.GetContentBlockByName(ctx, name)
	}
	return nil, errors.New("either ID or Name must be provided")
}

// UpdateContentBlock updates an existing content block.
func (uc *CmsUsecase) UpdateContentBlock(ctx context.Context, block *ContentBlock) (*ContentBlock, error) {
	return uc.repo.UpdateContentBlock(ctx, block)
}

// DeleteContentBlock deletes a content block.
func (uc *CmsUsecase) DeleteContentBlock(ctx context.Context, id uint) error {
	return uc.repo.DeleteContentBlock(ctx, id)
}
