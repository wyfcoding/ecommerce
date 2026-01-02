package persistence

import (
	"context"
	"errors"
	"time"

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
func (r *paymentRepository) FindByID(ctx context.Context, userID uint64, id uint64) (*domain.Payment, error) {
	db := r.getDB(userID)
	var entity domain.Payment
	if err := db.WithContext(ctx).Preload("Logs").First(&entity, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &entity, nil
}

// getDB 内部辅助方法，自动切换事务与普通连接
func (r *paymentRepository) getDB(userID uint64) *gorm.DB {
	if r.tx != nil {
		return r.tx
	}
	return r.sharding.GetDB(userID)
}

// Update 更新支付实体。
func (r *paymentRepository) Update(ctx context.Context, entity *domain.Payment) error {
	db := r.getDB(uint64(entity.UserID))
	return db.WithContext(ctx).Save(entity).Error
}

// Delete 根据ID从数据库删除支付记录。
func (r *paymentRepository) Delete(ctx context.Context, userID uint64, id uint64) error {
	db := r.getDB(userID)
	return db.WithContext(ctx).Delete(&domain.Payment{}, id).Error
}

// FindByPaymentNo 根据支付单号从数据库获取支付记录。
func (r *paymentRepository) FindByPaymentNo(ctx context.Context, userID uint64, paymentNo string) (*domain.Payment, error) {
	db := r.getDB(userID)
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
func (r *paymentRepository) FindByOrderID(ctx context.Context, userID uint64, orderID uint64) (*domain.Payment, error) {
	db := r.getDB(userID)
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
func (r *paymentRepository) SaveLog(ctx context.Context, log *domain.PaymentLog) error {
	// 真实实现：根据 UserID 实现分片路由。
	db := r.getDB(uint64(log.UserID))
	return db.WithContext(ctx).Create(log).Error
}

// FindLogsByPaymentID 根据支付ID从数据库获取所有支付日志。
func (r *paymentRepository) FindLogsByPaymentID(ctx context.Context, userID uint64, paymentID uint64) ([]*domain.PaymentLog, error) {
	db := r.getDB(userID)
	var logs []*domain.PaymentLog
	if err := db.WithContext(ctx).Where("payment_id = ?", paymentID).Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// FindSuccessPaymentsByDate 跨分片聚合指定日期的成功支付记录。
func (r *paymentRepository) FindSuccessPaymentsByDate(ctx context.Context, date time.Time) ([]*domain.Payment, error) {
	dbs := r.sharding.GetAllDBs()
	var allPayments []*domain.Payment

	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
	end := start.Add(24 * time.Hour)

	for _, db := range dbs {
		var list []*domain.Payment
		// 查找该分片内 PaidAt 在目标日期的记录
		err := db.WithContext(ctx).
			Where("status = ? AND paid_at >= ? AND paid_at < ?", domain.PaymentSuccess, start, end).
			Find(&list).Error
		if err != nil {
			return nil, err
		}
		allPayments = append(allPayments, list...)
	}

	return allPayments, nil
}

// SaveReconciliationRecord 保存对账结果到配置分片。
func (r *paymentRepository) SaveReconciliationRecord(ctx context.Context, record *domain.ReconciliationRecord) error {
	db := r.sharding.GetDB(0)
	return db.WithContext(ctx).Save(record).Error
}

// GetUserIDByPaymentNo 跨分片查找支付单归属的用户ID。
func (r *paymentRepository) GetUserIDByPaymentNo(ctx context.Context, paymentNo string) (uint64, error) {
	dbs := r.sharding.GetAllDBs()
	
	// 真实化执行：扫描所有分片寻找 PaymentNo
	// 顶级优化：如果 PaymentNo 包含分片信息，则无需全表扫描
	for _, db := range dbs {
		var p struct {
			UserID uint64
		}
		err := db.WithContext(ctx).Table("payments").
			Select("user_id").
			Where("payment_no = ?", paymentNo).
			First(&p).Error
		
		if err == nil {
			return p.UserID, nil
		}
	}

	return 0, errors.New("payment record not found in any shard")
}

func (r *paymentRepository) Transaction(ctx context.Context, userID uint64, fn func(tx any) error) error {
	db := r.sharding.GetDB(userID)
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
