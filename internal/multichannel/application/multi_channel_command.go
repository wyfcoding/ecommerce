package application

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/wyfcoding/ecommerce/internal/multichannel/domain"
)

// MultiChannelCommandService 处理渠道的写操作。
type MultiChannelCommandService struct {
	repo      domain.MultiChannelRepository
	publisher domain.EventPublisher
	logger    *slog.Logger
	adapters  map[string]domain.ChannelAdapter
}

// NewMultiChannelCommandService creates a new MultiChannelCommandService instance.
func NewMultiChannelCommandService(repo domain.MultiChannelRepository, publisher domain.EventPublisher, logger *slog.Logger) *MultiChannelCommandService {
	return &MultiChannelCommandService{
		repo:      repo,
		publisher: publisher,
		logger:    logger,
		adapters:  make(map[string]domain.ChannelAdapter),
	}
}

// RegisterAdapter 注册渠道适配器
func (m *MultiChannelCommandService) RegisterAdapter(channelType string, adapter domain.ChannelAdapter) {
	m.adapters[channelType] = adapter
}

// RegisterChannel 注册一个新的销售渠道。
func (m *MultiChannelCommandService) RegisterChannel(ctx context.Context, channel *domain.Channel) error {
	if err := m.repo.WithTx(ctx, func(tx any) error {
		if err := m.repo.SaveChannelInTx(ctx, tx, channel); err != nil {
			m.logger.Error("failed to register channel", "error", err)
			return err
		}
		if m.publisher == nil {
			return nil
		}
		return m.publisher.PublishInTx(ctx, tx, domain.ChannelRegisteredEventType, fmt.Sprintf("%d", channel.ID), &domain.ChannelRegisteredEvent{
			ChannelID: uint64(channel.ID),
			Name:      channel.Name,
			Type:      channel.Type,
			Timestamp: time.Now(),
		})
	}); err != nil {
		return err
	}
	return nil
}

// SyncOrders 同步指定渠道的订单数据。
func (m *MultiChannelCommandService) SyncOrders(ctx context.Context, channelID uint64) error {
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
		order.ChannelID = uint64(channel.ID)
		order.ChannelName = channel.Name

		exists, err := m.repo.GetOrderByChannelID(ctx, uint64(channel.ID), order.ChannelOrderID)
		if err != nil {
			failureCount++
			continue
		}

		if exists == nil {
			if err := m.repo.WithTx(ctx, func(tx any) error {
				if err := m.repo.SaveOrderInTx(ctx, tx, order); err != nil {
					return err
				}
				if m.publisher == nil {
					return nil
				}
				return m.publisher.PublishInTx(ctx, tx, domain.ChannelOrderCreatedEventType, fmt.Sprintf("%d", order.ID), &domain.ChannelOrderCreatedEvent{
					OrderID:     uint64(order.ID),
					ChannelID:   order.ChannelID,
					ExternalID:  order.ChannelOrderID,
					TotalAmount: order.TotalAmount,
					Timestamp:   time.Now(),
				})
			}); err != nil {
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
