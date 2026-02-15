package mysql

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wyfcoding/ecommerce/internal/payment/domain"
	"github.com/wyfcoding/pkg/database/sharding"
	"github.com/wyfcoding/pkg/eventsourcing"
	"gorm.io/gorm"
)

// eventStore 实现 domain.EventStore 接口。
type eventStore struct {
	sharding *sharding.Manager
}

// NewEventStore 创建 EventStore。
func NewEventStore(sharding *sharding.Manager) domain.EventStore {
	return &eventStore{sharding: sharding}
}

// EventRecord 数据库中的事件记录。
type EventRecord struct {
	gorm.Model
	AggregateID   string `gorm:"column:aggregate_id;not null;index:idx_agg_version,unique"`
	AggregateType string `gorm:"column:aggregate_type;not null"`
	EventType     string `gorm:"column:event_type;not null"`
	Version       int64  `gorm:"column:version;not null;index:idx_agg_version,unique"`
	Data          []byte `gorm:"column:data;type:json"`
}

func (s *eventStore) Save(ctx context.Context, events []eventsourcing.DomainEvent) error {
	if len(events) == 0 {
		return nil
	}

	// 假设使用 AggregateID 进行分片路由，或者存储在特定库中
	// 这里简化处理，存储在 AggregateID 对应的分片库中
	// 我们需要从 AggregateID 提取 userID 或者其他分片键，
	// 但 AggregateID 是 PaymentNo。所以我们可能需要一个全局事件库或者根据 PaymentNo 分片。
	// 这里为了简单，假设所有事件存在 shard 0 或者根据特定规则。
	// 在实际复杂场景中，应该有专门的事件库。
	db := s.sharding.GetDB(0)

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, e := range events {
			data, err := json.Marshal(e)
			if err != nil {
				return err
			}

			record := EventRecord{
				AggregateID:   e.AggregateID(),
				AggregateType: "Payment",
				EventType:     e.EventType(),
				Version:       e.Version(),
				Data:          data,
			}

			if err := tx.Table("payment_events").Create(&record).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *eventStore) Load(ctx context.Context, aggregateID string) ([]eventsourcing.DomainEvent, error) {
	db := s.sharding.GetDB(0)
	var records []EventRecord
	if err := db.WithContext(ctx).Table("payment_events").Where("aggregate_id = ?", aggregateID).Order("version asc").Find(&records).Error; err != nil {
		return nil, err
	}

	events := make([]eventsourcing.DomainEvent, len(records))
	for i, r := range records {
		var event eventsourcing.DomainEvent
		switch r.EventType {
		case domain.PaymentInitiatedEventType:
			event = &domain.PaymentInitiatedEvent{}
		case domain.PaymentAuthorizedEventType:
			event = &domain.PaymentAuthorizedEvent{}
		case domain.PaymentCapturedEventType:
			event = &domain.PaymentCapturedEvent{}
		case domain.PaymentPaidEventType:
			event = &domain.PaymentPaidEvent{}
		case domain.PaymentClosedEventType:
			event = &domain.PaymentClosedEvent{}
		case domain.RefundFinishedEventType:
			event = &domain.RefundFinishedEvent{}
		default:
			return nil, fmt.Errorf("unknown event type: %s", r.EventType)
		}

		if err := json.Unmarshal(r.Data, event); err != nil {
			return nil, err
		}
		events[i] = event
	}
	return events, nil
}

func (s *eventStore) GetHistory(ctx context.Context, aggregateID string) ([]eventsourcing.DomainEvent, error) {
	return s.Load(ctx, aggregateID)
}
