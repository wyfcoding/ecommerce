package application

import (
	"context"
	"time"

	"github.com/wyfcoding/ecommerce/internal/multi_channel/domain/entity"     // 导入多渠道领域的实体定义。
	"github.com/wyfcoding/ecommerce/internal/multi_channel/domain/repository" // 导入多渠道领域的仓储接口。

	"log/slog" // 导入结构化日志库。
)

// MultiChannelService 结构体定义了多渠道运营相关的应用服务。
// 它协调领域层和基础设施层，处理销售渠道的注册、外部订单的同步和管理等业务逻辑。
type MultiChannelService struct {
	repo   repository.MultiChannelRepository // 依赖MultiChannelRepository接口，用于数据持久化操作。
	logger *slog.Logger                      // 日志记录器，用于记录服务运行时的信息和错误。
}

// NewMultiChannelService 创建并返回一个新的 MultiChannelService 实例。
func NewMultiChannelService(repo repository.MultiChannelRepository, logger *slog.Logger) *MultiChannelService {
	return &MultiChannelService{
		repo:   repo,
		logger: logger,
	}
}

// RegisterChannel 注册一个新的销售渠道。
// ctx: 上下文。
// channel: 待注册的Channel实体。
// 返回可能发生的错误。
func (s *MultiChannelService) RegisterChannel(ctx context.Context, channel *entity.Channel) error {
	if err := s.repo.SaveChannel(ctx, channel); err != nil {
		s.logger.Error("failed to register channel", "error", err)
		return err
	}
	return nil
}

// SyncOrders 同步指定渠道的订单数据。
// ctx: 上下文。
// channelID: 待同步订单的渠道ID。
// 返回可能发生的错误。
func (s *MultiChannelService) SyncOrders(ctx context.Context, channelID uint64) error {
	// 获取渠道详情。
	channel, err := s.repo.GetChannel(ctx, channelID)
	if err != nil {
		return err
	}
	if channel == nil {
		return nil // 如果渠道不存在，则直接返回。
	}

	startTime := time.Now() // 记录同步开始时间。
	// TODO: 从外部渠道API（例如，淘宝、京东、亚马逊的API）获取订单数据。
	// 实际生产中，需要使用 channel.APIKey/Secret 等凭证来调用外部API。

	// Simulate 1 new order: 模拟从外部渠道获取到1个新订单。
	mockOrder := &entity.LocalOrder{
		ChannelID:      uint64(channel.ID),
		ChannelName:    channel.Name,
		ChannelOrderID: "MOCK-" + time.Now().Format("20060102150405"), // 模拟外部渠道订单号。
		Items: []*entity.OrderItem{ // 模拟订单商品。
			{ProductID: 1, ProductName: "Mock Product", Quantity: 1, Price: 1000, SKU: "MOCK-SKU"},
		},
		TotalAmount: 1000, // 模拟总金额。
		BuyerInfo: entity.BuyerInfo{ // 模拟买家信息。
			Name: "Mock Buyer",
		},
		Status: "pending", // 模拟订单状态。
	}

	// 检查模拟订单是否已存在于本地系统。
	exists, err := s.repo.GetOrderByChannelID(ctx, uint64(channel.ID), mockOrder.ChannelOrderID)
	if err != nil {
		return err
	}

	successCount := 0
	// 如果订单不存在，则保存到本地系统。
	if exists == nil {
		if err := s.repo.SaveOrder(ctx, mockOrder); err == nil {
			successCount = 1 // 记录成功同步的订单数量。
		}
	}

	// 记录同步结果日志。
	log := &entity.ChannelSyncLog{
		ChannelID:    uint64(channel.ID),
		ChannelName:  channel.Name,
		Type:         "order",             // 同步类型为订单。
		Status:       "success",           // 同步状态。
		ItemsCount:   1,                   // 外部渠道返回的总项数。
		SuccessCount: int32(successCount), // 成功处理的项数。
		StartTime:    startTime,
		EndTime:      time.Now(),
	}
	_ = s.repo.SaveSyncLog(ctx, log) // 保存同步日志。

	return nil
}

// ListChannels 获取销售渠道列表。
// ctx: 上下文。
// 返回销售渠道列表和可能发生的错误。
func (s *MultiChannelService) ListChannels(ctx context.Context) ([]*entity.Channel, error) {
	return s.repo.ListChannels(ctx, false) // false表示列出所有渠道（包括非活跃的）。
}

// ListOrders 获取本地化存储的外部渠道订单列表。
// ctx: 上下文。
// channelID: 筛选订单的渠道ID。
// status: 筛选订单的状态。
// page, pageSize: 分页参数。
// 返回本地订单列表、总数和可能发生的错误。
func (s *MultiChannelService) ListOrders(ctx context.Context, channelID uint64, status string, page, pageSize int) ([]*entity.LocalOrder, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.ListOrders(ctx, channelID, status, offset, pageSize)
}
