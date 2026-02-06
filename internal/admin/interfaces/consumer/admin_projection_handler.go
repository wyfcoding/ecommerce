// 生成摘要：新增管理后台事件消费处理器，用于驱动读模型投影更新。
package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/admin/application"
	"github.com/wyfcoding/ecommerce/internal/admin/domain"
)

// AdminProjectionHandler 处理管理后台事件并更新读模型。
type AdminProjectionHandler struct {
	projector *application.AdminProjectionService
	logger    *slog.Logger
}

// NewAdminProjectionHandler 创建事件消费处理器。
func NewAdminProjectionHandler(projector *application.AdminProjectionService, logger *slog.Logger) *AdminProjectionHandler {
	return &AdminProjectionHandler{
		projector: projector,
		logger:    logger,
	}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *AdminProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case domain.AdminUserCreatedEventType, domain.AdminUserUpdatedEventType, domain.AdminUserDisabledEventType, domain.RoleAssignedEventType:
		var event domain.AdminUserUpdatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal admin user event", "error", err)
			return err
		}
		return h.projector.OnAdminUserChanged(ctx, event.UserID)
	case domain.SystemSettingUpdatedEventType:
		var event domain.SystemSettingUpdatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal system setting updated event", "error", err)
			return err
		}
		return h.projector.OnSystemSettingUpdated(ctx, event.Key)
	case domain.AuditLogCreatedEventType:
		var event domain.AuditLogCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal audit log created event", "error", err)
			return err
		}
		return h.projector.OnAuditLogCreated(ctx, event.LogID)
	default:
		h.logger.WarnContext(ctx, "unknown admin event topic", "topic", msg.Topic)
		return nil
	}
}
