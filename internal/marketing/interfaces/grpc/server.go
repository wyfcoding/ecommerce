package grpc

import (
	"context"
	"encoding/json" // 导入JSON编码/解码库。
	"fmt"

	pb "github.com/wyfcoding/ecommerce/go-api/marketing/v1"           // 导入营销模块的protobuf定义。
	"github.com/wyfcoding/ecommerce/internal/marketing/application"   // 导入营销模块的应用服务。
	"github.com/wyfcoding/ecommerce/internal/marketing/domain/entity" // 导入营销模块的领域实体。

	"google.golang.org/grpc/codes"                       // gRPC状态码。
	"google.golang.org/grpc/status"                      // gRPC状态处理。
	"google.golang.org/protobuf/types/known/emptypb"     // 导入空消息类型。
	"google.golang.org/protobuf/types/known/timestamppb" // 导入时间戳消息类型。
)

// Server 结构体实现了 MarketingService 的 gRPC 服务端接口。
// 它是DDD分层架构中的接口层，负责接收gRPC请求，调用应用服务处理业务逻辑，并将结果封装为gRPC响应。
type Server struct {
	pb.UnimplementedMarketingServiceServer                               // 嵌入生成的UnimplementedMarketingServiceServer，确保前向兼容性。
	app                                    *application.MarketingService // 依赖Marketing应用服务，处理核心业务逻辑。
}

// NewServer 创建并返回一个新的 Marketing gRPC 服务端实例。
func NewServer(app *application.MarketingService) *Server {
	return &Server{app: app}
}

// CreateCampaign 处理创建营销活动的gRPC请求。
// req: 包含活动名称、类型、描述、时间范围、预算和规则（JSON字符串）的请求体。
// 返回创建成功的营销活动响应和可能发生的gRPC错误。
func (s *Server) CreateCampaign(ctx context.Context, req *pb.CreateCampaignRequest) (*pb.CreateCampaignResponse, error) {
	var rules map[string]interface{}
	// 如果Proto请求中提供了RulesJson，则尝试解析为map[string]interface{}。
	if req.RulesJson != "" {
		if err := json.Unmarshal([]byte(req.RulesJson), &rules); err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid rules_json format: %v", err))
		}
	}

	// 调用应用服务层创建营销活动。
	campaign, err := s.app.CreateCampaign(ctx, req.Name, entity.CampaignType(req.CampaignType), req.Description, req.StartTime.AsTime(), req.EndTime.AsTime(), req.Budget, rules)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create campaign: %v", err))
	}

	// 将领域实体转换为protobuf响应格式。
	return &pb.CreateCampaignResponse{
		Campaign: convertCampaignToProto(campaign),
	}, nil
}

// GetCampaign 处理获取营销活动详情的gRPC请求。
// req: 包含活动ID的请求体。
// 返回营销活动响应和可能发生的gRPC错误。
func (s *Server) GetCampaign(ctx context.Context, req *pb.GetCampaignRequest) (*pb.GetCampaignResponse, error) {
	campaign, err := s.app.GetCampaign(ctx, req.Id)
	if err != nil {
		// 如果营销活动未找到，返回NotFound状态码。
		return nil, status.Error(codes.NotFound, fmt.Sprintf("campaign not found: %v", err))
	}

	return &pb.GetCampaignResponse{
		Campaign: convertCampaignToProto(campaign),
	}, nil
}

// UpdateCampaignStatus 处理更新营销活动状态的gRPC请求。
// req: 包含活动ID和新状态的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) UpdateCampaignStatus(ctx context.Context, req *pb.UpdateCampaignStatusRequest) (*emptypb.Empty, error) {
	// 将Proto的Status（int32）转换为实体CampaignStatus。
	if err := s.app.UpdateCampaignStatus(ctx, req.Id, entity.CampaignStatus(req.Status)); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update campaign status: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ListCampaigns 处理列出营销活动的gRPC请求。
// req: 包含分页参数的请求体。
// 返回营销活动列表响应和可能发生的gRPC错误。
func (s *Server) ListCampaigns(ctx context.Context, req *pb.ListCampaignsRequest) (*pb.ListCampaignsResponse, error) {
	// 获取分页参数。
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 10
	}

	// 调用应用服务层获取营销活动列表。
	// 将Proto的Status（int32）转换为实体CampaignStatus。
	campaigns, total, err := s.app.ListCampaigns(ctx, entity.CampaignStatus(req.Status), page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list campaigns: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbCampaigns := make([]*pb.Campaign, len(campaigns))
	for i, c := range campaigns {
		pbCampaigns[i] = convertCampaignToProto(c)
	}

	return &pb.ListCampaignsResponse{
		Campaigns:  pbCampaigns,
		TotalCount: total, // 总记录数。
	}, nil
}

// RecordParticipation 处理记录用户参与营销活动的gRPC请求。
// req: 包含活动ID、用户ID、订单ID和优惠金额的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) RecordParticipation(ctx context.Context, req *pb.RecordParticipationRequest) (*emptypb.Empty, error) {
	if err := s.app.RecordParticipation(ctx, req.CampaignId, req.UserId, req.OrderId, req.Discount); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to record participation: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// CreateBanner 处理创建广告横幅的gRPC请求。
// req: 包含横幅标题、图片URL、链接、位置、优先级和时间范围的请求体。
// 返回创建成功的广告横幅响应和可能发生的gRPC错误。
func (s *Server) CreateBanner(ctx context.Context, req *pb.CreateBannerRequest) (*pb.CreateBannerResponse, error) {
	banner, err := s.app.CreateBanner(ctx, req.Title, req.ImageUrl, req.LinkUrl, req.Position, req.Priority, req.StartTime.AsTime(), req.EndTime.AsTime())
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create banner: %v", err))
	}

	return &pb.CreateBannerResponse{
		Banner: convertBannerToProto(banner),
	}, nil
}

// ListBanners 处理列出广告横幅的gRPC请求。
// req: 包含位置和是否只列出活跃横幅的请求体。
// 返回广告横幅列表响应和可能发生的gRPC错误。
func (s *Server) ListBanners(ctx context.Context, req *pb.ListBannersRequest) (*pb.ListBannersResponse, error) {
	banners, err := s.app.ListBanners(ctx, req.Position, req.ActiveOnly)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list banners: %v", err))
	}

	// 将领域实体列表转换为protobuf响应格式的列表。
	pbBanners := make([]*pb.Banner, len(banners))
	for i, b := range banners {
		pbBanners[i] = convertBannerToProto(b)
	}

	return &pb.ListBannersResponse{
		Banners: pbBanners,
	}, nil
}

// DeleteBanner 处理删除广告横幅的gRPC请求。
// req: 包含横幅ID的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) DeleteBanner(ctx context.Context, req *pb.DeleteBannerRequest) (*emptypb.Empty, error) {
	if err := s.app.DeleteBanner(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete banner: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// ClickBanner 处理记录广告横幅点击事件的gRPC请求。
// req: 包含横幅ID的请求体。
// 返回一个空响应和可能发生的gRPC错误。
func (s *Server) ClickBanner(ctx context.Context, req *pb.ClickBannerRequest) (*emptypb.Empty, error) {
	if err := s.app.ClickBanner(ctx, req.Id); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to record banner click: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// convertCampaignToProto 是一个辅助函数，将领域层的 Campaign 实体转换为 protobuf 的 Campaign 消息。
func convertCampaignToProto(c *entity.Campaign) *pb.Campaign {
	if c == nil {
		return nil
	}
	rulesJson, _ := json.Marshal(c.Rules) // 将Campaign的Rules字段（JSONMap）转换为JSON字符串。
	return &pb.Campaign{
		Id:           uint64(c.ID),                 // 活动ID。
		Name:         c.Name,                       // 名称。
		CampaignType: string(c.CampaignType),       // 活动类型。
		Description:  c.Description,                // 描述。
		StartTime:    timestamppb.New(c.StartTime), // 开始时间。
		EndTime:      timestamppb.New(c.EndTime),   // 结束时间。
		Budget:       c.Budget,                     // 预算。
		Spent:        c.Spent,                      // 已花费。
		TargetUsers:  c.TargetUsers,                // 目标用户数。
		ReachedUsers: c.ReachedUsers,               // 触达用户数。
		Status:       int32(c.Status),              // 状态。
		RulesJson:    string(rulesJson),            // 规则配置（JSON字符串）。
		CreatedAt:    timestamppb.New(c.CreatedAt), // 创建时间。
		UpdatedAt:    timestamppb.New(c.UpdatedAt), // 更新时间。
	}
}

// convertBannerToProto 是一个辅助函数，将领域层的 Banner 实体转换为 protobuf 的 Banner 消息。
func convertBannerToProto(b *entity.Banner) *pb.Banner {
	if b == nil {
		return nil
	}
	return &pb.Banner{
		Id:         uint64(b.ID),                 // Banner ID。
		Title:      b.Title,                      // 标题。
		ImageUrl:   b.ImageURL,                   // 图片URL。
		LinkUrl:    b.LinkURL,                    // 跳转URL。
		Position:   b.Position,                   // 位置。
		Priority:   b.Priority,                   // 优先级。
		StartTime:  timestamppb.New(b.StartTime), // 开始时间。
		EndTime:    timestamppb.New(b.EndTime),   // 结束时间。
		ClickCount: b.ClickCount,                 // 点击数。
		Enabled:    b.Enabled,                    // 是否启用。
		CreatedAt:  timestamppb.New(b.CreatedAt), // 创建时间。
		UpdatedAt:  timestamppb.New(b.UpdatedAt), // 更新时间。
	}
}
