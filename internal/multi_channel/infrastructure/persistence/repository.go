package persistence

import (
	"context"
	"github.com/wyfcoding/ecommerce/internal/multi_channel/domain/entity"
	"github.com/wyfcoding/ecommerce/internal/multi_channel/domain/repository"
	"errors"

	"gorm.io/gorm"
)

type multiChannelRepository struct {
	db *gorm.DB
}

func NewMultiChannelRepository(db *gorm.DB) repository.MultiChannelRepository {
	return &multiChannelRepository{db: db}
}

// 渠道
func (r *multiChannelRepository) SaveChannel(ctx context.Context, channel *entity.Channel) error {
	return r.db.WithContext(ctx).Save(channel).Error
}

func (r *multiChannelRepository) GetChannel(ctx context.Context, id uint64) (*entity.Channel, error) {
	var channel entity.Channel
	if err := r.db.WithContext(ctx).First(&channel, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &channel, nil
}

func (r *multiChannelRepository) ListChannels(ctx context.Context, activeOnly bool) ([]*entity.Channel, error) {
	var channels []*entity.Channel
	db := r.db.WithContext(ctx)
	if activeOnly {
		db = db.Where("is_enabled = ?", true)
	}
	if err := db.Find(&channels).Error; err != nil {
		return nil, err
	}
	return channels, nil
}

// 订单
func (r *multiChannelRepository) SaveOrder(ctx context.Context, order *entity.LocalOrder) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *multiChannelRepository) GetOrderByChannelID(ctx context.Context, channelID uint64, channelOrderID string) (*entity.LocalOrder, error) {
	var order entity.LocalOrder
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

func (r *multiChannelRepository) ListOrders(ctx context.Context, channelID uint64, status string, offset, limit int) ([]*entity.LocalOrder, int64, error) {
	var list []*entity.LocalOrder
	var total int64

	db := r.db.WithContext(ctx).Model(&entity.LocalOrder{})
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

// 日志
func (r *multiChannelRepository) SaveSyncLog(ctx context.Context, log *entity.ChannelSyncLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}
