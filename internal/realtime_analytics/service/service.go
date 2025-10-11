package service

import (
	"context"

	v1 "ecommerce/api/realtime_analytics/v1"
	"ecommerce/internal/realtime_analytics/biz"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RealtimeAnalyticsService is the gRPC service implementation for real-time analytics.
type RealtimeAnalyticsService struct {
	v1.UnimplementedRealtimeAnalyticsServiceServer
	uc *biz.RealtimeAnalyticsUsecase
}

// NewRealtimeAnalyticsService creates a new RealtimeAnalyticsService.
func NewRealtimeAnalyticsService(uc *biz.RealtimeAnalyticsUsecase) *RealtimeAnalyticsService {
	return &RealtimeAnalyticsService{uc: uc}
}

// GetRealtimeSalesMetrics implements the GetRealtimeSalesMetrics RPC.
func (s *RealtimeAnalyticsService) GetRealtimeSalesMetrics(ctx context.Context, req *v1.GetRealtimeSalesMetricsRequest) (*v1.RealtimeSalesMetricsResponse, error) {
	metrics, err := s.uc.GetRealtimeSalesMetrics(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get real-time sales metrics: %v", err)
	}

	return &v1.RealtimeSalesMetricsResponse{
			Metrics: &v1.RealtimeSalesMetrics{
				CurrentGmv:    metrics.CurrentGmv,
				CurrentOrders: metrics.CurrentOrders,
				ActiveUsers:   metrics.ActiveUsers,
				LastUpdated:   metrics.LastUpdated,
			},
		},
		nil
}

// GetRealtimeUserActivity implements the GetRealtimeUserActivity RPC.
func (s *RealtimeAnalyticsService) GetRealtimeUserActivity(ctx context.Context, req *v1.GetRealtimeUserActivityRequest) (*v1.RealtimeUserActivityResponse, error) {
	activity, err := s.uc.GetRealtimeUserActivity(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get real-time user activity: %v", err)
	}

	return &v1.RealtimeUserActivityResponse{
			Activity: &v1.RealtimeUserActivity{
				OnlineUsers:        activity.OnlineUsers,
				NewUsersLastHour:   activity.NewUsersLastHour,
				PageViewsPerMinute: activity.PageViewsPerMinute,
			},
		},
		nil
}
