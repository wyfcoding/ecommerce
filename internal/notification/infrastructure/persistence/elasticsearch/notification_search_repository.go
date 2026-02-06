// 生成摘要：实现通知搜索仓储（Elasticsearch），支持按用户与状态查询。
package elasticsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/notification/domain"
	"github.com/wyfcoding/pkg/search"
)

type notificationSearchRepository struct {
	client *search.Client
	index  string
}

// NewNotificationSearchRepository 创建通知搜索仓储实现。
func NewNotificationSearchRepository(client *search.Client, index string) domain.NotificationSearchRepository {
	if index == "" {
		index = "notifications"
	}
	return &notificationSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *notificationSearchRepository) Index(ctx context.Context, notification *domain.Notification) error {
	if notification == nil {
		return nil
	}
	return r.client.Index(ctx, r.index, r.documentID(notification.ID), notification)
}

func (r *notificationSearchRepository) Delete(ctx context.Context, notificationID uint64) error {
	if notificationID == 0 {
		return nil
	}
	return r.client.Delete(ctx, r.index, r.documentID(notificationID))
}

func (r *notificationSearchRepository) Search(ctx context.Context, userID uint64, status *domain.NotificationStatus, offset, limit int) ([]*domain.Notification, int64, error) {
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
				Source domain.Notification `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, query, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	items := make([]*domain.Notification, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		item := hit.Source
		items[i] = &item
	}
	return items, searchRes.Hits.Total.Value, nil
}

func (r *notificationSearchRepository) documentID(id uint64) string {
	return strings.TrimSpace(fmt.Sprintf("%d", id))
}
