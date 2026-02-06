// 生成摘要：实现通知模板搜索仓储（Elasticsearch）。
package elasticsearch

import (
	"context"
	"fmt"
	"strings"

	"github.com/wyfcoding/ecommerce/internal/notification/domain"
	"github.com/wyfcoding/pkg/search"
)

type notificationTemplateSearchRepository struct {
	client *search.Client
	index  string
}

// NewNotificationTemplateSearchRepository 创建模板搜索仓储实现。
func NewNotificationTemplateSearchRepository(client *search.Client, index string) domain.NotificationTemplateSearchRepository {
	if index == "" {
		index = "notification_templates"
	}
	return &notificationTemplateSearchRepository{
		client: client,
		index:  index,
	}
}

func (r *notificationTemplateSearchRepository) Index(ctx context.Context, template *domain.NotificationTemplate) error {
	if template == nil {
		return nil
	}
	return r.client.Index(ctx, r.index, r.documentID(template.ID), template)
}

func (r *notificationTemplateSearchRepository) Delete(ctx context.Context, templateID uint64) error {
	if templateID == 0 {
		return nil
	}
	return r.client.Delete(ctx, r.index, r.documentID(templateID))
}

func (r *notificationTemplateSearchRepository) Search(ctx context.Context, offset, limit int) ([]*domain.NotificationTemplate, int64, error) {
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

	var searchRes struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source domain.NotificationTemplate `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := r.client.Search(ctx, r.index, query, &searchRes); err != nil {
		return nil, 0, fmt.Errorf("es search failed: %w", err)
	}

	items := make([]*domain.NotificationTemplate, len(searchRes.Hits.Hits))
	for i, hit := range searchRes.Hits.Hits {
		item := hit.Source
		items[i] = &item
	}
	return items, searchRes.Hits.Total.Value, nil
}

func (r *notificationTemplateSearchRepository) documentID(id uint64) string {
	return strings.TrimSpace(fmt.Sprintf("%d", id))
}
