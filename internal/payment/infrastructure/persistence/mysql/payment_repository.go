package mysql

import (
	"context"
	"errors"
	"time"

	pb "github.com/wyfcoding/ecommerce/goapi/payment/v1"
	"github.com/wyfcoding/ecommerce/internal/payment/domain"
	"github.com/wyfcoding/pkg/contextx"
	"github.com/wyfcoding/pkg/database/sharding"
	"github.com/wyfcoding/pkg/dtm"

	"gorm.io/gorm"
)

type paymentRepository struct {
	sharding *sharding.Manager
	tx       *gorm.DB
}

// NewPaymentRepository 创建并返回一个新的 paymentRepository 实例。
func NewPaymentRepository(sharding *sharding.Manager) domain.PaymentRepository {
	return &paymentRepository{sharding: sharding}
}

func (r *paymentRepository) getDB(userID uint64) *gorm.DB {
	if r.tx != nil {
		return r.tx
	}
	return r.sharding.GetDB(userID)
}

// Save 将支付实体保存到数据库。
func (r *paymentRepository) Save(ctx context.Context, entity *domain.Payment) error {
	if entity == nil {
		return nil
	}
	db := r.getDB(uint64(entity.UserID)).WithContext(ctx)
	return db.Transaction(func(tx *gorm.DB) error {
		return r.saveInTx(ctx, tx, entity, false)
	})
}

// Update 更新支付实体，通过版本号实现乐观锁。
func (r *paymentRepository) Update(ctx context.Context, entity *domain.Payment) error {
	if entity == nil {
		return nil
	}
	db := r.getDB(uint64(entity.UserID)).WithContext(ctx)
	return db.Transaction(func(tx *gorm.DB) error {
		return r.saveInTx(ctx, tx, entity, true)
	})
}

func (r *paymentRepository) saveInTx(ctx context.Context, tx *gorm.DB, entity *domain.Payment, optimistic bool) error {
	model := toPaymentModel(entity)
	if model == nil {
		return nil
	}

	gormTx := tx.WithContext(ctx)
	if optimistic {
		currentVer := entity.PersistenceVer
		entity.PersistenceVer++
		model.Version = entity.PersistenceVer

		res := gormTx.Model(&PaymentModel{}).
			Where("id = ? AND version = ?", model.ID, currentVer).
			Select("*").
			Updates(model)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("optimistic lock failed")
		}
	} else {
		if model.ID == 0 {
			if err := gormTx.Create(model).Error; err != nil {
				return err
			}
		} else {
			if err := gormTx.Save(model).Error; err != nil {
				return err
			}
		}
	}

	paymentID := uint64(model.ID)
	for i := range model.Splits {
		if model.Splits[i].PaymentID == 0 {
			model.Splits[i].PaymentID = paymentID
		}
		if err := gormTx.Save(&model.Splits[i]).Error; err != nil {
			return err
		}
	}
	for i := range model.Logs {
		if model.Logs[i].PaymentID == 0 {
			model.Logs[i].PaymentID = paymentID
		}
		if err := gormTx.Save(&model.Logs[i]).Error; err != nil {
			return err
		}
	}

	if entity != nil {
		if synced := toDomainPayment(model); synced != nil {
			*entity = *synced
		}
	}
	return nil
}

// FindByID 根据ID从数据库获取支付记录。
func (r *paymentRepository) FindByID(ctx context.Context, userID uint64, id uint64) (*domain.Payment, error) {
	db := r.getDB(userID)
	var entity PaymentModel
	if err := db.WithContext(ctx).Preload("Logs").Preload("Splits").First(&entity, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainPayment(&entity), nil
}

// FindByPaymentNo 根据支付单号从数据库获取支付记录。
func (r *paymentRepository) FindByPaymentNo(ctx context.Context, userID uint64, paymentNo string) (*domain.Payment, error) {
	db := r.getDB(userID)
	var entity PaymentModel
	if err := db.WithContext(ctx).Where("payment_no = ?", paymentNo).Preload("Logs").Preload("Splits").First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainPayment(&entity), nil
}

// FindByOrderID 根据订单ID从数据库获取支付记录。
func (r *paymentRepository) FindByOrderID(ctx context.Context, userID uint64, orderID uint64) (*domain.Payment, error) {
	db := r.getDB(userID)
	var entity PaymentModel
	if err := db.WithContext(ctx).Where("order_id = ?", orderID).Preload("Logs").Preload("Splits").First(&entity).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomainPayment(&entity), nil
}

// SaveLog 将支付日志实体保存到数据库。
func (r *paymentRepository) SaveLog(ctx context.Context, log *domain.PaymentLog) error {
	if log == nil {
		return nil
	}
	db := r.getDB(uint64(log.UserID))
	model := PaymentLogModel{
		Model: gorm.Model{
			ID:        log.ID,
			CreatedAt: log.CreatedAt,
			UpdatedAt: log.UpdatedAt,
		},
		PaymentID: log.PaymentID,
		UserID:    log.UserID,
		Action:    log.Action,
		OldStatus: log.OldStatus,
		NewStatus: log.NewStatus,
		Remark:    log.Remark,
	}
	return db.WithContext(ctx).Create(&model).Error
}

// FindLogsByPaymentID 根据支付ID从数据库获取所有支付日志。
func (r *paymentRepository) FindLogsByPaymentID(ctx context.Context, userID uint64, paymentID uint64) ([]*domain.PaymentLog, error) {
	db := r.getDB(userID)
	var logs []PaymentLogModel
	if err := db.WithContext(ctx).Where("payment_id = ?", paymentID).Find(&logs).Error; err != nil {
		return nil, err
	}
	result := make([]*domain.PaymentLog, 0, len(logs))
	for _, log := range logs {
		logCopy := log
		result = append(result, &domain.PaymentLog{
			ID:        logCopy.ID,
			CreatedAt: logCopy.CreatedAt,
			UpdatedAt: logCopy.UpdatedAt,
			PaymentID: logCopy.PaymentID,
			UserID:    logCopy.UserID,
			Action:    logCopy.Action,
			OldStatus: logCopy.OldStatus,
			NewStatus: logCopy.NewStatus,
			Remark:    logCopy.Remark,
		})
	}
	return result, nil
}

// FindSuccessPaymentsByDate 跨分片聚合指定日期的成功支付记录。
func (r *paymentRepository) FindSuccessPaymentsByDate(ctx context.Context, date time.Time) ([]*domain.Payment, error) {
	dbs := r.sharding.GetAllDBs()
	var allPayments []*domain.Payment
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
	end := start.Add(24 * time.Hour)
	for _, db := range dbs {
		var list []PaymentModel
		err := db.WithContext(ctx).
			Where("status = ? AND paid_at >= ? AND paid_at < ?", pb.PaymentStatus_SUCCESS, start, end).
			Find(&list).Error
		if err != nil {
			return nil, err
		}
		for i := range list {
			payment := toDomainPayment(&list[i])
			if payment != nil {
				allPayments = append(allPayments, payment)
			}
		}
	}
	return allPayments, nil
}

// SaveReconciliationRecord 保存对账结果。
func (r *paymentRepository) SaveReconciliationRecord(ctx context.Context, record *domain.ReconciliationRecord) error {
	if record == nil {
		return nil
	}
	// 对账记录通常存储在管理分片/全局片 (shard 0)
	db := r.sharding.GetDB(0)
	model := toReconciliationRecordModel(record)
	return db.WithContext(ctx).Save(model).Error
}

// GetUserIDByPaymentNo 跨分片查找用户ID。
func (r *paymentRepository) GetUserIDByPaymentNo(ctx context.Context, paymentNo string) (uint64, error) {
	dbs := r.sharding.GetAllDBs()
	for _, db := range dbs {
		var p struct{ UserID uint64 }
		err := db.WithContext(ctx).Table("payments").Select("user_id").Where("payment_no = ?", paymentNo).First(&p).Error
		if err == nil {
			return p.UserID, nil
		}
	}
	return 0, errors.New("payment record not found")
}

func (r *paymentRepository) Transaction(ctx context.Context, userID uint64, fn func(tx any) error) error {
	db := r.sharding.GetDB(userID)
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

func (r *paymentRepository) WithTx(tx any) domain.PaymentRepository {
	if db, ok := tx.(*gorm.DB); ok {
		return &paymentRepository{
			sharding: r.sharding,
			tx:       db,
		}
	}
	return r
}

// ExecWithBarrier 在分布式事务屏障下执行业务逻辑。
func (r *paymentRepository) ExecWithBarrier(ctx context.Context, barrier any, fn func(ctx context.Context) error) error {
	db := r.sharding.GetDB(0)
	return dtm.CallWithGorm(ctx, barrier, db, func(tx *gorm.DB) error {
		txCtx := contextx.WithTx(ctx, tx)
		return fn(txCtx)
	})
}
