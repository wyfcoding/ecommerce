package application

import (
	"context"
	"fmt"
	"math"

	"github.com/wyfcoding/ecommerce/internal/warehouse/domain"
	"github.com/wyfcoding/pkg/utils/geo"
)

// WarehouseQueryService 处理所有仓库相关的查询操作 (Queries)。
type WarehouseQueryService struct {
	repo domain.WarehouseRepository
}

// NewWarehouseQueryService 构造函数。
func NewWarehouseQueryService(repo domain.WarehouseRepository) *WarehouseQueryService {
	return &WarehouseQueryService{repo: repo}
}

// GetWarehouse 获取详情。
func (s *WarehouseQueryService) GetWarehouse(ctx context.Context, id uint64) (*domain.Warehouse, error) {
	return s.repo.FindByID(ctx, id)
}

// ListWarehouses 列表。
func (s *WarehouseQueryService) ListWarehouses(ctx context.Context, page, pageSize int) ([]*domain.Warehouse, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, offset, pageSize)
}

// GetStock 查看当前库存。
func (s *WarehouseQueryService) GetStock(ctx context.Context, warehouseID, skuID uint64) (*domain.WarehouseStock, error) {
	return s.repo.GetStock(ctx, warehouseID, skuID)
}

// GetOptimalWarehouse 调度模型：根据地理位置寻找最近且有货的仓库。
func (s *WarehouseQueryService) GetOptimalWarehouse(ctx context.Context, skuID uint64, qty int32, lat, lon float64) (*domain.Warehouse, float64, int32, error) {
	// 1. 获取所有可用仓库。实际生产中应当先通过省市区或经纬度范围初步过滤。
	warehouses, _, err := s.repo.List(ctx, 0, 100)
	if err != nil {
		return nil, 0, 0, err
	}

	var bestW *domain.Warehouse
	minDistance := math.MaxFloat64
	var availQty int32

	for _, w := range warehouses {
		if w.Status != domain.WarehouseStatusActive {
			continue
		}

		// 2. 检查库存是否满足。
		stock, err := s.repo.GetStock(ctx, uint64(w.ID), skuID)
		if err != nil || stock == nil || stock.AvailableStock() < qty {
			continue
		}

		// 3. 计算地理距离 (Haversine)。
		dist := geo.HaversineDistance(lat, lon, w.Latitude, w.Longitude)

		// 4. 择优策略：距离优先。
		if dist < minDistance {
			minDistance = dist
			bestW = w
			availQty = stock.AvailableStock()
		}
	}

	if bestW == nil {
		return nil, 0, 0, fmt.Errorf("no active warehouse found with sufficient stock for SKU %d", skuID)
	}

	return bestW, minDistance, availQty, nil
}

// GetTransfer 获取调拨单。
func (s *WarehouseQueryService) GetTransfer(ctx context.Context, id uint64) (*domain.StockTransfer, error) {
	return s.repo.FindTransferByNo(ctx, fmt.Sprintf("%d", id))
}

// ListTransfers 列表分页。
func (s *WarehouseQueryService) ListTransfers(ctx context.Context, fromID, toID uint64, status *string, page, pageSize int) ([]*domain.StockTransfer, int64, error) {
	// 简化实现
	return nil, 0, nil
}
