package grpc

import (
	"context"
	"log/slog"

	pb "github.com/wyfcoding/ecommerce/go-api/cms/v1"
	"github.com/wyfcoding/ecommerce/internal/cms/application"
	"github.com/wyfcoding/ecommerce/internal/cms/domain"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedCMSServiceServer
	svc    *application.CMSService
	logger *slog.Logger
}

func NewServer(svc *application.CMSService, logger *slog.Logger) pb.CMSServiceServer {
	return &Server{
		svc:    svc,
		logger: logger.With("component", "cms_grpc"),
	}
}

func (s *Server) CreatePage(ctx context.Context, req *pb.CreatePageRequest) (*pb.Page, error) {
	page, err := s.svc.CreatePage(ctx, req)
	if err != nil {
		return nil, err
	}
	return toPBPage(page), nil
}

func (s *Server) GetPage(ctx context.Context, req *pb.GetPageRequest) (*pb.Page, error) {
	page, err := s.svc.GetPage(ctx, req)
	if err != nil {
		return nil, err
	}
	return toPBPage(page), nil
}

func (s *Server) UpdatePage(ctx context.Context, req *pb.UpdatePageRequest) (*pb.Page, error) {
	page, err := s.svc.UpdatePage(ctx, req)
	if err != nil {
		return nil, err
	}
	return toPBPage(page), nil
}

func (s *Server) DeletePage(ctx context.Context, req *pb.DeletePageRequest) (*emptypb.Empty, error) {
	if err := s.svc.DeletePage(ctx, req.Id); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListPages(ctx context.Context, req *pb.ListPagesRequest) (*pb.ListPagesResponse, error) {
	pages, total, err := s.svc.ListPages(ctx, req)
	if err != nil {
		return nil, err
	}
	pbPages := make([]*pb.Page, len(pages))
	for i, p := range pages {
		pbPages[i] = toPBPage(p)
	}
	return &pb.ListPagesResponse{Pages: pbPages, Total: int32(total)}, nil
}

func (s *Server) PublishPage(ctx context.Context, req *pb.PublishPageRequest) (*pb.Page, error) {
	page, err := s.svc.PublishPage(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return toPBPage(page), nil
}

// Template
func (s *Server) CreateTemplate(ctx context.Context, req *pb.CreateTemplateRequest) (*pb.Template, error) {
	tmpl, err := s.svc.CreateTemplate(ctx, req)
	if err != nil {
		return nil, err
	}
	return toPBTemplate(tmpl), nil
}

func (s *Server) GetTemplate(ctx context.Context, req *pb.GetTemplateRequest) (*pb.Template, error) {
	tmpl, err := s.svc.GetTemplate(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return toPBTemplate(tmpl), nil
}

func (s *Server) ListTemplates(ctx context.Context, req *pb.ListTemplatesRequest) (*pb.ListTemplatesResponse, error) {
	tmpls, total, err := s.svc.ListTemplates(ctx, req)
	if err != nil {
		return nil, err
	}
	pbTmpls := make([]*pb.Template, len(tmpls))
	for i, t := range tmpls {
		pbTmpls[i] = toPBTemplate(t)
	}
	return &pb.ListTemplatesResponse{Templates: pbTmpls, Total: int32(total)}, nil
}

// Asset
func (s *Server) CreateAsset(ctx context.Context, req *pb.CreateAssetRequest) (*pb.Asset, error) {
	asset, err := s.svc.CreateAsset(ctx, req)
	if err != nil {
		return nil, err
	}
	return toPBAsset(asset), nil
}

func (s *Server) GetAsset(ctx context.Context, req *pb.GetAssetRequest) (*pb.Asset, error) {
	asset, err := s.svc.GetAsset(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return toPBAsset(asset), nil
}

func (s *Server) ListAssets(ctx context.Context, req *pb.ListAssetsRequest) (*pb.ListAssetsResponse, error) {
	assets, total, err := s.svc.ListAssets(ctx, req)
	if err != nil {
		return nil, err
	}
	pbAssets := make([]*pb.Asset, len(assets))
	for i, a := range assets {
		pbAssets[i] = toPBAsset(a)
	}
	return &pb.ListAssetsResponse{Assets: pbAssets, Total: int32(total)}, nil
}

func (s *Server) DeleteAsset(ctx context.Context, req *pb.DeleteAssetRequest) (*emptypb.Empty, error) {
	if err := s.svc.DeleteAsset(ctx, req.Id); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// Mappings
func toPBPage(p *domain.Page) *pb.Page {
	res := &pb.Page{
		Id:          p.ID,
		Title:       p.Title,
		Slug:        p.Slug,
		TemplateId:  p.TemplateID,
		ContentJson: p.ContentJSON,
		Status:      p.Status,
		CreatedBy:   p.CreatedBy,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
		Metadata:    p.Metadata,
	}
	if p.PublishedAt != nil {
		res.PublishedAt = timestamppb.New(*p.PublishedAt)
	}
	return res
}

func toPBTemplate(t *domain.Template) *pb.Template {
	return &pb.Template{
		Id:            t.ID,
		Name:          t.Name,
		Description:   t.Description,
		StructureJson: t.StructureJSON,
		CreatedAt:     timestamppb.New(t.CreatedAt),
	}
}

func toPBAsset(a *domain.Asset) *pb.Asset {
	return &pb.Asset{
		Id:        a.ID,
		Name:      a.Name,
		FileType:  a.FileType,
		Url:       a.URL,
		Size:      a.Size,
		Bucket:    a.Bucket,
		Key:       a.Key,
		CreatedAt: timestamppb.New(a.CreatedAt),
	}
}
