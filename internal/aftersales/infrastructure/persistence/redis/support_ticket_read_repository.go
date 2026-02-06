// 生成摘要：实现售后客服工单读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/aftersales/domain"
)

const supportTicketPrefix = "aftersales:support:ticket:"

type supportTicketReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewSupportTicketReadRepository 创建客服工单读模型仓储。
func NewSupportTicketReadRepository(client redis.UniversalClient, ttl time.Duration) domain.SupportTicketReadRepository {
	return &supportTicketReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *supportTicketReadRepository) Save(ctx context.Context, ticket *domain.SupportTicket) error {
	if ticket == nil {
		return nil
	}
	data, err := json.Marshal(ticket)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, r.key(ticket.ID), data, r.ttl).Err()
}

func (r *supportTicketReadRepository) GetByID(ctx context.Context, id uint64) (*domain.SupportTicket, error) {
	data, err := r.client.Get(ctx, r.key(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var item domain.SupportTicket
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *supportTicketReadRepository) Delete(ctx context.Context, id uint64) error {
	return r.client.Del(ctx, r.key(id)).Err()
}

func (r *supportTicketReadRepository) key(id uint64) string {
	return fmt.Sprintf("%s%d", supportTicketPrefix, id)
}
