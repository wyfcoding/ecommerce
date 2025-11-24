package application

import (
	"context"
	"ecommerce/internal/logistics/domain/entity"
	"ecommerce/internal/logistics/domain/repository"
	"ecommerce/pkg/algorithm"
	"encoding/json"
	"time"

	"log/slog"
)

type LogisticsService struct {
	repo      repository.LogisticsRepository
	optimizer *algorithm.RouteOptimizer
	logger    *slog.Logger
}

func NewLogisticsService(repo repository.LogisticsRepository, logger *slog.Logger) *LogisticsService {
	return &LogisticsService{
		repo:      repo,
		optimizer: algorithm.NewRouteOptimizer(),
		logger:    logger,
	}
}

// CreateLogistics 创建物流单
func (s *LogisticsService) CreateLogistics(ctx context.Context, orderID uint64, orderNo, trackingNo, carrier, carrierCode string,
	senderName, senderPhone, senderAddress string, senderLat, senderLon float64,
	receiverName, receiverPhone, receiverAddress string, receiverLat, receiverLon float64) (*entity.Logistics, error) {

	logistics := entity.NewLogistics(orderID, orderNo, trackingNo, carrier, carrierCode,
		senderName, senderPhone, senderAddress, senderLat, senderLon,
		receiverName, receiverPhone, receiverAddress, receiverLat, receiverLon)

	if err := s.repo.Save(ctx, logistics); err != nil {
		s.logger.Error("failed to save logistics", "error", err)
		return nil, err
	}
	return logistics, nil
}

// GetLogistics 获取物流信息
func (s *LogisticsService) GetLogistics(ctx context.Context, id uint64) (*entity.Logistics, error) {
	return s.repo.GetByID(ctx, id)
}

// GetLogisticsByTrackingNo 根据运单号获取物流信息
func (s *LogisticsService) GetLogisticsByTrackingNo(ctx context.Context, trackingNo string) (*entity.Logistics, error) {
	return s.repo.GetByTrackingNo(ctx, trackingNo)
}

// UpdateStatus 更新物流状态
func (s *LogisticsService) UpdateStatus(ctx context.Context, id uint64, status entity.LogisticsStatus, location, description string) error {
	logistics, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	switch status {
	case entity.LogisticsStatusPickedUp:
		logistics.PickUp()
	case entity.LogisticsStatusInTransit:
		logistics.Transit(location)
	case entity.LogisticsStatusDelivering:
		logistics.Deliver()
	case entity.LogisticsStatusDelivered:
		logistics.Complete()
	case entity.LogisticsStatusReturning:
		logistics.Return()
	case entity.LogisticsStatusReturned:
		logistics.ReturnComplete()
	case entity.LogisticsStatusException:
		logistics.Exception(description)
	default:
		return entity.ErrInvalidStatus
	}

	logistics.AddTrace(location, description, "") // Status string could be mapped from enum if needed

	return s.repo.Save(ctx, logistics)
}

// AddTrace 添加物流轨迹
func (s *LogisticsService) AddTrace(ctx context.Context, id uint64, location, description, status string) error {
	logistics, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	logistics.AddTrace(location, description, status)
	logistics.UpdateLocation(location)

	return s.repo.Save(ctx, logistics)
}

// SetEstimatedTime 设置预计送达时间
func (s *LogisticsService) SetEstimatedTime(ctx context.Context, id uint64, estimatedTime time.Time) error {
	logistics, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	logistics.SetEstimatedTime(estimatedTime)
	return s.repo.Save(ctx, logistics)
}

// ListLogistics 获取物流列表
func (s *LogisticsService) ListLogistics(ctx context.Context, page, pageSize int) ([]*entity.Logistics, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, offset, pageSize)
}

// OptimizeDeliveryRoute 优化配送路线
func (s *LogisticsService) OptimizeDeliveryRoute(ctx context.Context, logisticsID uint64, destinations []algorithm.Location) (*entity.DeliveryRoute, error) {
	logistics, err := s.repo.GetByID(ctx, logisticsID)
	if err != nil {
		return nil, err
	}

	start := algorithm.Location{
		ID:  0, // Start point (e.g., warehouse or current location)
		Lat: logistics.SenderLat,
		Lon: logistics.SenderLon,
	}

	// If current location is set and has coordinates (parsing needed in real app), use it.
	// For now, use SenderLat/Lon as start.

	route := s.optimizer.OptimizeRoute(start, destinations)

	routeData, err := json.Marshal(route.Locations)
	if err != nil {
		return nil, err
	}

	deliveryRoute := &entity.DeliveryRoute{
		LogisticsID: logisticsID,
		RouteData:   string(routeData),
		Distance:    route.Distance,
	}

	// Save route (assuming repository has SaveRoute or we save it via Logistics update)
	// Since DeliveryRoute is associated with Logistics, we might need to save it separately or update Logistics.
	// Let's assume we need to add SaveRoute to repository or just save Logistics if it cascades.
	// GORM default association handling might work if we set logistics.Route = deliveryRoute and Save(logistics).
	logistics.Route = deliveryRoute
	if err := s.repo.Save(ctx, logistics); err != nil {
		return nil, err
	}

	return deliveryRoute, nil
}
