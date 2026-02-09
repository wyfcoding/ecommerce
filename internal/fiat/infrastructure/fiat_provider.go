package infrastructure

import (
	"context"
	"sync"
)

// MockFiatProvider 模拟法币汇率提供者
type MockFiatProvider struct {
	rates map[string]float64
	mu    sync.RWMutex
}

func NewMockFiatProvider() *MockFiatProvider {
	return &MockFiatProvider{
		rates: map[string]float64{
			"USD:CNY": 7.23,
			"CNY:USD": 0.138,
			"USD:EUR": 0.92,
			"EUR:USD": 1.08,
		},
	}
}

func (p *MockFiatProvider) GetRate(ctx context.Context, from, to string) (float64, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	key := from + ":" + to
	if rate, ok := p.rates[key]; ok {
		return rate, nil
	}
	return 1.0, nil
}
