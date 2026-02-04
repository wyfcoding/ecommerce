package application

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// WebhookSender 实现 Sender 接口，用于发送 HTTP Webhook。
type WebhookSender struct {
	client *http.Client
}

// NewWebhookSender 创建 Webhook 发送器。
func NewWebhookSender() *WebhookSender {
	return &WebhookSender{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Send 发送 Webhook。
// target: Webhook URL。
// subject: 忽略 (Webhook 通常只关心 payload)。
// content: JSON payload 字符串。
func (s *WebhookSender) Send(ctx context.Context, target, subject, content string) error {
	if target == "" {
		return errors.New("webhook url is empty")
	}

	// 尝试验证 content 是否为有效 JSON，若不是则封装
	if !json.Valid([]byte(content)) {
		payload := map[string]string{
			"subject": subject,
			"content": content,
		}
		b, _ := json.Marshal(payload)
		content = string(b)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", target, bytes.NewBufferString(content))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Ecommerce-Notification-Service/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return errors.New("webhook request failed with status: " + resp.Status)
	}

	return nil
}
