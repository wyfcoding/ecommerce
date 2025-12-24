package grpc

import (
	"context"
	"encoding/json"
	"fmt"

	pb "github.com/wyfcoding/ecommerce/goapi/marketing/v1"
	"github.com/wyfcoding/ecommerce/internal/marketing/application"
	"github.com/wyfcoding/ecommerce/internal/marketing/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server 结构体定义。
type Server struct {
	pb.UnimplementedMarketingServiceServer
	app *application.MarketingService
}

// NewServer 函数。
func NewServer(app *application.MarketingService) *Server {
	return &Server{app: app}
}

func (s *Server) CreateCampaign(ctx context.Context, req *pb.CreateCampaignRequest) (*pb.CreateCampaignResponse, error) {
	var rules map[string]any
	if req.RulesJson != "" {
		if err := json.Unmarshal([]byte(req.RulesJson), &rules); err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid rules_json format: %v", err))
		}
	}

	campaign, err := s.app.CreateCampaign(ctx, req.Name, domain.CampaignType(req.CampaignType), req.Description, req.StartTime.AsTime(), req.EndTime.AsTime(), req.Budget, rules)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create campaign: %v", err))
	}

	return &pb.CreateCampaignResponse{
		Campaign: convertCampaignToProto(campaign),
	}, nil
}

func (s *Server) GetCampaign(ctx context.Context, req *pb.GetCampaignRequest) (*pb.GetCampaignResponse, error) {
	campaign, err := s.app.GetCampaign(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("campaign not found: %v", err))
	}

	return &pb.GetCampaignResponse{
		Campaign: convertCampaignToProto(campaign),
	}, nil
}

func (s *Server) UpdateCampaignStatus(ctx context.Context, req *pb.UpdateCampaignStatusRequest) (*emptypb.Empty, error) {
	if err := s.app.UpdateCampaignStatus(ctx, req.Id, domain.CampaignStatus(req.Status)); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update campaign status: %v", err))
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ListCampaigns(ctx context.Context, req *pb.ListCampaignsRequest) (*pb.ListCampaignsResponse, error) {
	page := max(int(req.Page), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	campaigns, total, err := s.app.ListCampaigns(ctx, domain.CampaignStatus(req.Status), page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list campaigns: %v", err))
	}

	pbCampaigns := make([]*pb.Campaign, len(campaigns))
	for i, c := range campaigns {
		pbCampaigns[i] = convertCampaignToProto(c)
	}

	return &pb.ListCampaignsResponse{
		Campaigns:  pbCampaigns,
		TotalCount: total,
	}, nil
}

func (s *Server) RecordParticipation(ctx context.Context, req *pb.RecordParticipationRequest) (*emptypb.Empty, error) {
	if err := s.app.RecordParticipation(ctx, req.CampaignId, req.UserId, req.OrderId, req.Discount); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to record participation: %v", err))
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) CreateBanner(ctx context.Context, req *pb.CreateBannerRequest) (*pb.CreateBannerResponse, error) {
	banner, err := s.app.CreateBanner(ctx, req.Title, req.ImageUrl, req.LinkUrl, req.Position, req.Priority, req.StartTime.AsTime(), req.EndTime.AsTime())
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create banner: %v", err))
	}

	return &pb.CreateBannerResponse{
		Banner: convertBannerToProto(banner),
	}, nil
}

func (s *Server) ListBanners(ctx context.Context, req *pb.ListBannersRequest) (*pb.ListBannersResponse, error) {
	banners, err := s.app.ListBanners(ctx, req.Position, req.ActiveOnly)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list banners: %v", err))
	}

	pbBanners := make([]*pb.Banner, len(banners))
	for i, b := range banners {
		pbBanners[i] = convertBannerToProto(b)
	}

	return &pb.ListBannersResponse{
		Banners: pbBanners,
	}, nil
}

func (s *Server) DeleteBanner(ctx context.Context, req *pb.DeleteBannerRequest) (*emptypb.Empty, error) {
	if err := s.app.DeleteBanner(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete banner: %v", err))
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ClickBanner(ctx context.Context, req *pb.ClickBannerRequest) (*emptypb.Empty, error) {
	if err := s.app.ClickBanner(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to record banner click: %v", err))
	}
	return &emptypb.Empty{}, nil
}

func convertCampaignToProto(c *domain.Campaign) *pb.Campaign {
	if c == nil {
		return nil
	}
	rulesJson, _ := json.Marshal(c.Rules)
	return &pb.Campaign{
		Id:           uint64(c.ID),
		Name:         c.Name,
		CampaignType: string(c.CampaignType),
		Description:  c.Description,
		StartTime:    timestamppb.New(c.StartTime),
		EndTime:      timestamppb.New(c.EndTime),
		Budget:       c.Budget,
		Spent:        c.Spent,
		TargetUsers:  c.TargetUsers,
		ReachedUsers: c.ReachedUsers,
		Status:       int32(c.Status),
		RulesJson:    string(rulesJson),
		CreatedAt:    timestamppb.New(c.CreatedAt),
		UpdatedAt:    timestamppb.New(c.UpdatedAt),
	}
}

func convertBannerToProto(b *domain.Banner) *pb.Banner {
	if b == nil {
		return nil
	}
	return &pb.Banner{
		Id:         uint64(b.ID),
		Title:      b.Title,
		ImageUrl:   b.ImageURL,
		LinkUrl:    b.LinkURL,
		Position:   b.Position,
		Priority:   b.Priority,
		StartTime:  timestamppb.New(b.StartTime),
		EndTime:    timestamppb.New(b.EndTime),
		ClickCount: b.ClickCount,
		Enabled:    b.Enabled,
		CreatedAt:  timestamppb.New(b.CreatedAt),
		UpdatedAt:  timestamppb.New(b.UpdatedAt),
	}
}
