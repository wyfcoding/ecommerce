package domain

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrConversationNotFound = errors.New("conversation not found")
	ErrKnowledgeBaseNotFound = errors.New("knowledge base not found")
	ErrArticleNotFound      = errors.New("article not found")
	ErrConversationEnded    = errors.New("conversation has ended")
	ErrTransferFailed       = errors.New("transfer to human failed")
)

type ConversationStatus string

const (
	ConversationStatusActive     ConversationStatus = "ACTIVE"
	ConversationStatusEnded      ConversationStatus = "ENDED"
	ConversationStatusTransferred ConversationStatus = "TRANSFERRED"
)

type MessageSender string

const (
	MessageSenderUser  MessageSender = "USER"
	MessageSenderBot   MessageSender = "BOT"
	MessageSenderAgent MessageSender = "AGENT"
)

type Sentiment string

const (
	SentimentVeryNegative Sentiment = "VERY_NEGATIVE"
	SentimentNegative     Sentiment = "NEGATIVE"
	SentimentNeutral      Sentiment = "NEUTRAL"
	SentimentPositive     Sentiment = "POSITIVE"
	SentimentVeryPositive Sentiment = "VERY_POSITIVE"
)

type IntentCategory string

const (
	IntentOrderInquiry   IntentCategory = "ORDER_INQUIRY"
	IntentProductQuestion IntentCategory = "PRODUCT_QUESTION"
	IntentReturnRefund   IntentCategory = "RETURN_REFUND"
	IntentPaymentIssue   IntentCategory = "PAYMENT_ISSUE"
	IntentShippingTracking IntentCategory = "SHIPPING_TRACKING"
	IntentAccountIssue   IntentCategory = "ACCOUNT_ISSUE"
	IntentComplaint      IntentCategory = "COMPLAINT"
	IntentGeneralInquiry IntentCategory = "GENERAL_INQUIRY"
	IntentOther          IntentCategory = "OTHER"
)

type Conversation struct {
	ID               string            `json:"id"`
	UserID           uint64            `json:"user_id"`
	Status           ConversationStatus `json:"status"`
	Messages         []*Message        `json:"messages"`
	PrimaryIntent    IntentCategory    `json:"primary_intent"`
	OverallSentiment Sentiment         `json:"overall_sentiment"`
	AssignedAgentID  string            `json:"assigned_agent_id"`
	StartedAt        time.Time         `json:"started_at"`
	EndedAt          *time.Time        `json:"ended_at"`
	Metadata         map[string]string `json:"metadata"`
}

type Message struct {
	ID           string        `json:"id"`
	ConversationID string      `json:"conversation_id"`
	Sender       MessageSender `json:"sender"`
	Content      string        `json:"content"`
	Attachments  []*Attachment `json:"attachments"`
	Sentiment    Sentiment     `json:"sentiment"`
	Intent       IntentCategory `json:"intent"`
	Confidence   float64       `json:"confidence"`
	CreatedAt    time.Time     `json:"created_at"`
}

type Attachment struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	URL  string `json:"url"`
	Name string `json:"name"`
}

type KnowledgeBase struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Language     string    `json:"language"`
	IsActive     bool      `json:"is_active"`
	ArticleCount int       `json:"article_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type KnowledgeArticle struct {
	ID               string    `json:"id"`
	KnowledgeBaseID  string    `json:"knowledge_base_id"`
	Title            string    `json:"title"`
	Content          string    `json:"content"`
	Keywords         []string  `json:"keywords"`
	Tags             []string  `json:"tags"`
	Category         string    `json:"category"`
	ViewCount        int       `json:"view_count"`
	HelpfulnessScore float64   `json:"helpfulness_score"`
	IsActive         bool      `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type QuickReply struct {
	Text    string `json:"text"`
	Payload string `json:"payload"`
	Type    string `json:"type"`
}

type ChatbotResponse struct {
	ResponseText      string       `json:"response_text"`
	SuggestedArticles []string     `json:"suggested_articles"`
	QuickReplies      []*QuickReply `json:"quick_replies"`
	NeedsHumanTransfer bool        `json:"needs_human_transfer"`
	TransferReason    string       `json:"transfer_reason"`
	DetectedIntent    IntentCategory `json:"detected_intent"`
	Confidence        float64      `json:"confidence"`
}

type SentimentAnalysis struct {
	Sentiment   Sentiment `json:"sentiment"`
	Confidence  float64   `json:"confidence"`
	Keywords    []string  `json:"keywords"`
	Explanation string    `json:"explanation"`
}

type IntentResult struct {
	Intent            IntentCategory `json:"intent"`
	Confidence        float64        `json:"confidence"`
	Entities          []*Entity      `json:"entities"`
	SuggestedResponse string         `json:"suggested_response"`
}

type Entity struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

type Feedback struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	MessageID      string    `json:"message_id"`
	WasHelpful     bool      `json:"was_helpful"`
	Rating         int       `json:"rating"`
	Comment        string    `json:"comment"`
	CreatedAt      time.Time `json:"created_at"`
}

func NewConversation(userID uint64) *Conversation {
	return &Conversation{
		UserID:    userID,
		Status:    ConversationStatusActive,
		Messages:  []*Message{},
		Metadata:  make(map[string]string),
		StartedAt: time.Now(),
	}
}

func (c *Conversation) AddMessage(sender MessageSender, content string) *Message {
	msg := &Message{
		ID:             fmt.Sprintf("MSG%d", time.Now().UnixNano()),
		ConversationID: c.ID,
		Sender:         sender,
		Content:        content,
		Attachments:    []*Attachment{},
		CreatedAt:      time.Now(),
	}
	c.Messages = append(c.Messages, msg)
	return msg
}

func (c *Conversation) End(summary string) {
	c.Status = ConversationStatusEnded
	now := time.Now()
	c.EndedAt = &now
}

func (c *Conversation) Transfer(agentID string) {
	c.Status = ConversationStatusTransferred
	c.AssignedAgentID = agentID
}

func (c *Conversation) IsActive() bool {
	return c.Status == ConversationStatusActive
}

func (c *Conversation) UpdateSentiment(sentiment Sentiment) {
	c.OverallSentiment = sentiment
}

func (c *Conversation) UpdateIntent(intent IntentCategory) {
	c.PrimaryIntent = intent
}

func NewKnowledgeBase(name, language string) *KnowledgeBase {
	return &KnowledgeBase{
		Name:     name,
		Language: language,
		IsActive: true,
	}
}

func NewKnowledgeArticle(knowledgeBaseID, title, content string) *KnowledgeArticle {
	return &KnowledgeArticle{
		KnowledgeBaseID:  knowledgeBaseID,
		Title:           title,
		Content:         content,
		Keywords:        []string{},
		Tags:            []string{},
		IsActive:        true,
		HelpfulnessScore: 0.0,
	}
}

func (a *KnowledgeArticle) IncrementViewCount() {
	a.ViewCount++
}

func (a *KnowledgeArticle) UpdateHelpfulnessScore(wasHelpful bool) {
	totalFeedback := a.ViewCount
	if totalFeedback == 0 {
		return
	}
	
	helpfulCount := int(a.HelpfulnessScore * float64(totalFeedback) / 100)
	if wasHelpful {
		helpfulCount++
	}
	
	a.HelpfulnessScore = float64(helpfulCount) / float64(totalFeedback+1) * 100
}

func (a *KnowledgeArticle) MatchesQuery(query string) bool {
	query = strings.ToLower(query)
	if strings.Contains(strings.ToLower(a.Title), query) {
		return true
	}
	if strings.Contains(strings.ToLower(a.Content), query) {
		return true
	}
	for _, keyword := range a.Keywords {
		if strings.Contains(strings.ToLower(keyword), query) {
			return true
		}
	}
	return false
}

func NewFeedback(conversationID, messageID string, wasHelpful bool, rating int, comment string) *Feedback {
	return &Feedback{
		ConversationID: conversationID,
		MessageID:      messageID,
		WasHelpful:     wasHelpful,
		Rating:         rating,
		Comment:        comment,
		CreatedAt:      time.Now(),
	}
}

type ConversationRepository interface {
	Save(ctx context.Context, conversation *Conversation) error
	FindByID(ctx context.Context, id string) (*Conversation, error)
	FindByUserID(ctx context.Context, userID uint64, limit, offset int) ([]*Conversation, error)
	Update(ctx context.Context, conversation *Conversation) error
}

type KnowledgeBaseRepository interface {
	Save(ctx context.Context, kb *KnowledgeBase) error
	FindByID(ctx context.Context, id string) (*KnowledgeBase, error)
	FindAll(ctx context.Context, activeOnly bool, limit, offset int) ([]*KnowledgeBase, int64, error)
	Update(ctx context.Context, kb *KnowledgeBase) error
}

type KnowledgeArticleRepository interface {
	Save(ctx context.Context, article *KnowledgeArticle) error
	FindByID(ctx context.Context, id string) (*KnowledgeArticle, error)
	FindByKnowledgeBaseID(ctx context.Context, kbID string, limit, offset int) ([]*KnowledgeArticle, error)
	Search(ctx context.Context, kbID, query string, tags []string, category string, limit int) ([]*KnowledgeArticle, []float64, error)
	Update(ctx context.Context, article *KnowledgeArticle) error
	Delete(ctx context.Context, id string) error
}

type FeedbackRepository interface {
	Save(ctx context.Context, feedback *Feedback) error
	FindByConversationID(ctx context.Context, conversationID string) ([]*Feedback, error)
}

type NLPService interface {
	AnalyzeSentiment(text string) (*SentimentAnalysis, error)
	DetectIntent(text string, context []*Message) (*IntentResult, error)
	ExtractEntities(text string) ([]*Entity, error)
}

type ChatbotEngine interface {
	GenerateResponse(ctx context.Context, conversationID string, userMessage string, context []*Message) (*ChatbotResponse, error)
	GetQuickReplies(intent IntentCategory) []*QuickReply
	ShouldTransferToHuman(conversation *Conversation) (bool, string)
}

type AgentAssignmentService interface {
	FindAvailableAgent(department string) (string, error)
	GetEstimatedWaitTime(department string) (int, error)
}

type SimpleNLPService struct{}

func NewSimpleNLPService() *SimpleNLPService {
	return &SimpleNLPService{}
}

func (s *SimpleNLPService) AnalyzeSentiment(text string) (*SentimentAnalysis, error) {
	text = strings.ToLower(text)
	
	negativeWords := []string{"bad", "terrible", "awful", "horrible", "disappointed", "angry", "frustrated", "worst", "hate", "problem", "issue", "broken", "failed", "error"}
	positiveWords := []string{"good", "great", "excellent", "amazing", "wonderful", "fantastic", "best", "love", "happy", "satisfied", "thanks", "thank", "helpful"}
	
	var negCount, posCount int
	var foundKeywords []string
	
	words := strings.Fields(text)
	for _, word := range words {
		for _, neg := range negativeWords {
			if strings.Contains(word, neg) {
				negCount++
				foundKeywords = append(foundKeywords, word)
			}
		}
		for _, pos := range positiveWords {
			if strings.Contains(word, pos) {
				posCount++
				foundKeywords = append(foundKeywords, word)
			}
		}
	}
	
	var sentiment Sentiment
	var confidence float64
	
	switch {
	case negCount > posCount+1:
		if negCount > 2 {
			sentiment = SentimentVeryNegative
		} else {
			sentiment = SentimentNegative
		}
		confidence = float64(negCount-posCount) / float64(negCount+posCount+1)
	case posCount > negCount+1:
		if posCount > 2 {
			sentiment = SentimentVeryPositive
		} else {
			sentiment = SentimentPositive
		}
		confidence = float64(posCount-negCount) / float64(negCount+posCount+1)
	default:
		sentiment = SentimentNeutral
		confidence = 0.5
	}
	
	return &SentimentAnalysis{
		Sentiment:   sentiment,
		Confidence:  confidence,
		Keywords:    foundKeywords,
		Explanation: fmt.Sprintf("Detected %d negative and %d positive keywords", negCount, posCount),
	}, nil
}

func (s *SimpleNLPService) DetectIntent(text string, context []*Message) (*IntentResult, error) {
	text = strings.ToLower(text)
	
	intentPatterns := map[IntentCategory][]string{
		IntentOrderInquiry:    {"order", "订单", "purchase", "购买", "my order", "我的订单"},
		IntentProductQuestion: {"product", "商品", "item", "物品", "price", "价格", "specification", "规格"},
		IntentReturnRefund:    {"return", "退货", "refund", "退款", "exchange", "换货", "money back", "退钱"},
		IntentPaymentIssue:    {"payment", "支付", "pay", "付款", "charge", "扣款", "transaction", "交易"},
		IntentShippingTracking: {"shipping", "发货", "delivery", "配送", "track", "物流", "package", "包裹"},
		IntentAccountIssue:    {"account", "账户", "login", "登录", "password", "密码", "profile", "资料"},
		IntentComplaint:       {"complaint", "投诉", "report", "举报", "issue", "问题", "problem", "麻烦"},
		IntentGeneralInquiry:  {"help", "帮助", "question", "问题", "how", "怎么", "what", "什么"},
	}
	
	var bestIntent IntentCategory = IntentOther
	var bestScore int
	
	for intent, patterns := range intentPatterns {
		score := 0
		for _, pattern := range patterns {
			if strings.Contains(text, pattern) {
				score++
			}
		}
		if score > bestScore {
			bestScore = score
			bestIntent = intent
		}
	}
	
	confidence := 0.5
	if bestScore > 0 {
		confidence = float64(bestScore) / float64(len(intentPatterns[bestIntent]))
	}
	
	entities := s.extractSimpleEntities(text)
	
	return &IntentResult{
		Intent:     bestIntent,
		Confidence: confidence,
		Entities:   entities,
	}, nil
}

func (s *SimpleNLPService) ExtractEntities(text string) ([]*Entity, error) {
	return s.extractSimpleEntities(text), nil
}

func (s *SimpleNLPService) extractSimpleEntities(text string) []*Entity {
	var entities []*Entity
	
	orderPatterns := []string{"order #", "订单号", "order number"}
	for _, pattern := range orderPatterns {
		if idx := strings.Index(strings.ToLower(text), pattern); idx != -1 {
			start := idx + len(pattern)
			end := start + 10
			if end > len(text) {
				end = len(text)
			}
			value := strings.TrimSpace(text[start:end])
			entities = append(entities, &Entity{
				Type:  "ORDER_ID",
				Value: value,
				Start: start,
				End:   end,
			})
		}
	}
	
	return entities
}

type SimpleChatbotEngine struct {
	articleRepo KnowledgeArticleRepository
	nlpService  NLPService
}

func NewSimpleChatbotEngine(articleRepo KnowledgeArticleRepository, nlpService NLPService) *SimpleChatbotEngine {
	return &SimpleChatbotEngine{
		articleRepo: articleRepo,
		nlpService:  nlpService,
	}
}

func (e *SimpleChatbotEngine) GenerateResponse(ctx context.Context, conversationID string, userMessage string, context []*Message) (*ChatbotResponse, error) {
	intentResult, err := e.nlpService.DetectIntent(userMessage, context)
	if err != nil {
		return nil, err
	}
	
	sentiment, err := e.nlpService.AnalyzeSentiment(userMessage)
	if err != nil {
		sentiment = &SentimentAnalysis{Sentiment: SentimentNeutral}
	}
	
	responseText := e.generateResponseText(intentResult.Intent, sentiment.Sentiment)
	
	articles, _, err := e.articleRepo.Search(ctx, "", userMessage, nil, "", 3)
	if err != nil {
		articles = nil
	}
	
	var suggestedArticleIDs []string
	for _, a := range articles {
		suggestedArticleIDs = append(suggestedArticleIDs, a.ID)
	}
	
	needsTransfer, transferReason := e.ShouldTransferToHuman(&Conversation{
		PrimaryIntent:    intentResult.Intent,
		OverallSentiment: sentiment.Sentiment,
	})
	
	return &ChatbotResponse{
		ResponseText:       responseText,
		SuggestedArticles:  suggestedArticleIDs,
		QuickReplies:       e.GetQuickReplies(intentResult.Intent),
		NeedsHumanTransfer: needsTransfer,
		TransferReason:     transferReason,
		DetectedIntent:     intentResult.Intent,
		Confidence:         intentResult.Confidence,
	}, nil
}

func (e *SimpleChatbotEngine) generateResponseText(intent IntentCategory, sentiment Sentiment) string {
	responses := map[IntentCategory]string{
		IntentOrderInquiry:    "I can help you with your order. Could you please provide your order number?",
		IntentProductQuestion: "I'd be happy to help with product information. What would you like to know?",
		IntentReturnRefund:    "I understand you want to process a return or refund. Let me guide you through the process.",
		IntentPaymentIssue:    "I can help with payment issues. What specific problem are you experiencing?",
		IntentShippingTracking: "I can help track your package. Please provide your tracking number or order ID.",
		IntentAccountIssue:    "I can assist with account-related issues. What do you need help with?",
		IntentComplaint:       "I'm sorry to hear about your issue. Let me help resolve this for you.",
		IntentGeneralInquiry:  "Hello! How can I assist you today?",
		IntentOther:           "I'm here to help. Could you please tell me more about what you need?",
	}
	
	response := responses[intent]
	
	if sentiment == SentimentNegative || sentiment == SentimentVeryNegative {
		response = "I apologize for any inconvenience. " + response
	}
	
	return response
}

func (e *SimpleChatbotEngine) GetQuickReplies(intent IntentCategory) []*QuickReply {
	replies := map[IntentCategory][]*QuickReply{
		IntentOrderInquiry: {
			{Text: "Track my order", Payload: "TRACK_ORDER", Type: "action"},
			{Text: "Order status", Payload: "ORDER_STATUS", Type: "action"},
			{Text: "Cancel order", Payload: "CANCEL_ORDER", Type: "action"},
		},
		IntentReturnRefund: {
			{Text: "Start return", Payload: "START_RETURN", Type: "action"},
			{Text: "Check refund status", Payload: "REFUND_STATUS", Type: "action"},
			{Text: "Return policy", Payload: "RETURN_POLICY", Type: "info"},
		},
		IntentPaymentIssue: {
			{Text: "Payment methods", Payload: "PAYMENT_METHODS", Type: "info"},
			{Text: "Failed payment", Payload: "FAILED_PAYMENT", Type: "action"},
			{Text: "Refund to card", Payload: "REFUND_CARD", Type: "action"},
		},
		IntentGeneralInquiry: {
			{Text: "Shipping info", Payload: "SHIPPING_INFO", Type: "info"},
			{Text: "Product help", Payload: "PRODUCT_HELP", Type: "info"},
			{Text: "Talk to agent", Payload: "TRANSFER_HUMAN", Type: "action"},
		},
	}
	
	if r, ok := replies[intent]; ok {
		return r
	}
	return []*QuickReply{
		{Text: "Help", Payload: "HELP", Type: "info"},
		{Text: "Talk to agent", Payload: "TRANSFER_HUMAN", Type: "action"},
	}
}

func (e *SimpleChatbotEngine) ShouldTransferToHuman(conversation *Conversation) (bool, string) {
	if conversation.OverallSentiment == SentimentVeryNegative {
		return true, "Customer is very frustrated"
	}
	
	if conversation.PrimaryIntent == IntentComplaint {
		return true, "Customer has a complaint"
	}
	
	if len(conversation.Messages) > 10 {
		botMessageCount := 0
		for _, m := range conversation.Messages {
			if m.Sender == MessageSenderBot {
				botMessageCount++
			}
		}
		if botMessageCount > 5 {
			return true, "Extended conversation without resolution"
		}
	}
	
	return false, ""
}
