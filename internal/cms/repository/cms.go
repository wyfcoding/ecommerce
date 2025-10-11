package data

import (
	"context"
	"ecommerce/internal/cms/biz"

	"gorm.io/gorm"
)

// cmsRepo is the data layer implementation for CmsRepo.
type cmsRepo struct {
	data *Data
	// log  *log.Helper
}

// toBiz converts a data.ContentPage model to a biz.ContentPage entity.
func (p *ContentPage) toBiz() *biz.ContentPage {
	if p == nil {
		return nil
	}
	return &biz.ContentPage{
		ID:          p.ID,
		Title:       p.Title,
		Slug:        p.Slug,
		ContentHTML: p.ContentHTML,
		Status:      p.Status,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// fromBiz converts a biz.ContentPage entity to a data.ContentPage model.
func fromBizContentPage(b *biz.ContentPage) *ContentPage {
	if b == nil {
		return nil
	}
	return &ContentPage{
		Title:       b.Title,
		Slug:        b.Slug,
		ContentHTML: b.ContentHTML,
		Status:      b.Status,
	}
}

// toBiz converts a data.ContentBlock model to a biz.ContentBlock entity.
func (b *ContentBlock) toBiz() *biz.ContentBlock {
	if b == nil {
		return nil
	}
	return &biz.ContentBlock{
		ID:          b.ID,
		Name:        b.Name,
		ContentHTML: b.ContentHTML,
		Type:        b.Type,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
	}
}

// fromBiz converts a biz.ContentBlock entity to a data.ContentBlock model.
func fromBizContentBlock(b *biz.ContentBlock) *ContentBlock {
	if b == nil {
		return nil
	}
	return &ContentBlock{
		Name:        b.Name,
		ContentHTML: b.ContentHTML,
		Type:        b.Type,
	}
}

// CreateContentPage creates a new content page in the database.
func (r *cmsRepo) CreateContentPage(ctx context.Context, b *biz.ContentPage) (*biz.ContentPage, error) {
	page := fromBizContentPage(b)
	if err := r.data.db.WithContext(ctx).Create(page).Error; err != nil {
		return nil, err
	}
	return page.toBiz(), nil
}

// GetContentPageByID retrieves a content page by ID from the database.
func (r *cmsRepo) GetContentPageByID(ctx context.Context, id uint) (*biz.ContentPage, error) {
	var page ContentPage
	if err := r.data.db.WithContext(ctx).First(&page, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, biz.ErrContentPageNotFound
		}
		return nil, err
	}
	return page.toBiz(), nil
}

// GetContentPageBySlug retrieves a content page by slug from the database.
func (r *cmsRepo) GetContentPageBySlug(ctx context.Context, slug string) (*biz.ContentPage, error) {
	var page ContentPage
	if err := r.data.db.WithContext(ctx).Where("slug = ?", slug).First(&page).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, biz.ErrContentPageNotFound
		}
		return nil, err
	}
	return page.toBiz(), nil
}

// UpdateContentPage updates an existing content page in the database.
func (r *cmsRepo) UpdateContentPage(ctx context.Context, b *biz.ContentPage) (*biz.ContentPage, error) {
	page := fromBizContentPage(b)
	page.ID = b.ID // Ensure ID is set for update
	if err := r.data.db.WithContext(ctx).Save(page).Error; err != nil {
		return nil, err
	}
	return page.toBiz(), nil
}

// DeleteContentPage deletes a content page from the database.
func (r *cmsRepo) DeleteContentPage(ctx context.Context, id uint) error {
	result := r.data.db.WithContext(ctx).Delete(&ContentPage{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return biz.ErrContentPageNotFound
	}
	return nil
}

// ListContentPages lists content pages from the database with optional filters.
func (r *cmsRepo) ListContentPages(ctx context.Context, statusFilter string, pageSize, pageToken int32) ([]*biz.ContentPage, int32, error) {
	var pages []ContentPage
	var totalCount int32

	query := r.data.db.WithContext(ctx).Model(&ContentPage{})

	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	// Get total count
	query.Count(int64(&totalCount))

	// Apply pagination
	if pageSize > 0 {
		query = query.Limit(int(pageSize)).Offset(int(pageToken * pageSize))
	}

	if err := query.Find(&pages).Error; err != nil {
		return nil, 0, err
	}

	bizPages := make([]*biz.ContentPage, len(pages))
	for i, p := range pages {
		bizPages[i] = p.toBiz()
	}

	return bizPages, totalCount, nil
}

// CreateContentBlock creates a new content block in the database.
func (r *cmsRepo) CreateContentBlock(ctx context.Context, b *biz.ContentBlock) (*biz.ContentBlock, error) {
	block := fromBizContentBlock(b)
	if err := r.data.db.WithContext(ctx).Create(block).Error; err != nil {
		return nil, err
	}
	return block.toBiz(), nil
}

// GetContentBlockByID retrieves a content block by ID from the database.
func (r *cmsRepo) GetContentBlockByID(ctx context.Context, id uint) (*biz.ContentBlock, error) {
	var block ContentBlock
	if err := r.data.db.WithContext(ctx).First(&block, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, biz.ErrContentBlockNotFound
		}
		return nil, err
	}
	return block.toBiz(), nil
}

// GetContentBlockByName retrieves a content block by name from the database.
func (r *cmsRepo) GetContentBlockByName(ctx context.Context, name string) (*biz.ContentBlock, error) {
	var block ContentBlock
	if err := r.data.db.WithContext(ctx).Where("name = ?", name).First(&block).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, biz.ErrContentBlockNotFound
		}
		return nil, err
	}
	return block.toBiz(), nil
}

// UpdateContentBlock updates an existing content block in the database.
func (r *cmsRepo) UpdateContentBlock(ctx context.Context, b *biz.ContentBlock) (*biz.ContentBlock, error) {
	block := fromBizContentBlock(b)
	block.ID = b.ID // Ensure ID is set for update
	if err := r.data.db.WithContext(ctx).Save(block).Error; err != nil {
		return nil, err
	}
	return block.toBiz(), nil
}

// DeleteContentBlock deletes a content block from the database.
func (r *cmsRepo) DeleteContentBlock(ctx context.Context, id uint) error {
	result := r.data.db.WithContext(ctx).Delete(&ContentBlock{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return biz.ErrContentBlockNotFound
	}
	return nil
}
