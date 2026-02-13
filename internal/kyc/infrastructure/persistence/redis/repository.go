// 变更说明：KYC Redis 读模型仓储实现
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/kyc/domain"
)

// KYCReadRepository KYC读模型仓储
type KYCReadRepository struct {
	client *redis.Client
	ttl    time.Duration
}

// NewKYCReadRepository 创建KYC读模型仓储
func NewKYCReadRepository(client *redis.Client, ttl time.Duration) domain.KYCReadRepository {
	return &KYCReadRepository{
		client: client,
		ttl:    ttl,
	}
}

// cacheKey 生成缓存键
func (r *KYCReadRepository) cacheKey(userID uint64) string {
	return fmt.Sprintf("kyc:user:%d:status", userID)
}

// Save 保存KYC状态到缓存
func (r *KYCReadRepository) Save(ctx context.Context, app *domain.KYCApplication) error {
	data := &kycCacheData{
		ApplicationID: app.ApplicationID,
		UserID:        app.UserID,
		Status:        int(app.Status),
		Level:         int(app.Level),
		VerifiedAt:    app.VerifiedAt,
		ExpiresAt:     app.ExpiresAt,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, r.cacheKey(app.UserID), jsonData, r.ttl).Err()
}

// FindByUserID 从缓存查询用户KYC状态
func (r *KYCReadRepository) FindByUserID(ctx context.Context, userID uint64) (*domain.KYCApplication, error) {
	data, err := r.client.Get(ctx, r.cacheKey(userID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var cacheData kycCacheData
	if err := json.Unmarshal(data, &cacheData); err != nil {
		return nil, err
	}

	return &domain.KYCApplication{
		ApplicationID: cacheData.ApplicationID,
		UserID:        cacheData.UserID,
		Status:        domain.KYCStatus(cacheData.Status),
		Level:         domain.KYCLevel(cacheData.Level),
		VerifiedAt:    cacheData.VerifiedAt,
		ExpiresAt:     cacheData.ExpiresAt,
	}, nil
}

// Delete 删除缓存
func (r *KYCReadRepository) Delete(ctx context.Context, userID uint64) error {
	return r.client.Del(ctx, r.cacheKey(userID)).Err()
}

// kycCacheData KYC缓存数据
type kycCacheData struct {
	ApplicationID string     `json:"application_id"`
	UserID        uint64     `json:"user_id"`
	Status        int        `json:"status"`
	Level         int        `json:"level"`
	VerifiedAt    *time.Time `json:"verified_at"`
	ExpiresAt     *time.Time `json:"expires_at"`
}
