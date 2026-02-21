package domain

import "time"

// KnowledgeBase 代表知识库。
type KnowledgeBase struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Language    string    `json:"language"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// KnowledgeArticle 代表知识库中的文章。
type KnowledgeArticle struct {
	ID              string    `json:"id"`
	KnowledgeBaseID string    `json:"knowledge_base_id"`
	Title           string    `json:"title"`
	Content         string    `json:"content"`
	Keywords        []string  `json:"keywords"`
	Category        string    `json:"category"`
	IsActive        bool      `json:"is_active"`
	Score           float64   `json:"score"` // 搜索匹配分值
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// AIConversation 代表 AI 客服会话。
type AIConversation struct {
	ID               string             `json:"id"`
	UserID           uint64             `json:"user_id"`
	Status           ConversationStatus `json:"status"`
	PrimaryIntent    IntentCategory     `json:"primary_intent"`
	OverallSentiment Sentiment          `json:"overall_sentiment"`
	StartedAt        time.Time          `json:"started_at"`
	EndedAt          *time.Time         `json:"ended_at"`
}

// AIMessage 代表 AI 会话中的消息。
type AIMessage struct {
	ID             string         `json:"id"`
	ConversationID string         `json:"conversation_id"`
	Sender         MessageSender  `json:"sender"`
	Content        string         `json:"content"`
	Sentiment      Sentiment      `json:"sentiment"`
	Intent         IntentCategory `json:"intent"`
	Confidence     float64        `json:"confidence"`
	CreatedAt      time.Time      `json:"created_at"`
}

// ConversationStatus 定义会话状态。
type ConversationStatus int

const (
	ConversationStatusActive      ConversationStatus = 1
	ConversationStatusEnded       ConversationStatus = 2
	ConversationStatusTransferred ConversationStatus = 3
)

// MessageSender 定义消息发送者类型。
type MessageSender int

const (
	MessageSenderUser  MessageSender = 1
	MessageSenderBot   MessageSender = 2
	MessageSenderAgent MessageSender = 3
)

// Sentiment 定义情感倾向。
type Sentiment int

const (
	SentimentVeryNegative Sentiment = 1
	SentimentNegative     Sentiment = 2
	SentimentNeutral      Sentiment = 3
	SentimentPositive     Sentiment = 4
	SentimentVeryPositive Sentiment = 5
)

// IntentCategory 定义意图类别。
type IntentCategory int

const (
	IntentOrderInquiry     IntentCategory = 1
	IntentProductQuestion  IntentCategory = 2
	IntentReturnRefund     IntentCategory = 3
	IntentPaymentIssue     IntentCategory = 4
	IntentShippingTracking IntentCategory = 5
	IntentAccountIssue     IntentCategory = 6
	IntentComplaint        IntentCategory = 7
	IntentGeneralInquiry   IntentCategory = 8
	IntentOther            IntentCategory = 9
)

// IntentResult 意图识别结果
type IntentResult struct {
	Intent            IntentCategory `json:"intent"`
	Confidence        float64        `json:"confidence"`
	SuggestedResponse string         `json:"suggested_response"`
}

// SentimentAnalysis 情感分析结果
type SentimentAnalysis struct {
	Sentiment  Sentiment `json:"sentiment"`
	Confidence float64   `json:"confidence"`
}
