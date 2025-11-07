package service

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// SchedulerService 定时任务服务接口
type SchedulerService interface {
	Start() error
	Stop() error
	AddJob(spec string, cmd func()) (cron.EntryID, error)
	RemoveJob(id cron.EntryID)
}

type schedulerService struct {
	cron         *cron.Cron
	orderClient  OrderClient
	couponClient CouponClient
	cacheClient  CacheClient
	logger       *zap.Logger
}

// OrderClient 订单服务客户端接口
type OrderClient interface {
	CancelUnpaidOrders(ctx context.Context, timeout time.Duration) (int, error)
	AutoConfirmOrders(ctx context.Context, days int) (int, error)
}

// CouponClient 优惠券服务客户端接口
type CouponClient interface {
	ExpireCoupons(ctx context.Context) (int, error)
	SendExpiryReminder(ctx context.Context, days int) (int, error)
}

// CacheClient 缓存客户端接口
type CacheClient interface {
	SyncInventoryToCache(ctx context.Context) error
	SyncPricesToCache(ctx context.Context) error
}

// NewSchedulerService 创建定时任务服务实例
func NewSchedulerService(
	orderClient OrderClient,
	couponClient CouponClient,
	cacheClient CacheClient,
	logger *zap.Logger,
) SchedulerService {
	// 创建带秒级精度的 cron
	c := cron.New(cron.WithSeconds())
	
	return &schedulerService{
		cron:         c,
		orderClient:  orderClient,
		couponClient: couponClient,
		cacheClient:  cacheClient,
		logger:       logger,
	}
}

// Start 启动定时任务服务
func (s *schedulerService) Start() error {
	s.logger.Info("启动定时任务服务")

	// 1. 每分钟取消超时未支付订单
	_, err := s.cron.AddFunc("0 * * * * *", func() {
		ctx := context.Background()
		count, err := s.orderClient.CancelUnpaidOrders(ctx, 30*time.Minute)
		if err != nil {
			s.logger.Error("取消超时订单失败", zap.Error(err))
		} else if count > 0 {
			s.logger.Info("取消超时订单成功", zap.Int("count", count))
		}
	})
	if err != nil {
		return err
	}

	// 2. 每小时自动确认收货（7天后）
	_, err = s.cron.AddFunc("0 0 * * * *", func() {
		ctx := context.Background()
		count, err := s.orderClient.AutoConfirmOrders(ctx, 7)
		if err != nil {
			s.logger.Error("自动确认收货失败", zap.Error(err))
		} else if count > 0 {
			s.logger.Info("自动确认收货成功", zap.Int("count", count))
		}
	})
	if err != nil {
		return err
	}

	// 3. 每天凌晨1点处理过期优惠券
	_, err = s.cron.AddFunc("0 0 1 * * *", func() {
		ctx := context.Background()
		count, err := s.couponClient.ExpireCoupons(ctx)
		if err != nil {
			s.logger.Error("处理过期优惠券失败", zap.Error(err))
		} else if count > 0 {
			s.logger.Info("处理过期优惠券成功", zap.Int("count", count))
		}
	})
	if err != nil {
		return err
	}

	// 4. 每天上午10点发送优惠券即将过期提醒（3天内过期）
	_, err = s.cron.AddFunc("0 0 10 * * *", func() {
		ctx := context.Background()
		count, err := s.couponClient.SendExpiryReminder(ctx, 3)
		if err != nil {
			s.logger.Error("发送优惠券过期提醒失败", zap.Error(err))
		} else if count > 0 {
			s.logger.Info("发送优惠券过期提醒成功", zap.Int("count", count))
		}
	})
	if err != nil {
		return err
	}

	// 5. 每10分钟同步库存到缓存
	_, err = s.cron.AddFunc("0 */10 * * * *", func() {
		ctx := context.Background()
		if err := s.cacheClient.SyncInventoryToCache(ctx); err != nil {
			s.logger.Error("同步库存到缓存失败", zap.Error(err))
		}
	})
	if err != nil {
		return err
	}

	// 6. 每30分钟同步价格到缓存
	_, err = s.cron.AddFunc("0 */30 * * * *", func() {
		ctx := context.Background()
		if err := s.cacheClient.SyncPricesToCache(ctx); err != nil {
			s.logger.Error("同步价格到缓存失败", zap.Error(err))
		}
	})
	if err != nil {
		return err
	}

	// 7. 每天凌晨2点生成日报
	_, err = s.cron.AddFunc("0 0 2 * * *", func() {
		s.generateDailyReport()
	})
	if err != nil {
		return err
	}

	// 8. 每月1号凌晨3点生成月报
	_, err = s.cron.AddFunc("0 0 3 1 * *", func() {
		s.generateMonthlyReport()
	})
	if err != nil {
		return err
	}

	// 启动 cron
	s.cron.Start()
	s.logger.Info("定时任务服务启动成功")

	return nil
}

// Stop 停止定时任务服务
func (s *schedulerService) Stop() error {
	s.logger.Info("停止定时任务服务")
	ctx := s.cron.Stop()
	<-ctx.Done()
	s.logger.Info("定时任务服务已停止")
	return nil
}

// AddJob 添加自定义定时任务
func (s *schedulerService) AddJob(spec string, cmd func()) (cron.EntryID, error) {
	return s.cron.AddFunc(spec, cmd)
}

// RemoveJob 移除定时任务
func (s *schedulerService) RemoveJob(id cron.EntryID) {
	s.cron.Remove(id)
}

// generateDailyReport 生成日报
func (s *schedulerService) generateDailyReport() {
	s.logger.Info("开始生成日报")
	
	ctx := context.Background()
	yesterday := time.Now().AddDate(0, 0, -1)
	
	// TODO: 调用分析服务生成日报
	// 1. 销售数据统计
	// 2. 用户活跃度统计
	// 3. 商品销售排行
	// 4. 异常订单统计
	
	s.logger.Info("日报生成完成", zap.Time("date", yesterday))
}

// generateMonthlyReport 生成月报
func (s *schedulerService) generateMonthlyReport() {
	s.logger.Info("开始生成月报")
	
	ctx := context.Background()
	lastMonth := time.Now().AddDate(0, -1, 0)
	
	// TODO: 调用分析服务生成月报
	// 1. 月度销售总结
	// 2. 用户增长分析
	// 3. 商品销售分析
	// 4. 财务数据汇总
	
	s.logger.Info("月报生成完成", zap.Time("month", lastMonth))
}
