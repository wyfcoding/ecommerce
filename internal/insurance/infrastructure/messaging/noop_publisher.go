package messaging

import (
	"context"
	"log/slog"
)

type NoOpEventPublisher struct{}

func (p *NoOpEventPublisher) Publish(ctx context.Context, topic string, key string, event any) error {
	slog.InfoContext(ctx, "NoOpEventPublisher: event published", "topic", topic, "key", key, "event", event)
	return nil
}

func (p *NoOpEventPublisher) PublishInTx(ctx context.Context, tx any, topic string, key string, event any) error {
	slog.InfoContext(ctx, "NoOpEventPublisher: event published in tx", "topic", topic, "key", key, "event", event)
	return nil
}
