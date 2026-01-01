package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/multichannel/domain"
)

// ChannelManager 处理渠道的写操作。
type ChannelManager struct {
	repo     domain.MultiChannelRepository
	logger   *slog.Logger
	adapters map[string]domain.ChannelAdapter
}

// NewChannelManager creates a new ChannelManager instance.
func NewChannelManager(repo domain.MultiChannelRepository, logger *slog.Logger) *ChannelManager {
	return &ChannelManager{
		repo:     repo,
		logger:   logger,
		adapters: make(map[string]domain.ChannelAdapter),
	}
}

// RegisterAdapter 注册渠道适配器
func (m *ChannelManager) RegisterAdapter(channelType string, adapter domain.ChannelAdapter) {
	m.adapters[channelType] = adapter
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
	if channel == nil || !channel.IsEnabled {
		return nil
	}

	adapter, ok := m.adapters[channel.Type]
	if !ok {
		return fmt.Errorf("no adapter found for channel type: %s", channel.Type)
	}

	startTime := time.Now().Add(-24 * time.Hour) // 默认同步过去 24 小时
	endTime := time.Now()
	syncStartTime := time.Now()

	// 1. 调用真实适配器拉取数据
	externalOrders, err := adapter.FetchOrders(ctx, channel, startTime, endTime)
	if err != nil {
		m.logger.ErrorContext(ctx, "failed to fetch external orders", "channel", channel.Name, "error", err)
		return err
	}

	var (
		successCount int32
		failureCount int32
	)

	// 2. 遍历并入库
	for _, order := range externalOrders {
		exists, err := m.repo.GetOrderByChannelID(ctx, uint64(channel.ID), order.ChannelOrderID)
		if err != nil {
			failureCount++
			continue
		}

		if exists == nil {
			if err := m.repo.SaveOrder(ctx, order); err != nil {
				m.logger.ErrorContext(ctx, "failed to save synced order", "channel_order_id", order.ChannelOrderID, "error", err)
				failureCount++
			} else {
				successCount++
			}
		}
	}

	// 3. 记录同步日志
	log := &domain.ChannelSyncLog{
		ChannelID:    uint64(channel.ID),
		ChannelName:  channel.Name,
		Type:         "order",
		Status:       "success",
		ItemsCount:   int32(len(externalOrders)),
		SuccessCount: successCount,
		FailureCount: failureCount,
		StartTime:    syncStartTime,
		EndTime:      time.Now(),
	}
	if failureCount > 0 && successCount == 0 {
		log.Status = "failed"
	}

	if err := m.repo.SaveSyncLog(ctx, log); err != nil {
		m.logger.ErrorContext(ctx, "failed to save channel sync log", "channel_id", channel.ID, "error", err)
	}

	return nil
}
