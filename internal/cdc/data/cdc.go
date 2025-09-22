package data

import (
	"context"
	"ecommerce/internal/cdc/biz"
	"ecommerce/internal/cdc/data/model"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go" // Assuming Kafka client
	"gorm.io/gorm"
)

const (
	cdcTopic = "cdc_events" // Kafka topic for CDC events
)

type cdcRepo struct {
	data *Data
	kafkaProducer *kafka.Writer
}

// NewCdcRepo creates a new CdcRepo.
func NewCdcRepo(data *Data, kafkaProducer *kafka.Writer) biz.CdcRepo {
	return &cdcRepo{data: data, kafkaProducer: kafkaProducer}
}

// CaptureChangeEvent simulates capturing a database change event and pushing it to Kafka.
func (r *cdcRepo) CaptureChangeEvent(ctx context.Context, event *biz.ChangeEvent) (*biz.ChangeEvent, error) {
	// 1. Save event metadata to database (optional, for audit/replay)
	po := &model.ChangeEvent{
		EventID:         event.EventID,
		TableName:       event.TableName,
		OperationType:   event.OperationType,
		PrimaryKeyValue: event.PrimaryKeyValue,
		OldData:         event.OldData,
		NewData:         event.NewData,
		EventTimestamp:  event.EventTimestamp,
	}
	if err := r.data.db.WithContext(ctx).Create(po).Error; err != nil {
		return nil, fmt.Errorf("failed to save CDC event metadata: %w", err)
	}
	event.ID = po.ID

	// 2. Publish event to Kafka
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal CDC event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(event.EventID),
		Value: eventBytes,
		Time:  event.EventTimestamp,
	}

	err = r.kafkaProducer.WriteMessages(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to publish CDC event to Kafka: %w", err)
	}

	fmt.Printf("CDC event captured and published: EventID=%s, Table=%s, Op=%s\n", event.EventID, event.TableName, event.OperationType)
	return event, nil
}
