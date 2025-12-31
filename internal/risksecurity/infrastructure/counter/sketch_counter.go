package counter

import (
	"context"
	"fmt"
	"sync"

	"github.com/fynnwu/all/pkg/algorithm"
	"github.com/fynnwu/all/ecommerce/internal/risksecurity/domain/repository"
)

// SketchCounter 基于 Count-Min Sketch 的频率统计实现
// 适用于高并发、内存敏感的场景
type SketchCounter struct {
	sketch *algorithm.CountMinSketch
	mu     sync.Mutex
}

// NewSketchCounter 创建一个新的 SketchCounter
// epsilon: 误差率 (如 0.01)
// delta: 失败概率 (如 0.01)
func NewSketchCounter(epsilon, delta float64) (*SketchCounter, error) {
	cms, err := algorithm.NewCountMinSketch(epsilon, delta)
	if err != nil {
		return nil, err
	}
	return &SketchCounter{
		sketch: cms,
	}, nil
}

// Add 实现 FrequencyRepository 接口
func (s *SketchCounter) Add(ctx context.Context, key string, delta uint64) error {
	// 注意：CountMinSketch 本身加了锁，或者是纯内存操作，这里直接调用即可。
	// 如果需要持久化或分布式同步，这里需要额外的逻辑。
	s.sketch.AddString(key, delta)
	return nil
}

// Estimate 实现 FrequencyRepository 接口
func (s *SketchCounter) Estimate(ctx context.Context, key string) (uint64, error) {
	count := s.sketch.EstimateString(key)
	return count, nil
}

// Reset 重置
func (s *SketchCounter) Reset(ctx context.Context) error {
	s.sketch.Reset()
	return nil
}

// Ensure implementation
var _ repository.FrequencyRepository = (*SketchCounter)(nil)
