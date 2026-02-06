// 生成摘要：实现营销活动读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/marketing/domain"
)

const campaignPrefix = "marketing:campaign:detail:"

type campaignReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewCampaignReadRepository 创建营销活动读模型仓储。
func NewCampaignReadRepository(client redis.UniversalClient, ttl time.Duration) domain.CampaignReadRepository {
	return &campaignReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *campaignReadRepository) Save(ctx context.Context, campaign *domain.Campaign) error {
	if campaign == nil {
		return nil
	}
	data, err := json.Marshal(campaign)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(campaign.ID), data, r.ttl).Err()
}

func (r *campaignReadRepository) GetByID(ctx context.Context, id uint64) (*domain.Campaign, error) {
	data, err := r.client.Get(ctx, r.key(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var campaign domain.Campaign
	if err := json.Unmarshal(data, &campaign); err != nil {
		return nil, err
	}
	return &campaign, nil
}

func (r *campaignReadRepository) Delete(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.key(id)).Err()
}

func (r *campaignReadRepository) key(id uint64) string {
	return fmt.Sprintf("%s%d", campaignPrefix, id)
}
