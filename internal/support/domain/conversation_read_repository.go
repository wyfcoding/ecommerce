// 生成摘要：定义会话读模型仓储接口（Redis）。
package domain

import "context"

// ConversationReadRepository 定义会话读模型接口。
type ConversationReadRepository interface {
	Save(ctx context.Context, conversation *Conversation) error
	GetByID(ctx context.Context, id uint64) (*Conversation, error)
	Delete(ctx context.Context, id uint64) error
}
