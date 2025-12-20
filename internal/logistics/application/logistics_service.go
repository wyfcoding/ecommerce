package application

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/logistics/domain"
	"github.com/wyfcoding/pkg/algorithm"
)

// LogisticsService 是物流应用服务的门面。
type LogisticsService struct {
	Manager *LogisticsManager
	Query   *LogisticsQuery
}

// NewLogisticsService 定义了 NewLogistics 相关的服务逻辑。
func NewLogisticsService(manager *LogisticsManager, query *LogisticsQuery) *LogisticsService {
	return &LogisticsService{
		Manager: manager,
		Query:   query,
	}
}

func (s *LogisticsService) CreateLogistics(ctx context.Context, orderID uint64, orderNo, trackingNo, carrier, carrierCode string,
	senderName, senderPhone, senderAddress string, senderLat, senderLon float64,
	receiverName, receiverPhone, receiverAddress string, receiverLat, receiverLon float64) (*domain.Logistics, error) {
	return s.Manager.CreateLogistics(ctx, orderID, orderNo, trackingNo, carrier, carrierCode,
		senderName, senderPhone, senderAddress, senderLat, senderLon,
		receiverName, receiverPhone, receiverAddress, receiverLat, receiverLon)
}

func (s *LogisticsService) GetLogistics(ctx context.Context, id uint64) (*domain.Logistics, error) {
	return s.Query.GetLogistics(ctx, id)
}

func (s *LogisticsService) GetLogisticsByTrackingNo(ctx context.Context, trackingNo string) (*domain.Logistics, error) {
	return s.Query.GetLogisticsByTrackingNo(ctx, trackingNo)
}

func (s *LogisticsService) UpdateStatus(ctx context.Context, id uint64, status domain.LogisticsStatus, location, description string) error {
	return s.Manager.UpdateStatus(ctx, id, status, location, description)
}

func (s *LogisticsService) AddTrace(ctx context.Context, id uint64, location, description, status string) error {
	return s.Manager.AddTrace(ctx, id, location, description, status)
}

func (s *LogisticsService) SetEstimatedTime(ctx context.Context, id uint64, estimatedTime time.Time) error {
	return s.Manager.SetEstimatedTime(ctx, id, estimatedTime)
}

func (s *LogisticsService) ListLogistics(ctx context.Context, page, pageSize int) ([]*domain.Logistics, int64, error) {
	return s.Query.ListLogistics(ctx, page, pageSize)
}

func (s *LogisticsService) OptimizeDeliveryRoute(ctx context.Context, logisticsID uint64, destinations []algorithm.Location) (*domain.DeliveryRoute, error) {
	return s.Manager.OptimizeDeliveryRoute(ctx, logisticsID, destinations)
}
