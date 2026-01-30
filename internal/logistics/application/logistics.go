package application

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/logistics/domain"
	algorithm "github.com/wyfcoding/pkg/algorithm/optimization"
)

// Logistics 是物流应用服务的门面。
type Logistics struct {
	Command *LogisticsCommandService
	Query   *LogisticsQueryService
}

// NewLogistics 创建物流服务门面实例。
func NewLogistics(command *LogisticsCommandService, query *LogisticsQueryService) *Logistics {
	return &Logistics{
		Command: command,
		Query:   query,
	}
}

// CreateLogistics 创建物流运单记录。
func (s *Logistics) CreateLogistics(ctx context.Context, orderID uint64, orderNo, trackingNo, carrier, carrierCode string,
	senderName, senderPhone, senderAddress string, senderLat, senderLon float64,
	receiverName, receiverPhone, receiverAddress string, receiverLat, receiverLon float64,
) (*domain.Logistics, error) {
	return s.Command.CreateLogistics(ctx, orderID, orderNo, trackingNo, carrier, carrierCode,
		senderName, senderPhone, senderAddress, senderLat, senderLon,
		receiverName, receiverPhone, receiverAddress, receiverLat, receiverLon)
}

// GetLogistics 根据ID获取物流信息详情。
func (s *Logistics) GetLogistics(ctx context.Context, id uint64) (*domain.Logistics, error) {
	return s.Query.GetLogistics(ctx, id)
}

// GetLogisticsByTrackingNo 根据物流单号获取物流信息。
func (s *Logistics) GetLogisticsByTrackingNo(ctx context.Context, trackingNo string) (*domain.Logistics, error) {
	return s.Query.GetLogisticsByTrackingNo(ctx, trackingNo)
}

// UpdateStatus 更新物流状态。
func (s *Logistics) UpdateStatus(ctx context.Context, id uint64, status domain.LogisticsStatus, location, description string) error {
	return s.Command.UpdateStatus(ctx, id, status, location, description)
}

// AddTrace 添加一条新的物流轨迹记录。
func (s *Logistics) AddTrace(ctx context.Context, id uint64, location, description, status string) error {
	return s.Command.AddTrace(ctx, id, location, description, status)
}

// SetEstimatedTime 设置或更新预计送达时间。
func (s *Logistics) SetEstimatedTime(ctx context.Context, id uint64, estimatedTime time.Time) error {
	return s.Command.SetEstimatedTime(ctx, id, estimatedTime)
}

// ListLogistics 列出所有物流记录（分页）。
func (s *Logistics) ListLogistics(ctx context.Context, page, pageSize int) ([]*domain.Logistics, int64, error) {
	return s.Query.ListLogistics(ctx, page, pageSize)
}

// OptimizeDeliveryRoute 核心算法：路径规划优化。
func (s *Logistics) OptimizeDeliveryRoute(ctx context.Context, logisticsID uint64, destinations []algorithm.Location) (*domain.DeliveryRoute, error) {
	return s.Command.OptimizeDeliveryRoute(ctx, logisticsID, destinations)
}

// AssignRidersToOrders 指派并分配骑手。
func (s *Logistics) AssignRidersToOrders(ctx context.Context, riders []RiderInfo, logisticsIDs []uint64) (map[string]uint64, error) {
	return s.Command.AssignRidersToOrders(ctx, riders, logisticsIDs)
}
