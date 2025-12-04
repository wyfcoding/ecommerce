package repository

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/message/domain/entity" // 导入消息领域的实体定义。
)

// MessageRepository 是消息模块的仓储接口。
// 它定义了对消息和会话实体进行数据持久化操作的契约。
// 仓储接口属于领域层，旨在将领域逻辑与数据存储的实现细节解耦。
type MessageRepository interface {
	// --- 消息管理 (Message methods) ---

	// SaveMessage 将消息实体保存到数据存储中。
	// ctx: 上下文。
	// message: 待保存的消息实体。
	SaveMessage(ctx context.Context, message *entity.Message) error
	// GetMessage 根据ID获取消息实体。
	GetMessage(ctx context.Context, id uint64) (*entity.Message, error)
	// ListMessages 列出指定用户ID的所有消息实体，支持通过状态过滤和分页。
	ListMessages(ctx context.Context, userID uint64, status *entity.MessageStatus, offset, limit int) ([]*entity.Message, int64, error)
	// CountUnreadMessages 统计指定用户ID的未读消息数量。
	CountUnreadMessages(ctx context.Context, userID uint64) (int64, error)

	// --- 会话管理 (Conversation methods) ---

	// SaveConversation 将会话实体保存到数据存储中。
	SaveConversation(ctx context.Context, conversation *entity.Conversation) error
	// GetConversation 根据两个用户ID获取会话实体。
	// 会话通常在 user1ID 和 user2ID 之间，不区分顺序。
	GetConversation(ctx context.Context, user1ID, user2ID uint64) (*entity.Conversation, error)
	// ListConversations 列出指定用户ID参与的所有会话实体，支持分页。
	ListConversations(ctx context.Context, userID uint64, offset, limit int) ([]*entity.Conversation, int64, error)
}
