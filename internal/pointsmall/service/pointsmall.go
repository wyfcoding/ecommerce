package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ecommerce/internal/pointsmall/model"
	"ecommerce/internal/pointsmall/repository"
	"ecommerce/pkg/idgen"
)

var (
	ErrProductNotFound      = errors.New("积分商品不存在")
	ErrProductOffSale       = errors.New("商品已下架")
	ErrProductSoldOut       = errors.New("商品已售罄")
	ErrInsufficientStock    = errors.New("库存不足")
	ErrInsufficientPoints   = errors.New("积分不足")
	ErrExchangeLimitReached = errors.New("超过兑换限制")
	ErrActivityNotFound     = errors.New("抽奖活动不存在")
	ErrActivityNotActive    = errors.New("抽奖活动未激活")
	ErrDrawLimitReached     = errors.New("超过抽奖次数限制")
	ErrTaskNotFound         = errors.New("任务不存在")
	ErrTaskNotActive        = errors.New("任务未激活")
	ErrTaskAlreadyCompleted = errors.New("任务已完成")
)

// PointsMallService 积分商城服务接口
type PointsMallService interface {
	// 积分商品管理
	CreateProduct(ctx context.Context, product *model.PointsProduct) (*model.PointsProduct, error)
	UpdateProduct(ctx context.Context, product *model.PointsProduct) (*model.PointsProduct, error)
	GetProduct(ctx context.Context, id uint64) (*model.PointsProduct, error)
	ListProducts(ctx context.Context, categoryID uint64, status string, pageSize, pageNum int32) ([]*model.PointsProduct, int64, error)
	
	// 商品分类
	CreateCategory(ctx context.Context, category *model.PointsCategory) (*model.PointsCategory, error)
	ListCategories(ctx context.Context) ([]*model.PointsCategory, error)
	
	// 兑换订单
	ExchangeProduct(ctx context.Context, userID, productID uint64, quantity int32, addressID uint64) (*model.ExchangeOrder, error)
	GetExchangeOrder(ctx context.Context, orderNo string) (*model.ExchangeOrder, error)
	ListUserExchangeOrders(ctx context.Context, userID uint64, pageSize, pageNum int32) ([]*model.ExchangeOrder, int64, error)
	ShipExchangeOrder(ctx context.Context, orderNo, expressCompany, expressNo string) error
	
	// 抽奖活动
	CreateLotteryActivity(ctx context.Context, activity *model.LotteryActivity, prizes []*model.LotteryPrize) (*model.LotteryActivity, error)
	GetLotteryActivity(ctx context.Context, id uint64) (*model.LotteryActivity, []*model.LotteryPrize, error)
	DrawLottery(ctx context.Context, userID, activityID uint64) (*model.LotteryRecord, error)
	ListUserLotteryRecords(ctx context.Context, userID, activityID uint64, pageSize, pageNum int32) ([]*model.LotteryRecord, int64, error)
	
	// 积分任务
	CreateTask(ctx context.Context, task *model.PointsTask) (*model.PointsTask, error)
	ListTasks(ctx context.Context, taskType string) ([]*model.PointsTask, error)
	GetUserTaskProgress(ctx context.Context, userID uint64) ([]*model.UserTaskProgress, error)
	CompleteTask(ctx context.Context, userID, taskID uint64) (*model.UserTaskProgress, int64, error)
}

type pointsMallService struct {
	repo        repository.PointsMallRepo
	loyaltyRepo repository.LoyaltyRepo // 积分服务仓储
	redisClient *redis.Client
	logger      *zap.Logger
}

// NewPointsMallService 创建积分商城服务实例
func NewPointsMallService(
	repo repository.PointsMallRepo,
	loyaltyRepo repository.LoyaltyRepo,
	redisClient *redis.Client,
	logger *zap.Logger,
) PointsMallService {
	return &pointsMallService{
		repo:        repo,
		loyaltyRepo: loyaltyRepo,
		redisClient: redisClient,
		logger:      logger,
	}
}

// CreateProduct 创建积分商品
func (s *pointsMallService) CreateProduct(ctx context.Context, product *model.PointsProduct) (*model.PointsProduct, error) {
	product.ProductNo = fmt.Sprintf("PM%d", idgen.GenID())
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()

	if err := s.repo.CreateProduct(ctx, product); err != nil {
		s.logger.Error("创建积分商品失败", zap.Error(err))
		return nil, err
	}

	s.logger.Info("创建积分商品成功", zap.Uint64("productID", product.ID))
	return product, nil
}

// UpdateProduct 更新积分商品
func (s *pointsMallService) UpdateProduct(ctx context.Context, product *model.PointsProduct) (*model.PointsProduct, error) {
	product.UpdatedAt = time.Now()

	if err := s.repo.UpdateProduct(ctx, product); err != nil {
		s.logger.Error("更新积分商品失败", zap.Error(err))
		return nil, err
	}

	return product, nil
}

// GetProduct 获取积分商品详情
func (s *pointsMallService) GetProduct(ctx context.Context, id uint64) (*model.PointsProduct, error) {
	product, err := s.repo.GetProductByID(ctx, id)
	if err != nil {
		return nil, ErrProductNotFound
	}
	return product, nil
}

// ListProducts 获取积分商品列表
func (s *pointsMallService) ListProducts(ctx context.Context, categoryID uint64, status string, pageSize, pageNum int32) ([]*model.PointsProduct, int64, error) {
	products, total, err := s.repo.ListProducts(ctx, categoryID, status, pageSize, pageNum)
	if err != nil {
		s.logger.Error("获取积分商品列表失败", zap.Error(err))
		return nil, 0, err
	}
	return products, total, nil
}

// CreateCategory 创建商品分类
func (s *pointsMallService) CreateCategory(ctx context.Context, category *model.PointsCategory) (*model.PointsCategory, error) {
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()

	if err := s.repo.CreateCategory(ctx, category); err != nil {
		s.logger.Error("创建商品分类失败", zap.Error(err))
		return nil, err
	}

	return category, nil
}

// ListCategories 获取分类列表
func (s *pointsMallService) ListCategories(ctx context.Context) ([]*model.PointsCategory, error) {
	categories, err := s.repo.ListCategories(ctx)
	if err != nil {
		s.logger.Error("获取分类列表失败", zap.Error(err))
		return nil, err
	}
	return categories, nil
}

// ExchangeProduct 兑换商品
func (s *pointsMallService) ExchangeProduct(ctx context.Context, userID, productID uint64, quantity int32, addressID uint64) (*model.ExchangeOrder, error) {
	// 1. 获取商品信息
	product, err := s.repo.GetProductByID(ctx, productID)
	if err != nil {
		return nil, ErrProductNotFound
	}

	// 2. 验证商品状态
	if product.Status != model.PointsProductStatusOnSale {
		return nil, ErrProductOffSale
	}

	// 3. 验证库存
	if product.Stock < quantity {
		return nil, ErrInsufficientStock
	}

	// 4. 检查用户兑换限制
	userExchangeCount, err := s.repo.GetUserExchangeCount(ctx, userID, productID)
	if err != nil {
		s.logger.Error("获取用户兑换记录失败", zap.Error(err))
	}
	if userExchangeCount+int32(quantity) > product.LimitPerUser {
		return nil, ErrExchangeLimitReached
	}

	// 5. 计算所需积分和现金
	totalPoints := product.PointsPrice * int64(quantity)
	totalCash := product.CashPrice * int64(quantity)

	// 6. 验证用户积分
	userPoints, err := s.loyaltyRepo.GetUserPoints(ctx, userID)
	if err != nil {
		s.logger.Error("获取用户积分失败", zap.Error(err))
		return nil, err
	}
	if userPoints < totalPoints {
		return nil, ErrInsufficientPoints
	}

	// 7. 使用分布式锁确保并发安全
	lockKey := fmt.Sprintf("pointsmall:lock:product:%d", productID)
	locked, err := s.redisClient.SetNX(ctx, lockKey, 1, 5*time.Second).Result()
	if err != nil || !locked {
		return nil, fmt.Errorf("系统繁忙，请稍后重试")
	}
	defer s.redisClient.Del(ctx, lockKey)

	// 8. 再次检查库存（防止并发）
	product, err = s.repo.GetProductByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	if product.Stock < quantity {
		return nil, ErrInsufficientStock
	}

	// 9. 创建兑换订单
	orderNo := fmt.Sprintf("EX%d", idgen.GenID())
	order := &model.ExchangeOrder{
		OrderNo:      orderNo,
		UserID:       userID,
		ProductID:    productID,
		ProductName:  product.Name,
		ProductImage: product.MainImageURL,
		Quantity:     quantity,
		PointsPrice:  product.PointsPrice,
		CashPrice:    product.CashPrice,
		TotalPoints:  totalPoints,
		TotalCash:    totalCash,
		Status:       model.ExchangeOrderStatusPending,
		AddressID:    addressID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 10. 在事务中执行：扣减积分、扣减库存、创建订单
	err = s.repo.InTx(ctx, func(ctx context.Context) error {
		// 扣减积分
		if err := s.loyaltyRepo.DeductPoints(ctx, userID, totalPoints, "积分兑换", orderNo); err != nil {
			return err
		}

		// 扣减库存
		if err := s.repo.DeductProductStock(ctx, productID, quantity); err != nil {
			return err
		}

		// 创建订单
		if err := s.repo.CreateExchangeOrder(ctx, order); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		s.logger.Error("兑换商品失败", zap.Error(err))
		return nil, err
	}

	s.logger.Info("兑换商品成功",
		zap.String("orderNo", orderNo),
		zap.Uint64("userID", userID),
		zap.Uint64("productID", productID))

	return order, nil
}

// GetExchangeOrder 获取兑换订单详情
func (s *pointsMallService) GetExchangeOrder(ctx context.Context, orderNo string) (*model.ExchangeOrder, error) {
	order, err := s.repo.GetExchangeOrderByNo(ctx, orderNo)
	if err != nil {
		return nil, err
	}
	return order, nil
}

// ListUserExchangeOrders 获取用户兑换订单列表
func (s *pointsMallService) ListUserExchangeOrders(ctx context.Context, userID uint64, pageSize, pageNum int32) ([]*model.ExchangeOrder, int64, error) {
	orders, total, err := s.repo.ListUserExchangeOrders(ctx, userID, pageSize, pageNum)
	if err != nil {
		s.logger.Error("获取用户兑换订单失败", zap.Error(err))
		return nil, 0, err
	}
	return orders, total, nil
}

// ShipExchangeOrder 发货
func (s *pointsMallService) ShipExchangeOrder(ctx context.Context, orderNo, expressCompany, expressNo string) error {
	order, err := s.repo.GetExchangeOrderByNo(ctx, orderNo)
	if err != nil {
		return err
	}

	if order.Status != model.ExchangeOrderStatusPending {
		return fmt.Errorf("订单状态不允许发货")
	}

	now := time.Now()
	order.Status = model.ExchangeOrderStatusShipped
	order.ExpressCompany = expressCompany
	order.ExpressNo = expressNo
	order.ShippedAt = &now
	order.UpdatedAt = now

	if err := s.repo.UpdateExchangeOrder(ctx, order); err != nil {
		s.logger.Error("更新订单状态失败", zap.Error(err))
		return err
	}

	return nil
}

// CreateLotteryActivity 创建抽奖活动
func (s *pointsMallService) CreateLotteryActivity(ctx context.Context, activity *model.LotteryActivity, prizes []*model.LotteryPrize) (*model.LotteryActivity, error) {
	activity.CreatedAt = time.Now()
	activity.UpdatedAt = time.Now()

	// 验证概率总和
	var totalProbability float64
	for _, prize := range prizes {
		totalProbability += prize.Probability
	}
	if totalProbability > 1.0 {
		return nil, fmt.Errorf("奖品概率总和不能超过1.0")
	}

	err := s.repo.InTx(ctx, func(ctx context.Context) error {
		// 创建活动
		if err := s.repo.CreateLotteryActivity(ctx, activity); err != nil {
			return err
		}

		// 创建奖品
		for _, prize := range prizes {
			prize.ActivityID = activity.ID
			prize.RemainCount = prize.TotalCount
			prize.CreatedAt = time.Now()
			prize.UpdatedAt = time.Now()
			if err := s.repo.CreateLotteryPrize(ctx, prize); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		s.logger.Error("创建抽奖活动失败", zap.Error(err))
		return nil, err
	}

	return activity, nil
}

// GetLotteryActivity 获取抽奖活动详情
func (s *pointsMallService) GetLotteryActivity(ctx context.Context, id uint64) (*model.LotteryActivity, []*model.LotteryPrize, error) {
	activity, err := s.repo.GetLotteryActivityByID(ctx, id)
	if err != nil {
		return nil, nil, ErrActivityNotFound
	}

	prizes, err := s.repo.ListLotteryPrizes(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	return activity, prizes, nil
}

// DrawLottery 抽奖
func (s *pointsMallService) DrawLottery(ctx context.Context, userID, activityID uint64) (*model.LotteryRecord, error) {
	// 1. 获取活动信息
	activity, err := s.repo.GetLotteryActivityByID(ctx, activityID)
	if err != nil {
		return nil, ErrActivityNotFound
	}

	// 2. 验证活动状态
	now := time.Now()
	if !activity.IsActive || now.Before(activity.StartTime) || now.After(activity.EndTime) {
		return nil, ErrActivityNotActive
	}

	// 3. 检查用户抽奖次数
	if activity.MaxDrawsPerUser > 0 {
		userDrawCount, err := s.repo.GetUserDrawCount(ctx, userID, activityID)
		if err != nil {
			s.logger.Error("获取用户抽奖次数失败", zap.Error(err))
		}
		if userDrawCount >= activity.MaxDrawsPerUser {
			return nil, ErrDrawLimitReached
		}
	}

	// 4. 验证用户积分
	userPoints, err := s.loyaltyRepo.GetUserPoints(ctx, userID)
	if err != nil {
		return nil, err
	}
	if userPoints < activity.PointsPerDraw {
		return nil, ErrInsufficientPoints
	}

	// 5. 获取奖品列表
	prizes, err := s.repo.ListLotteryPrizes(ctx, activityID)
	if err != nil {
		return nil, err
	}

	// 6. 抽奖算法
	prize := s.drawPrize(prizes)

	// 7. 创建抽奖记录
	record := &model.LotteryRecord{
		ActivityID: activityID,
		UserID:     userID,
		Points:     activity.PointsPerDraw,
		IsWinning:  prize != nil,
		CreatedAt:  time.Now(),
	}

	if prize != nil {
		record.PrizeID = prize.ID
		record.PrizeName = prize.Name
	}

	// 8. 在事务中执行：扣减积分、扣减奖品库存、创建记录、发放奖励
	err = s.repo.InTx(ctx, func(ctx context.Context) error {
		// 扣减积分
		if err := s.loyaltyRepo.DeductPoints(ctx, userID, activity.PointsPerDraw, "抽奖消耗", fmt.Sprintf("LOTTERY_%d", activityID)); err != nil {
			return err
		}

		// 如果中奖，扣减奖品库存
		if prize != nil {
			if err := s.repo.DeductPrizeStock(ctx, prize.ID, 1); err != nil {
				return err
			}

			// 发放奖励
			if err := s.awardPrize(ctx, userID, prize); err != nil {
				return err
			}
		}

		// 创建记录
		if err := s.repo.CreateLotteryRecord(ctx, record); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		s.logger.Error("抽奖失败", zap.Error(err))
		return nil, err
	}

	s.logger.Info("抽奖成功",
		zap.Uint64("userID", userID),
		zap.Uint64("activityID", activityID),
		zap.Bool("isWinning", record.IsWinning))

	return record, nil
}

// drawPrize 抽奖算法
func (s *pointsMallService) drawPrize(prizes []*model.LotteryPrize) *model.LotteryPrize {
	// 过滤掉库存为0的奖品
	availablePrizes := make([]*model.LotteryPrize, 0)
	for _, prize := range prizes {
		if prize.RemainCount > 0 {
			availablePrizes = append(availablePrizes, prize)
		}
	}

	if len(availablePrizes) == 0 {
		return nil
	}

	// 生成随机数
	rand.Seed(time.Now().UnixNano())
	random := rand.Float64()

	// 根据概率抽奖
	var cumulative float64
	for _, prize := range availablePrizes {
		cumulative += prize.Probability
		if random <= cumulative {
			return prize
		}
	}

	return nil
}

// awardPrize 发放奖励
func (s *pointsMallService) awardPrize(ctx context.Context, userID uint64, prize *model.LotteryPrize) error {
	switch prize.Type {
	case "POINTS":
		// 发放积分
		// points, _ := strconv.ParseInt(prize.Value, 10, 64)
		// return s.loyaltyRepo.AddPoints(ctx, userID, points, "抽奖奖励", fmt.Sprintf("PRIZE_%d", prize.ID))
		// TODO: 实现积分发放
		return nil
	case "COUPON":
		// 发放优惠券
		// TODO: 调用优惠券服务
		return nil
	case "PRODUCT":
		// 发放商品
		// TODO: 创建兑换订单
		return nil
	default:
		return nil
	}
}

// ListUserLotteryRecords 获取用户抽奖记录
func (s *pointsMallService) ListUserLotteryRecords(ctx context.Context, userID, activityID uint64, pageSize, pageNum int32) ([]*model.LotteryRecord, int64, error) {
	records, total, err := s.repo.ListUserLotteryRecords(ctx, userID, activityID, pageSize, pageNum)
	if err != nil {
		s.logger.Error("获取用户抽奖记录失败", zap.Error(err))
		return nil, 0, err
	}
	return records, total, nil
}

// CreateTask 创建积分任务
func (s *pointsMallService) CreateTask(ctx context.Context, task *model.PointsTask) (*model.PointsTask, error) {
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	if err := s.repo.CreateTask(ctx, task); err != nil {
		s.logger.Error("创建积分任务失败", zap.Error(err))
		return nil, err
	}

	return task, nil
}

// ListTasks 获取任务列表
func (s *pointsMallService) ListTasks(ctx context.Context, taskType string) ([]*model.PointsTask, error) {
	tasks, err := s.repo.ListTasks(ctx, taskType)
	if err != nil {
		s.logger.Error("获取任务列表失败", zap.Error(err))
		return nil, err
	}
	return tasks, nil
}

// GetUserTaskProgress 获取用户任务进度
func (s *pointsMallService) GetUserTaskProgress(ctx context.Context, userID uint64) ([]*model.UserTaskProgress, error) {
	progress, err := s.repo.GetUserTaskProgress(ctx, userID)
	if err != nil {
		s.logger.Error("获取用户任务进度失败", zap.Error(err))
		return nil, err
	}
	return progress, nil
}

// CompleteTask 完成任务
func (s *pointsMallService) CompleteTask(ctx context.Context, userID, taskID uint64) (*model.UserTaskProgress, int64, error) {
	// 1. 获取任务信息
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, 0, ErrTaskNotFound
	}

	if !task.IsActive {
		return nil, 0, ErrTaskNotActive
	}

	// 2. 获取或创建用户任务进度
	progress, err := s.repo.GetUserTaskProgressByTaskID(ctx, userID, taskID)
	if err != nil {
		// 创建新进度
		progress = &model.UserTaskProgress{
			UserID:    userID,
			TaskID:    taskID,
			Progress:  0,
			ResetAt:   s.calculateResetTime(task.Type),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	// 3. 检查是否需要重置
	if time.Now().After(progress.ResetAt) {
		progress.Progress = 0
		progress.IsCompleted = false
		progress.CompletedAt = nil
		progress.ResetAt = s.calculateResetTime(task.Type)
	}

	// 4. 检查是否已完成
	if progress.IsCompleted {
		return nil, 0, ErrTaskAlreadyCompleted
	}

	// 5. 更新进度
	progress.Progress++
	progress.UpdatedAt = time.Now()

	// 6. 检查是否完成
	var points int64
	if progress.Progress >= task.Target {
		progress.IsCompleted = true
		now := time.Now()
		progress.CompletedAt = &now
		points = task.Points

		// 发放积分
		err = s.repo.InTx(ctx, func(ctx context.Context) error {
			// 更新进度
			if err := s.repo.UpdateUserTaskProgress(ctx, progress); err != nil {
				return err
			}

			// 发放积分
			if err := s.loyaltyRepo.AddPoints(ctx, userID, points, "完成任务", fmt.Sprintf("TASK_%d", taskID)); err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			s.logger.Error("完成任务失败", zap.Error(err))
			return nil, 0, err
		}
	} else {
		// 只更新进度
		if err := s.repo.UpdateUserTaskProgress(ctx, progress); err != nil {
			s.logger.Error("更新任务进度失败", zap.Error(err))
			return nil, 0, err
		}
	}

	return progress, points, nil
}

// calculateResetTime 计算重置时间
func (s *pointsMallService) calculateResetTime(taskType string) time.Time {
	now := time.Now()
	switch taskType {
	case "DAILY":
		// 明天0点
		return time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	case "WEEKLY":
		// 下周一0点
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		daysUntilMonday := 8 - weekday
		return time.Date(now.Year(), now.Month(), now.Day()+daysUntilMonday, 0, 0, 0, 0, now.Location())
	case "ONCE":
		// 永不重置
		return time.Date(9999, 12, 31, 23, 59, 59, 0, now.Location())
	default:
		return now.AddDate(0, 0, 1)
	}
}
