package application

import (
	"context"
	"encoding/json" // 导入JSON编码/解码库。
	"time"          // 导入时间库。

	"github.com/wyfcoding/ecommerce/internal/logistics/domain/entity"     // 导入物流领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/logistics/domain/repository" // 导入物流领域的仓储接口。
	"github.com/wyfcoding/ecommerce/pkg/algorithm"                        // 导入算法包，用于路线优化。

	"log/slog" // 导入结构化日志库。
)

// LogisticsService 结构体定义了物流管理相关的应用服务。
// 它协调领域层和基础设施层，处理物流单的创建、状态更新、轨迹追踪和配送路线优化等业务逻辑。
type LogisticsService struct {
	repo      repository.LogisticsRepository // 依赖LogisticsRepository接口，用于数据持久化操作。
	optimizer *algorithm.RouteOptimizer      // 依赖路由优化器，用于计算最优配送路线。
	logger    *slog.Logger                   // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewLogisticsService 创建并返回一个新的 LogisticsService 实例。
func NewLogisticsService(repo repository.LogisticsRepository, logger *slog.Logger) *LogisticsService {
	return &LogisticsService{
		repo:      repo,
		optimizer: algorithm.NewRouteOptimizer(), // 初始化路由优化器。
		logger:    logger,
	}
}

// CreateLogistics 创建一个新的物流单。
// ctx: 上下文。
// orderID, orderNo: 关联的订单ID和订单号。
// trackingNo, carrier, carrierCode: 运单号、承运商和承运商编码。
// senderName, senderPhone, senderAddress, senderLat, senderLon: 发件人信息及位置。
// receiverName, receiverPhone, receiverAddress, receiverLat, receiverLon: 收件人信息及位置。
// 返回created successfully的Logistics实体和可能发生的错误。
func (s *LogisticsService) CreateLogistics(ctx context.Context, orderID uint64, orderNo, trackingNo, carrier, carrierCode string,
	senderName, senderPhone, senderAddress string, senderLat, senderLon float64,
	receiverName, receiverPhone, receiverAddress string, receiverLat, receiverLon float64) (*entity.Logistics, error) {

	logistics := entity.NewLogistics(orderID, orderNo, trackingNo, carrier, carrierCode,
		senderName, senderPhone, senderAddress, senderLat, senderLon,
		receiverName, receiverPhone, receiverAddress, receiverLat, receiverLon) // 创建Logistics实体。

	// 通过仓储接口保存物流单。
	if err := s.repo.Save(ctx, logistics); err != nil {
		s.logger.ErrorContext(ctx, "failed to save logistics", "order_id", orderID, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "logistics created successfully", "logistics_id", logistics.ID, "tracking_no", trackingNo)
	return logistics, nil
}

// GetLogistics 获取指定ID的物流信息。
// ctx: 上下文。
// id: 物流单ID。
// 返回Logistics实体和可能发生的错误。
func (s *LogisticsService) GetLogistics(ctx context.Context, id uint64) (*entity.Logistics, error) {
	return s.repo.GetByID(ctx, id)
}

// GetLogisticsByTrackingNo 根据运单号获取物流信息。
// ctx: 上下文。
// trackingNo: 运单号。
// 返回Logistics实体和可能发生的错误。
func (s *LogisticsService) GetLogisticsByTrackingNo(ctx context.Context, trackingNo string) (*entity.Logistics, error) {
	return s.repo.GetByTrackingNo(ctx, trackingNo)
}

// UpdateStatus 更新物流单状态。
// ctx: 上下文。
// id: 物流单ID。
// status: 新的物流状态。
// location: 状态更新时的位置信息。
// description: 状态更新的描述。
// 返回可能发生的错误。
func (s *LogisticsService) UpdateStatus(ctx context.Context, id uint64, status entity.LogisticsStatus, location, description string) error {
	// 获取物流单实体。
	logistics, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 根据新的状态调用实体的方法进行状态转换。
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
		return entity.ErrInvalidStatus // 无效的状态转换。
	}

	// 记录物流轨迹。
	// TODO: status参数应从entity.LogisticsStatus映射为字符串。
	logistics.AddTrace(location, description, "")

	// 通过仓储接口保存更新后的物流单。
	if err := s.repo.Save(ctx, logistics); err != nil {
		s.logger.ErrorContext(ctx, "failed to update logistics status", "logistics_id", id, "status", status, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "logistics status updated successfully", "logistics_id", id, "status", status)
	return nil
}

// AddTrace 添加物流轨迹记录。
// ctx: 上下文。
// id: 物流单ID。
// location: 轨迹发生的位置。
// description: 轨迹描述。
// status: 轨迹发生时的物流状态（字符串形式）。
// 返回可能发生的错误。
func (s *LogisticsService) AddTrace(ctx context.Context, id uint64, location, description, status string) error {
	// 获取物流单实体。
	logistics, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 调用实体方法添加轨迹。
	logistics.AddTrace(location, description, status)
	// 更新物流单的当前位置。
	logistics.UpdateLocation(location)

	// 通过仓储接口保存更新后的物流单。
	if err := s.repo.Save(ctx, logistics); err != nil {
		s.logger.ErrorContext(ctx, "failed to add trace", "logistics_id", id, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "logistics trace added successfully", "logistics_id", id, "location", location)
	return nil
}

// SetEstimatedTime 设置物流单的预计送达时间。
// ctx: 上下文。
// id: 物流单ID。
// estimatedTime: 预计送达时间。
// 返回可能发生的错误。
func (s *LogisticsService) SetEstimatedTime(ctx context.Context, id uint64, estimatedTime time.Time) error {
	// 获取物流单实体。
	logistics, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 调用实体方法设置预计送达时间。
	logistics.SetEstimatedTime(estimatedTime)
	// 通过仓储接口保存更新后的物流单。
	if err := s.repo.Save(ctx, logistics); err != nil {
		s.logger.ErrorContext(ctx, "failed to set estimated time", "logistics_id", id, "error", err)
		return err
	}
	s.logger.InfoContext(ctx, "logistics estimated time set successfully", "logistics_id", id, "estimated_time", estimatedTime)
	return nil
}

// ListLogistics 获取物流单列表。
// ctx: 上下文。
// page, pageSize: 分页参数。
// 返回物流单列表、总数和可能发生的错误。
func (s *LogisticsService) ListLogistics(ctx context.Context, page, pageSize int) ([]*entity.Logistics, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, offset, pageSize)
}

// OptimizeDeliveryRoute 优化配送路线。
// ctx: 上下文。
// logisticsID: 关联的物流单ID。
// destinations: 待访问的目的地列表。
// 返回计算出的DeliveryRoute实体和可能发生的错误。
func (s *LogisticsService) OptimizeDeliveryRoute(ctx context.Context, logisticsID uint64, destinations []algorithm.Location) (*entity.DeliveryRoute, error) {
	// 获取物流单实体，以获取发件人位置作为起始点。
	logistics, err := s.repo.GetByID(ctx, logisticsID)
	if err != nil {
		return nil, err
	}

	// 构建起始点位置信息。
	start := algorithm.Location{
		ID:  0, // 起始点ID通常为0或特定值。
		Lat: logistics.SenderLat,
		Lon: logistics.SenderLon,
	}

	// 调用路由优化器计算最优路线。
	route := s.optimizer.OptimizeRoute(start, destinations)

	// 将优化后的路线数据序列化为JSON。
	routeData, err := json.Marshal(route.Locations)
	if err != nil {
		return nil, err
	}

	// 创建DeliveryRoute实体。
	deliveryRoute := &entity.DeliveryRoute{
		LogisticsID: logisticsID,
		RouteData:   string(routeData),
		Distance:    route.Distance,
	}

	// 将优化后的路线关联到物流单并保存。
	// GORM会根据关联关系自动保存或更新 DeliveryRoute 实体。
	logistics.Route = deliveryRoute
	if err := s.repo.Save(ctx, logistics); err != nil {
		s.logger.ErrorContext(ctx, "failed to save optimized route", "logistics_id", logisticsID, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "delivery route optimized successfully", "logistics_id", logisticsID)

	return deliveryRoute, nil
}
