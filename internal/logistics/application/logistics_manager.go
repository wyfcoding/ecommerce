package application

import (
	"context"
	"encoding/json"
	"log/slog"
	"math"
	"time"

	"github.com/wyfcoding/ecommerce/internal/logistics/domain"
	"github.com/wyfcoding/pkg/algorithm"
)

// LogisticsManager 处理物流的写操作（创建、状态更新、轨迹追踪、路线优化）。
type LogisticsManager struct {
	repo             domain.LogisticsRepository
	optimizer        *algorithm.RouteOptimizer
	packingOptimizer *algorithm.BinPackingOptimizer
	logger           *slog.Logger
}

// RiderInfo 骑手信息
type RiderInfo struct {
	ID  string
	Lat float64
	Lon float64
}

// OrderInfo 订单配送信息
type OrderInfo struct {
	ID  string
	Lat float64
	Lon float64
}

// AssignRidersToOrders 将骑手分配给订单 (最小总距离指派)，并更新数据库。
func (m *LogisticsManager) AssignRidersToOrders(ctx context.Context, riders []RiderInfo, logisticsIDs []uint64) (map[string]uint64, error) {
	if len(riders) == 0 || len(logisticsIDs) == 0 {
		return nil, nil
	}

	// 1. 获取物流单信息
	var logisticsList []*domain.Logistics
	var orders []OrderInfo
	// 维护索引映射: orders index -> logisticsList index
	var validIndices []int

	for _, id := range logisticsIDs {
		logistics, err := m.repo.GetByID(ctx, id)
		if err != nil {
			m.logger.WarnContext(ctx, "logistics not found for assignment", "id", id, "error", err)
			continue
		}
		// 仅分配待揽收状态的订单
		if logistics.Status != domain.LogisticsStatusPending && logistics.Status != domain.LogisticsStatusPickedUp {
			continue 
		}

		logisticsList = append(logisticsList, logistics)
		orders = append(orders, OrderInfo{
			ID:  logistics.OrderNo, // Assuming OrderNo is unique enough for logic, but we use index mostly
			Lat: logistics.SenderLat, // 骑手前往发件人位置
			Lon: logistics.SenderLon,
		})
		validIndices = append(validIndices, len(logisticsList)-1)
	}

	if len(orders) == 0 {
		return nil, nil
	}

	// 2. 构建二分图并求解 (KM算法)
	nx := len(riders)
	ny := len(orders)
	size := nx
	if ny > size {
		size = ny
	}

	bg := algorithm.NewWeightedBipartiteGraph(size, size)

	for i, rider := range riders {
		for j, order := range orders {
			dist := m.calculateDistance(rider.Lat, rider.Lon, order.Lat, order.Lon)
			bg.SetWeight(i, j, -dist) // 负权重求最大匹配 = 最小距离
		}
	}

	bg.Solve()
	match := bg.GetMatch()

	// 3. 应用分配结果
	result := make(map[string]uint64)
	for rIdx, oIdx := range match {
		if rIdx < len(riders) && oIdx < len(orders) {
			riderID := riders[rIdx].ID
			// 找回对应的 Logistics 实体
			// match oIdx corresponds to orders[oIdx] which corresponds to logisticsList[oIdx] (since we appended linearly)
			logistics := logisticsList[oIdx]
			
			logistics.AssignRider(riderID)
			logistics.Status = domain.LogisticsStatusPickedUp // 假设分配即揽收，或改为 PendingPickup
			
			if err := m.repo.Save(ctx, logistics); err != nil {
				m.logger.ErrorContext(ctx, "failed to save assigned rider", "logistics_id", logistics.ID, "rider_id", riderID, "error", err)
				continue
			}
			
			result[riderID] = uint64(logistics.ID)
		}
	}

	return result, nil
}

// calculateDistance 计算两点间的欧几里得距离 (简化版)
func (m *LogisticsManager) calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	dx := lat1 - lat2
	dy := lon1 - lon2
	return math.Sqrt(dx*dx + dy*dy)
}

// NewLogisticsManager 负责处理 NewLogistics 相关的写操作和业务逻辑。
func NewLogisticsManager(repo domain.LogisticsRepository, logger *slog.Logger) *LogisticsManager {
	return &LogisticsManager{
		repo:             repo,
		optimizer:        algorithm.NewRouteOptimizer(),
		packingOptimizer: algorithm.NewBinPackingOptimizer(1000.0), // 假设标准箱体积为 1000
		logger:           logger,
	}
}

// CalculatePackaging 计算订单的打包方案
func (m *LogisticsManager) CalculatePackaging(items []algorithm.Item) []*algorithm.Bin {
	return m.packingOptimizer.FFD(items)
}

// CreateLogistics 创建一个新的物流单。
func (m *LogisticsManager) CreateLogistics(ctx context.Context, orderID uint64, orderNo, trackingNo, carrier, carrierCode string,
	senderName, senderPhone, senderAddress string, senderLat, senderLon float64,
	receiverName, receiverPhone, receiverAddress string, receiverLat, receiverLon float64,
) (*domain.Logistics, error) {
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
