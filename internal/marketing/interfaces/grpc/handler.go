package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

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
	app *application.Marketing
}

// NewServer 函数。
func NewServer(app *application.Marketing) *Server {
	return &Server{app: app}
}

func (s *Server) CreateCampaign(ctx context.Context, req *pb.CreateCampaignRequest) (*pb.CreateCampaignResponse, error) {
	start := time.Now()
	slog.Info("gRPC CreateCampaign received", "name", req.Name, "type", req.CampaignType)

	var rules map[string]any
	if req.RulesJson != "" {
		if err := json.Unmarshal([]byte(req.RulesJson), &rules); err != nil {
			slog.Warn("gRPC CreateCampaign invalid rules json", "name", req.Name, "error", err)
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid rules_json format: %v", err))
		}
	}

	campaign, err := s.app.CreateCampaign(ctx, req.Name, domain.CampaignType(req.CampaignType), req.Description, req.StartTime.AsTime(), req.EndTime.AsTime(), req.Budget, rules)
	if err != nil {
		slog.Error("gRPC CreateCampaign failed", "name", req.Name, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create campaign: %v", err))
	}

	slog.Info("gRPC CreateCampaign successful", "campaign_id", campaign.ID, "duration", time.Since(start))
	return &pb.CreateCampaignResponse{
		Campaign: convertCampaignToProto(campaign),
	}, nil
}

func (s *Server) GetCampaign(ctx context.Context, req *pb.GetCampaignRequest) (*pb.GetCampaignResponse, error) {
	start := time.Now()
	slog.Debug("gRPC GetCampaign received", "id", req.Id)

	campaign, err := s.app.GetCampaign(ctx, req.Id)
	if err != nil {
		slog.Error("gRPC GetCampaign failed", "id", req.Id, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.NotFound, fmt.Sprintf("campaign not found: %v", err))
	}

	slog.Debug("gRPC GetCampaign successful", "id", req.Id, "duration", time.Since(start))
	return &pb.GetCampaignResponse{
		Campaign: convertCampaignToProto(campaign),
	}, nil
}

func (s *Server) UpdateCampaignStatus(ctx context.Context, req *pb.UpdateCampaignStatusRequest) (*emptypb.Empty, error) {
	start := time.Now()
	slog.Info("gRPC UpdateCampaignStatus received", "id", req.Id, "status", req.Status)

	if err := s.app.UpdateCampaignStatus(ctx, req.Id, domain.CampaignStatus(req.Status)); err != nil {
		slog.Error("gRPC UpdateCampaignStatus failed", "id", req.Id, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update campaign status: %v", err))
	}

	slog.Info("gRPC UpdateCampaignStatus successful", "id", req.Id, "duration", time.Since(start))
	return &emptypb.Empty{}, nil
}

func (s *Server) ListCampaigns(ctx context.Context, req *pb.ListCampaignsRequest) (*pb.ListCampaignsResponse, error) {
	start := time.Now()
	slog.Debug("gRPC ListCampaigns received", "page", req.Page, "status", req.Status)

	page := max(int(req.Page), 1)
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	campaigns, total, err := s.app.ListCampaigns(ctx, domain.CampaignStatus(req.Status), page, pageSize)
	if err != nil {
		slog.Error("gRPC ListCampaigns failed", "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list campaigns: %v", err))
	}

	pbCampaigns := make([]*pb.Campaign, len(campaigns))
	for i, c := range campaigns {
		pbCampaigns[i] = convertCampaignToProto(c)
	}

	slog.Debug("gRPC ListCampaigns successful", "count", len(pbCampaigns), "duration", time.Since(start))
	return &pb.ListCampaignsResponse{
		Campaigns:  pbCampaigns,
		TotalCount: total,
	}, nil
}

func (s *Server) RecordParticipation(ctx context.Context, req *pb.RecordParticipationRequest) (*emptypb.Empty, error) {
	start := time.Now()
	slog.Info("gRPC RecordParticipation received", "campaign_id", req.CampaignId, "user_id", req.UserId, "order_id", req.OrderId)

	if err := s.app.RecordParticipation(ctx, req.CampaignId, req.UserId, req.OrderId, req.Discount); err != nil {
		slog.Error("gRPC RecordParticipation failed", "campaign_id", req.CampaignId, "user_id", req.UserId, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to record participation: %v", err))
	}

	slog.Info("gRPC RecordParticipation successful", "campaign_id", req.CampaignId, "user_id", req.UserId, "duration", time.Since(start))
	return &emptypb.Empty{}, nil
}

func (s *Server) CreateBanner(ctx context.Context, req *pb.CreateBannerRequest) (*pb.CreateBannerResponse, error) {
	start := time.Now()
	slog.Info("gRPC CreateBanner received", "title", req.Title, "position", req.Position)

	banner, err := s.app.CreateBanner(ctx, req.Title, req.ImageUrl, req.LinkUrl, req.Position, req.Priority, req.StartTime.AsTime(), req.EndTime.AsTime())
	if err != nil {
		slog.Error("gRPC CreateBanner failed", "title", req.Title, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create banner: %v", err))
	}

	slog.Info("gRPC CreateBanner successful", "banner_id", banner.ID, "duration", time.Since(start))
	return &pb.CreateBannerResponse{
		Banner: convertBannerToProto(banner),
	}, nil
}

func (s *Server) ListBanners(ctx context.Context, req *pb.ListBannersRequest) (*pb.ListBannersResponse, error) {
	start := time.Now()
	slog.Debug("gRPC ListBanners received", "position", req.Position)

	banners, err := s.app.ListBanners(ctx, req.Position, req.ActiveOnly)
	if err != nil {
		slog.Error("gRPC ListBanners failed", "position", req.Position, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list banners: %v", err))
	}

	pbBanners := make([]*pb.Banner, len(banners))
	for i, b := range banners {
		pbBanners[i] = convertBannerToProto(b)
	}

	slog.Debug("gRPC ListBanners successful", "count", len(pbBanners), "duration", time.Since(start))
	return &pb.ListBannersResponse{
		Banners: pbBanners,
	}, nil
}

func (s *Server) DeleteBanner(ctx context.Context, req *pb.DeleteBannerRequest) (*emptypb.Empty, error) {
	start := time.Now()
	slog.Info("gRPC DeleteBanner received", "id", req.Id)

	if err := s.app.DeleteBanner(ctx, req.Id); err != nil {
		slog.Error("gRPC DeleteBanner failed", "id", req.Id, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete banner: %v", err))
	}

	slog.Info("gRPC DeleteBanner successful", "id", req.Id, "duration", time.Since(start))
	return &emptypb.Empty{}, nil
}

func (s *Server) ClickBanner(ctx context.Context, req *pb.ClickBannerRequest) (*emptypb.Empty, error) {
	start := time.Now()
	slog.Debug("gRPC ClickBanner received", "id", req.Id)

	if err := s.app.ClickBanner(ctx, req.Id); err != nil {
		slog.Error("gRPC ClickBanner failed", "id", req.Id, "error", err, "duration", time.Since(start))
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to record banner click: %v", err))
	}

	slog.Debug("gRPC ClickBanner successful", "id", req.Id, "duration", time.Since(start))
	return &emptypb.Empty{}, nil
}

func convertCampaignToProto(c *domain.Campaign) *pb.Campaign {
	if c == nil {
		return nil
	}
	rulesJson, err := json.Marshal(c.Rules)
	if err != nil {
		rulesJson = []byte("{}")
	}
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
