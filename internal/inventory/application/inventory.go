package application

import (
	"context"

	orderv1 "github.com/wyfcoding/ecommerce/goapi/order/v1"
	"github.com/wyfcoding/ecommerce/internal/inventory/domain"
	algorithm "github.com/wyfcoding/pkg/algorithm/optimization"
)

// Inventory 门面服务，整合 CommandService 和 Query。
type Inventory struct {
	command *InventoryCommandService
	query   *InventoryQuery
}

// NewInventory 构造函数。
func NewInventory(command *InventoryCommandService, query *InventoryQuery) *Inventory {
	return &Inventory{
		command: command,
		query:   query,
	}
}

func (s *Inventory) SetRemoteOrderClient(cli orderv1.OrderServiceClient) {
	s.command.SetRemoteOrderClient(cli)
}

// --- Commands (Writes) ---

func (s *Inventory) CreateInventory(ctx context.Context, skuID, productID, warehouseID uint64, totalStock, warningThreshold int32) (*domain.Inventory, error) {
	return s.command.CreateInventory(ctx, skuID, productID, warehouseID, totalStock, warningThreshold)
}

func (s *Inventory) DeleteInventory(ctx context.Context, skuID uint64) error {
	return s.command.DeleteInventory(ctx, skuID)
}

func (s *Inventory) AddStock(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	return s.command.AddStock(ctx, skuID, quantity, reason)
}

func (s *Inventory) DeductStock(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	return s.command.DeductStock(ctx, skuID, quantity, reason)
}

func (s *Inventory) LockStock(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	return s.command.LockStock(ctx, skuID, quantity, reason)
}

func (s *Inventory) UnlockStock(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	return s.command.UnlockStock(ctx, skuID, quantity, reason)
}

func (s *Inventory) ConfirmDeduction(ctx context.Context, skuID uint64, quantity int32, reason string) error {
	return s.command.ConfirmDeduction(ctx, skuID, quantity, reason)
}

func (s *Inventory) AllocateStock(ctx context.Context, lat, lon float64, items []algorithm.OrderItem) ([]algorithm.AllocationResult, error) {
	return s.command.AllocateStock(ctx, lat, lon, items)
}

// --- Query (Reads) ---

func (s *Inventory) GetInventory(ctx context.Context, skuID uint64) (*domain.Inventory, error) {
	return s.query.GetInventory(ctx, skuID)
}

func (s *Inventory) ListInventories(ctx context.Context, page, pageSize int) ([]*domain.Inventory, int64, error) {
	return s.query.ListInventories(ctx, page, pageSize)
}

func (s *Inventory) GetInventoryLogs(ctx context.Context, skuID uint64, inventoryID uint64, page, pageSize int) ([]*domain.InventoryLog, int64, error) {
	return s.query.GetInventoryLogs(ctx, skuID, inventoryID, page, pageSize)
}
