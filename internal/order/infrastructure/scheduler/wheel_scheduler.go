package scheduler

import (
	"fmt"
	"time"

	"github.com/fynnwu/all/pkg/algorithm"
	"github.com/fynnwu/all/ecommerce/internal/order/domain/service"
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
		return nil, err
	}
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
				fmt.Printf("Recovered from panic in timeout task for order %s: %v\n", orderID, r)
			}
		}()
		callback(orderID)
	}

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

var _ service.TimeoutScheduler = (*WheelScheduler)(nil)
