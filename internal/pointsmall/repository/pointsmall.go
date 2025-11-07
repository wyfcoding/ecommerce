package repository

import (
	"context"

	"gorm.io/gorm"

	"ecommerce/internal/pointsmall/model"
)

// PointsMallRepo 积分商城仓储接口
type PointsMallRepo interface {
	// 积分商品
	CreateProduct(ctx context.Context, product *model.PointsProduct) error
	UpdateProduct(ctx context.Context, product *model.PointsProduct) error
	GetProductByID(ctx context.Context, id uint64) (*model.PointsProduct, error)
	ListProducts(ctx context.Context, categoryID uint64, status string, pageSize, pageNum int32) ([]*model.PointsProduct, int64, error)
	DeductProductStock(ctx context.Context, productID uint64, quantity int32) error
	
	// 商品分类
	CreateCategory(ctx context.Context, category *model.PointsCategory) error
	ListCategories(ctx context.Context) ([]*model.PointsCategory, error)
	
	// 兑换订单
	CreateExchangeOrder(ctx context.Context, order *model.ExchangeOrder) error
	UpdateExchangeOrder(ctx context.Context, order *model.ExchangeOrder) error
	GetExchangeOrderByNo(ctx context.Context, orderNo string) (*model.ExchangeOrder, error)
	ListUserExchangeOrders(ctx context.Context, userID uint64, pageSize, pageNum int32) ([]*model.ExchangeOrder, int64, error)
	GetUserExchangeCount(ctx context.Context, userID, productID uint64) (int32, error)
	
	// 抽奖活动
	CreateLotteryActivity(ctx context.Context, activity *model.LotteryActivity) error
	GetLotteryActivityByID(ctx context.Context, id uint64) (*model.LotteryActivity, error)
	CreateLotteryPrize(ctx context.Context, prize *model.LotteryPrize) error
	ListLotteryPrizes(ctx context.Context, activityID uint64) ([]*model.LotteryPrize, error)
	DeductPrizeStock(ctx context.Context, prizeID uint64, quantity int32) error
	CreateLotteryRecord(ctx context.Context, record *model.LotteryRecord) error
	ListUserLotteryRecords(ctx context.Context, userID, activityID uint64, pageSize, pageNum int32) ([]*model.LotteryRecord, int64, error)
	GetUserDrawCount(ctx context.Context, userID, activityID uint64) (int32, error)
	
	// 积分任务
	CreateTask(ctx context.Context, task *model.PointsTask) error
	GetTaskByID(ctx context.Context, id uint64) (*model.PointsTask, error)
	ListTasks(ctx context.Context, taskType string) ([]*model.PointsTask, error)
	GetUserTaskProgressByTaskID(ctx context.Context, userID, taskID uint64) (*model.UserTaskProgress, error)
	GetUserTaskProgress(ctx context.Context, userID uint64) ([]*model.UserTaskProgress, error)
	UpdateUserTaskProgress(ctx context.Context, progress *model.UserTaskProgress) error
	
	// 事务
	InTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// LoyaltyRepo 积分服务仓储接口（需要从loyalty服务导入）
type LoyaltyRepo interface {
	GetUserPoints(ctx context.Context, userID uint64) (int64, error)
	DeductPoints(ctx context.Context, userID uint64, points int64, reason, refNo string) error
	AddPoints(ctx context.Context, userID uint64, points int64, reason, refNo string) error
}

type pointsMallRepo struct {
	db *gorm.DB
}

// NewPointsMallRepo 创建积分商城仓储实例
func NewPointsMallRepo(db *gorm.DB) PointsMallRepo {
	return &pointsMallRepo{db: db}
}

// CreateProduct 创建积分商品
func (r *pointsMallRepo) CreateProduct(ctx context.Context, product *model.PointsProduct) error {
	return r.db.WithContext(ctx).Create(product).Error
}

// UpdateProduct 更新积分商品
func (r *pointsMallRepo) UpdateProduct(ctx context.Context, product *model.PointsProduct) error {
	return r.db.WithContext(ctx).Save(product).Error
}

// GetProductByID 根据ID获取积分商品
func (r *pointsMallRepo) GetProductByID(ctx context.Context, id uint64) (*model.PointsProduct, error) {
	var product model.PointsProduct
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// ListProducts 获取积分商品列表
func (r *pointsMallRepo) ListProducts(ctx context.Context, categoryID uint64, status string, pageSize, pageNum int32) ([]*model.PointsProduct, int64, error) {
	var products []*model.PointsProduct
	var total int64

	query := r.db.WithContext(ctx).Model(&model.PointsProduct{})
	
	if categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}
	
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	err := query.Offset(int(offset)).Limit(int(pageSize)).Order("sort_order DESC, created_at DESC").Find(&products).Error
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

// DeductProductStock 扣减商品库存
func (r *pointsMallRepo) DeductProductStock(ctx context.Context, productID uint64, quantity int32) error {
	return r.db.WithContext(ctx).Model(&model.PointsProduct{}).
		Where("id = ? AND stock >= ?", productID, quantity).
		Updates(map[string]interface{}{
			"stock":      gorm.Expr("stock - ?", quantity),
			"sold_count": gorm.Expr("sold_count + ?", quantity),
		}).Error
}

// CreateCategory 创建商品分类
func (r *pointsMallRepo) CreateCategory(ctx context.Context, category *model.PointsCategory) error {
	return r.db.WithContext(ctx).Create(category).Error
}

// ListCategories 获取分类列表
func (r *pointsMallRepo) ListCategories(ctx context.Context) ([]*model.PointsCategory, error) {
	var categories []*model.PointsCategory
	err := r.db.WithContext(ctx).
		Where("is_visible = ?", true).
		Order("sort_order DESC, created_at DESC").
		Find(&categories).Error
	if err != nil {
		return nil, err
	}
	return categories, nil
}

// CreateExchangeOrder 创建兑换订单
func (r *pointsMallRepo) CreateExchangeOrder(ctx context.Context, order *model.ExchangeOrder) error {
	return r.db.WithContext(ctx).Create(order).Error
}

// UpdateExchangeOrder 更新兑换订单
func (r *pointsMallRepo) UpdateExchangeOrder(ctx context.Context, order *model.ExchangeOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

// GetExchangeOrderByNo 根据订单号获取兑换订单
func (r *pointsMallRepo) GetExchangeOrderByNo(ctx context.Context, orderNo string) (*model.ExchangeOrder, error) {
	var order model.ExchangeOrder
	err := r.db.WithContext(ctx).Where("order_no = ?", orderNo).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// ListUserExchangeOrders 获取用户兑换订单列表
func (r *pointsMallRepo) ListUserExchangeOrders(ctx context.Context, userID uint64, pageSize, pageNum int32) ([]*model.ExchangeOrder, int64, error) {
	var orders []*model.ExchangeOrder
	var total int64

	query := r.db.WithContext(ctx).Model(&model.ExchangeOrder{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	err := query.Offset(int(offset)).Limit(int(pageSize)).Order("created_at DESC").Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// GetUserExchangeCount 获取用户兑换次数
func (r *pointsMallRepo) GetUserExchangeCount(ctx context.Context, userID, productID uint64) (int32, error) {
	var count int32
	err := r.db.WithContext(ctx).Model(&model.ExchangeOrder{}).
		Where("user_id = ? AND product_id = ?", userID, productID).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&count).Error
	return count, err
}

// CreateLotteryActivity 创建抽奖活动
func (r *pointsMallRepo) CreateLotteryActivity(ctx context.Context, activity *model.LotteryActivity) error {
	return r.db.WithContext(ctx).Create(activity).Error
}

// GetLotteryActivityByID 根据ID获取抽奖活动
func (r *pointsMallRepo) GetLotteryActivityByID(ctx context.Context, id uint64) (*model.LotteryActivity, error) {
	var activity model.LotteryActivity
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&activity).Error
	if err != nil {
		return nil, err
	}
	return &activity, nil
}

// CreateLotteryPrize 创建抽奖奖品
func (r *pointsMallRepo) CreateLotteryPrize(ctx context.Context, prize *model.LotteryPrize) error {
	return r.db.WithContext(ctx).Create(prize).Error
}

// ListLotteryPrizes 获取抽奖奖品列表
func (r *pointsMallRepo) ListLotteryPrizes(ctx context.Context, activityID uint64) ([]*model.LotteryPrize, error) {
	var prizes []*model.LotteryPrize
	err := r.db.WithContext(ctx).
		Where("activity_id = ?", activityID).
		Order("sort_order ASC").
		Find(&prizes).Error
	if err != nil {
		return nil, err
	}
	return prizes, nil
}

// DeductPrizeStock 扣减奖品库存
func (r *pointsMallRepo) DeductPrizeStock(ctx context.Context, prizeID uint64, quantity int32) error {
	return r.db.WithContext(ctx).Model(&model.LotteryPrize{}).
		Where("id = ? AND remain_count >= ?", prizeID, quantity).
		Update("remain_count", gorm.Expr("remain_count - ?", quantity)).Error
}

// CreateLotteryRecord 创建抽奖记录
func (r *pointsMallRepo) CreateLotteryRecord(ctx context.Context, record *model.LotteryRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

// ListUserLotteryRecords 获取用户抽奖记录
func (r *pointsMallRepo) ListUserLotteryRecords(ctx context.Context, userID, activityID uint64, pageSize, pageNum int32) ([]*model.LotteryRecord, int64, error) {
	var records []*model.LotteryRecord
	var total int64

	query := r.db.WithContext(ctx).Model(&model.LotteryRecord{}).Where("user_id = ?", userID)
	
	if activityID > 0 {
		query = query.Where("activity_id = ?", activityID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (pageNum - 1) * pageSize
	err := query.Offset(int(offset)).Limit(int(pageSize)).Order("created_at DESC").Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// GetUserDrawCount 获取用户抽奖次数
func (r *pointsMallRepo) GetUserDrawCount(ctx context.Context, userID, activityID uint64) (int32, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.LotteryRecord{}).
		Where("user_id = ? AND activity_id = ?", userID, activityID).
		Count(&count).Error
	return int32(count), err
}

// CreateTask 创建积分任务
func (r *pointsMallRepo) CreateTask(ctx context.Context, task *model.PointsTask) error {
	return r.db.WithContext(ctx).Create(task).Error
}

// GetTaskByID 根据ID获取积分任务
func (r *pointsMallRepo) GetTaskByID(ctx context.Context, id uint64) (*model.PointsTask, error) {
	var task model.PointsTask
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// ListTasks 获取积分任务列表
func (r *pointsMallRepo) ListTasks(ctx context.Context, taskType string) ([]*model.PointsTask, error) {
	var tasks []*model.PointsTask
	
	query := r.db.WithContext(ctx).Where("is_active = ?", true)
	
	if taskType != "" {
		query = query.Where("type = ?", taskType)
	}

	err := query.Order("sort_order DESC, created_at DESC").Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetUserTaskProgressByTaskID 根据任务ID获取用户任务进度
func (r *pointsMallRepo) GetUserTaskProgressByTaskID(ctx context.Context, userID, taskID uint64) (*model.UserTaskProgress, error) {
	var progress model.UserTaskProgress
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND task_id = ?", userID, taskID).
		First(&progress).Error
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

// GetUserTaskProgress 获取用户任务进度列表
func (r *pointsMallRepo) GetUserTaskProgress(ctx context.Context, userID uint64) ([]*model.UserTaskProgress, error) {
	var progress []*model.UserTaskProgress
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&progress).Error
	if err != nil {
		return nil, err
	}
	return progress, nil
}

// UpdateUserTaskProgress 更新用户任务进度
func (r *pointsMallRepo) UpdateUserTaskProgress(ctx context.Context, progress *model.UserTaskProgress) error {
	return r.db.WithContext(ctx).Save(progress).Error
}

// InTx 在事务中执行
func (r *pointsMallRepo) InTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, "tx", tx))
	})
}
