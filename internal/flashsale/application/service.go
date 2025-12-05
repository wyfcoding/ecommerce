package application

import (
	"context"
	"encoding/json" // 导入JSON编码/解码库。
	"errors"        // 导入标准错误处理库。
	"fmt"           // 导入格式化库。
	"time"          // 导入时间库。

	"github.com/wyfcoding/ecommerce/internal/flashsale/domain/entity"     // 导入秒杀领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/flashsale/domain/repository" // 导入秒杀领域的仓储和缓存接口。
	"github.com/wyfcoding/ecommerce/pkg/idgen"                            // 导入ID生成器接口。
	"github.com/wyfcoding/ecommerce/pkg/messagequeue/kafka"               // 导入Kafka消息生产者。

	"log/slog" // 导入结构化日志库。
)

// FlashSaleService 结构体定义了秒杀活动相关的应用服务。
// 它协调领域层和基础设施层，处理秒杀活动的创建、库存管理、订单放置和取消等高并发业务逻辑。
type FlashSaleService struct {
	repo     repository.FlashSaleRepository // 依赖FlashSaleRepository接口，用于秒杀活动的持久化存储。
	cache    repository.FlashSaleCache      // 依赖FlashSaleCache接口，用于高并发库存扣减和限购检查。
	producer *kafka.Producer                // 依赖Kafka生产者，用于发布订单事件。
	idGen    idgen.Generator                // 依赖ID生成器，用于生成订单ID。
	logger   *slog.Logger                   // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewFlashSaleService 创建并返回一个新的 FlashSaleService 实例。
func NewFlashSaleService(
	repo repository.FlashSaleRepository,
	cache repository.FlashSaleCache,
	producer *kafka.Producer,
	idGen idgen.Generator,
	logger *slog.Logger,
) *FlashSaleService {
	return &FlashSaleService{
		repo:     repo,
		cache:    cache,
		producer: producer,
		idGen:    idGen,
		logger:   logger,
	}
}

// CreateFlashsale 创建一个新的秒杀活动。
// ctx: 上下文。
// name: 秒杀活动名称。
// productID, skuID: 关联的商品ID和SKU ID。
// originalPrice, flashPrice: 商品原价和秒杀价格。
// totalStock, limitPerUser: 秒杀总库存和每用户限购数量。
// startTime, endTime: 秒杀活动的开始和结束时间。
// 返回created successfully的Flashsale实体和可能发生的错误。
func (s *FlashSaleService) CreateFlashsale(ctx context.Context, name string, productID, skuID uint64, originalPrice, flashPrice int64, totalStock, limitPerUser int32, startTime, endTime time.Time) (*entity.Flashsale, error) {
	flashsale := entity.NewFlashsale(name, productID, skuID, originalPrice, flashPrice, totalStock, limitPerUser, startTime, endTime) // 创建Flashsale实体。
	// 通过仓储接口保存秒杀活动。
	if err := s.repo.SaveFlashsale(ctx, flashsale); err != nil {
		s.logger.ErrorContext(ctx, "failed to save flashsale", "name", name, "error", err)
		return nil, err
	}
	s.logger.InfoContext(ctx, "flashsale created successfully", "flashsale_id", flashsale.ID, "name", name)

	// 预热缓存：将秒杀库存加载到缓存中，以便在高并发下进行快速扣减。
	if err := s.cache.SetStock(ctx, uint64(flashsale.ID), totalStock); err != nil {
		s.logger.ErrorContext(ctx, "failed to pre-warm cache", "flashsale_id", flashsale.ID, "error", err)
		// TODO: 预热缓存失败的处理策略。是应该继续（让后续请求慢慢加载）还是直接失败？
		// 当前实现选择失败。
		return nil, fmt.Errorf("failed to pre-warm cache: %w", err)
	}

	return flashsale, nil
}

// GetFlashsale 获取指定ID的秒杀活动详情。
// ctx: 上下文。
// id: 秒杀活动的ID。
// 返回Flashsale实体和可能发生的错误。
func (s *FlashSaleService) GetFlashsale(ctx context.Context, id uint64) (*entity.Flashsale, error) {
	return s.repo.GetFlashsale(ctx, id)
}

// ListFlashsales 获取秒杀活动列表。
// ctx: 上下文。
// status: 筛选秒杀活动的状态。
// page, pageSize: 分页参数。
// 返回Flashsale列表、总数和可能发生的错误。
func (s *FlashSaleService) ListFlashsales(ctx context.Context, status *entity.FlashsaleStatus, page, pageSize int) ([]*entity.Flashsale, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListFlashsales(ctx, status, offset, pageSize)
}

// PlaceOrder 下达一个秒杀订单（高并发版本）。
// 该方法在高并发场景下，通过Redis缓存进行库存扣减和限购检查，并将订单事件发布到Kafka。
// ctx: 上下文。
// userID: 下单用户ID。
// flashsaleID: 关联的秒杀活动ID。
// quantity: 购买数量。
// 返回创建的FlashsaleOrder实体和可能发生的错误。
func (s *FlashSaleService) PlaceOrder(ctx context.Context, userID, flashsaleID uint64, quantity int32) (*entity.FlashsaleOrder, error) {
	// 1. 获取秒杀活动详情（可以从缓存中获取以提高性能）。
	flashsale, err := s.repo.GetFlashsale(ctx, flashsaleID)
	if err != nil {
		return nil, err
	}

	// 2. 验证秒杀活动是否在有效时间段内。
	now := time.Now()
	if now.Before(flashsale.StartTime) || now.After(flashsale.EndTime) {
		return nil, errors.New("flashsale is not active")
	}

	// 3. 在Redis缓存中进行库存扣减和每用户限购检查（原子操作）。
	// 这是高并发秒杀系统的核心逻辑，通过Redis的原子操作保证库存和限购的准确性。
	success, err := s.cache.DeductStock(ctx, flashsaleID, userID, quantity, flashsale.LimitPerUser)
	if err != nil {
		s.logger.ErrorContext(ctx, "failed to deduct stock in cache", "flashsale_id", flashsaleID, "user_id", userID, "error", err)
		return nil, err
	}
	if !success {
		return nil, entity.ErrFlashsaleSoldOut // 库存不足或达到用户限购。
	}

	// 4. 生成订单ID。
	orderID := s.idGen.Generate()

	// 5. 创建订单对象（状态为Pending）。
	// 注意：订单创建时，商品名称、SKU名称等信息未在参数中提供，通常需要从商品服务获取。
	order := entity.NewFlashsaleOrder(flashsaleID, userID, flashsale.ProductID, flashsale.SkuID, quantity, flashsale.FlashPrice)
	order.ID = uint(orderID) // 使用生成的订单ID。
	order.Status = entity.FlashsaleOrderStatusPending

	// 6. 发布订单事件到Kafka消息队列。
	// 将订单信息封装为JSON事件。
	event := map[string]interface{}{
		"order_id":     orderID,
		"flashsale_id": flashsaleID,
		"user_id":      userID,
		"product_id":   flashsale.ProductID,
		"sku_id":       flashsale.SkuID,
		"quantity":     quantity,
		"price":        flashsale.FlashPrice,
		"created_at":   order.CreatedAt,
	}
	payload, _ := json.Marshal(event) // 将事件转换为JSON字节。

	// 使用订单ID作为Kafka消息的键，以便于分区和顺序处理（如果需要）。
	if err := s.producer.Publish(ctx, []byte(fmt.Sprintf("%d", orderID)), payload); err != nil {
		s.logger.ErrorContext(ctx, "failed to publish order event", "order_id", orderID, "error", err)
		// 如果发布消息失败，需要回滚Redis中已扣减的库存。
		_ = s.cache.RevertStock(ctx, flashsaleID, userID, quantity)
		return nil, fmt.Errorf("failed to publish order: %w", err)
	}

	return order, nil
}

// CancelOrder 取消一个秒杀订单。
// ctx: 上下文。
// orderID: 待取消的订单ID。
// 返回可能发生的错误。
func (s *FlashSaleService) CancelOrder(ctx context.Context, orderID uint64) error {
	// 获取订单详情。
	order, err := s.repo.GetOrder(ctx, orderID)
	if err != nil {
		return err
	}

	// 如果订单不是Pending状态，则可能已被处理，无需取消。
	if order.Status != entity.FlashsaleOrderStatusPending {
		return nil
	}

	// 调用实体方法取消订单。
	order.Cancel()
	// 保存更新后的订单状态。
	if err := s.repo.SaveOrder(ctx, order); err != nil {
		return err
	}

	// 恢复Redis缓存中的库存。
	return s.cache.RevertStock(ctx, order.FlashsaleID, order.UserID, order.Quantity)
}
