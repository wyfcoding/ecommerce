package application

import (
	"context"
	"fmt"
	"log/slog"
	"math"

	"github.com/wyfcoding/ecommerce/internal/warehouse/domain"
	"github.com/wyfcoding/pkg/utils/geo"
)

// WarehouseQueryService 处理所有仓库相关的查询操作 (Queries)。
type WarehouseQueryService struct {
	repo       domain.WarehouseRepository
	readRepo   domain.WarehouseReadRepository
	searchRepo domain.WarehouseSearchRepository
	logger     *slog.Logger
}

// NewWarehouseQueryService 构造函数。
func NewWarehouseQueryService(repo domain.WarehouseRepository, readRepo domain.WarehouseReadRepository, searchRepo domain.WarehouseSearchRepository, logger *slog.Logger) *WarehouseQueryService {
	return &WarehouseQueryService{
		repo:       repo,
		readRepo:   readRepo,
		searchRepo: searchRepo,
		logger:     logger,
	}
}

// GetWarehouse 获取详情。
func (s *WarehouseQueryService) GetWarehouse(ctx context.Context, id uint64) (*domain.Warehouse, error) {
	if s.readRepo != nil {
		if w, err := s.readRepo.GetWarehouse(ctx, id); err == nil && w != nil {
			return w, nil
		}
	}
	return s.repo.FindByID(ctx, id)
}

// ListWarehouses 列表。
func (s *WarehouseQueryService) ListWarehouses(ctx context.Context, page, pageSize int) ([]*domain.Warehouse, int64, error) {
	offset := (page - 1) * pageSize
	if s.searchRepo != nil {
		list, total, err := s.searchRepo.SearchWarehouses(ctx, nil, nil, nil, nil, nil, offset, pageSize, nil, nil, "priority")
		if err == nil {
			return list, total, nil
		}
		if s.logger != nil {
			s.logger.WarnContext(ctx, "warehouse search fallback to mysql", "error", err)
		}
	}
	return s.repo.List(ctx, offset, pageSize)
}

// GetStock 查看当前库存。
func (s *WarehouseQueryService) GetStock(ctx context.Context, warehouseID, skuID uint64) (*domain.WarehouseStock, error) {
	if s.readRepo != nil {
		if stock, err := s.readRepo.GetStock(ctx, warehouseID, skuID); err == nil && stock != nil {
			return stock, nil
		}
	}
	return s.repo.GetStock(ctx, warehouseID, skuID)
}

// GetOptimalWarehouse 调度模型：根据地理位置寻找最近且有货的仓库。
func (s *WarehouseQueryService) GetOptimalWarehouse(ctx context.Context, skuID uint64, qty int32, lat, lon float64) (*domain.Warehouse, float64, int32, error) {
	// 1. 获取所有可用仓库。实际生产中应当先通过省市区或经纬度范围初步过滤。
	var (
		warehouses []*domain.Warehouse
		err        error
	)
	if s.searchRepo != nil {
		warehouses, _, err = s.searchRepo.SearchWarehouses(ctx, nil, nil, nil, nil, nil, 0, 200, nil, nil, "priority")
		if err != nil && s.logger != nil {
			s.logger.WarnContext(ctx, "warehouse search fallback to mysql", "error", err)
		}
	}
	if len(warehouses) == 0 {
		warehouses, _, err = s.repo.List(ctx, 0, 200)
		if err != nil {
			return nil, 0, 0, err
		}
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
	if s.readRepo != nil {
		if transfer, err := s.readRepo.GetTransfer(ctx, id); err == nil && transfer != nil {
			return transfer, nil
		}
	}
	return s.repo.FindTransferByID(ctx, id)
}

// ListTransfers 列表分页。
func (s *WarehouseQueryService) ListTransfers(ctx context.Context, fromID, toID uint64, status *string, page, pageSize int) ([]*domain.StockTransfer, int64, error) {
	offset := (page - 1) * pageSize
	if s.searchRepo != nil {
		var fromPtr *uint64
		var toPtr *uint64
		if fromID > 0 {
			fromPtr = &fromID
		}
		if toID > 0 {
			toPtr = &toID
		}
		list, total, err := s.searchRepo.SearchTransfers(ctx, fromPtr, toPtr, status, offset, pageSize, nil, nil, "created_at")
		if err == nil {
			return list, total, nil
		}
		if s.logger != nil {
			s.logger.WarnContext(ctx, "transfer search fallback to mysql", "error", err)
		}
	}
	return s.repo.ListTransfers(ctx, fromID, toID, status, offset, pageSize)
}
