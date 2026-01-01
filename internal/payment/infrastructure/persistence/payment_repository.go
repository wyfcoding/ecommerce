package persistence

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/payment/domain" // 导入支付领域的领域层接口和实体。
	"github.com/wyfcoding/pkg/databases/sharding"            // 导入分库分表管理器。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

type paymentRepository struct {
	sharding *sharding.Manager // 分库分表管理器实例。
	tx       *gorm.DB          // 事务 DB 实例 (可选)。
}

// NewPaymentRepository 创建并返回一个新的 paymentRepository 实例。
func NewPaymentRepository(sharding *sharding.Manager) domain.PaymentRepository {
	return &paymentRepository{sharding: sharding}
}

// Save 将支付实体保存到数据库。
// 如果实体已存在，则更新；如果不存在，则创建。
// 此方法根据 UserID 决定写入哪个分库。
func (r *paymentRepository) Save(ctx context.Context, entity *domain.Payment) error {
	// 优先使用事务中的 DB 实例。
	var db *gorm.DB
	if r.tx != nil {
		db = r.tx
	} else {
		db = r.sharding.GetDB(uint64(entity.UserID))
	}
	return db.WithContext(ctx).Create(entity).Error
}

// FindByID 根据ID从数据库获取支付记录。
// 注意：由于是分库分表设计，通过支付ID直接查询可能需要遍历所有分库，效率较低。
// 更优化的方案是：
// 1. 在支付ID中嵌入分片信息（如Snowflake ID）。
// 2. 要求调用方提供 userID 或 orderID 等分片键。
// 3. 维护一个ID到分库的映射表。
// 当前实现为了兼容接口，暂时默认从第一个分库（shard 0）查询。这在实际生产环境中是不可接受的。
func (r *paymentRepository) FindByID(ctx context.Context, id uint64) (*domain.Payment, error) {
	var db *gorm.DB
	if r.tx != nil {
		db = r.tx
	} else {
		db = r.sharding.GetDB(0) // 临时使用 shard 0。
	}
	var entity domain.Payment
	if err := db.WithContext(ctx).First(&entity, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &entity, nil
}

// Update 更新支付实体。
// 此方法根据 UserID 决定更新哪个分库。
func (r *paymentRepository) Update(ctx context.Context, entity *domain.Payment) error {
	var db *gorm.DB
	if r.tx != nil {
		db = r.tx
	} else {
		db = r.sharding.GetDB(uint64(entity.UserID))
	}
	return db.WithContext(ctx).Save(entity).Error
}

// Delete 根据ID从数据库删除支付记录。
// 存在与FindByID相同的分库问题。
func (r *paymentRepository) Delete(ctx context.Context, id uint64) error {
	// TODO: 优化此方法以支持分库删除。目前的实现是硬编码从 shard 0 删除，这在实际生产环境中可能导致数据找不到或删除错误。
	db := r.sharding.GetDB(0) // 临时使用 shard 0。
	return db.WithContext(ctx).Delete(&domain.Payment{}, id).Error
}

// FindByPaymentNo 根据支付单号从数据库获取支付记录。
// 存在与FindByID相同的分库问题。
func (r *paymentRepository) FindByPaymentNo(ctx context.Context, paymentNo string) (*domain.Payment, error) {
	var db *gorm.DB
	if r.tx != nil {
		db = r.tx
	} else {
		db = r.sharding.GetDB(0) // 临时使用 shard 0。
	}
	var entity domain.Payment
	if err := db.WithContext(ctx).Where("payment_no = ?", paymentNo).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &entity, nil
}

// FindByOrderID 根据订单ID从数据库获取支付记录。
// 存在与FindByID相同的分库问题。
func (r *paymentRepository) FindByOrderID(ctx context.Context, orderID uint64) (*domain.Payment, error) {
	var db *gorm.DB
	if r.tx != nil {
		db = r.tx
	} else {
		db = r.sharding.GetDB(0) // 临时使用 shard 0。
	}
	var entity domain.Payment
	if err := db.WithContext(ctx).Where("order_id = ?", orderID).First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &entity, nil
}

// ListByUserID 从数据库列出指定用户ID的所有支付记录，支持分页。
// 此方法利用了 userID 进行分库查询，因此效率较高。
func (r *paymentRepository) ListByUserID(ctx context.Context, userID uint64, offset, limit int) ([]*domain.Payment, int64, error) {
	var entities []*domain.Payment
	var total int64

	// 根据UserID获取对应的分库DB实例。
	db := r.sharding.GetDB(userID)

	// 统计总记录数。
	if err := db.WithContext(ctx).Model(&domain.Payment{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页。
	if err := db.WithContext(ctx).Where("user_id = ?", userID).Offset(offset).Limit(limit).Find(&entities).Error; err != nil {
		return nil, 0, err
	}

	return entities, total, nil
}

// SaveLog 将支付日志实体保存到数据库。
// 存在与FindByID相同的问题：PaymentLog通常不直接包含UserID，需要额外的逻辑来确定分库。
func (r *paymentRepository) SaveLog(ctx context.Context, log *domain.PaymentLog) error {
	// TODO: 优化此方法以支持分库保存。PaymentLog 通常没有直接的 UserID 字段来确定分库。
	// 理想的方案是：
	// 1. 将 UserID 反范式化到 PaymentLog 中。
	// 2. 通过 PaymentLog 的 PaymentID 查找对应的 Payment 实体，从而获取 UserID 来确定分库。
	// 目前临时硬编码保存到 shard 0。这在实际生产环境中是不可接受的。
	db := r.sharding.GetDB(0) // 临时使用 shard 0。
	return db.WithContext(ctx).Create(log).Error
}

// FindLogsByPaymentID 根据支付ID从数据库获取所有支付日志。
// 存在与FindByID相同的问题。
func (r *paymentRepository) FindLogsByPaymentID(ctx context.Context, paymentID uint64) ([]*domain.PaymentLog, error) {
	var db *gorm.DB
	if r.tx != nil {
		db = r.tx
	} else {
		db = r.sharding.GetDB(0) // 临时使用 shard 0。
	}
	var logs []*domain.PaymentLog
	if err := db.WithContext(ctx).Where("payment_id = ?", paymentID).Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *paymentRepository) Transaction(ctx context.Context, fn func(tx any) error) error {
	// 简单起见，这里假设事务在 shard 0 上执行（或由分片管理器支持跨库事务）。
	// 在分库分表环境下，事务通常需要由应用层决定分片键。
	db := r.sharding.GetDB(0)
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

func (r *paymentRepository) WithTx(tx any) domain.PaymentRepository {
	return &paymentRepository{
		sharding: r.sharding,
		tx:       tx.(*gorm.DB),
	}
}
