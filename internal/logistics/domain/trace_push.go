package domain

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"
)

var (
	ErrTraceSubscriptionNotFound = errors.New("trace subscription not found")
	ErrTraceSubscriptionInactive = errors.New("trace subscription is inactive")
	ErrPushFailed                = errors.New("push notification failed")
)

type PushChannel int8

const (
	PushChannelWebSocket PushChannel = 1
	PushChannelSSE       PushChannel = 2
	PushChannelMQTT      PushChannel = 3
	PushChannelWebhook   PushChannel = 4
	PushChannelSMS       PushChannel = 5
	PushChannelEmail     PushChannel = 6
	PushChannelAppPush   PushChannel = 7
)

func (c PushChannel) String() string {
	switch c {
	case PushChannelWebSocket:
		return "WEBSOCKET"
	case PushChannelSSE:
		return "SSE"
	case PushChannelMQTT:
		return "MQTT"
	case PushChannelWebhook:
		return "WEBHOOK"
	case PushChannelSMS:
		return "SMS"
	case PushChannelEmail:
		return "EMAIL"
	case PushChannelAppPush:
		return "APP_PUSH"
	default:
		return "UNKNOWN"
	}
}

type SubscriptionStatus int8

const (
	SubscriptionStatusActive    SubscriptionStatus = 1
	SubscriptionStatusInactive  SubscriptionStatus = 2
	SubscriptionStatusExpired   SubscriptionStatus = 3
	SubscriptionStatusCancelled SubscriptionStatus = 4
)

func (s SubscriptionStatus) String() string {
	switch s {
	case SubscriptionStatusActive:
		return "ACTIVE"
	case SubscriptionStatusInactive:
		return "INACTIVE"
	case SubscriptionStatusExpired:
		return "EXPIRED"
	case SubscriptionStatusCancelled:
		return "CANCELLED"
	default:
		return "UNKNOWN"
	}
}

type TraceSubscription struct {
	ID            uint               `json:"id"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	SubscriptionNo string            `json:"subscription_no"`
	TrackingNo    string             `json:"tracking_no"`
	LogisticsID   uint64             `json:"logistics_id"`
	UserID        uint64             `json:"user_id"`
	MerchantID    uint64             `json:"merchant_id"`
	Channels      []PushChannel      `json:"channels"`
	Status        SubscriptionStatus `json:"status"`
	LastPushAt    *time.Time         `json:"last_push_at"`
	LastTraceID   uint64             `json:"last_trace_id"`
	PushCount     int                `json:"push_count"`
	ExpiresAt     *time.Time         `json:"expires_at"`
	CancelledAt   *time.Time         `json:"cancelled_at"`
	CancelReason  string             `json:"cancel_reason"`
	WebhookURL    string             `json:"webhook_url"`
	ExtraConfig   map[string]any     `json:"extra_config"`
}

type TracePushEvent struct {
	ID           uint      `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	TrackingNo   string    `json:"tracking_no"`
	LogisticsID  uint64    `json:"logistics_id"`
	TraceID      uint64    `json:"trace_id"`
	Location     string    `json:"location"`
	Description  string    `json:"description"`
	Status       string    `json:"status"`
	Timestamp    time.Time `json:"timestamp"`
	Priority     int       `json:"priority"`
	PushedTo     []uint64  `json:"pushed_to"`
	FailedTo     []uint64  `json:"failed_to"`
}

type TracePushRecord struct {
	ID             uint       `json:"id"`
	CreatedAt      time.Time  `json:"created_at"`
	EventID        uint       `json:"event_id"`
	SubscriptionID uint       `json:"subscription_id"`
	UserID         uint64     `json:"user_id"`
	Channel        PushChannel `json:"channel"`
	Status         string     `json:"status"`
	SentAt         *time.Time `json:"sent_at"`
	DeliveredAt    *time.Time `json:"delivered_at"`
	ReadAt         *time.Time `json:"read_at"`
	ErrorMessage   string     `json:"error_message"`
	RetryCount     int        `json:"retry_count"`
}

type PushConfig struct {
	MaxRetries          int           `json:"max_retries"`
	RetryInterval       time.Duration `json:"retry_interval"`
	BatchSize           int           `json:"batch_size"`
	EnableBatchPush     bool          `json:"enable_batch_push"`
	DefaultExpiration   time.Duration `json:"default_expiration"`
	MaxSubscriptionsPerUser int        `json:"max_subscriptions_per_user"`
}

func DefaultPushConfig() *PushConfig {
	return &PushConfig{
		MaxRetries:          3,
		RetryInterval:       time.Second * 5,
		BatchSize:           100,
		EnableBatchPush:     true,
		DefaultExpiration:   time.Hour * 24 * 30,
		MaxSubscriptionsPerUser: 50,
	}
}

func NewTraceSubscription(subscriptionNo, trackingNo string, logisticsID, userID, merchantID uint64, channels []PushChannel) *TraceSubscription {
	return &TraceSubscription{
		SubscriptionNo: subscriptionNo,
		TrackingNo:     trackingNo,
		LogisticsID:    logisticsID,
		UserID:         userID,
		MerchantID:     merchantID,
		Channels:       channels,
		Status:         SubscriptionStatusActive,
		PushCount:      0,
		ExtraConfig:    make(map[string]any),
	}
}

func (s *TraceSubscription) SetExpiration(duration time.Duration) {
	t := time.Now().Add(duration)
	s.ExpiresAt = &t
}

func (s *TraceSubscription) IsExpired() bool {
	if s.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*s.ExpiresAt)
}

func (s *TraceSubscription) IsActive() bool {
	return s.Status == SubscriptionStatusActive && !s.IsExpired()
}

func (s *TraceSubscription) RecordPush(traceID uint64) {
	now := time.Now()
	s.LastPushAt = &now
	s.LastTraceID = traceID
	s.PushCount++
}

func (s *TraceSubscription) Cancel(reason string) {
	now := time.Now()
	s.Status = SubscriptionStatusCancelled
	s.CancelledAt = &now
	s.CancelReason = reason
}

func (s *TraceSubscription) Deactivate() {
	s.Status = SubscriptionStatusInactive
}

func (s *TraceSubscription) Activate() {
	s.Status = SubscriptionStatusActive
}

func (s *TraceSubscription) HasChannel(channel PushChannel) bool {
	for _, c := range s.Channels {
		if c == channel {
			return true
		}
	}
	return false
}

type TracePusher struct {
	logger       *slog.Logger
	config       *PushConfig
	subscriptions map[string][]*TraceSubscription
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	repository   TracePushRepository
	notifier     PushNotifier
}

type PushNotifier interface {
	Push(ctx context.Context, channel PushChannel, userID uint64, data map[string]any) error
	PushBatch(ctx context.Context, channel PushChannel, userIDs []uint64, data map[string]any) error
}

func NewTracePusher(logger *slog.Logger, config *PushConfig, repo TracePushRepository, notifier PushNotifier) *TracePusher {
	ctx, cancel := context.WithCancel(context.Background())
	return &TracePusher{
		logger:       logger,
		config:       config,
		subscriptions: make(map[string][]*TraceSubscription),
		ctx:          ctx,
		cancel:       cancel,
		repository:   repo,
		notifier:     notifier,
	}
}

func (p *TracePusher) Subscribe(ctx context.Context, subscription *TraceSubscription) error {
	if subscription.ExpiresAt == nil {
		subscription.SetExpiration(p.config.DefaultExpiration)
	}

	p.mu.Lock()
	p.subscriptions[subscription.TrackingNo] = append(p.subscriptions[subscription.TrackingNo], subscription)
	p.mu.Unlock()

	if p.repository != nil {
		if err := p.repository.SaveSubscription(ctx, subscription); err != nil {
			p.logger.Error("failed to save subscription", "tracking_no", subscription.TrackingNo, "error", err)
			return err
		}
	}

	p.logger.Info("subscribed to trace updates", "tracking_no", subscription.TrackingNo, "user_id", subscription.UserID)
	return nil
}

func (p *TracePusher) Unsubscribe(ctx context.Context, subscriptionID uint64) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for trackingNo, subs := range p.subscriptions {
		for i, sub := range subs {
			if sub.ID == uint(subscriptionID) {
				p.subscriptions[trackingNo] = append(subs[:i], subs[i+1:]...)
				sub.Cancel("user unsubscribed")

				if p.repository != nil {
					p.repository.UpdateSubscription(ctx, sub)
				}

				p.logger.Info("unsubscribed from trace updates", "subscription_id", subscriptionID)
				return nil
			}
		}
	}

	return ErrTraceSubscriptionNotFound
}

func (p *TracePusher) PushTrace(ctx context.Context, logistics *Logistics, trace *LogisticsTrace) error {
	p.mu.RLock()
	subs, exists := p.subscriptions[logistics.TrackingNo]
	p.mu.RUnlock()

	if !exists || len(subs) == 0 {
		return nil
	}

	event := &TracePushEvent{
		TrackingNo:  logistics.TrackingNo,
		LogisticsID: uint64(logistics.ID),
		TraceID:     uint64(trace.ID),
		Location:    trace.Location,
		Description: trace.Description,
		Status:      trace.Status,
		Timestamp:   time.Now(),
		Priority:    0,
		PushedTo:    make([]uint64, 0),
		FailedTo:    make([]uint64, 0),
	}

	if p.repository != nil {
		if err := p.repository.SaveEvent(ctx, event); err != nil {
			p.logger.Error("failed to save push event", "tracking_no", logistics.TrackingNo, "error", err)
		}
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, sub := range subs {
		if !sub.IsActive() {
			continue
		}

		if sub.LastTraceID >= uint64(trace.ID) {
			continue
		}

		wg.Add(1)
		go func(s *TraceSubscription) {
			defer wg.Done()

			data := map[string]any{
				"tracking_no": logistics.TrackingNo,
				"location":    trace.Location,
				"description": trace.Description,
				"status":      trace.Status,
				"timestamp":   event.Timestamp,
			}

			var pushErr error
			for _, channel := range s.Channels {
				if p.notifier != nil {
					pushErr = p.notifier.Push(ctx, channel, s.UserID, data)
					if pushErr == nil {
						break
					}
				}
			}

			mu.Lock()
			if pushErr != nil {
				event.FailedTo = append(event.FailedTo, s.UserID)
				p.logger.Error("failed to push trace", "tracking_no", logistics.TrackingNo, "user_id", s.UserID, "error", pushErr)
			} else {
				event.PushedTo = append(event.PushedTo, s.UserID)
				s.RecordPush(uint64(trace.ID))
			}
			mu.Unlock()

			if p.repository != nil {
				record := &TracePushRecord{
					EventID:        event.ID,
					SubscriptionID: s.ID,
					UserID:         s.UserID,
					Status:         "SUCCESS",
				}
				if pushErr != nil {
					record.Status = "FAILED"
					record.ErrorMessage = pushErr.Error()
				}
				now := time.Now()
				record.SentAt = &now
				p.repository.SavePushRecord(ctx, record)
				p.repository.UpdateSubscription(ctx, s)
			}
		}(sub)
	}

	wg.Wait()

	p.logger.Info("pushed trace update", "tracking_no", logistics.TrackingNo, "pushed_count", len(event.PushedTo), "failed_count", len(event.FailedTo))
	return nil
}

func (p *TracePusher) GetSubscriptions(trackingNo string) []*TraceSubscription {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.subscriptions[trackingNo]
}

func (p *TracePusher) LoadSubscriptions(ctx context.Context) error {
	if p.repository == nil {
		return nil
	}

	subs, err := p.repository.FindActiveSubscriptions(ctx, 10000)
	if err != nil {
		return err
	}

	p.mu.Lock()
	for _, sub := range subs {
		p.subscriptions[sub.TrackingNo] = append(p.subscriptions[sub.TrackingNo], sub)
	}
	p.mu.Unlock()

	p.logger.Info("loaded active subscriptions", "count", len(subs))
	return nil
}

func (p *TracePusher) CleanupExpired(ctx context.Context) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for trackingNo, subs := range p.subscriptions {
		var activeSubs []*TraceSubscription
		for _, sub := range subs {
			if sub.IsExpired() {
				sub.Status = SubscriptionStatusExpired
				if p.repository != nil {
					p.repository.UpdateSubscription(ctx, sub)
				}
			} else {
				activeSubs = append(activeSubs, sub)
			}
		}
		p.subscriptions[trackingNo] = activeSubs
	}
}

func (p *TracePusher) Stop() {
	p.cancel()
	p.logger.Info("trace pusher stopped")
}

type TracePushRepository interface {
	SaveSubscription(ctx context.Context, subscription *TraceSubscription) error
	UpdateSubscription(ctx context.Context, subscription *TraceSubscription) error
	FindSubscriptionByID(ctx context.Context, id uint64) (*TraceSubscription, error)
	FindSubscriptionsByTrackingNo(ctx context.Context, trackingNo string) ([]*TraceSubscription, error)
	FindSubscriptionsByUserID(ctx context.Context, userID uint64) ([]*TraceSubscription, error)
	FindActiveSubscriptions(ctx context.Context, limit int) ([]*TraceSubscription, error)
	DeleteSubscription(ctx context.Context, id uint64) error

	SaveEvent(ctx context.Context, event *TracePushEvent) error
	FindEventByID(ctx context.Context, id uint64) (*TracePushEvent, error)
	FindEventsByTrackingNo(ctx context.Context, trackingNo string, limit int) ([]*TracePushEvent, error)

	SavePushRecord(ctx context.Context, record *TracePushRecord) error
	FindPushRecordsByEventID(ctx context.Context, eventID uint) ([]*TracePushRecord, error)
}
