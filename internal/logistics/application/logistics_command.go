package application

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/logistics/domain"
	"github.com/wyfcoding/pkg/algorithm/graph"
	"github.com/wyfcoding/pkg/algorithm/math"
	algorithm "github.com/wyfcoding/pkg/algorithm/optimization"
	"github.com/wyfcoding/pkg/messagequeue"
)

// LogisticsCommandService 处理物流的写操作（创建、状态更新、轨迹追踪、路线优化）。
type LogisticsCommandService struct {
	repo             domain.LogisticsRepository
	publisher        messagequeue.EventPublisher
	optimizer        *algorithm.RouteOptimizer
	packingOptimizer *algorithm.BinPackingOptimizer
	logger           *slog.Logger
}

const DefaultVehicleCapacity = 500.0

type RiderInfo struct {
	ID  string
	Lat float64
	Lon float64
}

type OrderInfo struct {
	ID  string
	Lat float64
	Lon float64
}

// NewLogisticsCommandService 构造函数。
func NewLogisticsCommandService(
	repo domain.LogisticsRepository,
	publisher messagequeue.EventPublisher,
	logger *slog.Logger,
) *LogisticsCommandService {
	return &LogisticsCommandService{
		repo:             repo,
		publisher:        publisher,
		optimizer:        algorithm.NewRouteOptimizer(),
		packingOptimizer: algorithm.NewBinPackingOptimizer(1000.0),
		logger:           logger,
	}
}

// CreateLogistics 创建一个新的物流单。
func (s *LogisticsCommandService) CreateLogistics(ctx context.Context, orderID uint64, orderNo, trackingNo, carrier, carrierCode string,
	senderName, senderPhone, senderAddress string, senderLat, senderLon float64,
	receiverName, receiverPhone, receiverAddress string, receiverLat, receiverLon float64,
) (*domain.Logistics, error) {
	logistics := domain.NewLogistics(orderID, orderNo, trackingNo, carrier, carrierCode,
		senderName, senderPhone, senderAddress, senderLat, senderLon,
		receiverName, receiverPhone, receiverAddress, receiverLat, receiverLon)

	err := s.repo.WithTx(ctx, func(tx any) error {
		if err := s.repo.SaveInTx(ctx, tx, logistics); err != nil {
			return err
		}

		event := &domain.LogisticsCreatedEvent{
			LogisticsID: uint(logistics.ID),
			OrderID:     orderID,
			OrderNo:     orderNo,
			TrackingNo:  trackingNo,
			Timestamp:   time.Now(),
		}
		return s.publisher.PublishInTx(ctx, tx, domain.LogisticsCreatedEventType, fmt.Sprintf("%d", orderID), event)
	})
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to create logistics", "order_id", orderID, "error", err)
		return nil, err
	}

	s.logger.InfoContext(ctx, "logistics created", "logistics_id", logistics.ID, "tracking_no", trackingNo)
	return logistics, nil
}

// AssignRidersToOrders 将骑手分配给订单。
func (s *LogisticsCommandService) AssignRidersToOrders(ctx context.Context, riders []RiderInfo, logisticsIDs []uint64) (map[string]uint64, error) {
	if len(riders) == 0 || len(logisticsIDs) == 0 {
		return nil, nil
	}

	var logisticsList []*domain.Logistics
	var orders []OrderInfo

	for _, id := range logisticsIDs {
		logistics, err := s.repo.GetByID(ctx, id)
		if err != nil {
			s.logger.WarnContext(ctx, "logistics not found for assignment", "id", id, "error", err)
			continue
		}
		if logistics.Status != domain.LogisticsStatusPending && logistics.Status != domain.LogisticsStatusPickedUp {
			continue
		}

		logisticsList = append(logisticsList, logistics)
		orders = append(orders, OrderInfo{
			ID:  logistics.OrderNo,
			Lat: logistics.SenderLat,
			Lon: logistics.SenderLon,
		})
	}

	if len(orders) == 0 {
		return nil, nil
	}

	nx := len(riders)
	ny := len(orders)
	size := max(ny, nx)
	bg := graph.NewWeightedBipartiteGraph(size, size)

	for i, rider := range riders {
		for j, order := range orders {
			dist := math.HaversineDistance(rider.Lat, rider.Lon, order.Lat, order.Lon)
			bg.AddEdge(i, j, int(-dist))
		}
	}

	bg.MaxWeightMatch()
	match := bg.GetMatches()

	result := make(map[string]uint64)
	for rIdx, oIdx := range match {
		if rIdx < len(riders) && oIdx < len(orders) {
			riderID := riders[rIdx].ID
			logistics := logisticsList[oIdx]

			err := s.repo.WithTx(ctx, func(tx any) error {
				logistics.AssignRider(riderID)
				logistics.Status = domain.LogisticsStatusPickedUp
				if err := s.repo.SaveInTx(ctx, tx, logistics); err != nil {
					return err
				}

				event := &domain.RiderAssignedEvent{
					LogisticsID: uint(logistics.ID),
					OrderID:     logistics.OrderID,
					RiderID:     riderID,
					Timestamp:   time.Now(),
				}
				return s.publisher.PublishInTx(ctx, tx, domain.LogisticsRiderAssignedEventType, fmt.Sprintf("%d", logistics.ID), event)
			})
			if err != nil {
				s.logger.ErrorContext(ctx, "failed to assign rider", "logistics_id", logistics.ID, "rider_id", riderID, "error", err)
				continue
			}

			result[riderID] = uint64(logistics.ID)
		}
	}

	return result, nil
}

// UpdateStatus 更新物流单状态。
func (s *LogisticsCommandService) UpdateStatus(ctx context.Context, id uint64, status domain.LogisticsStatus, location, description string) error {
	logistics, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	return s.repo.WithTx(ctx, func(tx any) error {
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

		if err := s.repo.SaveInTx(ctx, tx, logistics); err != nil {
			return err
		}

		event := &domain.LogisticsStatusUpdatedEvent{
			LogisticsID: uint(logistics.ID),
			OrderID:     logistics.OrderID,
			Status:      status,
			Location:    location,
			Description: description,
			Timestamp:   time.Now(),
		}
		return s.publisher.PublishInTx(ctx, tx, domain.LogisticsStatusUpdatedEventType, fmt.Sprintf("%d", id), event)
	})
}

// AddTrace 添加物流轨迹记录。
func (s *LogisticsCommandService) AddTrace(ctx context.Context, id uint64, location, description, status string) error {
	logistics, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	return s.repo.WithTx(ctx, func(tx any) error {
		logistics.AddTrace(location, description, status)
		logistics.UpdateLocation(location)

		if err := s.repo.SaveInTx(ctx, tx, logistics); err != nil {
			return err
		}

		event := &domain.LogisticsTraceAddedEvent{
			LogisticsID: uint(logistics.ID),
			TrackingNo:  logistics.TrackingNo,
			Location:    location,
			Description: description,
			Status:      status,
			Timestamp:   time.Now(),
		}
		return s.publisher.PublishInTx(ctx, tx, domain.LogisticsTraceAddedEventType, fmt.Sprintf("%d", id), event)
	})
}

// SetEstimatedTime 设置物流单的预计送达时间。
func (s *LogisticsCommandService) SetEstimatedTime(ctx context.Context, id uint64, estimatedTime time.Time) error {
	logistics, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	logistics.SetEstimatedTime(estimatedTime)
	return s.repo.Save(ctx, logistics)
}

// OptimizeDeliveryRoute 优化配送路线。
func (s *LogisticsCommandService) OptimizeDeliveryRoute(ctx context.Context, logisticsID uint64, destinations []algorithm.Location) (*domain.DeliveryRoute, error) {
	logistics, err := s.repo.GetByID(ctx, logisticsID)
	if err != nil {
		return nil, err
	}

	start := algorithm.Location{
		ID:     0,
		Lat:    logistics.SenderLat,
		Lon:    logistics.SenderLon,
		Demand: 0,
	}

	for i := range destinations {
		if destinations[i].Demand == 0 {
			destinations[i].Demand = 10.0
		}
	}

	route := s.optimizer.OptimizeRoute(start, destinations)
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
	if err := s.repo.Save(ctx, logistics); err != nil {
		return nil, err
	}

	return deliveryRoute, nil
}

// CalculatePackaging 计算订单的打包方案
func (s *LogisticsCommandService) CalculatePackaging(items []algorithm.Item) []*algorithm.Bin {
	return s.packingOptimizer.FFD(items)
}
