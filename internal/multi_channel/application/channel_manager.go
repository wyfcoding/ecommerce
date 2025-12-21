package application

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/multi_channel/domain"

	"log/slog"
)

// ChannelManager 处理渠道的写操作。
type ChannelManager struct {
	repo   domain.MultiChannelRepository
	logger *slog.Logger
}

// NewChannelManager creates a new ChannelManager instance.
func NewChannelManager(repo domain.MultiChannelRepository, logger *slog.Logger) *ChannelManager {
	return &ChannelManager{
		repo:   repo,
		logger: logger,
	}
}

// RegisterChannel 注册一个新的销售渠道。
func (m *ChannelManager) RegisterChannel(ctx context.Context, channel *domain.Channel) error {
	if err := m.repo.SaveChannel(ctx, channel); err != nil {
		m.logger.Error("failed to register channel", "error", err)
		return err
	}
	return nil
}

// SyncOrders 同步指定渠道的订单数据。
func (m *ChannelManager) SyncOrders(ctx context.Context, channelID uint64) error {
	channel, err := m.repo.GetChannel(ctx, channelID)
	if err != nil {
		return err
	}
	if channel == nil {
		return nil
	}

	startTime := time.Now()

	// 模拟1 new order
	mockOrder := &domain.LocalOrder{
		ChannelID:      uint64(channel.ID),
		ChannelName:    channel.Name,
		ChannelOrderID: "MOCK-" + time.Now().Format("20060102150405"),
		Items: []*domain.OrderItem{
			{ProductID: 1, ProductName: "Mock Product", Quantity: 1, Price: 1000, SKU: "MOCK-SKU"},
		},
		TotalAmount: 1000,
		BuyerInfo: domain.BuyerInfo{
			Name: "Mock Buyer",
		},
		Status: "pending",
	}

	exists, err := m.repo.GetOrderByChannelID(ctx, uint64(channel.ID), mockOrder.ChannelOrderID)
	if err != nil {
		return err
	}

	successCount := 0
	if exists == nil {
		if err := m.repo.SaveOrder(ctx, mockOrder); err == nil {
			successCount = 1
		}
	}

	log := &domain.ChannelSyncLog{
		ChannelID:    uint64(channel.ID),
		ChannelName:  channel.Name,
		Type:         "order",
		Status:       "success",
		ItemsCount:   1,
		SuccessCount: int32(successCount),
		StartTime:    startTime,
		EndTime:      time.Now(),
	}
	_ = m.repo.SaveSyncLog(ctx, log)

	return nil
}
