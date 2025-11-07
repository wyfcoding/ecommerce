package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"ecommerce/internal/scheduler/model"
)

// SchedulerRepo 定时任务仓储接口
type SchedulerRepo interface {
	// 任务配置
	CreateTask(ctx context.Context, task *model.ScheduledTask) error
	UpdateTask(ctx context.Context, task *model.ScheduledTask) error
	GetTaskByID(ctx context.Context, id uint64) (*model.ScheduledTask, error)
	GetTaskByName(ctx context.Context, name string) (*model.ScheduledTask, error)
	ListTasks(ctx context.Context, status string, pageSize, pageNum int32) ([]*model.ScheduledTask, int64, error)
	ListActiveTasks(ctx context.Context) ([]*model.ScheduledTask, error)
	
	// 任务执行记录
	CreateTaskLog(ctx context.Context, log *model.TaskLog) error
	UpdateTaskLog(ctx context.Context, log *model.TaskLog) error
	GetTaskLog(ctx context.Context, id uint64) (*model.TaskLog, error)
	ListTaskLogs(ctx context.Context, taskID uint64, pageSize, pageNum int32) ([]*model.TaskLog, int64, error)
	
	// 任务锁
	AcquireLock(ctx context.Context, taskName string, ttl time.Duration) (bool, error)
	ReleaseLock(ctx context.Context, taskName string) error
	
	// 订单相关查询
	GetUnpaidOrders(ctx context.Context, expireMinutes int) ([]uint64, error)
	GetUnconfirmedOrders(ctx context.Context, autoConfirmDays int) ([]uint64, error)
	
	// 优惠券相关查询
	GetExpiredCoupons(ctx context.Context) ([]uint64, error)
	GetExpiringCoupons(ctx context.Context, days int) ([]uint64, error)
}

type schedulerRepo struct {
	db *gorm.DB
}

// NewSchedulerRepo 创建定时任务仓储实例
func NewSchedulerRepo(db *gorm.DB) SchedulerRepo {
	return &schedulerRepo{db: db}
}

// CreateTask 创建任务配置
func (r *schedulerRepo) CreateTask(ctx context.Context, task *model.ScheduledTask) error {
	return r.db.WithContext(ctx).Create(task).Error
}

// UpdateTask 更新任务配置
func (r *schedulerRepo) UpdateTask(ctx context.Context, task *model.ScheduledTask) error {
	return r.db.WithContext(ctx).Save(task).Error
}

// GetTaskByID 根据ID获取任务
func (r *schedulerRepo) GetTaskByID(ctx context.Context, id uint64) (*model.ScheduledTask, error) {
	var task model.ScheduledTask
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetTaskByName 根据名称获取任务
func (r *schedulerRepo) GetTaskByName(ctx context.Context, name string) (*model.ScheduledTask, error) {
	var task model.ScheduledTask
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// ListTasks 获取任务列表
func (r *schedulerRepo) ListTasks(ctx context.Context, status string, pageSize, pageNum int32) ([]*model.ScheduledTask, int64, error) {
	var tasks []*model.ScheduledTask
	var total int64

	query := r.db.WithContext(ctx).Model(&model.ScheduledTask{})
	
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	err := query.Offset(int(offset)).Limit(int(pageSize)).Order("created_at DESC").Find(&tasks).Error
	if err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// ListActiveTasks 获取所有激活的任务
func (r *schedulerRepo) ListActiveTasks(ctx context.Context) ([]*model.ScheduledTask, error) {
	var tasks []*model.ScheduledTask
	err := r.db.WithContext(ctx).
		Where("status = ?", model.TaskStatusActive).
		Order("created_at ASC").
		Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// CreateTaskLog 创建任务执行记录
func (r *schedulerRepo) CreateTaskLog(ctx context.Context, log *model.TaskLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// UpdateTaskLog 更新任务执行记录
func (r *schedulerRepo) UpdateTaskLog(ctx context.Context, log *model.TaskLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}

// GetTaskLog 获取任务执行记录
func (r *schedulerRepo) GetTaskLog(ctx context.Context, id uint64) (*model.TaskLog, error) {
	var log model.TaskLog
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// ListTaskLogs 获取任务执行记录列表
func (r *schedulerRepo) ListTaskLogs(ctx context.Context, taskID uint64, pageSize, pageNum int32) ([]*model.TaskLog, int64, error) {
	var logs []*model.TaskLog
	var total int64

	query := r.db.WithContext(ctx).Model(&model.TaskLog{})
	
	if taskID > 0 {
		query = query.Where("task_id = ?", taskID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	err := query.Offset(int(offset)).Limit(int(pageSize)).Order("created_at DESC").Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// AcquireLock 获取任务锁
func (r *schedulerRepo) AcquireLock(ctx context.Context, taskName string, ttl time.Duration) (bool, error) {
	lock := &model.TaskLock{
		TaskName:  taskName,
		LockedAt:  time.Now(),
		ExpiresAt: time.Now().Add(ttl),
	}

	// 尝试插入锁记录
	result := r.db.WithContext(ctx).Create(lock)
	if result.Error != nil {
		// 如果已存在，检查是否过期
		var existingLock model.TaskLock
		err := r.db.WithContext(ctx).Where("task_name = ?", taskName).First(&existingLock).Error
		if err != nil {
			return false, err
		}

		// 如果已过期，删除旧锁并重新获取
		if time.Now().After(existingLock.ExpiresAt) {
			r.db.WithContext(ctx).Delete(&existingLock)
			return r.AcquireLock(ctx, taskName, ttl)
		}

		return false, nil
	}

	return true, nil
}

// ReleaseLock 释放任务锁
func (r *schedulerRepo) ReleaseLock(ctx context.Context, taskName string) error {
	return r.db.WithContext(ctx).Where("task_name = ?", taskName).Delete(&model.TaskLock{}).Error
}

// GetUnpaidOrders 获取超时未支付的订单
func (r *schedulerRepo) GetUnpaidOrders(ctx context.Context, expireMinutes int) ([]uint64, error) {
	var orderIDs []uint64
	
	// 计算过期时间
	expireTime := time.Now().Add(-time.Duration(expireMinutes) * time.Minute)
	
	// 查询订单表（需要根据实际表结构调整）
	err := r.db.WithContext(ctx).
		Table("orders").
		Select("id").
		Where("status = ? AND created_at < ?", "PENDING", expireTime).
		Pluck("id", &orderIDs).Error
	
	return orderIDs, err
}

// GetUnconfirmedOrders 获取超时未确认收货的订单
func (r *schedulerRepo) GetUnconfirmedOrders(ctx context.Context, autoConfirmDays int) ([]uint64, error) {
	var orderIDs []uint64
	
	// 计算自动确认时间
	confirmTime := time.Now().Add(-time.Duration(autoConfirmDays) * 24 * time.Hour)
	
	// 查询订单表
	err := r.db.WithContext(ctx).
		Table("orders").
		Select("id").
		Where("status = ? AND shipped_at < ?", "SHIPPED", confirmTime).
		Pluck("id", &orderIDs).Error
	
	return orderIDs, err
}

// GetExpiredCoupons 获取已过期的优惠券
func (r *schedulerRepo) GetExpiredCoupons(ctx context.Context) ([]uint64, error) {
	var couponIDs []uint64
	
	err := r.db.WithContext(ctx).
		Table("user_coupons").
		Select("id").
		Where("status = ? AND expire_time < ?", "UNUSED", time.Now()).
		Pluck("id", &couponIDs).Error
	
	return couponIDs, err
}

// GetExpiringCoupons 获取即将过期的优惠券
func (r *schedulerRepo) GetExpiringCoupons(ctx context.Context, days int) ([]uint64, error) {
	var couponIDs []uint64
	
	// 计算即将过期时间范围
	startTime := time.Now()
	endTime := time.Now().Add(time.Duration(days) * 24 * time.Hour)
	
	err := r.db.WithContext(ctx).
		Table("user_coupons").
		Select("id").
		Where("status = ? AND expire_time BETWEEN ? AND ?", "UNUSED", startTime, endTime).
		Pluck("id", &couponIDs).Error
	
	return couponIDs, err
}
