package interfaces

import (
	"context"

	pb "github.com/wyfcoding/ecommerce/go-api/influencer/v1"
	"github.com/wyfcoding/ecommerce/internal/influencer/application"
)

type InfluencerHandler struct {
	pb.UnimplementedInfluencerServiceServer
	app *application.InfluencerService
}

func NewInfluencerHandler(app *application.InfluencerService) *InfluencerHandler {
	return &InfluencerHandler{app: app}
}

func (h *InfluencerHandler) RegisterInfluencer(ctx context.Context, req *pb.RegisterInfluencerRequest) (*pb.RegisterInfluencerResponse, error) {
	return h.app.RegisterInfluencer(ctx, req)
}

func (h *InfluencerHandler) CreateCampaign(ctx context.Context, req *pb.CreateCampaignRequest) (*pb.CreateCampaignResponse, error) {
	return h.app.CreateCampaign(ctx, req)
}

func (h *InfluencerHandler) GetEarnings(ctx context.Context, req *pb.GetEarningsRequest) (*pb.GetEarningsResponse, error) {
	return h.app.GetEarnings(ctx, req.InfluencerId)
}

func (h *InfluencerHandler) ListCampaigns(ctx context.Context, req *pb.ListCampaignsRequest) (*pb.ListCampaignsResponse, error) {
	return h.app.ListCampaigns(ctx, req.InfluencerId)
}
