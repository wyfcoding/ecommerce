package scheduler

import (
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/order/domain"
	"github.com/wyfcoding/pkg/algorithm"
)

// WheelScheduler 基于时间轮的超时调度器
type WheelScheduler struct {
	wheel *algorithm.TimingWheel
}

// NewWheelScheduler 创建一个新的调度器
// tick: 时间粒度 (如 1s)
// buckets: 槽位数量 (如 3600)
func NewWheelScheduler(tick time.Duration, buckets int) (*WheelScheduler, error) {
	tw, err := algorithm.NewTimingWheel(tick, buckets)
	if err != nil {
		slog.Error("Failed to initialize TimingWheel", "error", err)
		return nil, err
	}
	slog.Info("WheelScheduler initialized", "tick", tick, "buckets", buckets)
	return &WheelScheduler{
			wheel: tw,
		},
		nil
}

// ScheduleTimeout 调度任务
func (s *WheelScheduler) ScheduleTimeout(orderID string, timeout time.Duration, callback func(orderID string)) error {
	// 将业务 callback 封装为无参函数
	task := func() {
		// 这里可以加一些恢复逻辑 (recover)
		defer func() {
			if r := recover(); r != nil {
				slog.Error("Recovered from panic in timeout task", "order_id", orderID, "recover", r)
			}
		}()
		slog.Debug("Timeout task triggered", "order_id", orderID)
		callback(orderID)
	}

	slog.Debug("Scheduling timeout task", "order_id", orderID, "timeout", timeout)
	return s.wheel.AddTask(timeout, task)
}

// Start 启动
func (s *WheelScheduler) Start() {
	s.wheel.Start()
}

// Stop 停止
func (s *WheelScheduler) Stop() {
	s.wheel.Stop()
}

var _ domain.TimeoutScheduler = (*WheelScheduler)(nil)
