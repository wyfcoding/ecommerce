package application

import (
	"context"
	"log"
	"strconv"

	"github.com/wyfcoding/ecommerce/internal/inventory/domain"
)

type InventoryService struct {
	repo domain.InventoryRepository
	// 这里可以注入 MQ Producer
}

func NewInventoryService(repo domain.InventoryRepository) *InventoryService {
	return &InventoryService{
		repo: repo,
	}
}

// ReserveStock 核心高并发接口
func (s *InventoryService) ReserveStock(ctx context.Context, req domain.DeductRequest) error {
	skuID, _ := strconv.ParseUint(req.SKUID, 10, 64)
	// 1. Redis 预占 (高性能，抗并发)
	err := s.repo.Reserve(ctx, skuID, int32(req.Quantity))
	if err != nil {
		return err
	}

	// 2. 异步回写 (Write-Behind)
	// 实际生产中应发送 Kafka 消息，这里简单起见启动 Goroutine 直接写 DB (存在丢失风险，仅演示)
	// 或者写入 Redis List 供 Worker 消费
	go func() {
		// 模拟消费队列
		// 注意：这里需要一个可以同步到 DB 的方法
		log.Printf("INFO: Stock reserved for SKU: %d, quantity: %d\n", skuID, req.Quantity)
	}()

	return nil
}
