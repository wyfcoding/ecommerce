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

// NewLogisticsService 创建物流服务门面实例。
func NewLogisticsService(manager *LogisticsManager, query *LogisticsQuery) *LogisticsService {
	return &LogisticsService{
		Manager: manager,
		Query:   query,
	}
}

// CreateLogistics 创建物流运单记录。
func (s *LogisticsService) CreateLogistics(ctx context.Context, orderID uint64, orderNo, trackingNo, carrier, carrierCode string,
	senderName, senderPhone, senderAddress string, senderLat, senderLon float64,
	receiverName, receiverPhone, receiverAddress string, receiverLat, receiverLon float64,
) (*domain.Logistics, error) {
	return s.Manager.CreateLogistics(ctx, orderID, orderNo, trackingNo, carrier, carrierCode,
		senderName, senderPhone, senderAddress, senderLat, senderLon,
		receiverName, receiverPhone, receiverAddress, receiverLat, receiverLon)
}

// GetLogistics 根据ID获取物流信息详情。
func (s *LogisticsService) GetLogistics(ctx context.Context, id uint64) (*domain.Logistics, error) {
	return s.Query.GetLogistics(ctx, id)
}

// GetLogisticsByTrackingNo 根据物流单号获取物流信息。
func (s *LogisticsService) GetLogisticsByTrackingNo(ctx context.Context, trackingNo string) (*domain.Logistics, error) {
	return s.Query.GetLogisticsByTrackingNo(ctx, trackingNo)
}

// UpdateStatus 更新物流状态（如已揽收、运输中、已送达等）。
func (s *LogisticsService) UpdateStatus(ctx context.Context, id uint64, status domain.LogisticsStatus, location, description string) error {
	return s.Manager.UpdateStatus(ctx, id, status, location, description)
}

// AddTrace 添加一条新的物流轨迹记录。
func (s *LogisticsService) AddTrace(ctx context.Context, id uint64, location, description, status string) error {
	return s.Manager.AddTrace(ctx, id, location, description, status)
}

// SetEstimatedTime 设置或更新预计送达时间。
func (s *LogisticsService) SetEstimatedTime(ctx context.Context, id uint64, estimatedTime time.Time) error {
	return s.Manager.SetEstimatedTime(ctx, id, estimatedTime)
}

// ListLogistics 列出所有物流记录（分页）。
func (s *LogisticsService) ListLogistics(ctx context.Context, page, pageSize int) ([]*domain.Logistics, int64, error) {
	return s.Query.ListLogistics(ctx, page, pageSize)
}

// OptimizeDeliveryRoute 核心算法：为指定运单及其多个目的地优化配送路径。
func (s *LogisticsService) OptimizeDeliveryRoute(ctx context.Context, logisticsID uint64, destinations []algorithm.Location) (*domain.DeliveryRoute, error) {
	return s.Manager.OptimizeDeliveryRoute(ctx, logisticsID, destinations)
}
