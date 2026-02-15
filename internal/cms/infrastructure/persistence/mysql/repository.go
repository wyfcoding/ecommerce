package mysql

import (
	"context"
	"encoding/json"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/cms/v1"
	"github.com/wyfcoding/ecommerce/internal/cms/domain"
	"github.com/wyfcoding/pkg/database"
	"github.com/wyfcoding/pkg/logging"
	"gorm.io/gorm"
)

// PageModel GORM 模型，代表页面。
type PageModel struct {
	gorm.Model
	PageID      string     `gorm:"column:page_id;uniqueIndex;type:varchar(64);not null;comment:页面唯一ID"`
	Title       string     `gorm:"column:title;type:varchar(255);not null;comment:标题"`
	Slug        string     `gorm:"column:slug;uniqueIndex;type:varchar(255);not null;comment:路径别名"`
	TemplateID  string     `gorm:"column:template_id;index;type:varchar(64);comment:模版ID"`
	ContentJSON string     `gorm:"column:content_json;type:longtext;comment:页面内容JSON"`
	Status      int32      `gorm:"column:status;index;comment:状态"`
	CreatedBy   string     `gorm:"column:created_by;type:varchar(64);comment:创建者"`
	Metadata    string     `gorm:"column:metadata;type:text;comment:元数据JSON"`
	PublishedAt *time.Time `gorm:"column:published_at;comment:发布时间"`
}

func (PageModel) TableName() string { return "cms_pages" }

// TemplateModel GORM 模型，代表模版。
type TemplateModel struct {
	gorm.Model
	TemplateID    string `gorm:"column:template_id;uniqueIndex;type:varchar(64);not null;comment:模版唯一ID"`
	Name          string `gorm:"column:name;type:varchar(255);not null;comment:名称"`
	Description   string `gorm:"column:description;type:text;comment:描述"`
	StructureJSON string `gorm:"column:structure_json;type:longtext;comment:结构定义JSON"`
}

func (TemplateModel) TableName() string { return "cms_templates" }

// AssetModel GORM 模型，代表资源。
type AssetModel struct {
	gorm.Model
	AssetID  string `gorm:"column:asset_id;uniqueIndex;type:varchar(64);not null;comment:资源唯一ID"`
	Name     string `gorm:"column:name;type:varchar(255);not null;comment:名称"`
	FileType string `gorm:"column:file_type;index;type:varchar(50);comment:文件类型"`
	URL      string `gorm:"column:url;type:varchar(512);comment:访问地址"`
	Size     int64  `gorm:"column:size;comment:文件大小"`
	Bucket   string `gorm:"column:bucket;type:varchar(255);comment:存储桶"`
	Key      string `gorm:"column:key_path;type:varchar(512);comment:存储路径"`
}

func (AssetModel) TableName() string { return "cms_assets" }

type cmsRepository struct {
	db     *database.DB
	logger *logging.Logger
}

func NewCMSRepository(db *database.DB, logger *logging.Logger) domain.CMSRepository {
	return &cmsRepository{db: db, logger: logger}
}

// Page 实现
func (r *cmsRepository) SavePage(ctx context.Context, page *domain.Page) error {
	m := toPageModel(page)
	return r.db.RawDB().WithContext(ctx).Save(m).Error
}

func (r *cmsRepository) GetPageByID(ctx context.Context, id string) (*domain.Page, error) {
	var m PageModel
	if err := r.db.RawDB().WithContext(ctx).Where("page_id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return toPageDomain(&m), nil
}

func (r *cmsRepository) GetPageBySlug(ctx context.Context, slug string) (*domain.Page, error) {
	var m PageModel
	if err := r.db.RawDB().WithContext(ctx).Where("slug = ?", slug).First(&m).Error; err != nil {
		return nil, err
	}
	return toPageDomain(&m), nil
}

func (r *cmsRepository) ListPages(ctx context.Context, status pb.PageStatus, offset, limit int) ([]*domain.Page, int, error) {
	var models []*PageModel
	var total int64
	db := r.db.RawDB().WithContext(ctx).Model(&PageModel{})
	if status != pb.PageStatus_PAGE_STATUS_UNSPECIFIED {
		db = db.Where("status = ?", int32(status))
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(limit).Find(&models).Error; err != nil {
		return nil, 0, err
	}
	pages := make([]*domain.Page, len(models))
	for i, m := range models {
		pages[i] = toPageDomain(m)
	}
	return pages, int(total), nil
}

func (r *cmsRepository) DeletePage(ctx context.Context, id string) error {
	return r.db.RawDB().WithContext(ctx).Where("page_id = ?", id).Delete(&PageModel{}).Error
}

// Template 实现
func (r *cmsRepository) SaveTemplate(ctx context.Context, tmpl *domain.Template) error {
	m := toTemplateModel(tmpl)
	return r.db.RawDB().WithContext(ctx).Save(m).Error
}

func (r *cmsRepository) GetTemplateByID(ctx context.Context, id string) (*domain.Template, error) {
	var m TemplateModel
	if err := r.db.RawDB().WithContext(ctx).Where("template_id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return toTemplateDomain(&m), nil
}

func (r *cmsRepository) ListTemplates(ctx context.Context, offset, limit int) ([]*domain.Template, int, error) {
	var models []*TemplateModel
	var total int64
	db := r.db.RawDB().WithContext(ctx).Model(&TemplateModel{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(limit).Find(&models).Error; err != nil {
		return nil, 0, err
	}
	tmpls := make([]*domain.Template, len(models))
	for i, m := range models {
		tmpls[i] = toTemplateDomain(m)
	}
	return tmpls, int(total), nil
}

// Asset 实现
func (r *cmsRepository) SaveAsset(ctx context.Context, asset *domain.Asset) error {
	m := toAssetModel(asset)
	return r.db.RawDB().WithContext(ctx).Save(m).Error
}

func (r *cmsRepository) GetAssetByID(ctx context.Context, id string) (*domain.Asset, error) {
	var m AssetModel
	if err := r.db.RawDB().WithContext(ctx).Where("asset_id = ?", id).First(&m).Error; err != nil {
		return nil, err
	}
	return toAssetDomain(&m), nil
}

func (r *cmsRepository) ListAssets(ctx context.Context, fileType string, offset, limit int) ([]*domain.Asset, int, error) {
	var models []*AssetModel
	var total int64
	db := r.db.RawDB().WithContext(ctx).Model(&AssetModel{})
	if fileType != "" {
		db = db.Where("file_type = ?", fileType)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := db.Offset(offset).Limit(limit).Find(&models).Error; err != nil {
		return nil, 0, err
	}
	assets := make([]*domain.Asset, len(models))
	for i, m := range models {
		assets[i] = toAssetDomain(m)
	}
	return assets, int(total), nil
}

func (r *cmsRepository) DeleteAsset(ctx context.Context, id string) error {
	return r.db.RawDB().WithContext(ctx).Where("asset_id = ?", id).Delete(&AssetModel{}).Error
}

// Mappings
func toPageModel(p *domain.Page) *PageModel {
	meta, _ := json.Marshal(p.Metadata)
	return &PageModel{
		PageID:      p.ID,
		Title:       p.Title,
		Slug:        p.Slug,
		TemplateID:  p.TemplateID,
		ContentJSON: p.ContentJSON,
		Status:      int32(p.Status),
		CreatedBy:   p.CreatedBy,
		Metadata:    string(meta),
		PublishedAt: p.PublishedAt,
	}
}

func toPageDomain(m *PageModel) *domain.Page {
	var meta map[string]string
	_ = json.Unmarshal([]byte(m.Metadata), &meta)
	return &domain.Page{
		ID:          m.PageID,
		Title:       m.Title,
		Slug:        m.Slug,
		TemplateID:  m.TemplateID,
		ContentJSON: m.ContentJSON,
		Status:      pb.PageStatus(m.Status),
		CreatedBy:   m.CreatedBy,
		Metadata:    meta,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		PublishedAt: m.PublishedAt,
	}
}

func toTemplateModel(t *domain.Template) *TemplateModel {
	return &TemplateModel{
		TemplateID:    t.ID,
		Name:          t.Name,
		Description:   t.Description,
		StructureJSON: t.StructureJSON,
	}
}

func toTemplateDomain(m *TemplateModel) *domain.Template {
	return &domain.Template{
		ID:            m.TemplateID,
		Name:          m.Name,
		Description:   m.Description,
		StructureJSON: m.StructureJSON,
		CreatedAt:     m.CreatedAt,
	}
}

func toAssetModel(a *domain.Asset) *AssetModel {
	return &AssetModel{
		AssetID:  a.ID,
		Name:     a.Name,
		FileType: a.FileType,
		URL:      a.URL,
		Size:     a.Size,
		Bucket:   a.Bucket,
		Key:      a.Key,
	}
}

func toAssetDomain(m *AssetModel) *domain.Asset {
	return &domain.Asset{
		ID:        m.AssetID,
		Name:      m.Name,
		FileType:  m.FileType,
		URL:       m.URL,
		Size:      m.Size,
		Bucket:    m.Bucket,
		Key:       m.Key,
		CreatedAt: m.CreatedAt,
	}
}
