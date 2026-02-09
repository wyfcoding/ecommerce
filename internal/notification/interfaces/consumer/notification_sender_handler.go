package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	kg "github.com/segmentio/kafka-go"
	"github.com/wyfcoding/pkg/messagequeue/kafka"
	"github.com/wyfcoding/pkg/notification"
)

// NotificationSenderHandler 处理实际的消息发送逻辑。
type NotificationSenderHandler struct {
	sender notification.Sender
	logger *slog.Logger
}

// NewNotificationSenderHandler 创建发送处理器。
func NewNotificationSenderHandler(sender notification.Sender, logger *slog.Logger) *NotificationSenderHandler {
	return &NotificationSenderHandler{
		sender: sender,
		logger: logger,
	}
}

// Handle 处理 Kafka消息。
// 消息格式预期为 pkg/messagequeue/kafka.NotificationCommand (JSON)。
func (h *NotificationSenderHandler) Handle(ctx context.Context, msg kg.Message) error {
	var cmd kafka.NotificationCommand
	if err := json.Unmarshal(msg.Value, &cmd); err != nil {
		h.logger.ErrorContext(ctx, "failed to unmarshal notification command", "error", err)
		return nil // 无法解析的消息直接丢弃，避免死循环
	}

	// 转换为 pkg.notification.Message
	notifMsg := &notification.Message{
		ID:        fmt.Sprintf("%s-%d", cmd.Target, time.Now().UnixNano()), // 临时ID，实际可携带 TraceID
		Channel:   h.sender.Channel(),
		Recipient: cmd.Target,
		Subject:   cmd.Subject,
		Content:   cmd.Content,
		Priority:  notification.PriorityNormal,
	}

	// 发送通知
	_, err := h.sender.Send(ctx, notifMsg)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to send notification via handler",
			"channel", h.sender.Channel(),
			"recipient", cmd.Target,
			"error", err)
		return err // 返回错误以触发 Kafka 重试
	}

	h.logger.InfoContext(ctx, "notification handled and sent successfully",
		"channel", h.sender.Channel(),
		"recipient", cmd.Target)
	return nil
}
