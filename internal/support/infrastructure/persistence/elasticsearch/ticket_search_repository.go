// 生成摘要：实现工单搜索仓储（Elasticsearch）。
package elasticsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/support/domain"
	"github.com/wyfcoding/pkg/search"
)

type ticketSearchRepository struct {
	client *search.Client
	index  string
}

// NewTicketSearchRepository 创建工单搜索仓储实现。
func NewTicketSearchRepository(client *search.Client, index string) domain.TicketSearchRepository {
	if index == "" {
		index = "support_tickets"
	}
	return &ticketSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *ticketSearchRepository) Index(ctx context.Context, ticket *domain.Ticket) error {
	if ticket == nil {
		return nil
	}
	return r.client.Index(ctx, r.index, r.documentID(ticket.ID), ticket)
}

func (r *ticketSearchRepository) Delete(ctx context.Context, ticketID uint64) error {
	if ticketID == 0 {
		return nil
	}
	return r.client.Delete(ctx, r.index, r.documentID(ticketID))
}

func (r *ticketSearchRepository) Search(ctx context.Context, userID uint64, status *domain.TicketStatus, offset, limit int) ([]*domain.Ticket, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	query := map[string]any{
		"from":             offset,
		"size":             limit,
		"track_total_hits": true,
		"sort": []any{
			map[string]any{"created_at": map[string]any{"order": "desc"}},
		},
	}
	filters := make([]any, 0, 2)
	if userID != 0 {
		filters = append(filters, map[string]any{"term": map[string]any{"user_id": userID}})
	}
	if status != nil {
		filters = append(filters, map[string]any{"term": map[string]any{"status": *status}})
	}
	if len(filters) > 0 {
		query["query"] = map[string]any{
			"bool": map[string]any{
				"filter": filters,
			},
		}
	}

	var searchRes struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source domain.Ticket `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, query, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	items := make([]*domain.Ticket, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		item := hit.Source
		items[i] = &item
	}
	return items, searchRes.Hits.Total.Value, nil
}

func (r *ticketSearchRepository) documentID(id uint64) string {
	return strings.TrimSpace(fmt.Sprintf("%d", id))
}
