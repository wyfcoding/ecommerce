// 生成摘要：实现工单读模型 Redis 仓储。
package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wyfcoding/ecommerce/internal/support/domain"
)

const (
	ticketIDPrefix = "support:ticket:id:"
	ticketNoPrefix = "support:ticket:no:"
)

type ticketReadRepository struct {
	client redis.UniversalClient
	ttl    time.Duration
}

// NewTicketReadRepository 创建工单读模型仓储。
func NewTicketReadRepository(client redis.UniversalClient, ttl time.Duration) domain.TicketReadRepository {
	return &ticketReadRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *ticketReadRepository) Save(ctx context.Context, ticket *domain.Ticket) error {
	if ticket == nil {
		return nil
	}
	data, err := json.Marshal(ticket)
	if err != nil {
		return err
	}
	pipe := r.client.Pipeline()
	pipe.Set(ctx, r.keyByID(ticket.ID), data, r.ttl)
	if ticket.TicketNo != "" {
		pipe.Set(ctx, r.keyByNo(ticket.TicketNo), data, r.ttl)
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (r *ticketReadRepository) GetByID(ctx context.Context, id uint64) (*domain.Ticket, error) {
	data, err := r.client.Get(ctx, r.keyByID(id)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var ticket domain.Ticket
	if err := json.Unmarshal(data, &ticket); err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *ticketReadRepository) GetByNo(ctx context.Context, ticketNo string) (*domain.Ticket, error) {
	if ticketNo == "" {
		return nil, nil
	}
	data, err := r.client.Get(ctx, r.keyByNo(ticketNo)).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var ticket domain.Ticket
	if err := json.Unmarshal(data, &ticket); err != nil {
		return nil, err
	}
	return &ticket, nil
}

func (r *ticketReadRepository) Delete(ctx context.Context, id uint64, ticketNo string) error {
	keys := make([]string, 0, 2)
	if id != 0 {
		keys = append(keys, r.keyByID(id))
	}
	if ticketNo != "" {
		keys = append(keys, r.keyByNo(ticketNo))
	}
	if len(keys) == 0 {
		return nil
	}
	return r.client.Del(ctx, keys...).Err()
}

func (r *ticketReadRepository) keyByID(id uint64) string {
	return fmt.Sprintf("%s%d", ticketIDPrefix, id)
}

func (r *ticketReadRepository) keyByNo(ticketNo string) string {
	return ticketNoPrefix + ticketNo
}
