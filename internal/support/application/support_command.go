package application

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/support/domain"
	"github.com/wyfcoding/ecommerce/internal/support/infrastructure/ai"
	algorithm "github.com/wyfcoding/pkg/algos/ml"
	"github.com/wyfcoding/pkg/messagequeue"
)

// SupportCommandService 处理客户服务的写操作。
type SupportCommandService struct {
	repo      domain.TicketRepository
	aiAdapter ai.AIAdapter
	publisher messagequeue.EventPublisher
	logger    *slog.Logger
}

// NewSupportCommandService 创建并返回一个新的 SupportCommandService 实例。
func NewSupportCommandService(repo domain.TicketRepository, aiAdapter ai.AIAdapter, publisher messagequeue.EventPublisher, logger *slog.Logger) *SupportCommandService {
	return &SupportCommandService{
		repo:      repo,
		aiAdapter: aiAdapter,
		publisher: publisher,
		logger:    logger,
	}
}

// SegmentUsers 利用 K-Means 算法对用户进行分群。
func (m *SupportCommandService) SegmentUsers(ctx context.Context, k int) (map[uint64]int, error) {
	// 1. 从 Repository 获取真实的聚合统计数据
	userStats, err := m.repo.GetCustomerSegmentationStats(ctx)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to get customer stats for segmentation", "error", err)
		return nil, err
	}

	if len(userStats) < k {
		return nil, fmt.Errorf("not enough data points for k=%d (found %d)", k, len(userStats))
	}

	// 2. 构造 KMeans 输入数据
	data := make([][]float64, len(userStats))
	userIDs := make([]uint64, len(userStats))
	for i, stat := range userStats {
		data[i] = []float64{float64(stat.TicketCount), float64(stat.AvgPriority)}
		userIDs[i] = stat.UserID
	}

	// 3. 执行 K-Means 聚类
	kmeans := algorithm.NewKMeans(k, data)
	kmeans.Fit(100) // maxIter = 100

	// 4. 收集结果
	assignments := kmeans.GetAssignments()
	results := make(map[uint64]int)
	for i, clusterID := range assignments {
		results[userIDs[i]] = clusterID
	}

	m.logger.InfoContext(ctx, "user segmentation completed", "k", k, "user_count", len(userStats))
	return results, nil
}

// CreateTicket 创建一个新的客户服务工单。
func (m *SupportCommandService) CreateTicket(ctx context.Context, userID uint64, subject, description, category string, priority domain.TicketPriority) (*domain.Ticket, error) {
	ticketNo := fmt.Sprintf("TKT%d", time.Now().UnixNano())
	ticket := domain.NewTicket(ticketNo, userID, subject, description, category, priority)

	if err := m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveTicketInTx(ctx, tx, ticket); err != nil {
			m.logger.ErrorContext(ctx, "failed to create ticket", "user_id", userID, "subject", subject, "error", err)
			return err
		}
		if m.publisher != nil {
			event := &domain.TicketCreatedEvent{
				TicketID:  ticket.ID,
				TicketNo:  ticket.TicketNo,
				UserID:    ticket.UserID,
				Timestamp: time.Now(),
			}
			if err := m.publisher.PublishInTx(ctx, tx, domain.TicketCreatedEventType, fmt.Sprintf("%d", ticket.ID), event); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	m.logger.InfoContext(ctx, "ticket created successfully", "ticket_id", ticket.ID, "ticket_no", ticketNo)
	return ticket, nil
}

// ReplyTicket 回复一个工单。
func (m *SupportCommandService) ReplyTicket(ctx context.Context, ticketID, senderID uint64, senderType, content string, msgType domain.MessageType) (*domain.Message, error) {
	ticket, err := m.repo.GetTicket(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if ticket == nil {
		return nil, errors.New("ticket not found")
	}

	message := domain.NewMessage(ticketID, senderID, senderType, content, msgType, false)

	needTicketUpdate := senderType != "user" && ticket.Status == domain.TicketStatusOpen
	if needTicketUpdate {
		ticket.Status = domain.TicketStatusInProgress
	}

	if err := m.repo.WithTx(ctx, func(tx any) error {
		if needTicketUpdate {
			if err := m.repo.UpdateTicketInTx(ctx, tx, ticket); err != nil {
				return err
			}
		}
		if err := m.repo.SaveMessageInTx(ctx, tx, message); err != nil {
			m.logger.ErrorContext(ctx, "failed to save message", "ticket_id", ticketID, "sender_id", senderID, "error", err)
			return err
		}
		if m.publisher != nil {
			msgEvent := &domain.TicketMessageCreatedEvent{
				MessageID: message.ID,
				TicketID:  ticketID,
				SenderID:  senderID,
				Timestamp: time.Now(),
			}
			if err := m.publisher.PublishInTx(ctx, tx, domain.TicketMessageCreatedEventType, fmt.Sprintf("%d", message.ID), msgEvent); err != nil {
				return err
			}
			if needTicketUpdate {
				ticketEvent := &domain.TicketUpdatedEvent{
					TicketID:   ticket.ID,
					Status:     ticket.Status,
					AssigneeID: ticket.AssigneeID,
					Timestamp:  time.Now(),
				}
				if err := m.publisher.PublishInTx(ctx, tx, domain.TicketUpdatedEventType, fmt.Sprintf("%d", ticket.ID), ticketEvent); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	m.logger.InfoContext(ctx, "message saved successfully", "message_id", message.ID, "ticket_id", ticketID)

	return message, nil
}

// CloseTicket 关闭一个工单。
func (m *SupportCommandService) CloseTicket(ctx context.Context, id uint64) error {
	ticket, err := m.repo.GetTicket(ctx, id)
	if err != nil {
		return err
	}
	if ticket == nil {
		return errors.New("ticket not found")
	}

	ticket.Close()
	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.UpdateTicketInTx(ctx, tx, ticket); err != nil {
			return err
		}
		if m.publisher != nil {
			event := &domain.TicketUpdatedEvent{
				TicketID:   ticket.ID,
				Status:     ticket.Status,
				AssigneeID: ticket.AssigneeID,
				Timestamp:  time.Now(),
			}
			return m.publisher.PublishInTx(ctx, tx, domain.TicketUpdatedEventType, fmt.Sprintf("%d", ticket.ID), event)
		}
		return nil
	})
}

// ResolveTicket 解决一个工单。
func (m *SupportCommandService) ResolveTicket(ctx context.Context, id uint64) error {
	ticket, err := m.repo.GetTicket(ctx, id)
	if err != nil {
		return err
	}
	if ticket == nil {
		return errors.New("ticket not found")
	}

	ticket.Resolve()
	return m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.UpdateTicketInTx(ctx, tx, ticket); err != nil {
			return err
		}
		if m.publisher != nil {
			event := &domain.TicketUpdatedEvent{
				TicketID:   ticket.ID,
				Status:     ticket.Status,
				AssigneeID: ticket.AssigneeID,
				Timestamp:  time.Now(),
			}
			return m.publisher.PublishInTx(ctx, tx, domain.TicketUpdatedEventType, fmt.Sprintf("%d", ticket.ID), event)
		}
		return nil
	})
}

// StartConversation 开启一个新的私聊会话。
func (m *SupportCommandService) StartConversation(ctx context.Context, user1ID, user2ID uint64) (*domain.Conversation, error) {
	// 简单的唯一性检查逻辑略（如查询是否存在）
	conv := domain.NewConversation(user1ID, user2ID)
	if err := m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveConversationInTx(ctx, tx, conv); err != nil {
			m.logger.ErrorContext(ctx, "failed to create conversation", "u1", user1ID, "u2", user2ID, "error", err)
			return err
		}
		if m.publisher != nil {
			event := &domain.ConversationCreatedEvent{
				ConversationID: conv.ID,
				User1ID:        conv.User1ID,
				User2ID:        conv.User2ID,
				Timestamp:      time.Now(),
			}
			return m.publisher.PublishInTx(ctx, tx, domain.ConversationCreatedEventType, fmt.Sprintf("%d", conv.ID), event)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return conv, nil
}

// SendConversationMessage 发送私聊消息。
func (m *SupportCommandService) SendConversationMessage(ctx context.Context, convID, senderID, receiverID uint64, content string, msgType domain.MessageType) (*domain.ConversationMessage, error) {
	msg := domain.NewConversationMessage(convID, senderID, receiverID, content, msgType)
	if err := m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveConversationMessageInTx(ctx, tx, msg); err != nil {
			return err
		}
		if m.publisher != nil {
			event := &domain.ConversationMessageCreatedEvent{
				MessageID:      msg.ID,
				ConversationID: msg.ConversationID,
				SenderID:       msg.SenderID,
				ReceiverID:     msg.ReceiverID,
				Timestamp:      time.Now(),
			}
			return m.publisher.PublishInTx(ctx, tx, domain.ConversationMessageCreatedEventType, fmt.Sprintf("%d", msg.ID), event)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return msg, nil
}

// --- Intelligent Support Methods ---

// CreateKnowledgeBase 创建一个新的知识库。
func (m *SupportCommandService) CreateKnowledgeBase(ctx context.Context, name, description, language string) (*domain.KnowledgeBase, error) {
	kb := &domain.KnowledgeBase{
		ID:          fmt.Sprintf("KB%d", time.Now().UnixNano()),
		Name:        name,
		Description: description,
		Language:    language,
		IsActive:    true,
	}
	if err := m.repo.SaveKnowledgeBase(ctx, kb); err != nil {
		m.logger.ErrorContext(ctx, "failed to create knowledge base", "name", name, "error", err)
		return nil, err
	}
	return kb, nil
}

// StartAIConversation 开启一个新的 AI 客服会话。
func (m *SupportCommandService) StartAIConversation(ctx context.Context, userID uint64) (*domain.AIConversation, error) {
	conv := &domain.AIConversation{
		ID:        fmt.Sprintf("AI%d", time.Now().UnixNano()),
		UserID:    userID,
		Status:    domain.ConversationStatusActive,
		StartedAt: time.Now(),
	}
	if err := m.repo.SaveAIConversation(ctx, conv); err != nil {
		m.logger.ErrorContext(ctx, "failed to start AI conversation", "user_id", userID, "error", err)
		return nil, err
	}
	return conv, nil
}

// GetChatbotResponse 获取 AI 客服的回复。
func (m *SupportCommandService) GetChatbotResponse(ctx context.Context, conversationID, userMessage string) (*domain.IntentResult, error) {
	// 1. 保存用户消息
	userMsg := &domain.AIMessage{
		ID:             fmt.Sprintf("MSG%d", time.Now().UnixNano()),
		ConversationID: conversationID,
		Sender:         domain.MessageSenderUser,
		Content:        userMessage,
		CreatedAt:      time.Now(),
	}
	_ = m.repo.SaveAIMessage(ctx, userMsg)

	// 2. 调用 AI 适配器进行处理
	intent, err := m.aiAdapter.RecognizeIntent(ctx, userMessage)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to recognize intent", "msg", userMessage, "error", err)
		return nil, err
	}

	response, err := m.aiAdapter.GenerateChatbotResponse(ctx, conversationID, userMessage)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to generate chatbot response", "conv_id", conversationID, "error", err)
		return nil, err
	}
	intent.SuggestedResponse = response

	// 3. 保存 AI 回复
	botMsg := &domain.AIMessage{
		ID:             fmt.Sprintf("MSG%d", time.Now().UnixNano()+1),
		ConversationID: conversationID,
		Sender:         domain.MessageSenderBot,
		Content:        response,
		Intent:         intent.Intent,
		Confidence:     intent.Confidence,
		CreatedAt:      time.Now(),
	}
	_ = m.repo.SaveAIMessage(ctx, botMsg)

	return intent, nil
}
