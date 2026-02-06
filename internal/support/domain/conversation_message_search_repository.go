// 生成摘要：定义会话消息搜索仓储接口（Elasticsearch）。
package domain

import "context"

// ConversationMessageSearchRepository 定义会话消息搜索接口。
type ConversationMessageSearchRepository interface {
	Index(ctx context.Context, message *ConversationMessage) error
	Delete(ctx context.Context, messageID uint64) error
	Search(ctx context.Context, conversationID uint64, offset, limit int) ([]*ConversationMessage, int64, error)
}
