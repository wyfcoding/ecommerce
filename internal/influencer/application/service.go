package application

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	pb "github.com/wyfcoding/ecommerce/goapi/influencer/v1"
	"github.com/wyfcoding/ecommerce/internal/influencer/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type InfluencerService struct {
	repo domain.InfluencerRepository
}

func NewInfluencerService(repo domain.InfluencerRepository) *InfluencerService {
	return &InfluencerService{repo: repo}
}

func (s *InfluencerService) RegisterInfluencer(ctx context.Context, req *pb.RegisterInfluencerRequest) (*pb.RegisterInfluencerResponse, error) {
	infID := fmt.Sprintf("inf_%d", time.Now().UnixNano())
	inf := &domain.Influencer{
		UserID:        req.UserId,
		InfluencerID:  infID,
		Name:          req.Name,
		Platform:      req.Platform,
		Handle:        req.Handle,
		FollowerCount: req.FollowerCount,
		Status:        domain.StatusPending,
		TotalEarnings: decimal.Zero,
	}

	if err := s.repo.SaveInfluencer(ctx, inf); err != nil {
		return nil, err
	}

	return &pb.RegisterInfluencerResponse{
		InfluencerId: infID,
		Status:       string(domain.StatusPending),
	}, nil
}

func (s *InfluencerService) CreateCampaign(ctx context.Context, req *pb.CreateCampaignRequest) (*pb.CreateCampaignResponse, error) {
	rate, err := decimal.NewFromString(req.CommissionRate)
	if err != nil {
		return nil, err
	}

	campID := fmt.Sprintf("camp_%d", time.Now().UnixNano())
	camp := &domain.Campaign{
		CampaignID:     campID,
		InfluencerID:   req.InfluencerId,
		ProductID:      req.ProductId,
		CommissionRate: rate,
		StartAt:        req.StartAt.AsTime(),
		EndAt:          req.EndAt.AsTime(),
		Status:         "ACTIVE",
	}

	if err := s.repo.SaveCampaign(ctx, camp); err != nil {
		return nil, err
	}

	return &pb.CreateCampaignResponse{CampaignId: campID}, nil
}

func (s *InfluencerService) GetEarnings(ctx context.Context, infID string) (*pb.GetEarningsResponse, error) {
	inf, err := s.repo.GetInfluencer(ctx, infID)
	if err != nil {
		return nil, err
	}
	if inf == nil {
		return nil, fmt.Errorf("influencer %s not found", infID)
	}

	// 简单模拟统计
	return &pb.GetEarningsResponse{
		InfluencerId:    inf.InfluencerID,
		TotalEarnings:   inf.TotalEarnings.String(),
		PendingEarnings: "0.00",
		TotalSales:      123, // 演示数据
	}, nil
}

func (s *InfluencerService) ListCampaigns(ctx context.Context, infID string) (*pb.ListCampaignsResponse, error) {
	camps, err := s.repo.ListCampaigns(ctx, infID)
	if err != nil {
		return nil, err
	}

	var pbCamps []*pb.CampaignItem
	for _, c := range camps {
		pbCamps = append(pbCamps, &pb.CampaignItem{
			CampaignId:     c.CampaignID,
			ProductId:      c.ProductID,
			Status:         c.Status,
			CommissionRate: c.CommissionRate.String(),
			CreatedAt:      timestamppb.New(c.CreatedAt),
		})
	}

	return &pb.ListCampaignsResponse{Campaigns: pbCamps}, nil
}
