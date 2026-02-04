package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/support/domain"
	algorithm "github.com/wyfcoding/pkg/algorithm/ml"
)

// SupportCommandService 处理客户服务的写操作。
type SupportCommandService struct {
	repo      domain.TicketRepository
	publisher domain.EventPublisher
	logger    *slog.Logger
}

// NewSupportCommandService 创建并返回一个新的 SupportCommandService 实例。
func NewSupportCommandService(repo domain.TicketRepository, publisher domain.EventPublisher, logger *slog.Logger) *SupportCommandService {
	return &SupportCommandService{
		repo:      repo,
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

	if err := m.repo.SaveTicket(ctx, ticket); err != nil {
		m.logger.ErrorContext(ctx, "failed to create ticket", "user_id", userID, "subject", subject, "error", err)
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

	if senderType != "user" && ticket.Status == domain.TicketStatusOpen {
		ticket.Status = domain.TicketStatusInProgress
		if err := m.repo.UpdateTicket(ctx, ticket); err != nil {
			return nil, err
		}
	}

	message := domain.NewMessage(ticketID, senderID, senderType, content, msgType, false)
	if err := m.repo.SaveMessage(ctx, message); err != nil {
		m.logger.ErrorContext(ctx, "failed to save message", "ticket_id", ticketID, "sender_id", senderID, "error", err)
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

	ticket.Close()
	return m.repo.UpdateTicket(ctx, ticket)
}

// ResolveTicket 解决一个工单。
func (m *SupportCommandService) ResolveTicket(ctx context.Context, id uint64) error {
	ticket, err := m.repo.GetTicket(ctx, id)
	if err != nil {
		return err
	}

	ticket.Resolve()
	return m.repo.UpdateTicket(ctx, ticket)
}

// StartConversation 开启一个新的私聊会话。
func (m *SupportCommandService) StartConversation(ctx context.Context, user1ID, user2ID uint64) (*domain.Conversation, error) {
	// 简单的唯一性检查逻辑略（如查询是否存在）
	conv := domain.NewConversation(user1ID, user2ID)
	if err := m.repo.SaveConversation(ctx, conv); err != nil {
		m.logger.ErrorContext(ctx, "failed to create conversation", "u1", user1ID, "u2", user2ID, "error", err)
		return nil, err
	}
	return conv, nil
}

// SendConversationMessage 发送私聊消息。
func (m *SupportCommandService) SendConversationMessage(ctx context.Context, convID, senderID, receiverID uint64, content string, msgType domain.MessageType) (*domain.ConversationMessage, error) {
	msg := domain.NewConversationMessage(convID, senderID, receiverID, content, msgType)
	if err := m.repo.SaveConversationMessage(ctx, msg); err != nil {
		return nil, err
	}
	return msg, nil
}
