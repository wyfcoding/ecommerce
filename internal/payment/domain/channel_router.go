package domain

import (
	"context"
	"math/rand"
	"slices"
	"sort"
	"sync"
	"time"
)

type ChannelHealth string

const (
	ChannelHealthHealthy   ChannelHealth = "healthy"
	ChannelHealthDegraded  ChannelHealth = "degraded"
	ChannelHealthUnhealthy ChannelHealth = "unhealthy"
)

type PaymentChannel struct {
	Code            string        `json:"code"`
	Name            string        `json:"name"`
	Type            ChannelType   `json:"type"`
	Priority        int           `json:"priority"`
	Enabled         bool          `json:"enabled"`
	Health          ChannelHealth `json:"health"`
	SuccessRate     float64       `json:"success_rate"`
	AvgLatencyMs    int64         `json:"avg_latency_ms"`
	MinAmount       int64         `json:"min_amount"`
	MaxAmount       int64         `json:"max_amount"`
	SupportCurrency []string      `json:"support_currency"`
	SupportRegion   []string      `json:"support_region"`
	FeePercent      float64       `json:"fee_percent"`
	FixedFee        int64         `json:"fixed_fee"`
	Config          string        `json:"config"`
	LastUpdated     time.Time     `json:"last_updated"`
}

type RoutingFactor struct {
	Amount         int64
	Currency       string
	Region         string
	UserID         uint64
	UserPreference string
	DeviceType     string
	PaymentMethod  string
}

type RoutingResult struct {
	PrimaryChannel   *PaymentChannel   `json:"primary_channel"`
	BackupChannels   []*PaymentChannel `json:"backup_channels"`
	RoutingReason    string            `json:"routing_reason"`
	RoutingScore     float64           `json:"routing_score"`
	FallbackStrategy string            `json:"fallback_strategy"`
}

type RoutingStrategy string

const (
	RoutingStrategyCost     RoutingStrategy = "cost"
	RoutingStrategySuccess  RoutingStrategy = "success"
	RoutingStrategyLatency  RoutingStrategy = "latency"
	RoutingStrategyBalanced RoutingStrategy = "balanced"
	RoutingStrategyUserPref RoutingStrategy = "user_preference"
)

type ChannelRouter struct {
	channels   map[string]*PaymentChannel
	mu         sync.RWMutex
	statistics *ChannelStatistics
	strategy   RoutingStrategy
}

type ChannelStatistics struct {
	SuccessCount map[string]int64
	FailCount    map[string]int64
	TotalLatency map[string]int64
	TotalCount   map[string]int64
	mu           sync.RWMutex
}

func NewChannelStatistics() *ChannelStatistics {
	return &ChannelStatistics{
		SuccessCount: make(map[string]int64),
		FailCount:    make(map[string]int64),
		TotalLatency: make(map[string]int64),
		TotalCount:   make(map[string]int64),
	}
}

func (s *ChannelStatistics) RecordSuccess(channelCode string, latencyMs int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SuccessCount[channelCode]++
	s.TotalLatency[channelCode] += latencyMs
	s.TotalCount[channelCode]++
}

func (s *ChannelStatistics) RecordFailure(channelCode string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.FailCount[channelCode]++
	s.TotalCount[channelCode]++
}

func (s *ChannelStatistics) GetSuccessRate(channelCode string) float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	total := s.TotalCount[channelCode]
	if total == 0 {
		return 1.0
	}
	return float64(s.SuccessCount[channelCode]) / float64(total)
}

func (s *ChannelStatistics) GetAvgLatency(channelCode string) int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	total := s.TotalCount[channelCode]
	if total == 0 {
		return 0
	}
	return s.TotalLatency[channelCode] / total
}

func NewChannelRouter(strategy RoutingStrategy) *ChannelRouter {
	return &ChannelRouter{
		channels:   make(map[string]*PaymentChannel),
		statistics: NewChannelStatistics(),
		strategy:   strategy,
	}
}

func (r *ChannelRouter) AddChannel(channel *PaymentChannel) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.channels[channel.Code] = channel
}

func (r *ChannelRouter) RemoveChannel(code string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.channels, code)
}

func (r *ChannelRouter) UpdateChannelHealth(code string, health ChannelHealth) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if ch, ok := r.channels[code]; ok {
		ch.Health = health
		ch.LastUpdated = time.Now()
	}
}

func (r *ChannelRouter) Route(ctx context.Context, factor *RoutingFactor) (*RoutingResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var candidates []*PaymentChannel
	for _, ch := range r.channels {
		if r.isChannelEligible(ch, factor) {
			candidates = append(candidates, ch)
		}
	}

	if len(candidates) == 0 {
		return nil, ErrNoAvailableChannel
	}

	sort.Slice(candidates, func(i, j int) bool {
		return r.calculateScore(candidates[i], factor) > r.calculateScore(candidates[j], factor)
	})

	primary := candidates[0]
	var backups []*PaymentChannel
	if len(candidates) > 1 {
		backups = candidates[1:min(3, len(candidates))]
	}

	return &RoutingResult{
		PrimaryChannel:   primary,
		BackupChannels:   backups,
		RoutingReason:    r.getRoutingReason(primary, factor),
		RoutingScore:     r.calculateScore(primary, factor),
		FallbackStrategy: "sequential",
	}, nil
}

func (r *ChannelRouter) isChannelEligible(ch *PaymentChannel, factor *RoutingFactor) bool {
	if !ch.Enabled {
		return false
	}
	if ch.Health == ChannelHealthUnhealthy {
		return false
	}
	if factor.Amount < ch.MinAmount || factor.Amount > ch.MaxAmount {
		return false
	}
	if factor.Currency != "" {
		currencySupported := slices.Contains(ch.SupportCurrency, factor.Currency)
		if !currencySupported {
			return false
		}
	}
	if factor.Region != "" {
		regionSupported := false
		for _, reg := range ch.SupportRegion {
			if reg == factor.Region || reg == "*" {
				regionSupported = true
				break
			}
		}
		if !regionSupported {
			return false
		}
	}
	return true
}

func (r *ChannelRouter) calculateScore(ch *PaymentChannel, factor *RoutingFactor) float64 {
	var score float64 = 100.0

	switch r.strategy {
	case RoutingStrategySuccess:
		score = ch.SuccessRate * 100
		if ch.Health == ChannelHealthDegraded {
			score *= 0.8
		}
	case RoutingStrategyCost:
		score = 100 - ch.FeePercent
		if ch.FixedFee > 0 {
			score -= float64(ch.FixedFee) / float64(factor.Amount) * 100
		}
	case RoutingStrategyLatency:
		if ch.AvgLatencyMs > 0 {
			score = 1000 / float64(ch.AvgLatencyMs)
		}
	case RoutingStrategyBalanced:
		successScore := ch.SuccessRate * 40
		latencyScore := 30.0
		if ch.AvgLatencyMs > 0 && ch.AvgLatencyMs < 1000 {
			latencyScore = 30 * (1 - float64(ch.AvgLatencyMs)/1000)
		}
		costScore := 30 - ch.FeePercent
		score = successScore + latencyScore + costScore
		if ch.Health == ChannelHealthDegraded {
			score *= 0.85
		}
	}

	score += float64(ch.Priority)

	return score
}

func (r *ChannelRouter) getRoutingReason(ch *PaymentChannel, factor *RoutingFactor) string {
	switch r.strategy {
	case RoutingStrategySuccess:
		return "Selected for highest success rate"
	case RoutingStrategyCost:
		return "Selected for lowest transaction cost"
	case RoutingStrategyLatency:
		return "Selected for lowest latency"
	case RoutingStrategyBalanced:
		return "Selected based on balanced scoring"
	default:
		return "Selected based on routing rules"
	}
}

func (r *ChannelRouter) RecordResult(channelCode string, success bool, latencyMs int64) {
	if success {
		r.statistics.RecordSuccess(channelCode, latencyMs)
	} else {
		r.statistics.RecordFailure(channelCode)
	}

	r.mu.Lock()
	if ch, ok := r.channels[channelCode]; ok {
		ch.SuccessRate = r.statistics.GetSuccessRate(channelCode)
		ch.AvgLatencyMs = r.statistics.GetAvgLatency(channelCode)
		if ch.SuccessRate < 0.8 {
			ch.Health = ChannelHealthDegraded
		}
		if ch.SuccessRate < 0.5 {
			ch.Health = ChannelHealthUnhealthy
		}
	}
	r.mu.Unlock()
}

func (r *ChannelRouter) GetAvailableChannels() []*PaymentChannel {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*PaymentChannel
	for _, ch := range r.channels {
		if ch.Enabled && ch.Health != ChannelHealthUnhealthy {
			result = append(result, ch)
		}
	}
	return result
}

func (r *ChannelRouter) SelectFallback(primary *PaymentChannel, backups []*PaymentChannel, factor *RoutingFactor) *PaymentChannel {
	for _, backup := range backups {
		if r.isChannelEligible(backup, factor) && backup.Health != ChannelHealthUnhealthy {
			return backup
		}
	}
	return nil
}

type WeightedRandomRouter struct {
	*ChannelRouter
	rand *rand.Rand
}

func NewWeightedRandomRouter(strategy RoutingStrategy) *WeightedRandomRouter {
	return &WeightedRandomRouter{
		ChannelRouter: NewChannelRouter(strategy),
		rand:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (r *WeightedRandomRouter) Route(ctx context.Context, factor *RoutingFactor) (*RoutingResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var candidates []*PaymentChannel
	var weights []float64
	totalWeight := 0.0

	for _, ch := range r.channels {
		if r.isChannelEligible(ch, factor) {
			weight := r.calculateScore(ch, factor)
			candidates = append(candidates, ch)
			weights = append(weights, weight)
			totalWeight += weight
		}
	}

	if len(candidates) == 0 {
		return nil, ErrNoAvailableChannel
	}

	target := r.rand.Float64() * totalWeight
	cumulative := 0.0
	var selected *PaymentChannel
	for i, ch := range candidates {
		cumulative += weights[i]
		if cumulative >= target {
			selected = ch
			break
		}
	}

	if selected == nil {
		selected = candidates[0]
	}

	return &RoutingResult{
		PrimaryChannel: selected,
		RoutingReason:  "Selected by weighted random distribution",
		RoutingScore:   r.calculateScore(selected, factor),
	}, nil
}

var ErrNoAvailableChannel = NewPaymentError("NO_AVAILABLE_CHANNEL", "no available payment channel")

func NewPaymentError(code, message string) *PaymentError {
	return &PaymentError{Code: code, Message: message}
}

type PaymentError struct {
	Code    string
	Message string
}

func (e *PaymentError) Error() string {
	return e.Message
}
