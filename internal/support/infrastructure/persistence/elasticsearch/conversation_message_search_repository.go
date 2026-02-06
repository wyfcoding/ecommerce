// 生成摘要：实现会话消息搜索仓储（Elasticsearch）。
package elasticsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/support/domain"
	"github.com/wyfcoding/pkg/search"
)

type conversationMessageSearchRepository struct {
	client *search.Client
	index  string
}

// NewConversationMessageSearchRepository 创建会话消息搜索仓储实现。
func NewConversationMessageSearchRepository(client *search.Client, index string) domain.ConversationMessageSearchRepository {
	if index == "" {
		index = "support_conversation_messages"
	}
	return &conversationMessageSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *conversationMessageSearchRepository) Index(ctx context.Context, message *domain.ConversationMessage) error {
	if message == nil {
		return nil
	}
	return r.client.Index(ctx, r.index, r.documentID(message.ID), message)
}

func (r *conversationMessageSearchRepository) Delete(ctx context.Context, messageID uint64) error {
	if messageID == 0 {
		return nil
	}
	return r.client.Delete(ctx, r.index, r.documentID(messageID))
}

func (r *conversationMessageSearchRepository) Search(ctx context.Context, conversationID uint64, offset, limit int) ([]*domain.ConversationMessage, int64, error) {
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
			"term": map[string]any{"conversation_id": conversationID},
		},
	}

	var searchRes struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source domain.ConversationMessage `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, query, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	items := make([]*domain.ConversationMessage, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		item := hit.Source
		items[i] = &item
	}
	return items, searchRes.Hits.Total.Value, nil
}

func (r *conversationMessageSearchRepository) documentID(id uint64) string {
	return strings.TrimSpace(fmt.Sprintf("%d", id))
}
