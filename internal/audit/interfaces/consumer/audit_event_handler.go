package consumer

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"

	kg "github.com/segmentio/kafka-go"
	"github.com/wyfcoding/ecommerce/internal/audit/application"
	"github.com/wyfcoding/ecommerce/internal/audit/domain"
	pkgAudit "github.com/wyfcoding/pkg/audit"
	"github.com/wyfcoding/pkg/eventsourcing"
)

// AuditEventHandler 处理通用审计事件并持久化。
type AuditEventHandler struct {
	commandService *application.AuditCommandService
	logger         *slog.Logger
}

// NewAuditEventHandler 创建审计事件处理器。
func NewAuditEventHandler(commandService *application.AuditCommandService, logger *slog.Logger) *AuditEventHandler {
	return &AuditEventHandler{
		commandService: commandService,
		logger:         logger,
	}
}

// Handle 处理关联 audit.event 主题的 Kafka 消息。
// 消息结构为 eventsourcing.BaseEvent，Data 字段为 pkgAudit.Event。
func (h *AuditEventHandler) Handle(ctx context.Context, msg kg.Message) error {
	// 1. 解析 BaseEvent
	var baseEvent eventsourcing.BaseEvent
	if err := json.Unmarshal(msg.Value, &baseEvent); err != nil {
		h.logger.ErrorContext(ctx, "failed to unmarshal base event", "error", err)
		return nil // 格式错误，丢弃
	}

	// 2. 解析 pkgAudit.Event (BaseEvent.Data 是 interface{}，再次 unmarshal)
	// 注意：BaseEvent unmarshal 后 Data 可能是 map，需要重新转义
	// 这里简化处理：假设 Data 已经在 JSON 中是对象。
	// 更稳健的方式是将 BaseEvent.Data 定义为 json.RawMessage，或者二次序列化。
	// 由于 pkg/eventsourcing 定义 Data 为 interface{}，这里我们直接尝试将 Data 转回 bytes 再转 Event
	// 或者，如果 producer 发送的是嵌套 JSON，json.Unmarshal 会将 Data 解析为 map[string]interface{}

	// 为了确保正确，我们先将 msg.Value 解析到一个带 RawMessage 的临时结构
	var rawEvent struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(msg.Value, &rawEvent); err != nil {
		h.logger.ErrorContext(ctx, "failed to extract data from base event", "error", err)
		return nil
	}

	var auditEvent pkgAudit.Event
	if err := json.Unmarshal(rawEvent.Data, &auditEvent); err != nil {
		h.logger.ErrorContext(ctx, "failed to unmarshal audit event data", "error", err)
		return nil
	}

	// 3. 映射并记录日志
	userID := parseUserID(auditEvent.ActorID)
	username := "unknown" // 可以从 metadata 或用户服务获取，这里暂用 unknown
	if u, ok := auditEvent.Metadata["username"]; ok {
		username = u
	}

	// 映射 eventType
	// pkgAudit 没有 EventType，只有 Action。
	// 我们尝试从 Action 推断，或者将 Action 作为 EventType?
	// ecommerce domain LogEvent 需要 (eventType domain.AuditEventType, module, action string).
	// pkgAudit: Action, Resource.
	// 映射: module=Resource, action=Action, eventType="GENERAL" or derived.

	eventType := domain.AuditEventType("GENERAL")
	if auditEvent.Result == pkgAudit.ResultFailure {
		eventType = "FAILURE"
	}
	// 如果 metadata 中有 event_type 则优先使用
	if et, ok := auditEvent.Metadata["event_type"]; ok {
		eventType = domain.AuditEventType(et)
	}

	// 构造 Options
	opts := []application.LogOption{
		application.WithResource(auditEvent.Resource, auditEvent.ResourceID),
		application.WithClientInfo(auditEvent.IP, auditEvent.UserAgent),
	}

	if auditEvent.Duration > 0 {
		opts = append(opts, application.WithDuration(int64(auditEvent.Duration)))
	}
	if auditEvent.Error != "" {
		opts = append(opts, application.WithError(auditEvent.Error))
	}
	// 还可以处理 Metadata 到 Log 的扩展字段（如果有）

	// 4. 调用 CommandService
	// 注意：pkgAudit timestamp 忽略？CommandService 内部使用 time.Now()。
	// 如果需要保留原始时间，CommandService 需要支持 WithTimestamp Option。
	// 暂且使用当前时间记录。

	err := h.commandService.LogEvent(ctx, userID, username, eventType, auditEvent.Resource, auditEvent.Action, opts...)
	if err != nil {
		h.logger.ErrorContext(ctx, "failed to persist audit log from event", "error", err)
		return err // 返回错误触发重试
	}

	return nil
}

func parseUserID(actorID string) uint64 {
	uid, err := strconv.ParseUint(actorID, 10, 64)
	if err != nil {
		return 0 // 非数字 ID 视为系统或未登录
	}
	return uid
}
