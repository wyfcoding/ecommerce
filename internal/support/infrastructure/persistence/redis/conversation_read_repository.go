// 生成摘要：实现会话读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/support/domain"
)

const conversationPrefix = "support:conversation:"

type conversationReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewConversationReadRepository 创建会话读模型仓储。
func NewConversationReadRepository(client redis.UniversalClient, ttl time.Duration) domain.ConversationReadRepository {
	return &conversationReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *conversationReadRepository) Save(ctx context.Context, conversation *domain.Conversation) error {
	if conversation == nil {
		return nil
	}
	data, err := json.Marshal(conversation)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(conversation.ID), data, r.ttl).Err()
}

func (r *conversationReadRepository) GetByID(ctx context.Context, id uint64) (*domain.Conversation, error) {
	data, err := r.client.Get(ctx, r.key(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var conversation domain.Conversation
	if err := json.Unmarshal(data, &conversation); err != nil {
		return nil, err
	}
	return &conversation, nil
}

func (r *conversationReadRepository) Delete(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.key(id)).Err()
}

func (r *conversationReadRepository) key(id uint64) string {
	return fmt.Sprintf("%s%d", conversationPrefix, id)
}
