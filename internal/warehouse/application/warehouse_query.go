package application

import (
	"context"
	"errors"
	"math"

	"github.com/wyfcoding/ecommerce/internal/warehouse/domain"
)

// WarehouseQuery 处理仓库模块的查询操作。
type WarehouseQuery struct {
	repo domain.WarehouseRepository
}

// NewWarehouseQuery 创建并返回一个新的 WarehouseQuery 实例。
func NewWarehouseQuery(repo domain.WarehouseRepository) *WarehouseQuery {
	return &WarehouseQuery{repo: repo}
}

// GetWarehouseByID 根据ID获取仓库详情。
func (q *WarehouseQuery) GetWarehouseByID(ctx context.Context, id uint64) (*domain.Warehouse, error) {
	return q.repo.GetWarehouse(ctx, id)
}

// GetWarehouseByCode 根据代码获取仓库详情。
func (q *WarehouseQuery) GetWarehouseByCode(ctx context.Context, code string) (*domain.Warehouse, error) {
	return q.repo.GetWarehouseByCode(ctx, code)
}

// SearchWarehouses 搜索仓库。
func (q *WarehouseQuery) SearchWarehouses(ctx context.Context, status *domain.WarehouseStatus, offset, limit int) ([]*domain.Warehouse, int64, error) {
	return q.repo.ListWarehouses(ctx, status, offset, limit)
}

// GetStock 获取库存信息。
func (q *WarehouseQuery) GetStock(ctx context.Context, warehouseID, skuID uint64) (*domain.WarehouseStock, error) {
	return q.repo.GetStock(ctx, warehouseID, skuID)
}

// ListWarehouseStocks 列出仓库的所有库存。
func (q *WarehouseQuery) ListWarehouseStocks(ctx context.Context, warehouseID uint64, offset, limit int) ([]*domain.WarehouseStock, int64, error) {
	return q.repo.ListStocks(ctx, warehouseID, offset, limit)
}

// GetTransferByID 获取调拨单详情。
func (q *WarehouseQuery) GetTransferByID(ctx context.Context, id uint64) (*domain.StockTransfer, error) {
	return q.repo.GetTransfer(ctx, id)
}

// ListTransfers 列出调拨单。
func (q *WarehouseQuery) ListTransfers(ctx context.Context, fromWH, toWH uint64, status *domain.StockTransferStatus, offset, limit int) ([]*domain.StockTransfer, int64, error) {
	return q.repo.ListTransfers(ctx, fromWH, toWH, status, offset, limit)
}

// GetOptimalWarehouse 根据地理优先级寻找到最优的仓库。
func (q *WarehouseQuery) GetOptimalWarehouse(ctx context.Context, skuID uint64, qty int32, lat, lon float64) (*domain.Warehouse, float64, int32, error) {
	warehouses, stocks, err := q.repo.ListWarehousesWithStock(ctx, skuID, qty)
	if err != nil {
		return nil, 0, 0, err
	}

	if len(warehouses) == 0 {
		return nil, 0, 0, errors.New("no warehouse available with sufficient stock")
	}

	var bestWH *domain.Warehouse
	var minDistance = -1.0
	var bestStock int32

	for i, wh := range warehouses {
		dist := calculateDistance(lat, lon, wh.Latitude, wh.Longitude)
		// 结合距离和优先级（权重）进行选择，此处简化为仅限距离
		if minDistance < 0 || dist < minDistance {
			minDistance = dist
			bestWH = wh
			bestStock = stocks[i]
		}
	}

	return bestWH, minDistance, bestStock, nil
}

// calculateDistance 使用 Haversine 公式计算两点间的距离（单位：公里）。
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0 // 地球半径，单位公里
	rad := math.Pi / 180.0
	dLat := (lat2 - lat1) * rad
	dLon := (lon2 - lon1) * rad
	phi1 := lat1 * rad
	phi2 := lat2 * rad

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(phi1)*math.Cos(phi2)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}
