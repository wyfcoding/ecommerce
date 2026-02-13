package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/intelligentsupport/domain"
	"github.com/wyfcoding/pkg/idgen"
)

type IntelligentSupportService struct {
	conversationRepo domain.ConversationRepository
	knowledgeBaseRepo domain.KnowledgeBaseRepository
	articleRepo      domain.KnowledgeArticleRepository
	feedbackRepo     domain.FeedbackRepository
	nlpService       domain.NLPService
	chatbotEngine    domain.ChatbotEngine
	agentService     domain.AgentAssignmentService
	idGenerator      idgen.Generator
	logger           *slog.Logger
}

func NewIntelligentSupportService(
	conversationRepo domain.ConversationRepository,
	knowledgeBaseRepo domain.KnowledgeBaseRepository,
	articleRepo domain.KnowledgeArticleRepository,
	feedbackRepo domain.FeedbackRepository,
	nlpService domain.NLPService,
	chatbotEngine domain.ChatbotEngine,
	agentService domain.AgentAssignmentService,
	idGenerator idgen.Generator,
	logger *slog.Logger,
) *IntelligentSupportService {
	return &IntelligentSupportService{
		conversationRepo: conversationRepo,
		knowledgeBaseRepo: knowledgeBaseRepo,
		articleRepo:      articleRepo,
		feedbackRepo:     feedbackRepo,
		nlpService:       nlpService,
		chatbotEngine:    chatbotEngine,
		agentService:     agentService,
		idGenerator:      idGenerator,
		logger:           logger,
	}
}

type StartConversationCommand struct {
	UserID          uint64
	InitialMessage  string
	Metadata        map[string]string
}

func (s *IntelligentSupportService) StartConversation(ctx context.Context, cmd *StartConversationCommand) (*domain.Conversation, error) {
	s.logger.InfoContext(ctx, "starting conversation", "user_id", cmd.UserID)

	conversation := domain.NewConversation(cmd.UserID)
	conversation.ID = fmt.Sprintf("CONV%d", s.idGenerator.Generate())
	
	if cmd.Metadata != nil {
		conversation.Metadata = cmd.Metadata
	}

	if cmd.InitialMessage != "" {
		msg := conversation.AddMessage(domain.MessageSenderUser, cmd.InitialMessage)
		
		intent, err := s.nlpService.DetectIntent(cmd.InitialMessage, nil)
		if err == nil {
			msg.Intent = intent.Intent
			msg.Confidence = intent.Confidence
			conversation.UpdateIntent(intent.Intent)
		}
		
		sentiment, err := s.nlpService.AnalyzeSentiment(cmd.InitialMessage)
		if err == nil {
			msg.Sentiment = sentiment.Sentiment
			conversation.UpdateSentiment(sentiment.Sentiment)
		}
	}

	if err := s.conversationRepo.Save(ctx, conversation); err != nil {
		return nil, fmt.Errorf("failed to save conversation: %w", err)
	}

	s.logger.InfoContext(ctx, "conversation started", "conversation_id", conversation.ID)
	return conversation, nil
}

type SendMessageCommand struct {
	ConversationID string
	Content        string
	Attachments    []*domain.Attachment
}

func (s *IntelligentSupportService) SendMessage(ctx context.Context, cmd *SendMessageCommand) (*domain.Message, error) {
	s.logger.DebugContext(ctx, "sending message", "conversation_id", cmd.ConversationID)

	conversation, err := s.conversationRepo.FindByID(ctx, cmd.ConversationID)
	if err != nil {
		return nil, err
	}

	if !conversation.IsActive() {
		return nil, domain.ErrConversationEnded
	}

	userMsg := conversation.AddMessage(domain.MessageSenderUser, cmd.Content)
	userMsg.Attachments = cmd.Attachments

	intent, err := s.nlpService.DetectIntent(cmd.Content, conversation.Messages)
	if err == nil {
		userMsg.Intent = intent.Intent
		userMsg.Confidence = intent.Confidence
		conversation.UpdateIntent(intent.Intent)
	}

	sentiment, err := s.nlpService.AnalyzeSentiment(cmd.Content)
	if err == nil {
		userMsg.Sentiment = sentiment.Sentiment
		conversation.UpdateSentiment(sentiment.Sentiment)
	}

	response, err := s.chatbotEngine.GenerateResponse(ctx, conversation.ID, cmd.Content, conversation.Messages)
	if err != nil {
		s.logger.WarnContext(ctx, "failed to generate bot response", "error", err)
		response = &domain.ChatbotResponse{
			ResponseText: "I apologize, I'm having trouble processing your request. Let me connect you with an agent.",
			NeedsHumanTransfer: true,
		}
	}

	botMsg := conversation.AddMessage(domain.MessageSenderBot, response.ResponseText)
	botMsg.Intent = response.DetectedIntent
	botMsg.Confidence = response.Confidence

	if response.NeedsHumanTransfer {
		s.handleTransfer(ctx, conversation, response.TransferReason)
	}

	if err := s.conversationRepo.Update(ctx, conversation); err != nil {
		return nil, fmt.Errorf("failed to update conversation: %w", err)
	}

	return userMsg, nil
}

func (s *IntelligentSupportService) handleTransfer(ctx context.Context, conversation *domain.Conversation, reason string) {
	if s.agentService != nil {
		agentID, err := s.agentService.FindAvailableAgent("")
		if err == nil && agentID != "" {
			conversation.Transfer(agentID)
			s.logger.InfoContext(ctx, "conversation transferred to agent", 
				"conversation_id", conversation.ID, 
				"agent_id", agentID,
				"reason", reason)
		}
	}
}

func (s *IntelligentSupportService) GetConversation(ctx context.Context, conversationID string) (*domain.Conversation, error) {
	return s.conversationRepo.FindByID(ctx, conversationID)
}

func (s *IntelligentSupportService) ListConversations(ctx context.Context, userID uint64, status domain.ConversationStatus, page, pageSize int) ([]*domain.Conversation, int, error) {
	offset := (page - 1) * pageSize
	conversations, err := s.conversationRepo.FindByUserID(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	var filtered []*domain.Conversation
	for _, c := range conversations {
		if status == "" || c.Status == status {
			filtered = append(filtered, c)
		}
	}

	return filtered, len(filtered), nil
}

type EndConversationCommand struct {
	ConversationID string
	Summary        string
}

func (s *IntelligentSupportService) EndConversation(ctx context.Context, cmd *EndConversationCommand) error {
	s.logger.InfoContext(ctx, "ending conversation", "conversation_id", cmd.ConversationID)

	conversation, err := s.conversationRepo.FindByID(ctx, cmd.ConversationID)
	if err != nil {
		return err
	}

	conversation.End(cmd.Summary)
	return s.conversationRepo.Update(ctx, conversation)
}

type TransferToHumanCommand struct {
	ConversationID string
	Reason         string
	Department     string
}

type TransferToHumanResult struct {
	Success            bool
	AgentID            string
	EstimatedWaitSeconds int
	Message            string
}

func (s *IntelligentSupportService) TransferToHuman(ctx context.Context, cmd *TransferToHumanCommand) (*TransferToHumanResult, error) {
	s.logger.InfoContext(ctx, "transferring to human", "conversation_id", cmd.ConversationID)

	conversation, err := s.conversationRepo.FindByID(ctx, cmd.ConversationID)
	if err != nil {
		return nil, err
	}

	if s.agentService == nil {
		return &TransferToHumanResult{
			Success: false,
			Message: "Agent service not available",
		}, nil
	}

	agentID, err := s.agentService.FindAvailableAgent(cmd.Department)
	if err != nil {
		return &TransferToHumanResult{
			Success: false,
			Message: fmt.Sprintf("No available agent: %v", err),
		}, nil
	}

	waitTime, _ := s.agentService.GetEstimatedWaitTime(cmd.Department)

	conversation.Transfer(agentID)
	if err := s.conversationRepo.Update(ctx, conversation); err != nil {
		return nil, err
	}

	return &TransferToHumanResult{
		Success:              true,
		AgentID:              agentID,
		EstimatedWaitSeconds: waitTime,
		Message:              "Transfer successful",
	}, nil
}

type GetChatbotResponseCommand struct {
	ConversationID string
	UserMessage    string
	Context        []*domain.Message
}

func (s *IntelligentSupportService) GetChatbotResponse(ctx context.Context, cmd *GetChatbotResponseCommand) (*domain.ChatbotResponse, error) {
	return s.chatbotEngine.GenerateResponse(ctx, cmd.ConversationID, cmd.UserMessage, cmd.Context)
}

func (s *IntelligentSupportService) GetQuickReplies(ctx context.Context, intent domain.IntentCategory) []*domain.QuickReply {
	return s.chatbotEngine.GetQuickReplies(intent)
}

type SubmitFeedbackCommand struct {
	ConversationID string
	MessageID      string
	WasHelpful     bool
	Rating         int
	Comment        string
}

func (s *IntelligentSupportService) SubmitFeedback(ctx context.Context, cmd *SubmitFeedbackCommand) error {
	s.logger.InfoContext(ctx, "submitting feedback", "conversation_id", cmd.ConversationID)

	feedback := domain.NewFeedback(cmd.ConversationID, cmd.MessageID, cmd.WasHelpful, cmd.Rating, cmd.Comment)
	feedback.ID = fmt.Sprintf("FB%d", s.idGenerator.Generate())

	if err := s.feedbackRepo.Save(ctx, feedback); err != nil {
		return fmt.Errorf("failed to save feedback: %w", err)
	}

	return nil
}

func (s *IntelligentSupportService) AnalyzeSentiment(ctx context.Context, text string) (*domain.SentimentAnalysis, error) {
	return s.nlpService.AnalyzeSentiment(text)
}

func (s *IntelligentSupportService) GetIntent(ctx context.Context, text string, context []*domain.Message) (*domain.IntentResult, error) {
	return s.nlpService.DetectIntent(text, context)
}

type CreateKnowledgeBaseCommand struct {
	Name        string
	Description string
	Language    string
}

func (s *IntelligentSupportService) CreateKnowledgeBase(ctx context.Context, cmd *CreateKnowledgeBaseCommand) (*domain.KnowledgeBase, error) {
	s.logger.InfoContext(ctx, "creating knowledge base", "name", cmd.Name)

	kb := domain.NewKnowledgeBase(cmd.Name, cmd.Language)
	kb.ID = fmt.Sprintf("KB%d", s.idGenerator.Generate())
	kb.Description = cmd.Description

	if err := s.knowledgeBaseRepo.Save(ctx, kb); err != nil {
		return nil, fmt.Errorf("failed to save knowledge base: %w", err)
	}

	return kb, nil
}

func (s *IntelligentSupportService) GetKnowledgeBase(ctx context.Context, id string) (*domain.KnowledgeBase, error) {
	return s.knowledgeBaseRepo.FindByID(ctx, id)
}

func (s *IntelligentSupportService) ListKnowledgeBases(ctx context.Context, activeOnly bool, page, pageSize int) ([]*domain.KnowledgeBase, int64, error) {
	offset := (page - 1) * pageSize
	return s.knowledgeBaseRepo.FindAll(ctx, activeOnly, pageSize, offset)
}

type CreateKnowledgeArticleCommand struct {
	KnowledgeBaseID string
	Title           string
	Content         string
	Keywords        []string
	Tags            []string
	Category        string
}

func (s *IntelligentSupportService) CreateKnowledgeArticle(ctx context.Context, cmd *CreateKnowledgeArticleCommand) (*domain.KnowledgeArticle, error) {
	s.logger.InfoContext(ctx, "creating knowledge article", "title", cmd.Title)

	article := domain.NewKnowledgeArticle(cmd.KnowledgeBaseID, cmd.Title, cmd.Content)
	article.ID = fmt.Sprintf("ART%d", s.idGenerator.Generate())
	article.Keywords = cmd.Keywords
	article.Tags = cmd.Tags
	article.Category = cmd.Category

	if err := s.articleRepo.Save(ctx, article); err != nil {
		return nil, fmt.Errorf("failed to save article: %w", err)
	}

	return article, nil
}

func (s *IntelligentSupportService) GetKnowledgeArticle(ctx context.Context, id string) (*domain.KnowledgeArticle, error) {
	article, err := s.articleRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	article.IncrementViewCount()
	_ = s.articleRepo.Update(ctx, article)

	return article, nil
}

type SearchKnowledgeArticlesCommand struct {
	KnowledgeBaseID string
	Query           string
	Tags            []string
	Category        string
	Limit           int
}

type SearchKnowledgeArticlesResult struct {
	Articles []*domain.KnowledgeArticle
	Scores   []float64
}

func (s *IntelligentSupportService) SearchKnowledgeArticles(ctx context.Context, cmd *SearchKnowledgeArticlesCommand) (*SearchKnowledgeArticlesResult, error) {
	limit := cmd.Limit
	if limit <= 0 {
		limit = 10
	}

	articles, scores, err := s.articleRepo.Search(ctx, cmd.KnowledgeBaseID, cmd.Query, cmd.Tags, cmd.Category, limit)
	if err != nil {
		return nil, err
	}

	return &SearchKnowledgeArticlesResult{
		Articles: articles,
		Scores:   scores,
	}, nil
}

type UpdateKnowledgeArticleCommand struct {
	ID        string
	Title     string
	Content   string
	Keywords  []string
	Tags      []string
	Category  string
	IsActive  bool
}

func (s *IntelligentSupportService) UpdateKnowledgeArticle(ctx context.Context, cmd *UpdateKnowledgeArticleCommand) (*domain.KnowledgeArticle, error) {
	article, err := s.articleRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	if cmd.Title != "" {
		article.Title = cmd.Title
	}
	if cmd.Content != "" {
		article.Content = cmd.Content
	}
	if cmd.Keywords != nil {
		article.Keywords = cmd.Keywords
	}
	if cmd.Tags != nil {
		article.Tags = cmd.Tags
	}
	if cmd.Category != "" {
		article.Category = cmd.Category
	}
	article.IsActive = cmd.IsActive
	article.UpdatedAt = time.Now()

	if err := s.articleRepo.Update(ctx, article); err != nil {
		return nil, fmt.Errorf("failed to update article: %w", err)
	}

	return article, nil
}

func (s *IntelligentSupportService) DeleteKnowledgeArticle(ctx context.Context, id string) error {
	return s.articleRepo.Delete(ctx, id)
}
