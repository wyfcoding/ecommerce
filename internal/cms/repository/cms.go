package repository

import (
	"context"

	"ecommerce/internal/cms/model"
)

// CmsRepo defines the data storage interface for CMS data.
// The business layer depends on this interface, not on a concrete data implementation.
type CmsRepo interface {
	CreateContentPage(ctx context.Context, page *model.ContentPage) (*model.ContentPage, error)
	GetContentPageByID(ctx context.Context, id uint) (*model.ContentPage, error)
	GetContentPageBySlug(ctx context.Context, slug string) (*model.ContentPage, error)
	UpdateContentPage(ctx context.Context, page *model.ContentPage) (*model.ContentPage, error)
	DeleteContentPage(ctx context.Context, id uint) error
	ListContentPages(ctx context.Context, statusFilter string, pageSize, pageToken int32) ([]*model.ContentPage, int32, error)

	CreateContentBlock(ctx context.Context, block *model.ContentBlock) (*model.ContentBlock, error)
	GetContentBlockByID(ctx context.Context, id uint) (*model.ContentBlock, error)
	GetContentBlockByName(ctx context.Context, name string) (*model.ContentBlock, error)
	UpdateContentBlock(ctx context.Context, block *model.ContentBlock) (*model.ContentBlock, error)
	DeleteContentBlock(ctx context.Context, id uint) error
}