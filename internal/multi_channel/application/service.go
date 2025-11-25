package application

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/multi_channel/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/multi_channel/domain/repository"
	"time"

	"log/slog"
)

type MultiChannelService struct {
	repo   repository.MultiChannelRepository
	logger *slog.Logger
}

func NewMultiChannelService(repo repository.MultiChannelRepository, logger *slog.Logger) *MultiChannelService {
	return &MultiChannelService{
		repo:   repo,
		logger: logger,
	}
}

// RegisterChannel 注册渠道
func (s *MultiChannelService) RegisterChannel(ctx context.Context, channel *entity.Channel) error {
	if err := s.repo.SaveChannel(ctx, channel); err != nil {
		s.logger.Error("failed to register channel", "error", err)
		return err
	}
	return nil
}

// SyncOrders 同步订单 (Mock logic)
func (s *MultiChannelService) SyncOrders(ctx context.Context, channelID uint64) error {
	channel, err := s.repo.GetChannel(ctx, channelID)
	if err != nil {
		return err
	}
	if channel == nil {
		return nil
	}

	startTime := time.Now()
	// Mock fetching orders from external channel API
	// In reality, we would use channel.APIKey/Secret to call external API

	// Simulate 1 new order
	mockOrder := &entity.LocalOrder{
		ChannelID:      uint64(channel.ID),
		ChannelName:    channel.Name,
		ChannelOrderID: "MOCK-" + time.Now().Format("20060102150405"),
		Items: []*entity.OrderItem{
			{ProductID: 1, ProductName: "Mock Product", Quantity: 1, Price: 1000, SKU: "MOCK-SKU"},
		},
		TotalAmount: 1000,
		BuyerInfo: entity.BuyerInfo{
			Name: "Mock Buyer",
		},
		Status: "pending",
	}

	// Check if exists
	exists, err := s.repo.GetOrderByChannelID(ctx, uint64(channel.ID), mockOrder.ChannelOrderID)
	if err != nil {
		return err
	}

	successCount := 0
	if exists == nil {
		if err := s.repo.SaveOrder(ctx, mockOrder); err == nil {
			successCount = 1
		}
	}

	// Log sync result
	log := &entity.ChannelSyncLog{
		ChannelID:    uint64(channel.ID),
		ChannelName:  channel.Name,
		Type:         "order",
		Status:       "success",
		ItemsCount:   1,
		SuccessCount: int32(successCount),
		StartTime:    startTime,
		EndTime:      time.Now(),
	}
	_ = s.repo.SaveSyncLog(ctx, log)

	return nil
}

// ListChannels 获取渠道列表
func (s *MultiChannelService) ListChannels(ctx context.Context) ([]*entity.Channel, error) {
	return s.repo.ListChannels(ctx, false)
}

// ListOrders 获取订单列表
func (s *MultiChannelService) ListOrders(ctx context.Context, channelID uint64, status string, page, pageSize int) ([]*entity.LocalOrder, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListOrders(ctx, channelID, status, offset, pageSize)
}
