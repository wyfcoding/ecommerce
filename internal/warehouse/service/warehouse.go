package service

import (
	"context"
	"ecommerce/pkg/algorithm"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ecommerce/internal/warehouse/model"
	"ecommerce/internal/warehouse/repository"
	"ecommerce/pkg/idgen"
)

var (
	ErrWarehouseNotFound     = errors.New("仓库不存在")
	ErrWarehouseInactive     = errors.New("仓库未启用")
	ErrInsufficientStock     = errors.New("库存不足")
	ErrStockTransferNotFound = errors.New("调拨单不存在")
	ErrInvalidTransferStatus = errors.New("无效的调拨状态")
	ErrStocktakingNotFound   = errors.New("盘点单不存在")
)

// WarehouseService 仓库服务接口
type WarehouseService interface {
	// 仓库管理
	CreateWarehouse(ctx context.Context, warehouse *model.Warehouse) (*model.Warehouse, error)
	UpdateWarehouse(ctx context.Context, warehouse *model.Warehouse) (*model.Warehouse, error)
	GetWarehouse(ctx context.Context, id uint64) (*model.Warehouse, error)
	ListWarehouses(ctx context.Context, status string, pageSize, pageNum int32) ([]*model.Warehouse, int64, error)

	// 仓库库存
	GetWarehouseStock(ctx context.Context, warehouseID, skuID uint64) (*model.WarehouseStock, error)
	ListWarehouseStocks(ctx context.Context, warehouseID uint64, pageSize, pageNum int32) ([]*model.WarehouseStock, int64, error)
	UpdateWarehouseStock(ctx context.Context, warehouseID, skuID uint64, quantity int32) error

	// 智能仓库选择
	SelectOptimalWarehouse(ctx context.Context, skuID uint64, quantity int32, province, city string) (*model.Warehouse, error)
	AllocateStock(ctx context.Context, orderID, orderItemID, skuID uint64, quantity int32, province, city string) (*model.StockAllocation, error)

	// 库存调拨
	CreateStockTransfer(ctx context.Context, transfer *model.StockTransfer) (*model.StockTransfer, error)
	ApproveStockTransfer(ctx context.Context, transferID, approverID uint64) error
	ShipStockTransfer(ctx context.Context, transferID uint64) error
	ReceiveStockTransfer(ctx context.Context, transferID uint64) error
	CancelStockTransfer(ctx context.Context, transferID uint64, reason string) error
	GetStockTransfer(ctx context.Context, transferNo string) (*model.StockTransfer, error)

	// 库存盘点
	CreateStocktaking(ctx context.Context, stocktaking *model.Stocktaking) (*model.Stocktaking, error)
	StartStocktaking(ctx context.Context, stocktakingID uint64) error
	RecordStocktakingItem(ctx context.Context, item *model.StocktakingItem) error
	CompleteStocktaking(ctx context.Context, stocktakingID uint64) error
	GetStocktaking(ctx context.Context, stockNo string) (*model.Stocktaking, error)
}

type warehouseService struct {
	repo        repository.WarehouseRepo
	redisClient *redis.Client
	logger      *zap.Logger
	allocator   *algorithm.WarehouseAllocator
}

// NewWarehouseService 创建仓库服务实例
func NewWarehouseService(
	repo repository.WarehouseRepo,
	redisClient *redis.Client,
	logger *zap.Logger,
) WarehouseService {
	return &warehouseService{
		repo:        repo,
		redisClient: redisClient,
		logger:      logger,
		allocator:   algorithm.NewWarehouseAllocator(),
	}
}

// CreateWarehouse 创建仓库
func (s *warehouseService) CreateWarehouse(ctx context.Context, warehouse *model.Warehouse) (*model.Warehouse, error) {
	warehouse.CreatedAt = time.Now()
	warehouse.UpdatedAt = time.Now()

	if err := s.repo.CreateWarehouse(ctx, warehouse); err != nil {
		s.logger.Error("创建仓库失败", zap.Error(err))
		return nil, err
	}

	s.logger.Info("创建仓库成功", zap.Uint64("warehouseID", warehouse.ID))
	return warehouse, nil
}

// UpdateWarehouse 更新仓库
func (s *warehouseService) UpdateWarehouse(ctx context.Context, warehouse *model.Warehouse) (*model.Warehouse, error) {
	warehouse.UpdatedAt = time.Now()

	if err := s.repo.UpdateWarehouse(ctx, warehouse); err != nil {
		s.logger.Error("更新仓库失败", zap.Error(err))
		return nil, err
	}

	return warehouse, nil
}

// GetWarehouse 获取仓库详情
func (s *warehouseService) GetWarehouse(ctx context.Context, id uint64) (*model.Warehouse, error) {
	warehouse, err := s.repo.GetWarehouseByID(ctx, id)
	if err != nil {
		return nil, ErrWarehouseNotFound
	}
	return warehouse, nil
}

// ListWarehouses 获取仓库列表
func (s *warehouseService) ListWarehouses(ctx context.Context, status string, pageSize, pageNum int32) ([]*model.Warehouse, int64, error) {
	warehouses, total, err := s.repo.ListWarehouses(ctx, status, pageSize, pageNum)
	if err != nil {
		s.logger.Error("获取仓库列表失败", zap.Error(err))
		return nil, 0, err
	}
	return warehouses, total, nil
}

// GetWarehouseStock 获取仓库库存
func (s *warehouseService) GetWarehouseStock(ctx context.Context, warehouseID, skuID uint64) (*model.WarehouseStock, error) {
	stock, err := s.repo.GetWarehouseStock(ctx, warehouseID, skuID)
	if err != nil {
		return nil, err
	}
	return stock, nil
}

// ListWarehouseStocks 获取仓库库存列表
func (s *warehouseService) ListWarehouseStocks(ctx context.Context, warehouseID uint64, pageSize, pageNum int32) ([]*model.WarehouseStock, int64, error) {
	stocks, total, err := s.repo.ListWarehouseStocks(ctx, warehouseID, pageSize, pageNum)
	if err != nil {
		s.logger.Error("获取仓库库存列表失败", zap.Error(err))
		return nil, 0, err
	}
	return stocks, total, nil
}

// UpdateWarehouseStock 更新仓库库存
func (s *warehouseService) UpdateWarehouseStock(ctx context.Context, warehouseID, skuID uint64, quantity int32) error {
	if err := s.repo.UpdateWarehouseStock(ctx, warehouseID, skuID, quantity); err != nil {
		s.logger.Error("更新仓库库存失败", zap.Error(err))
		return err
	}
	return nil
}

// SelectOptimalWarehouse 智能选择最优仓库
func (s *warehouseService) SelectOptimalWarehouse(ctx context.Context, skuID uint64, quantity int32, province, city string) (*model.Warehouse, error) {
	// 1. 获取所有启用的仓库
	warehouses, _, err := s.repo.ListWarehouses(ctx, string(model.WarehouseStatusActive), 100, 1)
	if err != nil {
		return nil, err
	}

	if len(warehouses) == 0 {
		return nil, ErrWarehouseNotFound
	}

	// 2. 筛选有库存的仓库
	var availableWarehouses []*model.Warehouse
	for _, warehouse := range warehouses {
		stock, err := s.repo.GetWarehouseStock(ctx, warehouse.ID, skuID)
		if err != nil {
			continue
		}
		if stock.Stock >= quantity {
			availableWarehouses = append(availableWarehouses, warehouse)
		}
	}

	if len(availableWarehouses) == 0 {
		return nil, ErrInsufficientStock
	}

	// 3. 选择最优仓库（优先级：同城 > 同省 > 优先级高 > 距离近）
	var bestWarehouse *model.Warehouse
	var bestScore float64 = -1

	for _, warehouse := range availableWarehouses {
		score := s.calculateWarehouseScore(warehouse, province, city)
		if score > bestScore {
			bestScore = score
			bestWarehouse = warehouse
		}
	}

	return bestWarehouse, nil
}

// calculateWarehouseScore 计算仓库评分
func (s *warehouseService) calculateWarehouseScore(warehouse *model.Warehouse, province, city string) float64 {
	score := float64(warehouse.Priority) * 10 // 优先级基础分

	// 同城加分
	if warehouse.City == city {
		score += 100
	}

	// 同省加分
	if warehouse.Province == province {
		score += 50
	}

	return score
}

// calculateDistance 计算距离（简化版，实际应使用地理位置计算）
func (s *warehouseService) calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	// 使用Haversine公式计算两点间距离
	const R = 6371 // 地球半径（公里）

	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// AllocateStock 分配库存
func (s *warehouseService) AllocateStock(ctx context.Context, orderID, orderItemID, skuID uint64, quantity int32, province, city string) (*model.StockAllocation, error) {
	// 1. 选择最优仓库
	warehouse, err := s.SelectOptimalWarehouse(ctx, skuID, quantity, province, city)
	if err != nil {
		return nil, err
	}

	// 2. 创建分配记录
	allocation := &model.StockAllocation{
		OrderID:     orderID,
		OrderItemID: orderItemID,
		WarehouseID: warehouse.ID,
		SkuID:       skuID,
		Quantity:    quantity,
		Status:      "ALLOCATED",
		AllocatedAt: time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.CreateStockAllocation(ctx, allocation); err != nil {
		s.logger.Error("创建库存分配记录失败", zap.Error(err))
		return nil, err
	}

	s.logger.Info("库存分配成功",
		zap.Uint64("orderID", orderID),
		zap.Uint64("warehouseID", warehouse.ID),
		zap.Uint64("skuID", skuID))

	return allocation, nil
}

// CreateStockTransfer 创建库存调拨
func (s *warehouseService) CreateStockTransfer(ctx context.Context, transfer *model.StockTransfer) (*model.StockTransfer, error) {
	// 1. 验证源仓库库存
	fromStock, err := s.repo.GetWarehouseStock(ctx, transfer.FromWarehouseID, transfer.SkuID)
	if err != nil {
		return nil, err
	}

	if fromStock.Stock < transfer.Quantity {
		return nil, ErrInsufficientStock
	}

	// 2. 生成调拨单号
	transfer.TransferNo = fmt.Sprintf("TF%d", idgen.GenID())
	transfer.Status = model.StockTransferStatusPending
	transfer.CreatedAt = time.Now()
	transfer.UpdatedAt = time.Now()

	// 3. 创建调拨单
	if err := s.repo.CreateStockTransfer(ctx, transfer); err != nil {
		s.logger.Error("创建库存调拨失败", zap.Error(err))
		return nil, err
	}

	s.logger.Info("创建库存调拨成功", zap.String("transferNo", transfer.TransferNo))
	return transfer, nil
}

// ApproveStockTransfer 审核调拨单
func (s *warehouseService) ApproveStockTransfer(ctx context.Context, transferID, approverID uint64) error {
	transfer, err := s.repo.GetStockTransferByID(ctx, transferID)
	if err != nil {
		return ErrStockTransferNotFound
	}

	if transfer.Status != model.StockTransferStatusPending {
		return ErrInvalidTransferStatus
	}

	now := time.Now()
	transfer.Status = model.StockTransferStatusApproved
	transfer.ApprovedBy = approverID
	transfer.ApprovedAt = &now
	transfer.UpdatedAt = now

	if err := s.repo.UpdateStockTransfer(ctx, transfer); err != nil {
		s.logger.Error("审核调拨单失败", zap.Error(err))
		return err
	}

	return nil
}

// ShipStockTransfer 发货
func (s *warehouseService) ShipStockTransfer(ctx context.Context, transferID uint64) error {
	transfer, err := s.repo.GetStockTransferByID(ctx, transferID)
	if err != nil {
		return ErrStockTransferNotFound
	}

	if transfer.Status != model.StockTransferStatusApproved {
		return ErrInvalidTransferStatus
	}

	// 在事务中执行：扣减源仓库库存、更新调拨状态
	err = s.repo.InTx(ctx, func(ctx context.Context) error {
		// 扣减源仓库库存
		if err := s.repo.DeductWarehouseStock(ctx, transfer.FromWarehouseID, transfer.SkuID, transfer.Quantity); err != nil {
			return err
		}

		// 更新调拨状态
		now := time.Now()
		transfer.Status = model.StockTransferStatusShipped
		transfer.ShippedAt = &now
		transfer.UpdatedAt = now

		if err := s.repo.UpdateStockTransfer(ctx, transfer); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		s.logger.Error("发货失败", zap.Error(err))
		return err
	}

	return nil
}

// ReceiveStockTransfer 收货
func (s *warehouseService) ReceiveStockTransfer(ctx context.Context, transferID uint64) error {
	transfer, err := s.repo.GetStockTransferByID(ctx, transferID)
	if err != nil {
		return ErrStockTransferNotFound
	}

	if transfer.Status != model.StockTransferStatusShipped {
		return ErrInvalidTransferStatus
	}

	// 在事务中执行：增加目标仓库库存、更新调拨状态
	err = s.repo.InTx(ctx, func(ctx context.Context) error {
		// 增加目标仓库库存
		if err := s.repo.AddWarehouseStock(ctx, transfer.ToWarehouseID, transfer.SkuID, transfer.Quantity); err != nil {
			return err
		}

		// 更新调拨状态
		now := time.Now()
		transfer.Status = model.StockTransferStatusReceived
		transfer.ReceivedAt = &now
		transfer.UpdatedAt = now

		if err := s.repo.UpdateStockTransfer(ctx, transfer); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		s.logger.Error("收货失败", zap.Error(err))
		return err
	}

	// 自动完成
	return s.completeStockTransfer(ctx, transferID)
}

// completeStockTransfer 完成调拨
func (s *warehouseService) completeStockTransfer(ctx context.Context, transferID uint64) error {
	transfer, err := s.repo.GetStockTransferByID(ctx, transferID)
	if err != nil {
		return err
	}

	now := time.Now()
	transfer.Status = model.StockTransferStatusCompleted
	transfer.CompletedAt = &now
	transfer.UpdatedAt = now

	if err := s.repo.UpdateStockTransfer(ctx, transfer); err != nil {
		s.logger.Error("完成调拨失败", zap.Error(err))
		return err
	}

	return nil
}

// CancelStockTransfer 取消调拨
func (s *warehouseService) CancelStockTransfer(ctx context.Context, transferID uint64, reason string) error {
	transfer, err := s.repo.GetStockTransferByID(ctx, transferID)
	if err != nil {
		return ErrStockTransferNotFound
	}

	if transfer.Status == model.StockTransferStatusCompleted || transfer.Status == model.StockTransferStatusCancelled {
		return ErrInvalidTransferStatus
	}

	// 如果已发货，需要恢复源仓库库存
	if transfer.Status == model.StockTransferStatusShipped {
		if err := s.repo.AddWarehouseStock(ctx, transfer.FromWarehouseID, transfer.SkuID, transfer.Quantity); err != nil {
			return err
		}
	}

	transfer.Status = model.StockTransferStatusCancelled
	transfer.Remark = reason
	transfer.UpdatedAt = time.Now()

	if err := s.repo.UpdateStockTransfer(ctx, transfer); err != nil {
		s.logger.Error("取消调拨失败", zap.Error(err))
		return err
	}

	return nil
}

// GetStockTransfer 获取调拨单详情
func (s *warehouseService) GetStockTransfer(ctx context.Context, transferNo string) (*model.StockTransfer, error) {
	transfer, err := s.repo.GetStockTransferByNo(ctx, transferNo)
	if err != nil {
		return nil, ErrStockTransferNotFound
	}
	return transfer, nil
}

// CreateStocktaking 创建库存盘点
func (s *warehouseService) CreateStocktaking(ctx context.Context, stocktaking *model.Stocktaking) (*model.Stocktaking, error) {
	stocktaking.StockNo = fmt.Sprintf("ST%d", idgen.GenID())
	stocktaking.Status = model.StocktakingStatusPending
	stocktaking.CreatedAt = time.Now()
	stocktaking.UpdatedAt = time.Now()

	if err := s.repo.CreateStocktaking(ctx, stocktaking); err != nil {
		s.logger.Error("创建库存盘点失败", zap.Error(err))
		return nil, err
	}

	s.logger.Info("创建库存盘点成功", zap.String("stockNo", stocktaking.StockNo))
	return stocktaking, nil
}

// StartStocktaking 开始盘点
func (s *warehouseService) StartStocktaking(ctx context.Context, stocktakingID uint64) error {
	stocktaking, err := s.repo.GetStocktakingByID(ctx, stocktakingID)
	if err != nil {
		return ErrStocktakingNotFound
	}

	if stocktaking.Status != model.StocktakingStatusPending {
		return fmt.Errorf("盘点单状态不允许开始")
	}

	stocktaking.Status = model.StocktakingStatusInProgress
	stocktaking.StartTime = time.Now()
	stocktaking.UpdatedAt = time.Now()

	if err := s.repo.UpdateStocktaking(ctx, stocktaking); err != nil {
		s.logger.Error("开始盘点失败", zap.Error(err))
		return err
	}

	return nil
}

// RecordStocktakingItem 记录盘点明细
func (s *warehouseService) RecordStocktakingItem(ctx context.Context, item *model.StocktakingItem) error {
	// 计算差异
	item.CalculateDifference()
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()

	if err := s.repo.CreateStocktakingItem(ctx, item); err != nil {
		s.logger.Error("记录盘点明细失败", zap.Error(err))
		return err
	}

	return nil
}

// CompleteStocktaking 完成盘点
func (s *warehouseService) CompleteStocktaking(ctx context.Context, stocktakingID uint64) error {
	stocktaking, err := s.repo.GetStocktakingByID(ctx, stocktakingID)
	if err != nil {
		return ErrStocktakingNotFound
	}

	if stocktaking.Status != model.StocktakingStatusInProgress {
		return fmt.Errorf("盘点单状态不允许完成")
	}

	// 获取盘点明细
	items, err := s.repo.ListStocktakingItems(ctx, stocktakingID)
	if err != nil {
		return err
	}

	// 在事务中执行：调整库存、更新盘点状态
	err = s.repo.InTx(ctx, func(ctx context.Context) error {
		// 根据盘点结果调整库存
		for _, item := range items {
			if item.Difference != 0 {
				// 更新仓库库存
				if err := s.repo.AdjustWarehouseStock(ctx, stocktaking.WarehouseID, item.SkuID, item.Difference); err != nil {
					return err
				}
			}
		}

		// 更新盘点状态
		now := time.Now()
		stocktaking.Status = model.StocktakingStatusCompleted
		stocktaking.EndTime = &now
		stocktaking.UpdatedAt = now

		if err := s.repo.UpdateStocktaking(ctx, stocktaking); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		s.logger.Error("完成盘点失败", zap.Error(err))
		return err
	}

	s.logger.Info("完成盘点成功", zap.Uint64("stocktakingID", stocktakingID))
	return nil
}

// GetStocktaking 获取盘点单详情
func (s *warehouseService) GetStocktaking(ctx context.Context, stockNo string) (*model.Stocktaking, error) {
	stocktaking, err := s.repo.GetStocktakingByNo(ctx, stockNo)
	if err != nil {
		return nil, ErrStocktakingNotFound
	}
	return stocktaking, nil
}

// AllocateStockOptimal 使用智能算法分配库存（支持多仓库拆单）
func (s *warehouseService) AllocateStockOptimal(
	ctx context.Context,
	orderID uint64,
	items []algorithm.OrderItem,
	userLat, userLon float64,
) ([]algorithm.AllocationResult, error) {

	// 1. 获取所有启用的仓库
	warehouses, _, err := s.repo.ListWarehouses(ctx, string(model.WarehouseStatusActive), 100, 1)
	if err != nil {
		return nil, err
	}

	if len(warehouses) == 0 {
		return nil, ErrWarehouseNotFound
	}

	// 2. 构建仓库库存数据
	warehouseData := make(map[uint64]map[uint64]*algorithm.WarehouseInfo)

	for _, warehouse := range warehouses {
		warehouseData[warehouse.ID] = make(map[uint64]*algorithm.WarehouseInfo)

		// 获取该仓库的所有库存
		stocks, _, err := s.repo.ListWarehouseStocks(ctx, warehouse.ID, 1000, 1)
		if err != nil {
			continue
		}

		for _, stock := range stocks {
			warehouseData[warehouse.ID][stock.SkuID] = &algorithm.WarehouseInfo{
				ID:       warehouse.ID,
				Lat:      warehouse.Latitude,
				Lon:      warehouse.Longitude,
				Stock:    stock.Stock,
				Priority: warehouse.Priority,
				ShipCost: 500, // TODO: 从配置或数据库获取实际配送成本
			}
		}
	}

	// 3. 使用智能算法分配
	results := s.allocator.AllocateOptimal(userLat, userLon, items, warehouseData)

	// 4. 记录分配结果到数据库
	for _, result := range results {
		for _, item := range result.Items {
			allocation := &model.StockAllocation{
				OrderID:     orderID,
				WarehouseID: result.WarehouseID,
				SkuID:       item.SkuID,
				Quantity:    item.Quantity,
				Status:      "ALLOCATED",
				AllocatedAt: time.Now(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			if err := s.repo.CreateStockAllocation(ctx, allocation); err != nil {
				s.logger.Error("创建库存分配记录失败", zap.Error(err))
				// 继续处理其他分配
			}
		}
	}

	s.logger.Info("智能库存分配成功",
		zap.Uint64("orderID", orderID),
		zap.Int("warehouseCount", len(results)),
		zap.Float64("userLat", userLat),
		zap.Float64("userLon", userLon))

	return results, nil
}
