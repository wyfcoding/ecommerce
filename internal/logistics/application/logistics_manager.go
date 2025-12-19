package application

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/logistics/domain"
	"github.com/wyfcoding/pkg/algorithm"
)

// LogisticsManager 处理物流的写操作（创建、状态更新、轨迹追踪、路线优化）。
type LogisticsManager struct {
	repo      domain.LogisticsRepository
	optimizer *algorithm.RouteOptimizer
	logger    *slog.Logger
}

func NewLogisticsManager(repo domain.LogisticsRepository, logger *slog.Logger) *LogisticsManager {
	return &LogisticsManager{
		repo:      repo,
		optimizer: algorithm.NewRouteOptimizer(),
		logger:    logger,
	}
}

// CreateLogistics 创建一个新的物流单。
func (m *LogisticsManager) CreateLogistics(ctx context.Context, orderID uint64, orderNo, trackingNo, carrier, carrierCode string,
	senderName, senderPhone, senderAddress string, senderLat, senderLon float64,
	receiverName, receiverPhone, receiverAddress string, receiverLat, receiverLon float64) (*domain.Logistics, error) {

	logistics := domain.NewLogistics(orderID, orderNo, trackingNo, carrier, carrierCode,
		senderName, senderPhone, senderAddress, senderLat, senderLon,
		receiverName, receiverPhone, receiverAddress, receiverLat, receiverLon)

	if err := m.repo.Save(ctx, logistics); err != nil {
		m.logger.ErrorContext(ctx, "failed to save logistics", "order_id", orderID, "error", err)
		return nil, err
	}
	m.logger.InfoContext(ctx, "logistics created successfully", "logistics_id", logistics.ID, "tracking_no", trackingNo)
	return logistics, nil
}

// UpdateStatus 更新物流单状态。
func (m *LogisticsManager) UpdateStatus(ctx context.Context, id uint64, status domain.LogisticsStatus, location, description string) error {
	logistics, err := m.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	switch status {
	case domain.LogisticsStatusPickedUp:
		logistics.PickUp()
	case domain.LogisticsStatusInTransit:
		logistics.Transit(location)
	case domain.LogisticsStatusDelivering:
		logistics.Deliver()
	case domain.LogisticsStatusDelivered:
		logistics.Complete()
	case domain.LogisticsStatusReturning:
		logistics.Return()
	case domain.LogisticsStatusReturned:
		logistics.ReturnComplete()
	case domain.LogisticsStatusException:
		logistics.Exception(description)
	default:
		return domain.ErrInvalidStatus
	}

	logistics.AddTrace(location, description, "")

	if err := m.repo.Save(ctx, logistics); err != nil {
		m.logger.ErrorContext(ctx, "failed to update logistics status", "logistics_id", id, "status", status, "error", err)
		return err
	}
	return nil
}

// AddTrace 添加物流轨迹记录。
func (m *LogisticsManager) AddTrace(ctx context.Context, id uint64, location, description, status string) error {
	logistics, err := m.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	logistics.AddTrace(location, description, status)
	logistics.UpdateLocation(location)

	if err := m.repo.Save(ctx, logistics); err != nil {
		m.logger.ErrorContext(ctx, "failed to add trace", "logistics_id", id, "error", err)
		return err
	}
	return nil
}

// SetEstimatedTime 设置物流单的预计送达时间。
func (m *LogisticsManager) SetEstimatedTime(ctx context.Context, id uint64, estimatedTime time.Time) error {
	logistics, err := m.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	logistics.SetEstimatedTime(estimatedTime)
	if err := m.repo.Save(ctx, logistics); err != nil {
		m.logger.ErrorContext(ctx, "failed to set estimated time", "logistics_id", id, "error", err)
		return err
	}
	return nil
}

// OptimizeDeliveryRoute 优化配送路线。
func (m *LogisticsManager) OptimizeDeliveryRoute(ctx context.Context, logisticsID uint64, destinations []algorithm.Location) (*domain.DeliveryRoute, error) {
	logistics, err := m.repo.GetByID(ctx, logisticsID)
	if err != nil {
		return nil, err
	}

	start := algorithm.Location{
		ID:  0,
		Lat: logistics.SenderLat,
		Lon: logistics.SenderLon,
	}

	route := m.optimizer.OptimizeRoute(start, destinations)

	routeData, err := json.Marshal(route.Locations)
	if err != nil {
		return nil, err
	}

	deliveryRoute := &domain.DeliveryRoute{
		LogisticsID: logisticsID,
		RouteData:   string(routeData),
		Distance:    route.Distance,
	}

	logistics.Route = deliveryRoute
	if err := m.repo.Save(ctx, logistics); err != nil {
		m.logger.ErrorContext(ctx, "failed to save optimized route", "logistics_id", logisticsID, "error", err)
		return nil, err
	}

	return deliveryRoute, nil
}
