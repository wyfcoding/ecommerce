package application

import (
	"context"
	"errors" // 导入标准错误处理库。

	"github.com/wyfcoding/ecommerce/internal/message/domain/entity"     // 导入消息领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/message/domain/repository" // 导入消息领域的仓储接口。

	"log/slog" // 导入结构化日志库。
)

// MessageService 结构体定义了消息管理相关的应用服务。
// 它协调领域层和基础设施层，处理消息的发送、获取、标记已读以及会话状态的管理等业务逻辑。
type MessageService struct {
	repo   repository.MessageRepository // 依赖MessageRepository接口，用于数据持久化操作。
	logger *slog.Logger                 // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewMessageService 创建并返回一个新的 MessageService 实例。
func NewMessageService(repo repository.MessageRepository, logger *slog.Logger) *MessageService {
	return &MessageService{
		repo:   repo,
		logger: logger,
	}
}

// SendMessage 发送一条消息。
// ctx: 上下文。
// senderID: 发送者用户ID。
// receiverID: 接收者用户ID。
// messageType: 消息类型。
// title: 消息标题。
// content: 消息内容。
// link: 消息关联的链接。
// 返回发送成功的Message实体和可能发生的错误。
func (s *MessageService) SendMessage(ctx context.Context, senderID, receiverID uint64, messageType entity.MessageType, title, content, link string) (*entity.Message, error) {
	message := entity.NewMessage(senderID, receiverID, messageType, title, content, link) // 创建Message实体。
	// 通过仓储接口保存消息。
	if err := s.repo.SaveMessage(ctx, message); err != nil {
		s.logger.Error("failed to save message", "error", err)
		return nil, err
	}

	// 如果是用户之间的消息（非系统消息），则更新会话。
	if senderID > 0 && receiverID > 0 && messageType != entity.MessageTypeSystem {
		if err := s.updateConversation(ctx, senderID, receiverID, message); err != nil {
			s.logger.Warn("failed to update conversation", "error", err)
			// 注意：更新会话失败不应该导致消息发送失败，只记录警告。
		}
	}

	return message, nil
}

// updateConversation 更新或创建会话记录。
// 它会根据发送者和接收者ID找到或创建一个会话，并更新最后一条消息。
// ctx: 上下文。
// senderID: 发送者用户ID。
// receiverID: 接收者用户ID。
// message: 关联的消息实体。
// 返回可能发生的错误。
func (s *MessageService) updateConversation(ctx context.Context, senderID, receiverID uint64, message *entity.Message) error {
	// 尝试获取现有会话。
	conv, err := s.repo.GetConversation(ctx, senderID, receiverID)
	if err != nil {
		return err
	}

	// 如果会话不存在，则创建一个新的会话。
	if conv == nil {
		conv = entity.NewConversation(senderID, receiverID)
	}

	// 更新会话的最后一条消息信息。
	conv.UpdateLastMessage(uint64(message.ID), message.Content, senderID)
	// 保存更新后的会话。
	return s.repo.SaveConversation(ctx, conv)
}

// GetMessage 获取指定ID的消息详情。
// ctx: 上下文。
// id: 消息ID。
// 返回Message实体和可能发生的错误。
func (s *MessageService) GetMessage(ctx context.Context, id uint64) (*entity.Message, error) {
	return s.repo.GetMessage(ctx, id)
}

// MarkAsRead 标记指定消息为已读。
// ctx: 上下文。
// id: 消息ID。
// userID: 操作用户ID，用于权限验证（确保只有接收者可以标记为已读）。
// 返回可能发生的错误。
func (s *MessageService) MarkAsRead(ctx context.Context, id uint64, userID uint64) error {
	message, err := s.repo.GetMessage(ctx, id)
	if err != nil {
		return err
	}
	if message == nil {
		return errors.New("message not found") // 消息不存在。
	}

	// 权限验证：确保当前用户是消息的接收者。
	if message.ReceiverID != userID {
		return errors.New("permission denied")
	}

	// 调用实体方法标记消息为已读。
	message.MarkAsRead()
	// 保存更新后的消息。
	return s.repo.SaveMessage(ctx, message)
}

// ListMessages 获取用户消息列表。
// ctx: 上下文。
// userID: 用户的唯一标识符。
// status: 消息状态（例如，未读、已读）。
// page, pageSize: 分页参数。
// 返回消息列表、总数和可能发生的错误。
func (s *MessageService) ListMessages(ctx context.Context, userID uint64, status *int, page, pageSize int) ([]*entity.Message, int64, error) {
	offset := (page - 1) * pageSize
	var msgStatus *entity.MessageStatus
	if status != nil {
		s := entity.MessageStatus(*status)
		msgStatus = &s
	}
	return s.repo.ListMessages(ctx, userID, msgStatus, offset, pageSize)
}

// GetUnreadCount 获取指定用户的未读消息数量。
// ctx: 上下文。
// userID: 用户的唯一标识符。
// 返回未读消息数量和可能发生的错误。
func (s *MessageService) GetUnreadCount(ctx context.Context, userID uint64) (int64, error) {
	return s.repo.CountUnreadMessages(ctx, userID)
}
