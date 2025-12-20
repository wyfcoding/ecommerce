package persistence

import (
	"context"
	"errors"

	"github.com/wyfcoding/ecommerce/internal/multi_channel/domain"

	"gorm.io/gorm"
)

type multiChannelRepository struct {
	db *gorm.DB
}

// NewMultiChannelRepository 创建并返回一个新的 multiChannelRepository 实例。
func NewMultiChannelRepository(db *gorm.DB) domain.MultiChannelRepository {
	return &multiChannelRepository{db: db}
}

// --- 渠道 (Channel methods) ---

func (r *multiChannelRepository) SaveChannel(ctx context.Context, channel *domain.Channel) error {
	return r.db.WithContext(ctx).Save(channel).Error
}

func (r *multiChannelRepository) GetChannel(ctx context.Context, id uint64) (*domain.Channel, error) {
	var channel domain.Channel
	if err := r.db.WithContext(ctx).First(&channel, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &channel, nil
}

func (r *multiChannelRepository) ListChannels(ctx context.Context, activeOnly bool) ([]*domain.Channel, error) {
	var channels []*domain.Channel
	db := r.db.WithContext(ctx)
	if activeOnly {
		db = db.Where("is_enabled = ?", true)
	}
	if err := db.Find(&channels).Error; err != nil {
		return nil, err
	}
	return channels, nil
}

// --- 订单 (LocalOrder methods) ---

func (r *multiChannelRepository) SaveOrder(ctx context.Context, order *domain.LocalOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *multiChannelRepository) GetOrderByChannelID(ctx context.Context, channelID uint64, channelOrderID string) (*domain.LocalOrder, error) {
	var order domain.LocalOrder
	err := r.db.WithContext(ctx).
		Where("channel_id = ? AND channel_order_id = ?", channelID, channelOrderID).
		First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &order, nil
}

func (r *multiChannelRepository) ListOrders(ctx context.Context, channelID uint64, status string, offset, limit int) ([]*domain.LocalOrder, int64, error) {
	var list []*domain.LocalOrder
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.LocalOrder{})
	if channelID > 0 {
		db = db.Where("channel_id = ?", channelID)
	}
	if status != "" {
		db = db.Where("status = ?", status)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Offset(offset).Limit(limit).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// --- 日志 (ChannelSyncLog methods) ---

func (r *multiChannelRepository) SaveSyncLog(ctx context.Context, log *domain.ChannelSyncLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}
