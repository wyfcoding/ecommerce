package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	pb "github.com/wyfcoding/ecommerce/go-api/cms/v1"
	"github.com/wyfcoding/ecommerce/internal/cms/domain"
	"github.com/wyfcoding/pkg/idgen"
)

type CMSService struct {
	repo   domain.CMSRepository
	idGen  idgen.Generator
	logger *slog.Logger
}

func NewCMSService(repo domain.CMSRepository, idGen idgen.Generator, logger *slog.Logger) *CMSService {
	return &CMSService{
		repo:   repo,
		idGen:  idGen,
		logger: logger.With("service", "cms_application"),
	}
}

// Page Use Cases
func (s *CMSService) CreatePage(ctx context.Context, req *pb.CreatePageRequest) (*domain.Page, error) {
	pageID := s.idGen.Generate()
	page := &domain.Page{
		ID:          fmt.Sprintf("pg_%d", pageID),
		Title:       req.Title,
		Slug:        req.Slug,
		TemplateID:  req.TemplateId,
		ContentJSON: req.ContentJson,
		Status:      pb.PageStatus_DRAFT,
		CreatedBy:   req.CreatedBy,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.SavePage(ctx, page); err != nil {
		return nil, fmt.Errorf("failed to save page: %w", err)
	}
	return page, nil
}

func (s *CMSService) GetPage(ctx context.Context, req *pb.GetPageRequest) (*domain.Page, error) {
	if req.GetId() != "" {
		return s.repo.GetPageByID(ctx, req.GetId())
	}
	if req.GetSlug() != "" {
		return s.repo.GetPageBySlug(ctx, req.GetSlug())
	}
	return nil, fmt.Errorf("id or slug must be provided")
}

func (s *CMSService) UpdatePage(ctx context.Context, req *pb.UpdatePageRequest) (*domain.Page, error) {
	page, err := s.repo.GetPageByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	if req.Title != "" {
		page.Title = req.Title
	}
	if req.ContentJson != "" {
		page.ContentJSON = req.ContentJson
	}
	if req.Status != pb.PageStatus_PAGE_STATUS_UNSPECIFIED {
		page.Status = req.Status
	}
	page.UpdatedAt = time.Now()

	if err := s.repo.SavePage(ctx, page); err != nil {
		return nil, err
	}
	return page, nil
}

func (s *CMSService) PublishPage(ctx context.Context, id string) (*domain.Page, error) {
	page, err := s.repo.GetPageByID(ctx, id)
	if err != nil {
		return nil, err
	}
	page.Publish()
	if err := s.repo.SavePage(ctx, page); err != nil {
		return nil, err
	}
	return page, nil
}

func (s *CMSService) ListPages(ctx context.Context, req *pb.ListPagesRequest) ([]*domain.Page, int, error) {
	offset := int((req.Page - 1) * req.PageSize)
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListPages(ctx, req.Status, offset, int(req.PageSize))
}

func (s *CMSService) DeletePage(ctx context.Context, id string) error {
	return s.repo.DeletePage(ctx, id)
}

// Template Use Cases
func (s *CMSService) CreateTemplate(ctx context.Context, req *pb.CreateTemplateRequest) (*domain.Template, error) {
	tmplID := s.idGen.Generate()
	tmpl := &domain.Template{
		ID:            fmt.Sprintf("tpl_%d", tmplID),
		Name:          req.Name,
		Description:   req.Description,
		StructureJSON: req.StructureJson,
		CreatedAt:     time.Now(),
	}
	if err := s.repo.SaveTemplate(ctx, tmpl); err != nil {
		return nil, err
	}
	return tmpl, nil
}

func (s *CMSService) GetTemplate(ctx context.Context, id string) (*domain.Template, error) {
	return s.repo.GetTemplateByID(ctx, id)
}

func (s *CMSService) ListTemplates(ctx context.Context, req *pb.ListTemplatesRequest) ([]*domain.Template, int, error) {
	offset := int((req.Page - 1) * req.PageSize)
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListTemplates(ctx, offset, int(req.PageSize))
}

// Asset Use Cases
func (s *CMSService) CreateAsset(ctx context.Context, req *pb.CreateAssetRequest) (*domain.Asset, error) {
	assetID := s.idGen.Generate()
	asset := &domain.Asset{
		ID:        fmt.Sprintf("as_%d", assetID),
		Name:      req.Name,
		FileType:  req.FileType,
		URL:       req.Url,
		Size:      req.Size,
		Bucket:    req.Bucket,
		Key:       req.Key,
		CreatedAt: time.Now(),
	}
	if err := s.repo.SaveAsset(ctx, asset); err != nil {
		return nil, err
	}
	return asset, nil
}

func (s *CMSService) GetAsset(ctx context.Context, id string) (*domain.Asset, error) {
	return s.repo.GetAssetByID(ctx, id)
}

func (s *CMSService) ListAssets(ctx context.Context, req *pb.ListAssetsRequest) ([]*domain.Asset, int, error) {
	offset := int((req.Page - 1) * req.PageSize)
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListAssets(ctx, req.FileType, offset, int(req.PageSize))
}

func (s *CMSService) DeleteAsset(ctx context.Context, id string) error {
	return s.repo.DeleteAsset(ctx, id)
}
