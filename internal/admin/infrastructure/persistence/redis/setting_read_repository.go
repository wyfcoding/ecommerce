// 生成摘要：实现系统配置读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/admin/domain"
)

const settingPrefix = "admin:setting:"

type settingReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewSettingReadRepository 创建系统配置读模型仓储。
func NewSettingReadRepository(client redis.UniversalClient, ttl time.Duration) domain.SettingReadRepository {
	return &settingReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *settingReadRepository) Save(ctx context.Context, setting *domain.SystemSetting) error {
	if setting == nil {
		return nil
	}
	data, err := json.Marshal(setting)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(setting.Key), data, r.ttl).Err()
}

func (r *settingReadRepository) GetByKey(ctx context.Context, key string) (*domain.SystemSetting, error) {
	data, err := r.client.Get(ctx, r.key(key)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var setting domain.SystemSetting
	if err := json.Unmarshal(data, &setting); err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *settingReadRepository) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, r.key(key)).Err()
}

func (r *settingReadRepository) key(key string) string {
	return fmt.Sprintf("%s%s", settingPrefix, key)
}
