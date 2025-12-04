package persistence

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/order/domain/entity"     // 导入订单模块的领域实体定义。
	"github.com/wyfcoding/ecommerce/internal/order/domain/repository" // 导入订单模块的领域仓储接口。
	"github.com/wyfcoding/ecommerce/pkg/databases/sharding"           // 导入分库分表管理器。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

type orderRepository struct {
	sharding *sharding.Manager // 分库分表管理器实例。
}

// NewOrderRepository 创建并返回一个新的 orderRepository 实例。
func NewOrderRepository(sharding *sharding.Manager) repository.OrderRepository {
	return &orderRepository{sharding: sharding}
}

// Save 将订单实体保存到数据库。
// 如果订单已存在，则更新；如果不存在，则创建。
// 此方法在一个事务中保存订单主实体、订单项和订单日志。
func (r *orderRepository) Save(ctx context.Context, order *entity.Order) error {
	// 根据UserID获取对应的分库DB实例。
	db := r.sharding.GetDB(order.UserID)
	// 使用事务确保订单主实体、订单项和订单日志的保存操作的原子性。
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 保存或更新订单主实体。
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		// 保存订单项。
		for _, item := range order.Items {
			if item.ID == 0 { // 仅保存新的订单项。
				item.OrderID = uint64(order.ID) // 关联订单ID。
			}
			if err := tx.Save(item).Error; err != nil {
				return err
			}
		}
		// 保存订单日志。
		for _, log := range order.Logs {
			if log.ID == 0 { // 仅保存新的订单日志。
				log.OrderID = uint64(order.ID) // 关联订单ID。
			}
			if err := tx.Save(log).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetByID 根据ID从数据库获取订单记录。
// 注意：由于是分库分表设计，通过订单ID直接查询可能需要遍历所有分库，效率较低。
// 更优化的方案是：
// 1. 在订单ID中嵌入分片信息（如Snowflake ID）。
// 2. 提供 userID 参数辅助查询，或者通过上下文获取当前用户ID。
// 3. 维护一个ID到分库的映射表。
// 当前实现为了兼容接口，暂时默认从主库（shard 0）查询。
func (r *orderRepository) GetByID(ctx context.Context, id uint64) (*entity.Order, error) {
	// TODO: 优化此方法以支持分库查询。目前的实现是硬编码从 shard 0 获取，这在实际生产环境中可能导致数据找不到或跨分库查询效率低下。
	// 理想情况下，`GetByID` 应该能够根据 `id` 确定所属的分库，或者 `order` 实体本身包含足够信息指导查询。
	db := r.sharding.GetDB(0) // 临时使用 shard 0。
	var order entity.Order
	// Preload "Items" 和 "Logs" 确保在获取订单时，同时加载所有关联的订单项和操作日志。
	if err := db.WithContext(ctx).Preload("Items").Preload("Logs").First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &order, nil
}

// GetByOrderNo 根据订单编号从数据库获取订单记录。
// 注意：与 GetByID 类似，也存在跨分库查询问题。
func (r *orderRepository) GetByOrderNo(ctx context.Context, orderNo string) (*entity.Order, error) {
	// TODO: 优化此方法以支持分库查询。目前的实现是硬编码从 shard 0 获取。
	db := r.sharding.GetDB(0) // 临时使用 shard 0。
	var order entity.Order
	// Preload "Items" 和 "Logs" 确保在获取订单时，同时加载所有关联的订单项和操作日志。
	if err := db.WithContext(ctx).Preload("Items").Preload("Logs").Where("order_no = ?", orderNo).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &order, nil
}

// List 从数据库列出指定用户ID的所有订单记录，支持通过状态过滤和分页。
// 此方法利用了 userID 进行分库查询，因此效率较高。
func (r *orderRepository) List(ctx context.Context, userID uint64, status *entity.OrderStatus, offset, limit int) ([]*entity.Order, int64, error) {
	var list []*entity.Order
	var total int64

	// 根据UserID获取对应的分库DB实例。
	db := r.sharding.GetDB(userID).WithContext(ctx).Model(&entity.Order{})

	if userID > 0 { // 如果提供了用户ID，则按用户ID过滤。
		db = db.Where("user_id = ?", userID)
	}
	if status != nil { // 如果提供了订单状态，则按状态过滤。
		db = db.Where("status = ?", *status)
	}

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Preload "Items" 确保在获取订单列表时，同时加载所有关联的订单项。
	// 应用分页和排序。
	if err := db.Preload("Items").Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}
