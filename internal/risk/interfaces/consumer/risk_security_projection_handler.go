// 生成摘要：新增风控事件消费处理器，用于驱动读模型投影更新。
package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/risk/application"
	"github.com/wyfcoding/ecommerce/internal/risk/domain"
)

// RiskSecurityProjectionHandler 处理风控事件并更新读模型。
type RiskSecurityProjectionHandler struct {
	projector *application.RiskSecurityProjectionService
	logger    *slog.Logger
}

// NewRiskSecurityProjectionHandler 创建事件消费处理器。
func NewRiskSecurityProjectionHandler(projector *application.RiskSecurityProjectionService, logger *slog.Logger) *RiskSecurityProjectionHandler {
	return &RiskSecurityProjectionHandler{
		projector: projector,
		logger:    logger,
	}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *RiskSecurityProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case domain.RiskAnalysisCreatedEventType:
		var event domain.RiskAnalysisCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal risk analysis created event", "error", err)
			return err
		}
		return h.projector.OnRiskAnalysisCreated(ctx, &event)
	case domain.BlacklistAddedEventType:
		var event domain.BlacklistAddedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal blacklist added event", "error", err)
			return err
		}
		return h.projector.OnBlacklistAdded(ctx, &event)
	case domain.BlacklistRemovedEventType:
		var event domain.BlacklistRemovedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal blacklist removed event", "error", err)
			return err
		}
		return h.projector.OnBlacklistRemoved(ctx, &event)
	case domain.UserBehaviorUpdatedEventType:
		var event domain.UserBehaviorUpdatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal user behavior updated event", "error", err)
			return err
		}
		return h.projector.OnUserBehaviorUpdated(ctx, &event)
	case domain.DeviceFingerprintSavedEventType:
		var event domain.DeviceFingerprintSavedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal device fingerprint saved event", "error", err)
			return err
		}
		return h.projector.OnDeviceFingerprintSaved(ctx, &event)
	default:
		h.logger.WarnContext(ctx, "unknown risk event topic", "topic", msg.Topic)
		return nil
	}
}
