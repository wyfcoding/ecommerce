// 生成摘要：实现工单消息搜索仓储（Elasticsearch）。
package elasticsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/support/domain"
	"github.com/wyfcoding/pkg/search"
)

type ticketMessageSearchRepository struct {
	client *search.Client
	index  string
}

// NewTicketMessageSearchRepository 创建工单消息搜索仓储实现。
func NewTicketMessageSearchRepository(client *search.Client, index string) domain.TicketMessageSearchRepository {
	if index == "" {
		index = "support_ticket_messages"
	}
	return &ticketMessageSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *ticketMessageSearchRepository) Index(ctx context.Context, message *domain.Message) error {
	if message == nil {
		return nil
	}
	return r.client.Index(ctx, r.index, r.documentID(message.ID), message)
}

func (r *ticketMessageSearchRepository) Delete(ctx context.Context, messageID uint64) error {
	if messageID == 0 {
		return nil
	}
	return r.client.Delete(ctx, r.index, r.documentID(messageID))
}

func (r *ticketMessageSearchRepository) Search(ctx context.Context, ticketID uint64, offset, limit int) ([]*domain.Message, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	query := map[string]any{
		"from":             offset,
		"size":             limit,
		"track_total_hits": true,
		"sort": []any{
			map[string]any{"created_at": map[string]any{"order": "asc"}},
		},
		"query": map[string]any{
			"term": map[string]any{"ticket_id": ticketID},
		},
	}

	var searchRes struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source domain.Message `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, query, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	items := make([]*domain.Message, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		item := hit.Source
		items[i] = &item
	}
	return items, searchRes.Hits.Total.Value, nil
}

func (r *ticketMessageSearchRepository) documentID(id uint64) string {
	return strings.TrimSpace(fmt.Sprintf("%d", id))
}
