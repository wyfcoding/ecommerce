package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"text/template"

	"go.uber.org/zap"

	"ecommerce/internal/notification/model"
	"ecommerce/internal/notification/repository"
	// 伪代码: 模拟各种通知渠道的客户端
	// "ecommerce/pkg/notification/email"
	// "ecommerce/pkg/notification/sms"
	// "ecommerce/pkg/notification/push"
)

// NotificationService 定义了通知服务的业务逻辑接口
type NotificationService interface {
	// ProcessEvent 是处理从消息队列接收到的事件的核心方法
	ProcessEvent(ctx context.Context, eventType string, payload []byte) error
	// SendImmediateNotification 提供一个同步发送通知的接口 (可选)
	SendImmediateNotification(ctx context.Context, channel model.NotificationChannel, recipient, subject, content string) error
}

// notificationService 是接口的具体实现
type notificationService struct {
	repo   repository.NotificationRepository
	logger *zap.Logger
	// emailSender email.Sender
	// smsSender   sms.Sender
	// pushSender  push.Sender
}

// NewNotificationService 创建一个新的 notificationService 实例
func NewNotificationService(repo repository.NotificationRepository, logger *zap.Logger) NotificationService {
	return &notificationService{repo: repo, logger: logger}
}

// ProcessEvent 处理异步事件
func (s *notificationService) ProcessEvent(ctx context.Context, eventType string, payload []byte) error {
	s.logger.Info("Processing event", zap.String("eventType", eventType))

	// 1. 根据事件类型确定要使用的模板ID
	templateID, ok := s.getTemplateIDForEvent(eventType)
	if !ok {
		return fmt.Errorf("未知的事件类型或没有对应的模板: %s", eventType)
	}

	// 2. 解析事件的 payload
	var data map[string]interface{}
	if err := json.Unmarshal(payload, &data); err != nil {
		return fmt.Errorf("解析事件 payload 失败: %w", err)
	}

	// 3. 从 payload 中提取必要信息
	recipient, _ := data["recipient"].(string) // 接收地址
	language, _ := data["language"].(string)   // 语言偏好
	if language == "" {
		language = "en-US" // 默认语言
	}

	// 4. 获取通知模板
	tmpl, err := s.repo.GetTemplate(ctx, templateID, language)
	if err != nil {
		return fmt.Errorf("获取通知模板失败: %w", err)
	}

	// 5. 渲染模板
	subject, content, err := s.renderTemplate(tmpl, data)
	if err != nil {
		return fmt.Errorf("渲染模板失败: %w", err)
	}

	// 6. 发送通知
	return s.dispatchNotification(ctx, tmpl.Channel, recipient, subject, content)
}

// SendImmediateNotification 同步发送一个即时通知
func (s *notificationService) SendImmediateNotification(ctx context.Context, channel model.NotificationChannel, recipient, subject, content string) error {
	return s.dispatchNotification(ctx, channel, recipient, subject, content)
}

// dispatchNotification 是实际的发送分发逻辑
func (s *notificationService) dispatchNotification(ctx context.Context, channel model.NotificationChannel, recipient, subject, content string) error {
	logEntry := &model.NotificationLog{
		Channel:   channel,
		Recipient: recipient,
		Subject:   subject,
		Content:   content,
		Status:    model.StatusPending,
	}

	// 先在数据库中创建日志
	if err := s.repo.CreateLog(ctx, logEntry); err != nil {
		s.logger.Error("Failed to create notification log", zap.Error(err))
		// 即使日志创建失败，也应尝试发送，但需要记录这个错误
	}

	var sendErr error
	switch channel {
	case model.ChannelEmail:
		// sendErr = s.emailSender.Send(recipient, subject, content)
		s.logger.Info("Simulating email send", zap.String("to", recipient))
	case model.ChannelSMS:
		// sendErr = s.smsSender.Send(recipient, content)
		s.logger.Info("Simulating SMS send", zap.String("to", recipient))
	case model.ChannelPush:
		// sendErr = s.pushSender.Send(recipient, subject, content)
		s.logger.Info("Simulating push notification send", zap.String("to", recipient))
	default:
		sendErr = fmt.Errorf("不支持的通知渠道: %s", channel)
	}

	// 更新日志状态
	if logEntry.ID > 0 {
		status := model.StatusSent
		failureReason := ""
		if sendErr != nil {
			status = model.StatusFailed
			failureReason = sendErr.Error()
		}
		if err := s.repo.UpdateLogStatus(context.Background(), logEntry.ID, status, failureReason); err != nil {
			s.logger.Error("Failed to update notification log status", zap.Error(err))
		}
	}

	return sendErr
}

// getTemplateIDForEvent 是一个映射，将事件类型转换为模板ID
func (s *notificationService) getTemplateIDForEvent(eventType string) (string, bool) {
	// 这个映射关系可以硬编码，也可以存储在数据库或配置中
	mapping := map[string]string{
		"user.registered":   "user_welcome_email",
		"order.shipped":     "order_shipped_sms",
		"payment.success":   "payment_success_push",
	}
	templateID, ok := mapping[eventType]
	return templateID, ok
}

// renderTemplate 使用 Go 的 text/template 引擎来渲染模板
func (s *notificationService) renderTemplate(tmpl *model.NotificationTemplate, data map[string]interface{}) (subject, body string, err error) {
	// 渲染标题
	subjTmpl, err := template.New("subject").Parse(tmpl.Subject)
	if err != nil {
		return "", "", err
	}
	var subjBuf bytes.Buffer
	if err := subjTmpl.Execute(&subjBuf, data); err != nil {
		return "", "", err
	}
	subject = subjBuf.String()

	// 渲染内容
	bodyTmpl, err := template.New("body").Parse(tmpl.Body)
	if err != nil {
		return "", "", err
	}
	var bodyBuf bytes.Buffer
	if err := bodyTmpl.Execute(&bodyBuf, data); err != nil {
		return "", "", err
	}
	body = bodyBuf.String()

	return subject, body, nil
}