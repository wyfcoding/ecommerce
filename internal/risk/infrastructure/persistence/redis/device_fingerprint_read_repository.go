// 生成摘要：实现设备指纹读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/risk/domain"
)

const deviceFingerprintPrefix = "risk:device:fingerprint:"

type deviceFingerprintReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewDeviceFingerprintReadRepository 创建设备指纹读模型仓储。
func NewDeviceFingerprintReadRepository(client redis.UniversalClient, ttl time.Duration) domain.DeviceFingerprintReadRepository {
	return &deviceFingerprintReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *deviceFingerprintReadRepository) Save(ctx context.Context, fp *domain.DeviceFingerprint) error {
	if fp == nil {
		return nil
	}
	data, err := json.Marshal(fp)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(fp.DeviceID), data, r.ttl).Err()
}

func (r *deviceFingerprintReadRepository) GetByDeviceID(ctx context.Context, deviceID string) (*domain.DeviceFingerprint, error) {
	data, err := r.client.Get(ctx, r.key(deviceID)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var fp domain.DeviceFingerprint
	if err := json.Unmarshal(data, &fp); err != nil {
		return nil, err
	}
	return &fp, nil
}

func (r *deviceFingerprintReadRepository) DeleteByDeviceID(ctx context.Context, deviceID string) error {
	return r.client.Del(ctx, r.key(deviceID)).Err()
}

func (r *deviceFingerprintReadRepository) key(deviceID string) string {
	return fmt.Sprintf("%s%s", deviceFingerprintPrefix, deviceID)
}
