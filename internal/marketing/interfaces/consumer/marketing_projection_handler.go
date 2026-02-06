// 生成摘要：新增营销事件消费处理器，用于驱动读模型投影更新。
package consumer

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/marketing/application"
	"github.com/wyfcoding/ecommerce/internal/marketing/domain"
)

// MarketingProjectionHandler 处理营销事件并更新读模型。
type MarketingProjectionHandler struct {
	projector *application.MarketingProjectionService
	logger    *slog.Logger
}

// NewMarketingProjectionHandler 创建事件消费处理器。
func NewMarketingProjectionHandler(projector *application.MarketingProjectionService, logger *slog.Logger) *MarketingProjectionHandler {
	return &MarketingProjectionHandler{
		projector: projector,
		logger:    logger,
	}
}

// Handle 处理 Kafka 消息并触发读模型投影。
func (h *MarketingProjectionHandler) Handle(ctx context.Context, msg kafka.Message) error {
	switch msg.Topic {
	case domain.CampaignCreatedEventType:
		var event domain.CampaignCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal campaign created event", "error", err)
			return err
		}
		return h.projector.OnCampaignCreated(ctx, &event)
	case domain.CampaignStatusUpdatedEventType:
		var event domain.CampaignStatusUpdatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal campaign status updated event", "error", err)
			return err
		}
		return h.projector.OnCampaignStatusUpdated(ctx, &event)
	case domain.ParticipationRecordedEventType:
		var event domain.ParticipationRecordedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal participation recorded event", "error", err)
			return err
		}
		return h.projector.OnParticipationRecorded(ctx, &event)
	case domain.BannerCreatedEventType:
		var event domain.BannerCreatedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal banner created event", "error", err)
			return err
		}
		return h.projector.OnBannerCreated(ctx, &event)
	case domain.BannerClickedEventType:
		var event domain.BannerClickedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal banner clicked event", "error", err)
			return err
		}
		return h.projector.OnBannerClicked(ctx, &event)
	case domain.BannerDeletedEventType:
		var event domain.BannerDeletedEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			h.logger.ErrorContext(ctx, "failed to unmarshal banner deleted event", "error", err)
			return err
		}
		return h.projector.OnBannerDeleted(ctx, &event)
	default:
		h.logger.WarnContext(ctx, "unknown marketing event topic", "topic", msg.Topic)
		return nil
	}
}
