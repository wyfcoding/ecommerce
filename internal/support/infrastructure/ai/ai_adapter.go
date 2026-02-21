package ai

import (
	"context"

	"github.com/wyfcoding/ecommerce/internal/support/domain"
)

// AIAdapter 定义了 AI 服务的适配器接口。
type AIAdapter interface {
	AnalyzeSentiment(ctx context.Context, text string) (*domain.SentimentAnalysis, error)
	RecognizeIntent(ctx context.Context, text string) (*domain.IntentResult, error)
	GenerateChatbotResponse(ctx context.Context, conversationID, userMessage string) (string, error)
}

// MockAIAdapter 提供了一个模拟的 AI 适配器实现。
type MockAIAdapter struct{}

func NewMockAIAdapter() AIAdapter {
	return &MockAIAdapter{}
}

func (a *MockAIAdapter) AnalyzeSentiment(ctx context.Context, text string) (*domain.SentimentAnalysis, error) {
	// 简单模拟逻辑：根据文本长度模拟情感
	sentiment := domain.SentimentNeutral
	if len(text) > 100 {
		sentiment = domain.SentimentPositive
	} else if len(text) < 10 {
		sentiment = domain.SentimentNegative
	}
	return &domain.SentimentAnalysis{
		Sentiment:  sentiment,
		Confidence: 0.85,
	}, nil
}

func (a *MockAIAdapter) RecognizeIntent(ctx context.Context, text string) (*domain.IntentResult, error) {
	// 简单模拟逻辑
	return &domain.IntentResult{
		Intent:            domain.IntentGeneralInquiry,
		Confidence:        0.9,
		SuggestedResponse: "您的意图已识别，正在为您处理。",
	}, nil
}

func (a *MockAIAdapter) GenerateChatbotResponse(ctx context.Context, conversationID, userMessage string) (string, error) {
	return "这是来自智能客服的回应：我已经收到您的消息，请稍等。", nil
}
