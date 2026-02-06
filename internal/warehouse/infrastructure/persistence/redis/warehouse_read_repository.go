package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/warehouse/domain"
)

const (
	warehouseDetailPrefix   = "warehouse:detail:"
	warehouseStockPrefix    = "warehouse:stock:"
	warehouseTransferPrefix = "warehouse:transfer:"
)

// warehouseReadRepository 基于 Redis 的仓库读模型仓储。
type warehouseReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewWarehouseReadRepository 创建仓库读模型仓储。
func NewWarehouseReadRepository(client redis.UniversalClient, ttl time.Duration) domain.WarehouseReadRepository {
	return &warehouseReadRepository{client: client, ttl: ttl}
}

func (r *warehouseReadRepository) SaveWarehouse(ctx context.Context, warehouse *domain.Warehouse) error {
	if warehouse == nil {
		return nil
	}
	data, err := json.Marshal(warehouse)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.warehouseKey(uint64(warehouse.ID)), data, r.ttl).Err()
}

func (r *warehouseReadRepository) GetWarehouse(ctx context.Context, id uint64) (*domain.Warehouse, error) {
	data, err := r.client.Get(ctx, r.warehouseKey(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var warehouse domain.Warehouse
	if err := json.Unmarshal(data, &warehouse); err != nil {
		return nil, err
	}
	return &warehouse, nil
}

func (r *warehouseReadRepository) DeleteWarehouse(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.warehouseKey(id)).Err()
}

func (r *warehouseReadRepository) SaveStock(ctx context.Context, stock *domain.WarehouseStock) error {
	if stock == nil {
		return nil
	}
	data, err := json.Marshal(stock)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.stockKey(stock.WarehouseID, stock.SkuID), data, r.ttl).Err()
}

func (r *warehouseReadRepository) GetStock(ctx context.Context, warehouseID, skuID uint64) (*domain.WarehouseStock, error) {
	data, err := r.client.Get(ctx, r.stockKey(warehouseID, skuID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var stock domain.WarehouseStock
	if err := json.Unmarshal(data, &stock); err != nil {
		return nil, err
	}
	return &stock, nil
}

func (r *warehouseReadRepository) DeleteStock(ctx context.Context, warehouseID, skuID uint64) error {
	return r.client.Del(ctx, r.stockKey(warehouseID, skuID)).Err()
}

func (r *warehouseReadRepository) SaveTransfer(ctx context.Context, transfer *domain.StockTransfer) error {
	if transfer == nil {
		return nil
	}
	data, err := json.Marshal(transfer)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.transferKey(uint64(transfer.ID)), data, r.ttl).Err()
}

func (r *warehouseReadRepository) GetTransfer(ctx context.Context, id uint64) (*domain.StockTransfer, error) {
	data, err := r.client.Get(ctx, r.transferKey(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var transfer domain.StockTransfer
	if err := json.Unmarshal(data, &transfer); err != nil {
		return nil, err
	}
	return &transfer, nil
}

func (r *warehouseReadRepository) DeleteTransfer(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.transferKey(id)).Err()
}

func (r *warehouseReadRepository) warehouseKey(id uint64) string {
	return fmt.Sprintf("%s%d", warehouseDetailPrefix, id)
}

func (r *warehouseReadRepository) stockKey(warehouseID, skuID uint64) string {
	return fmt.Sprintf("%s%d:%d", warehouseStockPrefix, warehouseID, skuID)
}

func (r *warehouseReadRepository) transferKey(id uint64) string {
	return fmt.Sprintf("%s%d", warehouseTransferPrefix, id)
}
