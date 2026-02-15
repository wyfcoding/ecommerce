package domain

import (
	"context"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/cms/v1"
)

// Page 聚合根，代表一个内容页面。
type Page struct {
	ID          string
	Title       string
	Slug        string
	TemplateID  string
	ContentJSON string // 存储组件布局的 JSON
	Status      pb.PageStatus
	CreatedBy   string
	Metadata    map[string]string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	PublishedAt *time.Time
}

// Template 定义页面布局和允许的组件。
type Template struct {
	ID            string
	Name          string
	Description   string
	StructureJSON string // 允许的组件定义
	CreatedAt     time.Time
}

// Asset 代表上传的媒体资源。
type Asset struct {
	ID        string
	Name      string
	FileType  string
	URL       string
	Size      int64
	Bucket    string
	Key       string
	CreatedAt time.Time
}

// CMSRepository 定义了 CMS 相关的仓储操作。
type CMSRepository interface {
	// Page 操作
	SavePage(ctx context.Context, page *Page) error
	GetPageByID(ctx context.Context, id string) (*Page, error)
	GetPageBySlug(ctx context.Context, slug string) (*Page, error)
	ListPages(ctx context.Context, status pb.PageStatus, offset, limit int) ([]*Page, int, error)
	DeletePage(ctx context.Context, id string) error

	// Template 操作
	SaveTemplate(ctx context.Context, tmpl *Template) error
	GetTemplateByID(ctx context.Context, id string) (*Template, error)
	ListTemplates(ctx context.Context, offset, limit int) ([]*Template, int, error)

	// Asset 操作
	SaveAsset(ctx context.Context, asset *Asset) error
	GetAssetByID(ctx context.Context, id string) (*Asset, error)
	ListAssets(ctx context.Context, fileType string, offset, limit int) ([]*Asset, int, error)
	DeleteAsset(ctx context.Context, id string) error
}

// Publish 转换页面状态为发布。
func (p *Page) Publish() {
	p.Status = pb.PageStatus_PUBLISHED
	now := time.Now()
	p.PublishedAt = &now
	p.UpdatedAt = now
}

// Archive 存档页面。
func (p *Page) Archive() {
	p.Status = pb.PageStatus_ARCHIVED
	p.UpdatedAt = time.Now()
}
