package repository

import (
	"context"
	"time"
)

// DistributedLocker defines the interface for a distributed locking mechanism.
type DistributedLocker interface {
	AcquireLock(ctx context.Context, key string, expiry time.Duration) (bool, error)
	ReleaseLock(ctx context.Context, key string) (bool, error)
}