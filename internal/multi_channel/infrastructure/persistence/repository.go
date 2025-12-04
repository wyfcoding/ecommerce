package persistence

import (
	"context"
	"errors"
	"github.com/wyfcoding/ecommerce/internal/multi_channel/domain/entity"     // 导入多渠道模块的领域实体定义。
	"github.com/wyfcoding/ecommerce/internal/multi_channel/domain/repository" // 导入多渠道模块的领域仓储接口。

	"gorm.io/gorm" // 导入GORM ORM框架。
)

type multiChannelRepository struct {
	db *gorm.DB // GORM数据库连接实例。
}

// NewMultiChannelRepository 创建并返回一个新的 multiChannelRepository 实例。
// db: GORM数据库连接实例。
func NewMultiChannelRepository(db *gorm.DB) repository.MultiChannelRepository {
	return &multiChannelRepository{db: db}
}

// --- 渠道 (Channel methods) ---

// SaveChannel 将销售渠道实体保存到数据库。
// 如果渠道已存在（通过ID），则更新其信息；如果不存在，则创建。
func (r *multiChannelRepository) SaveChannel(ctx context.Context, channel *entity.Channel) error {
	return r.db.WithContext(ctx).Save(channel).Error
}

// GetChannel 根据ID从数据库获取销售渠道记录。
// 如果记录未找到，则返回nil。
func (r *multiChannelRepository) GetChannel(ctx context.Context, id uint64) (*entity.Channel, error) {
	var channel entity.Channel
	if err := r.db.WithContext(ctx).First(&channel, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &channel, nil
}

// ListChannels 从数据库列出所有销售渠道记录。
// activeOnly: 布尔值，如果为true，则只列出启用的渠道。
func (r *multiChannelRepository) ListChannels(ctx context.Context, activeOnly bool) ([]*entity.Channel, error) {
	var channels []*entity.Channel
	db := r.db.WithContext(ctx)
	if activeOnly { // 根据activeOnly参数过滤。
		db = db.Where("is_enabled = ?", true)
	}
	if err := db.Find(&channels).Error; err != nil {
		return nil, err
	}
	return channels, nil
}

// --- 订单 (LocalOrder methods) ---

// SaveOrder 将本地订单实体保存到数据库。
func (r *multiChannelRepository) SaveOrder(ctx context.Context, order *entity.LocalOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

// GetOrderByChannelID 根据渠道ID和渠道订单ID获取本地订单记录。
// 如果记录未找到，则返回nil。
func (r *multiChannelRepository) GetOrderByChannelID(ctx context.Context, channelID uint64, channelOrderID string) (*entity.LocalOrder, error) {
	var order entity.LocalOrder
	err := r.db.WithContext(ctx).
		Where("channel_id = ? AND channel_order_id = ?", channelID, channelOrderID).
		First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 如果记录未找到，返回nil。
		}
		return nil, err // 其他错误则返回。
	}
	return &order, nil
}

// ListOrders 从数据库列出所有本地订单记录，支持通过渠道ID和状态过滤，并支持分页。
func (r *multiChannelRepository) ListOrders(ctx context.Context, channelID uint64, status string, offset, limit int) ([]*entity.LocalOrder, int64, error) {
	var list []*entity.LocalOrder
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.LocalOrder{})
	if channelID > 0 { // 如果提供了渠道ID，则按渠道ID过滤。
		db = db.Where("channel_id = ?", channelID)
	}
	if status != "" { // 如果提供了状态，则按状态过滤。
		db = db.Where("status = ?", status)
	}

	// 统计总记录数。
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 应用分页和排序。
	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// --- 日志 (ChannelSyncLog methods) ---

// SaveSyncLog 将渠道同步日志实体保存到数据库。
func (r *multiChannelRepository) SaveSyncLog(ctx context.Context, log *entity.ChannelSyncLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}
