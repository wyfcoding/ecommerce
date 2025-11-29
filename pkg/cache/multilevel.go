package cache

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// MultiLevelCache implements a multi-level cache (L1: Local, L2: Distributed).
type MultiLevelCache struct {
	l1     Cache
	l2     Cache
	tracer trace.Tracer
}

// NewMultiLevelCache creates a new MultiLevelCache.
func NewMultiLevelCache(l1, l2 Cache) *MultiLevelCache {
	return &MultiLevelCache{
		l1:     l1,
		l2:     l2,
		tracer: otel.Tracer("github.com/wyfcoding/ecommerce/pkg/cache"),
	}
}

func (c *MultiLevelCache) Get(ctx context.Context, key string, value interface{}) error {
	ctx, span := c.tracer.Start(ctx, "MultiLevelCache.Get", trace.WithAttributes(
		attribute.String("cache.key", key),
	))
	defer span.End()

	// 1. Try L1
	if err := c.l1.Get(ctx, key, value); err == nil {
		span.SetAttributes(attribute.String("cache.hit", "L1"))
		return nil
	}

	// 2. Try L2
	if err := c.l2.Get(ctx, key, value); err == nil {
		span.SetAttributes(attribute.String("cache.hit", "L2"))
		// Populate L1
		_ = c.l1.Set(ctx, key, value, 0)
		return nil
	}

	span.SetAttributes(attribute.String("cache.hit", "miss"))
	return fmt.Errorf("cache miss: %s", key)
}

func (c *MultiLevelCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	ctx, span := c.tracer.Start(ctx, "MultiLevelCache.Set", trace.WithAttributes(
		attribute.String("cache.key", key),
	))
	defer span.End()

	// 1. Set L2 (Master)
	if err := c.l2.Set(ctx, key, value, expiration); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "failed to set L2")
		return fmt.Errorf("failed to set L2: %w", err)
	}

	// 2. Set L1
	if err := c.l1.Set(ctx, key, value, expiration); err != nil {
		span.RecordError(err)
		// Don't fail the operation if L1 fails, just log/trace it
		span.AddEvent("failed to set L1", trace.WithAttributes(attribute.String("error", err.Error())))
	}

	return nil
}

func (c *MultiLevelCache) Delete(ctx context.Context, keys ...string) error {
	ctx, span := c.tracer.Start(ctx, "MultiLevelCache.Delete", trace.WithAttributes(
		attribute.StringSlice("cache.keys", keys),
	))
	defer span.End()

	// Delete from both
	err1 := c.l1.Delete(ctx, keys...)
	err2 := c.l2.Delete(ctx, keys...)

	if err1 != nil {
		span.RecordError(err1)
		return fmt.Errorf("failed to delete L1: %w", err1)
	}
	if err2 != nil {
		span.RecordError(err2)
		return fmt.Errorf("failed to delete L2: %w", err2)
	}
	return nil
}

func (c *MultiLevelCache) Exists(ctx context.Context, key string) (bool, error) {
	ctx, span := c.tracer.Start(ctx, "MultiLevelCache.Exists", trace.WithAttributes(
		attribute.String("cache.key", key),
	))
	defer span.End()

	// Check L1 first
	exists, err := c.l1.Exists(ctx, key)
	if err != nil {
		return false, err
	}
	if exists {
		span.SetAttributes(attribute.Bool("cache.exists", true), attribute.String("cache.layer", "L1"))
		return true, nil
	}
	// Check L2
	exists, err = c.l2.Exists(ctx, key)
	if err == nil {
		span.SetAttributes(attribute.Bool("cache.exists", exists), attribute.String("cache.layer", "L2"))
	}
	return exists, err
}

func (c *MultiLevelCache) Close() error {
	err1 := c.l1.Close()
	err2 := c.l2.Close()
	if err1 != nil {
		return err1
	}
	return err2
}
