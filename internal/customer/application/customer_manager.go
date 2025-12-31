package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/customer/domain"
	"github.com/wyfcoding/pkg/algorithm"
)

// CustomerManager 处理客户服务的写操作。
type CustomerManager struct {
	repo   domain.CustomerRepository
	logger *slog.Logger
}

// NewCustomerManager 创建并返回一个新的 CustomerManager 实例。
func NewCustomerManager(repo domain.CustomerRepository, logger *slog.Logger) *CustomerManager {
	return &CustomerManager{
		repo:   repo,
		logger: logger,
	}
}

// SegmentUsers 利用 K-Means 算法对用户进行分群。
// k: 期望的分群数量。
// 该功能通过分析用户的工单频率、处理时长等维度，识别不同价值或行为偏好的用户群体。
func (m *CustomerManager) SegmentUsers(ctx context.Context, k int) (map[uint64]int, error) {
	// 1. 获取所有活跃用户的统计数据
	// 这里简化模拟获取数据，实际应从 Repository 中进行聚合查询
	// 假设我们关心两个维度：1. 工单总数 (反映活跃/问题度) 2. 平均优先级 (反映紧迫度)
	userStats := []struct {
		UserID      uint64
		TicketCount float64
		AvgPriority float64
	}{
		{UserID: 1, TicketCount: 10, AvgPriority: 4},
		{UserID: 2, TicketCount: 2, AvgPriority: 1},
		{UserID: 3, TicketCount: 15, AvgPriority: 3},
		{UserID: 4, TicketCount: 1, AvgPriority: 2},
		{UserID: 5, TicketCount: 8, AvgPriority: 4},
	}

	if len(userStats) < k {
		return nil, fmt.Errorf("not enough data points for k=%d", k)
	}

	// 2. 构造 KMeans 输入点
	points := make([]*algorithm.KMeansPoint, len(userStats))
	for i, stat := range userStats {
		points[i] = &algorithm.KMeansPoint{
			ID:   stat.UserID,
			Data: []float64{stat.TicketCount, stat.AvgPriority},
		}
	}

	// 3. 执行 K-Means 聚类
	kmeans := algorithm.NewKMeans(k, 100, 0.01)
	kmeans.Fit(points)

	// 4. 收集结果
	results := make(map[uint64]int)
	for _, p := range points {
		results[p.ID] = p.Label
	}

	m.logger.InfoContext(ctx, "user segmentation completed", "k", k, "user_count", len(userStats))
	return results, nil
}

// CreateTicket 创建一个新的客户服务工单。
func (m *CustomerManager) CreateTicket(ctx context.Context, userID uint64, subject, description, category string, priority domain.TicketPriority) (*domain.Ticket, error) {
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
func (m *CustomerManager) ReplyTicket(ctx context.Context, ticketID, senderID uint64, senderType, content string, msgType domain.MessageType) (*domain.Message, error) {
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
func (m *CustomerManager) CloseTicket(ctx context.Context, id uint64) error {
	ticket, err := m.repo.GetTicket(ctx, id)
	if err != nil {
		return err
	}

	ticket.Close()
	return m.repo.UpdateTicket(ctx, ticket)
}

// ResolveTicket 解决一个工单。
func (m *CustomerManager) ResolveTicket(ctx context.Context, id uint64) error {
	ticket, err := m.repo.GetTicket(ctx, id)
	if err != nil {
		return err
	}

	ticket.Resolve()
	return m.repo.UpdateTicket(ctx, ticket)
}
